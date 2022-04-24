package adapter

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
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

	HostName() string
	Port() string

	Sub4Nico() string
	Sub4Clash() string
	Sub4V2ray() string

	ToNico() map[string]any

	UniqueId() string
	UniqueIdShort() string

	GenDialContext(u *url.URL) (constant.Conn, error)

	GetClient() *http.Client

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
	port string
	host string
}

func NewProxyAdapter(adapter constant.Proxy, opt any) (AdapterProxy, error) {
	p := &ProxyAdapter{
		Proxy: adapter,
		name:  adapter.Name(),
		opt:   map[string]any{},
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
		err := decode(p.opt, &v)
		if err != nil {
			log.Errorf("err:%v", err)
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unknown %s", reflect.TypeOf(opt))
	}

	p.opt["type"] = p.coverAdapterType(p.Type())

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

				unionId += fmt.Sprintf(`%s:"%v",`, key, updateUnionId(val))
			})

			unionId = fmt.Sprintf("{%s}", unionId)
		case map[any]any:
			var keys pie.Strings
			nopt := map[string]any{}
			for key := range opt {
				keys = append(keys, fmt.Sprintf("%v", key))
				nopt[fmt.Sprintf("%v", key)] = opt[key]
			}

			keys.Each(func(key string) {
				val := nopt[key]
				if value == nil {
					return
				}

				unionId += fmt.Sprintf(`%s:"%v"`, key, updateUnionId(val))
			})

			unionId = fmt.Sprintf("{%s}", unionId)
		case string, []byte,
			uint32, uint64, uint, uint8,
			float64, float32,
			int32, int64, int, int8,
			bool, json.Number:
			unionId = fmt.Sprintf("%v", opt)
		case []any:
			var vals pie.Strings
			for _, val := range opt {
				vals = append(vals, updateUnionId(val))
			}
			unionId = fmt.Sprintf(`["%s"]`, vals.Join(`","`))
		case constant.AdapterType:
			unionId = fmt.Sprintf("%s", opt)
		default:
			log.Panicf("unknown type:%v,value:%v", reflect.TypeOf(value), value)
		}

		return unionId
	}

	p.opt["name"] = ""

	p.uniqueId = utils.Sha512(updateUnionId(p.opt))
	// p.uniqueId = updateUnionId(p.opt)

	if p.name == "" {
		p.name = utils.ShortStr(p.uniqueId, 12)
	}

	p.opt["unique_id"] = p.uniqueId

	p.name = strings.TrimSuffix(p.name, "\n")
	p.name = strings.TrimSuffix(p.name, "\r")

	p.host, p.port = splitHostPort(p.Addr())

	return p, nil
}

func (p *ProxyAdapter) cloneOpt() map[string]any {
	o := map[string]any{}

	for k, v := range p.opt {
		o[k] = v
	}

	o["name"] = p.name

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

func (p *ProxyAdapter) ToNico() map[string]any {
	return p.cloneOpt()
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
		Scheme:      p.coverAdapterType(p.Type()),
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
		u.Fragment = p.Name()
		u.Host = p.Addr()

		u.User = url.User(fmt.Sprintf("%v", opt["password"]))

		setQuery("sni", opt["sni"])

		switch fmt.Sprintf("%v", opt["network"]) {
		case "ws":
			setQuery("type", "ws")
			if v, ok := opt["wsopts"]; ok {
				val := v.(map[string]any)
				setQuery("wspath", val["path"])
			}
		default:
			setQuery("type", opt["network"])
		}

		setQuery("security", "tls")
		setQuery("headerType", "none")
	case constant.Shadowsocks:
		u.Fragment = p.Name()
		u.User = url.UserPassword(
			fmt.Sprintf("%v", opt["cipher"]),
			fmt.Sprintf("%v", opt["password"]))
	case constant.Vmess:
		m := map[string]any{
			"ps":   opt["name"],
			"add":  opt["server"],
			"port": opt["port"],
		}

		buf, err := json.Marshal(m)
		if err != nil {
			log.Errorf("err:%v", err)
			return ""
		}

		u.Path = base64.StdEncoding.EncodeToString(buf)
	default:
		log.Errorf("unknown type %s", p.Type())
		return ""
	}

	u.RawQuery = query.Encode()

	return u.String()
}

