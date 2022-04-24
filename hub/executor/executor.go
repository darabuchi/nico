package executor

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/Dreamacro/clash/adapter/outbound"
	P "github.com/Dreamacro/clash/component/process"
	"github.com/Dreamacro/clash/constant"
	"github.com/Dreamacro/clash/listener/mixed"
	"github.com/darabuchi/log"
	"github.com/darabuchi/nico/adapter"
	"github.com/darabuchi/nico/hub/rule"
	"github.com/darabuchi/utils"
)

const (
	Alive    = "alive"
	Delay    = "delay"
	Speed    = "speed"
	SpeedStr = "speed_str"
)

type ExecutorCallback struct {
	OnNodeAdd    func(node adapter.AdapterProxy)
	OnNodeDel    func(node adapter.AdapterProxy)
	OnDelayCheck func(node adapter.AdapterProxy, delay time.Duration)
}

type Executor struct {
	callback             *ExecutorCallback
	lock                 sync.RWMutex
	allProxy, aliveProxy adapter.ProxyList

	event chan executorEvent

	connChan chan constant.ConnContext
	service  *mixed.Listener

	rule *rule.AdapterRule
}

type eventType int

const (
	eventCheckDelay eventType = iota
	eventCleanDead
)

type executorEvent struct {
	eventType eventType

	node adapter.AdapterProxy
}

func NewExecutor() *Executor {
	p := &Executor{
		callback: &ExecutorCallback{},
		connChan: make(chan constant.ConnContext),
		event:    make(chan executorEvent, 5),
		rule:     rule.GetAdapterRule(),
	}

	p.handleConn()
	p.handleNode()

	return p
}

// 事件处理

func (p *Executor) OnNodeAdd(logic func(node adapter.AdapterProxy)) {
	p.callback.OnNodeAdd = logic
}

func (p *Executor) onNodeAdd(node adapter.AdapterProxy) {
	defer utils.CachePanic()
	if p.callback.OnNodeAdd != nil {
		p.callback.OnNodeAdd(node)
	}
}

func (p *Executor) OnNodeDel(logic func(node adapter.AdapterProxy)) {
	p.callback.OnNodeDel = logic
}

func (p *Executor) onNodeDel(node adapter.AdapterProxy) {
	defer utils.CachePanic()
	if p.callback.OnNodeDel != nil {
		p.callback.OnNodeDel(node)
	}
}

func (p *Executor) OnDelayCheck(logic func(node adapter.AdapterProxy, delay time.Duration)) {
	p.callback.OnDelayCheck = logic
}

func (p *Executor) onDelayCheck(node adapter.AdapterProxy, delay time.Duration) {
	defer utils.CachePanic()
	if p.callback.OnDelayCheck != nil {
		p.callback.OnDelayCheck(node, delay)
	}
}

func (p *Executor) SetAdapterRule(ar *rule.AdapterRule) {
	p.rule = ar
}

// 节点处理
func (p *Executor) handleNode() {
	go func(sign chan os.Signal) {
		delayCheck := time.NewTicker(time.Minute * 5)
		defer delayCheck.Stop()

		speedCheck := time.NewTicker(time.Hour)
		defer speedCheck.Stop()

		for {
			select {
			case <-delayCheck.C:
				proxies := p.cloneProxyList()
				log.Infof("check delay for %d proxies", len(proxies))

				proxies.Each(p.checkDelay)

				p.proxySort()
				delayCheck.Reset(time.Minute * 5)
			case <-speedCheck.C:
				proxies := p.cloneProxyList()
				log.Infof("check spped for %d proxies", len(proxies))

				proxies.Each(p.checkSpeed)

				p.proxySort()
				speedCheck.Reset(time.Minute * 5)
			case e := <-p.event:
				switch e.eventType {
				case eventCheckDelay:
					n := e.node
					log.Infof("load new node %s[%s]", n.Name(), n.UniqueId())

					p.checkDelay(n)

					p.proxySort()
				case eventCleanDead:
					p.cleanDeadNode()
				}

			case <-sign:
				return
			}
		}
	}(utils.GetExitSign())
}

func (p *Executor) checkDelay(proxy adapter.AdapterProxy) {
	log.Infof("check delay for %s", proxy.Name())
	delay, err := proxy.URLTest(context.TODO(), "https://www.google.com")
	if err != nil {
		proxy.Store(Alive, false)
		log.Debugf("err:%v", err)
		p.onDelayCheck(proxy, -1)
	} else {
		proxy.Store(Alive, true)
		proxy.Store(Delay, delay)
		log.Infof("%s delay:%dms", proxy.Name(), delay)
		p.onDelayCheck(proxy, time.Duration(delay)*time.Millisecond)
	}
}

