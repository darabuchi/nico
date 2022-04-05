package adapter

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/Dreamacro/clash/adapter/outbound"
	"github.com/Dreamacro/clash/constant"
	"github.com/Luoxin/faker"
	"github.com/darabuchi/log"
	"github.com/darabuchi/utils"
	"github.com/elliotchance/pie/pie"
	"gopkg.in/yaml.v3"
)

type AdapterType int

const (
	Direct AdapterType = iota
	Reject
	Proxy
)

func (at AdapterType) String() string {
	switch at {
	case Direct:
		return "Direct"
	case Reject:
		return "Reject"
	case Proxy:
		return "Proxy"
	default:
		return "Unknown"
	}
}

func ParseAdapterType(at string) AdapterType {
	switch at {
	case "Direct":
		return Direct
	case "Reject":
		return Reject
	case "Proxy":
		return Proxy
	default:
		return -1
	}
}

func CoverAdapterType(at constant.AdapterType) AdapterType {
	switch at {
	case constant.Direct:
		return Direct
	case constant.Reject:
		return Reject
	case constant.Shadowsocks:
		fallthrough
	case constant.ShadowsocksR:
		fallthrough
	case constant.Snell:
		fallthrough
	case constant.Socks5:
		fallthrough
	case constant.Http:
		fallthrough
	case constant.Vmess:
		fallthrough
	case constant.Trojan:
		return Proxy
	default:
		return -1
	}
}

type AdapterProxy interface {
	constant.Proxy
	Cache

	Sub4Nico() string
	Sub4Clash() string
	Sub4V2ray() string
	UniqueId() string

	DoRequest(method, rawUrl string, body io.Reader, timeout time.Duration, headers map[string]string, logic func(resp *http.Response, start time.Time) error) error

	Get(url string, timeout time.Duration, headers map[string]string) ([]byte, error)

	Post(url string, body []byte, timeout time.Duration, headers map[string]string) ([]byte, error)
	PostJson(url string, reqBody, rspBody any, timeout time.Duration, headers map[string]string) error
}

//go:generate pie ProxyList.*
type ProxyList []AdapterProxy

type ProxyAdapter struct {
	constant.Proxy
	Cache
	ExtraInfo

	opt map[string]any

	uniqueId string

	name string
}

