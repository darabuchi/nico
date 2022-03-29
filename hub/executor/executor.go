package executor

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/Dreamacro/clash/adapter/outbound"
	"github.com/Dreamacro/clash/constant"
	"github.com/Dreamacro/clash/listener/mixed"
	"github.com/darabuchi/log"
	"github.com/darabuchi/nico/adapter"
	"github.com/darabuchi/utils"
)

const (
	Alive    = "alive"
	Delay    = "delay"
	Speed    = "speed"
	SpeedStr = "speed_str"
)

type Executor struct {
	lock                 sync.RWMutex
	allProxy, aliveProxy adapter.ProxyList

	checkProxyC chan adapter.Proxy

	connChan chan constant.ConnContext
	service  *mixed.Listener
}

func NewExecutor() *Executor {
	p := &Executor{
		connChan:    make(chan constant.ConnContext),
		checkProxyC: make(chan adapter.Proxy, 10),
	}

	p.handleConn()
	p.handleNode()

	return p
}

func (p *Executor) handleConn() {
	relay := func(l, r net.Conn) {
		go func() {
			_, _ = io.Copy(l, r)
		}()
		_, _ = io.Copy(r, l)
	}

	go func(sign chan os.Signal) {
		defer func() {
			if p.service != nil {
				p.service.Close()
			}
			log.Warn("stop service")
		}()

		direct := outbound.NewDirect()
		// reject := outbound.NewReject()

		for {
			select {
			case c := <-p.connChan:
				go func(conn constant.ConnContext) {
					log.SetTrace(conn.ID().String())
					defer log.DelTrace()

					defer utils.CachePanic()

					metadata := conn.Metadata()

					// key := "adapter.dmain." + metadata.String()

					var cc constant.ProxyAdapter
					cc = direct

					log.Infof("try to connect %v ues proxy %v-%v", metadata.RemoteAddress(), cc.Type(), cc.Name())

					remote, err := cc.DialContext(context.TODO(), metadata)
					if err != nil {
						log.Errorf("err:%v", err)
						return
					}

					log.Infof("%s use %v-%s", metadata.RemoteAddress(), cc.Type(), cc.Name())

					// packet, err := cc.ListenPacketContext(ctx, metadata)
					// if err != nil {
					//	log.Errorf("err:%v", err)
					//	return
					// }

					relay(remote, conn.Conn())
				}(c)

			case <-sign:
				return
			}
		}
	}(utils.GetExitSign())
}

func (p *Executor) handleNode() {
	go func(sign chan os.Signal) {
		delayCheck := time.NewTicker(time.Minute)
		defer delayCheck.Stop()

		speedCheck := time.NewTicker(time.Minute * 30)
		defer speedCheck.Stop()

		for {
			select {
			case <-delayCheck.C:
				proxies := p.cloneProxyList()
				log.Infof("check delay for %d proxies", len(proxies))

				proxies.Each(p.checkDelay)

				p.proxySort()

			case <-speedCheck.C:
				proxies := p.cloneProxyList()
				log.Infof("check spped for %d proxies", len(proxies))

				proxies.Each(p.checkSpeed)

				p.proxySort()

			case n := <-p.checkProxyC:
				log.Infof("load new node %s[%s]", n.Name(), n.UniqueId())

				p.checkDelay(n)
				p.checkSpeed(n)
				p.proxySort()

			case <-sign:
				return
			}
		}
	}(utils.GetExitSign())
}

func (p *Executor) checkDelay(proxy adapter.Proxy) {
	delay, err := proxy.URLTest(context.TODO(), "https://www.google.com")
	if err != nil {
		log.Debugf("err:%v", err)
		proxy.Store(Alive, false)
		proxy.Store(Delay, -1)
	} else {
		log.Infof("%s delay:%dms", proxy.Name(), delay)
		proxy.Store(Alive, true)
		proxy.Store(Delay, delay)
	}
}