func (p *ProxyAdapter) coverAdapterType(adapterType constant.AdapterType) string {
	switch adapterType {
	case constant.Direct:
		return "direct"
	case constant.Reject:
		return "reject"

	case constant.Shadowsocks:
		return "ss"
	case constant.ShadowsocksR:
		return "ssr"
	case constant.Snell:
		return "snell"
	case constant.Socks5:
		return "socks_5"
	case constant.Http:
		return "http"
	case constant.Vmess:
		return "vmess"
	case constant.Trojan:
		return "trojan"
	case constant.Vless:
		return "vless"
	default:
		return "Unknown"
	}
}

func (p *ProxyAdapter) UniqueId() string {
	return p.uniqueId
}

func (p *ProxyAdapter) UniqueIdShort() string {
	return utils.ShortStr(p.uniqueId, 12)
}

func (p *ProxyAdapter) GenDialContext(u *url.URL) (constant.Conn, error) {
	return p.DialContext(context.TODO(), &constant.Metadata{
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
}

func (p *ProxyAdapter) GetClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Proxy: nil,
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				host, port := splitHostPort(addr)

				metadata := &constant.Metadata{
					AddrType: constant.AtypDomainName,
					DstPort:  port,
					Host:     host,
				}

				return p.DialContext(ctx, metadata)
			},
			TLSHandshakeTimeout:   time.Second * 3,
			DisableKeepAlives:     false,
			DisableCompression:    false,
			MaxIdleConns:          100,
			MaxIdleConnsPerHost:   10,
			MaxConnsPerHost:       10,
			IdleConnTimeout:       time.Second * 3,
			ResponseHeaderTimeout: time.Second * 3,
			ExpectContinueTimeout: time.Minute,
		},
		// CheckRedirect: func(req *http.Request, via []*http.Request) error {
		// 	return http.ErrUseLastResponse
		// },
		Timeout: time.Second * 5,
	}
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

	instance, err := p.GenDialContext(u)
	if err != nil {
		log.Errorf("err:%v", err)
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
		// CheckRedirect: func(req *http.Request, via []*http.Request) error {
		// 	return http.ErrUseLastResponse
		// },
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
		Proxy:     p.Proxy,
		Cache:     NewAdapterCache(),
		ExtraInfo: p.ExtraInfo,
		opt:       p.opt,
		uniqueId:  p.uniqueId,
		name:      p.name,
		port:      p.port,
		host:      p.host,
	}

	return np
}

func (p *ProxyAdapter) Name() string {
	return p.name
}

func (p *ProxyAdapter) HostName() string {
	return p.host
}

func (p *ProxyAdapter) Port() string {
	return p.port
}

func decodeSlice(dst []any, src any) error {
	t := reflect.TypeOf(src)
	if t.Kind() != reflect.Slice {
		panic("src is not map")
	}

	v := reflect.ValueOf(src)

	for i := 0; i < v.Len(); i++ {
		lv := v.Index(i)

		switch lv.Kind() {
		case reflect.Bool:
			dst = append(dst, lv.Bool())
		case reflect.Int,
			reflect.Int8,
			reflect.Int16,
			reflect.Int32,
			reflect.Int64:
			dst = append(dst, lv.Int())
		case reflect.Uint,
			reflect.Uint8,
			reflect.Uint16,
			reflect.Uint32,
			reflect.Uint64:
			dst = append(dst, lv.Uint())
		case reflect.Float32, reflect.Float64:
			dst = append(dst, lv.Float())
		case reflect.Complex64, reflect.Complex128:
			dst = append(dst, lv.Complex())
		case reflect.Interface:
			dst = append(dst, lv.Interface())
		case reflect.Map:
			m := map[string]any{}
			dst = append(dst, m)
			err := decodeMap(m, lv.Interface())
			if err != nil {
				log.Debugf("err:%v", err)
				return err
			}
		case reflect.Slice:
			l := []any{}
			dst = append(dst, l)
			err := decodeSlice(l, lv.Interface())
			if err != nil {
				log.Debugf("err:%v", err)
				return err
			}
		case reflect.String:
			dst = append(dst, lv.String())
		case reflect.Struct, reflect.Ptr:
			m := map[string]any{}
			dst = append(dst, m)
			err := decode(m, lv.Interface())
			if err != nil {
				log.Debugf("err:%v", err)
				return err
			}
		default:
			log.Debugf("unknown kind %s", lv.Kind())
		}
	}

	return nil
}