func NewProxyAdapter(adapter constant.Proxy, opt any) (AdapterProxy, error) {
	p := &ProxyAdapter{
		Proxy: adapter,
		name:  adapter.Name(),
	}

	p.Cache = NewAdapterCache()

	switch v := opt.(type) {
	case map[string]any:
		p.opt = v
		p.ExtraInfo = ParseClash4Extra(v)
	case outbound.ShadowSocksOption,
		outbound.VlessOption,
		outbound.VmessOption,
		outbound.ShadowSocksROption,
		outbound.TrojanOption:
		buf, err := json.Marshal(v)
		if err != nil {
			log.Errorf("err:%v", err)
			return nil, err
		}

		err = json.Unmarshal(buf, &p.opt)
		if err != nil {
			log.Errorf("err:%v", err)
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unknown %s", reflect.TypeOf(opt))
	}

	delete(p.opt, "Name")

	var updateUnionId func(opt any) string
	updateUnionId = func(value any) string {
		if value == nil {
			return ""
		}

		var unionId string

		switch opt := value.(type) {
		case map[string]any:
			var keys pie.Strings
			for key := range opt {
				keys = append(keys, key)
			}

			keys.Each(func(key string) {
				val := opt[key]
				if value == nil {
					return
				}

				unionId = fmt.Sprintf("%s:%v", key, updateUnionId(val))
			})

			unionId = fmt.Sprintf("{%s}", unionId)
		case string, []byte,
			uint32, uint64, uint, uint8,
			float64, float32,
			int32, int64, int, int8,
			bool:
			unionId = fmt.Sprintf("%v", opt)
		case []any:
			var vals pie.Strings
			for _, val := range opt {
				vals = append(vals, updateUnionId(val))
			}
			unionId = fmt.Sprintf(`["%s"]`, vals.Join(`","`))
		default:
			log.Panicf("unknown type:%v,value:%v", reflect.TypeOf(value), value)
		}

		return unionId
	}

	p.uniqueId = utils.Sha384(updateUnionId(p.opt))

	if p.name == "" {
		p.name = utils.ShortStr(p.uniqueId, 12)
	}

	p.name = strings.TrimSuffix(p.name, "\n")
	p.name = strings.TrimSuffix(p.name, "\r")

	return p, nil
}

func (p *ProxyAdapter) cloneOpt() map[string]any {
	o := map[string]any{}

	for k, v := range p.opt {
		o[k] = v
	}

	o["Name"] = p.name

	return o
}

func (p *ProxyAdapter) Sub4Nico() string {
	buf, err := yaml.Marshal(p.cloneOpt())
	if err != nil {
		log.Errorf("err:%v", err)
		return ""
	}
	return string(buf)
}

func (p *ProxyAdapter) Sub4Clash() string {
	buf, err := yaml.Marshal(p.cloneOpt())
	if err != nil {
		log.Errorf("err:%v", err)
		return ""
	}
	return string(buf)
}

func (p *ProxyAdapter) Sub4V2ray() string {
	u := &url.URL{
		Scheme:      "",
		Opaque:      "",
		User:        nil,
		Host:        "",
		Path:        "",
		RawPath:     "",
		ForceQuery:  false,
		RawQuery:    "",
		Fragment:    "",
		RawFragment: "",
	}

	opt := p.cloneOpt()

	query := u.Query()

	setQuery := func(key string, value any) {
		if value == nil {
			return
		}

		query.Set(key, fmt.Sprintf("%v", value))
	}

	switch p.Type() {
	case constant.Trojan:
		u.Scheme = "trojan"
		u.Fragment = fmt.Sprintf("%v", opt["Name"])

		u.Host = fmt.Sprintf("%v", opt["Server"])
		if v, ok := opt["Port"]; ok {
			u.Host = fmt.Sprintf("%s:%v", u.Host, v)
		}

		u.User = url.User(fmt.Sprintf("%v", opt["Password"]))

		setQuery("sni", opt["SNI"])

		switch fmt.Sprintf("%v", opt["Network"]) {
		case "ws":
			setQuery("type", "ws")
			if v, ok := opt["WSOpts"]; ok {
				val := v.(map[string]any)
				setQuery("wspath", val["Path"])
			}
		default:
			setQuery("type", opt["Network"])
		}

		setQuery("security", "tls")
		setQuery("headerType", "none")

	default:
		log.Panicf("unknown type %s", p.Type())
	}

	u.RawQuery = query.Encode()

	return u.String()
}

func (p *ProxyAdapter) UniqueId() string {
	return p.uniqueId
}

func (p *ProxyAdapter) DoRequest(method, rawUrl string, body io.Reader, timeout time.Duration, headers map[string]string, logic func(resp *http.Response, start time.Time) error) error {

	if timeout == 0 {
		timeout = time.Second * 5
	}

	u, err := url.Parse(rawUrl)
	if err != nil {
		return err
	}

	request, err := http.NewRequest(method, rawUrl, body)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	request.Header.Set("User-Agent", faker.New().UserAgent())
	request.Header.Set("Upgrade-Insecure-Requests", "1")
	request.Header.Set("Host", "www.google.com")
	request.Header.Set("accept-language", "en-US,en;q=0.5")
	request.Header.Set("sec-fetch-dest", "document")
	request.Header.Set("sec-fetch-mode", "navigate")
	request.Header.Set("sec-fetch-site", "none")
	request.Header.Set("sec-fetch-user", "?1")
	request.Header.Set("sec-gpc", "1")
	request.Header.Set("Accept-Encoding", "gzip, deflate")
	request.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")

	for k, v := range headers {
		request.Header.Set(k, v)
	}

	start := time.Now()

	instance, err := p.DialContext(context.TODO(), &constant.Metadata{
		AddrType: constant.AtypDomainName,
		Host:     u.Hostname(),
		DstPort: func() string {
			if u.Port() != "" {
				return u.Port()
			}

			switch u.Scheme {
			case "http":
				return "80"
			case "https":
				return "443"
			}
			return ""
		}(),
	})
	if err != nil {
		return err
	}
	defer instance.Close()

	client := http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return instance, nil
			},
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			TLSHandshakeTimeout:   time.Second * 3,
			DisableCompression:    true,
			IdleConnTimeout:       time.Second * 3,
			DisableKeepAlives:     true,
			ResponseHeaderTimeout: time.Second * 3,
			ForceAttemptHTTP2:     false,
		},

		Timeout: timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = logic(resp, start)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	return nil
}

func (p *ProxyAdapter) Get(url string, timeout time.Duration, headers map[string]string) ([]byte, error) {
	var buf []byte
	err := p.DoRequest(http.MethodGet, url, nil, timeout, headers, func(resp *http.Response, start time.Time) error {
		var err error
		buf, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}

		return nil
	})
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	return buf, nil
}

func (p *ProxyAdapter) Post(url string, body []byte, timeout time.Duration, headers map[string]string) ([]byte, error) {
	var buf []byte
	err := p.DoRequest(http.MethodPost, url, bytes.NewBuffer(body), timeout, headers, func(resp *http.Response, start time.Time) error {
		var err error
		buf, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}

		return nil
	})
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	return buf, nil
}

func (p *ProxyAdapter) PostJson(url string, reqBody, rspBody any, timeout time.Duration, headers map[string]string) error {
	var reqBuf, rspBuf []byte

	var err error
	switch x := reqBody.(type) {
	case string:
		reqBuf = []byte(x)
	case []byte:
		reqBuf = x
	default:
		reqBuf, err = json.Marshal(x)
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}
	}

	rspBuf, err = p.Post(url, reqBuf, timeout, headers)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	err = json.Unmarshal(rspBuf, rspBody)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	return nil
}

func (p *ProxyAdapter) Clone() AdapterProxy {
	np := &ProxyAdapter{
		Proxy:    p.Proxy,
		Cache:    NewAdapterCache(),
		opt:      p.opt,
		uniqueId: p.uniqueId,
		name:     p.name,
	}

	return np
}

func (p *ProxyAdapter) Name() string {
	return p.name
}