func (p *Executor) checkSpeed(proxy adapter.AdapterProxy) {
	log.Infof("check speed for %s", proxy.Name())

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
		SortUsing(func(a, b adapter.AdapterProxy) bool {
			return a.LoadUint16(Delay) > b.LoadUint16(Delay)
		}).
		SortStableUsing(func(a, b adapter.AdapterProxy) bool {
			return a.LoadFloat64(Speed) > b.LoadFloat64(Speed)
		})

	p.aliveProxy = p.allProxy.Filter(func(proxy adapter.AdapterProxy) bool {
		return proxy.LoadBool(Alive)
	})
}

func (p *Executor) cloneProxyList() adapter.ProxyList {
	var proxyList adapter.ProxyList

	p.lock.RLock()
	defer p.lock.RUnlock()

	p.allProxy.Each(func(proxy adapter.AdapterProxy) {
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

func (p *Executor) AddNode(n adapter.AdapterProxy) {
	p.addNode(n)
}

func (p *Executor) addNode(n adapter.AdapterProxy) {
	n.Store(Delay, 0)
	n.Store(Speed, 0)
	n.Store(SpeedStr, "wait")

	p.lock.Lock()

	existed := p.aliveProxy.Any(func(value adapter.AdapterProxy) bool {
		return value.UniqueId() == n.UniqueId()
	})

	if !existed {
		p.allProxy = append(p.allProxy, n)
	}

	p.lock.Unlock()

	p.onNodeAdd(n)

	if !existed {
		p.event <- executorEvent{
			eventType: eventCheckDelay,
			node:      n,
		}
	}
}

func (p *Executor) cleanDeadNode() {
	var closeList adapter.ProxyList
	p.lock.Lock()
	p.allProxy = p.allProxy.Filter(func(proxy adapter.AdapterProxy) bool {
		alive := proxy.LoadBool(Alive)
		if !alive {
			closeList = append(closeList, proxy)
		}
		return alive
	})
	p.lock.Unlock()

	closeList.Each(func(proxy adapter.AdapterProxy) {
		p.onNodeDel(proxy)
	})
}

func (p *Executor) CleanDeadNodes() {
	p.event <- executorEvent{
		eventType: eventCleanDead,
	}
}

func (p *Executor) ChooseProxy() adapter.AdapterProxy {
	p.lock.RLock()
	defer p.lock.RUnlock()

	if len(p.aliveProxy) > 0 {
		for _, proxy := range p.aliveProxy {
			if proxy.LoadBool(Alive) {
				return proxy
			}
		}
	}

	return nil
}

// 监听端口
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
		reject := outbound.NewReject()

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

					switch p.rule.Match(metadata) {
					case adapter.Proxy:
						cc = p.ChooseProxy()
						if cc == nil {
							log.Warn("not found usable proxy")
							return
						}

					case adapter.Reject:
						cc = reject
					default:
						cc = direct
					}

					log.Infof("try to connect %v ues proxy %v-%v", metadata.RemoteAddress(),
						adapter.CoverAdapterType(cc.Type()), cc.Name())

					remote, err := cc.DialContext(context.TODO(), metadata)
					if err != nil {
						log.Errorf("err:%v", err)

						if adapter.CoverAdapterType(cc.Type()) == adapter.Reject {
							return
						}

						cc = p.ChooseProxy()
						if cc == nil {
							log.Warn("not found usable proxy")
							return
						}

						log.Infof("try to connect %v ues proxy %v-%v", metadata.RemoteAddress(), adapter.CoverAdapterType(cc.Type()), cc.Name())

						remote, err = cc.DialContext(context.TODO(), metadata)
						if err != nil {
							log.Errorf("err:%v", err)
							return
						}

						if metadata.DstIP != nil {
							r, err := rule.NewSrcIp(metadata.DstIP.String(), adapter.CoverAdapterType(cc.Type()))
							if err != nil {
								log.Errorf("err:%v", err)
							} else {
								p.rule.AddRule(r)
							}
						}

						if metadata.Host != metadata.DstIP.String() {
							r, err := rule.NewDomain(metadata.Host, adapter.CoverAdapterType(cc.Type()))
							if err != nil {
								log.Errorf("err:%v", err)
							} else {
								p.rule.AddRule(r)
							}
						}
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

func (p *Executor) match(metadata *constant.Metadata) {
	srcPort, err := strconv.Atoi(metadata.SrcPort)
	if err == nil {
		path, err := P.FindProcessName(metadata.NetWork.String(), metadata.SrcIP, srcPort)
		if err != nil {
			log.Debugf("[Process] find process %s: %v", metadata.String(), err)
		} else {
			log.Debugf("[Process] %s from process %s", metadata.String(), path)
			metadata.ProcessPath = path
		}
	}

}