func decodeMap(dst map[string]any, src any) error {
	t := reflect.TypeOf(src)
	if t.Kind() != reflect.Map {
		panic("src is not map")
	}

	v := reflect.ValueOf(src)

	for _, mk := range v.MapKeys() {
		mv := v.MapIndex(mk)
		mk := fmt.Sprintf("%v", mk.Interface())

		switch mv.Kind() {
		case reflect.Bool:
			dst[mk] = mv.Bool()
		case reflect.Int,
			reflect.Int8,
			reflect.Int16,
			reflect.Int32,
			reflect.Int64:
			dst[mk] = mv.Int()
		case reflect.Uint,
			reflect.Uint8,
			reflect.Uint16,
			reflect.Uint32,
			reflect.Uint64:
			dst[mk] = mv.Uint()
		case reflect.Float32, reflect.Float64:
			dst[mk] = mv.Float()
		case reflect.Complex64, reflect.Complex128:
			dst[mk] = mv.Complex()
		case reflect.Interface:
			dst[mk] = mv.Interface()
		case reflect.Map:
			m := map[string]any{}
			dst[mk] = m
			err := decodeMap(m, mv.Interface())
			if err != nil {
				log.Debugf("err:%v", err)
				return err
			}
		case reflect.Slice:
			l := []any{}
			dst[mk] = l
			err := decodeSlice(l, mv.Interface())
			if err != nil {
				log.Debugf("err:%v", err)
				return err
			}
		case reflect.String:
			dst[mk] = mv.String()
		case reflect.Struct, reflect.Ptr:
			m := map[string]any{}
			dst[mk] = m
			err := decode(m, mv.Interface())
			if err != nil {
				log.Debugf("err:%v", err)
				return err
			}
		default:
			log.Debugf("unknown kind %s", mv.Kind())
		}
	}

	return nil
}

func decode(dst map[string]any, src any) error {
	if src == nil {
		return nil
	}

	for reflect.TypeOf(src).Kind() == reflect.Ptr {
		src = reflect.ValueOf(src).Elem().Interface()
	}

	t := reflect.TypeOf(src)
	v := reflect.ValueOf(src)

	for idx := 0; idx < t.NumField(); idx++ {
		ft := t.Field(idx)
		fv := v.Field(idx)

		tag := strings.TrimSuffix(ft.Tag.Get("proxy"), ",omitempty")

		switch fv.Kind() {
		case reflect.Bool:
			dst[tag] = fv.Bool()
		case reflect.Int,
			reflect.Int8,
			reflect.Int16,
			reflect.Int32,
			reflect.Int64:
			dst[tag] = fv.Int()
		case reflect.Uint,
			reflect.Uint8,
			reflect.Uint16,
			reflect.Uint32,
			reflect.Uint64:
			dst[tag] = fv.Uint()
		case reflect.Float32, reflect.Float64:
			dst[tag] = fv.Float()
		case reflect.Complex64, reflect.Complex128:
			dst[tag] = fv.Complex()
		case reflect.Interface:
			dst[tag] = fv.Interface()
		case reflect.Map:
			m := map[string]any{}
			dst[tag] = m
			err := decodeMap(m, fv.Interface())
			if err != nil {
				log.Debugf("err:%v", err)
				return err
			}
		case reflect.Slice:
			l := []any{}
			dst[tag] = l
			err := decodeSlice(l, fv.Interface())
			if err != nil {
				log.Debugf("err:%v", err)
				return err
			}
		case reflect.String:
			dst[tag] = fv.String()
		case reflect.Struct, reflect.Ptr:
			m := map[string]any{}
			dst[tag] = m
			err := decode(m, fv.Interface())
			if err != nil {
				log.Debugf("err:%v", err)
				return err
			}
		default:
			log.Debugf("unknown kind %s", fv.Kind())
		}
	}

	return nil
}

func splitHostPort(hostPort string) (host, port string) {
	host = hostPort

	colon := strings.LastIndexByte(host, ':')
	if colon != -1 && validOptionalPort(host[colon:]) {
		host, port = host[:colon], host[colon+1:]
	}

	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		host = host[1 : len(host)-1]
	}

	return
}

func validOptionalPort(port string) bool {
	if port == "" {
		return true
	}
	if port[0] != ':' {
		return false
	}
	for _, b := range port[1:] {
		if b < '0' || b > '9' {
			return false
		}
	}
	return true
}