func (p *Executor) checkSpeed(proxy adapter.Proxy) {
	const bodySize = 1024 * 1024
	body := make([]byte, bodySize) // 1024*128

	err := proxy.DoRequest(http.MethodGet, "http://cachefly.cachefly.net/50mb.test", nil, time.Minute, map[string]string{}, func(resp *http.Response, start time.Time) error {
		_, err := resp.Body.Read(body)
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}

		spDelay := time.Since(start)

		speed := float64(bodySize)

		switch {
		case spDelay < time.Microsecond:
			speed = speed * 1000000000 / float64(spDelay.Nanoseconds())
		case spDelay < time.Millisecond:
			speed = speed * 1000000 / float64(spDelay.Microseconds())
		case spDelay < time.Second:
			speed = speed * 1000 / float64(spDelay.Milliseconds())
		default:
			speed = speed / spDelay.Seconds()
		}

		if speed > 0 {
			proxy.Store(Speed, speed/1024)
		}

		var speedStr string
		if speed == 0 {
			speedStr = "0bps"
		} else if speed < 1024 {
			speedStr = fmt.Sprintf("%.2fbps", speed*8)
		} else if speed < (1024 * 128) {
			speedStr = fmt.Sprintf("%.2fKbps", speed/128)
		} else if speed < (1024 * 1024 * 128) {
			speedStr = fmt.Sprintf("%.2fMbps", speed/(128*1024))
		} else if speed < (1024 * 1024 * 1024 * 128) {
			speedStr = fmt.Sprintf("%.2fGbps", speed/(128*1024*1024))
		} else {
			speedStr = fmt.Sprintf("%.2fTbps", speed/(128*1024*1024*1024))
		}
		proxy.Store(SpeedStr, speedStr)

		return nil
	})
	if err != nil {
		log.Errorf("err:%v", err)
		proxy.Store(Speed, -1)
		proxy.Store(SpeedStr, "0bps")
	}
}

func (p *Executor) proxySort() {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.allProxy = p.allProxy.
		SortUsing(func(a, b adapter.Proxy) bool {
			return a.LoadUint16(Delay) > b.LoadUint16(Delay)
		}).SortStableUsing(func(a, b adapter.Proxy) bool {
		return a.LoadFloat64(Speed) > b.LoadFloat64(Speed)
	})

	p.aliveProxy = p.aliveProxy.Filter(func(proxy adapter.Proxy) bool {
		return proxy.LoadBool(Alive)
	})
}

func (p *Executor) cloneProxyList() adapter.ProxyList {
	var proxyList adapter.ProxyList

	p.lock.RLock()
	defer p.lock.RUnlock()

	p.allProxy.Each(func(proxy adapter.Proxy) {
		proxyList = append(proxyList, proxy)
	})

	return proxyList
}

func (p *Executor) AddNodeByV2rayLink(s string) error {
	n, err := adapter.ParseV2ray(s)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	p.addNode(n)

	return nil
}

func (p *Executor) AddNodeByClash(m map[string]any) error {
	n, err := adapter.ParseClash(m)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	p.addNode(n)

	return nil
}

func (p *Executor) AddNode(n adapter.Proxy) {
	p.addNode(n)
}

func (p *Executor) addNode(n adapter.Proxy) {
	n.Store(Delay, 0)
	n.Store(Speed, 0)
	n.Store(SpeedStr, "wait")

	p.lock.Lock()

	existed := p.aliveProxy.Any(func(value adapter.Proxy) bool {
		return value.UniqueId() == n.UniqueId()
	})

	if !existed {
		p.allProxy = append(p.allProxy, n)
	}

	p.lock.Unlock()

	if !existed {
		p.checkProxyC <- n
	}
}

func (p *Executor) Listen(port string) error {
	var err error

	addr := ":" + port

	p.lock.Lock()
	defer p.lock.Unlock()

	if p.service != nil {
		// if p.service.Address() == addr {
		// 	log.Warnf("same addr, skip restart")
		// 	return nil
		// }

		err = p.service.Close()
		if err != nil {
			log.Debugf("err:%v", err)
			return err
		}
	}

	p.service, err = mixed.New(addr, p.connChan)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	return nil
}
