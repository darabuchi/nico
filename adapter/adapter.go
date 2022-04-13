package adapter

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/url"
	"strconv"
	"strings"

	"github.com/Dreamacro/clash/adapter"
	"github.com/Dreamacro/clash/adapter/outbound"
	"github.com/darabuchi/log"
	"github.com/elliotchance/pie/pie"
	"github.com/valyala/fastjson"
	"gopkg.in/yaml.v3"
)

var (
	ErrUnsupportedType = errors.New("unsupported type")
)

func ParseClash(m map[string]any) (AdapterProxy, error) {
	p, err := adapter.ParseProxy(m)
	if err != nil {
		log.Errorf("err:%v", err)

		if strings.Contains(err.Error(), "unsupport proxy type") {
			return nil, ErrUnsupportedType
		}

		return nil, err
	}

	return NewProxyAdapter(p, m)
}

func ParseClashJson(s []byte) (AdapterProxy, error) {
	var m map[string]any
	err := json.Unmarshal(s, &m)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	return ParseClash(m)
}

func ParseClashYaml(s []byte) (AdapterProxy, error) {
	var m map[string]any
	err := yaml.Unmarshal(s, &m)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	return ParseClash(m)
}

func ParseV2ray(s string) (AdapterProxy, error) {
	s = strings.TrimSuffix(s, "\n")
	s = strings.TrimSuffix(s, "\r")

	u, err := url.Parse(s)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	switch u.Scheme {
	case "trojan":
		return ParseLinkTrojan(s)
	case "vmess":
		return ParseLinkVmess(s)
	case "vless":
		return ParseLinkVless(s)
	case "ss":
		return ParseLinkSS(s)
	default:
		return nil, ErrUnsupportedType
	}
}

func ParseLinkSS(s string) (AdapterProxy, error) {
	urlStr := "ss://" + Base64Decode(strings.TrimPrefix(s, "ss://"))

	u, err := url.Parse(urlStr)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	port, _ := strconv.Atoi(u.Port())

	var cipher, password string

	// 对username解析
	userStr := Base64Decode(u.User.String())

	userSplit := strings.Split(userStr, ":")
	if len(userSplit) > 0 {
		cipher = userSplit[0]
	}

	if len(userSplit) > 1 {
		password = userSplit[1]
	}

	opt := outbound.ShadowSocksOption{
		BasicOption: outbound.BasicOption{},
		Name:        u.Fragment,
		Server:      u.Hostname(),
		Port:        port,
		Password:    password,
		Cipher:      cipher,
		UDP:         true,
		Plugin:      "",
		PluginOpts:  nil,
	}

	log.Debugf("ss opt:%+v", opt)

	at, err := outbound.NewShadowSocks(opt)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	return NewProxyAdapter(adapter.NewProxy(at), opt)
}

func ParseLinkTrojan(s string) (AdapterProxy, error) {
	u, err := url.Parse(s)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	sni := u.Query().Get("sni")
	if sni == "" {
		sni = u.Hostname()
	}

	// 处理ws
	var wsOpt outbound.WSOptions
	if u.Query().Get("ws") != "1" || strings.ToLower(u.Query().Get("ws")) == "true" {
		wsOpt = outbound.WSOptions{
			Path:                u.Query().Get("wspath"),
			Headers:             nil,
			MaxEarlyData:        0,
			EarlyDataHeaderName: "",
		}
	}

	transformType := u.Query().Get("type")
	transformType, _ = url.QueryUnescape(transformType)

	var alpn []string
	if transformType == "h2" {
		alpn = append(alpn, transformType)
	}

	for _, val := range strings.Split(u.Query().Get("alpn"), ",") {
		if val == "" {
			continue
		}
		alpn = append(alpn, val)
	}

	alpn = pie.Strings(alpn).Unique()

	port, _ := strconv.Atoi(u.Port())

	opt := outbound.TrojanOption{
		BasicOption: outbound.BasicOption{
			Interface:   "",
			RoutingMark: 0,
		},
		Name:           u.Fragment,
		Server:         u.Hostname(),
		Password:       u.User.String(),
		Port:           port,
		ALPN:           alpn,
		SNI:            sni,
		SkipCertVerify: true,
		UDP:            true,
		Network:        transformType,
		GrpcOpts: outbound.GrpcOptions{
			GrpcServiceName: "",
		},
		WSOpts: wsOpt,
	}

	log.Debugf("trojan opt:%+v", opt)

	at, err := outbound.NewTrojan(opt)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	return NewProxyAdapter(adapter.NewProxy(at), opt)
}

func ParseLinkVless(s string) (AdapterProxy, error) {
	u, err := url.Parse(s)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	sni := u.Query().Get("sni")
	if sni == "" {
		sni = u.Hostname()
	}

	transformType := u.Query().Get("type")
	transformType, _ = url.QueryUnescape(transformType)

	var alpn []string
	if transformType == "h2" {
		alpn = append(alpn, transformType)
	}

	port, _ := strconv.Atoi(u.Port())

	opt := outbound.VlessOption{
		Name:   u.Fragment,
		Server: u.Hostname(),
		Port:   port,
		UUID:   u.User.String(),
		UDP:    true,
		TLS: func() bool {
			return u.Query().Get("security") == "tls"
		}(),
		Network: func() string {
			if u.Query().Get("type") != "" {
				return u.Query().Get("type")
			}

			return "tcp"
		}(),
		WSPath:         u.Query().Get("path"),
		WSHeaders:      nil,
		SkipCertVerify: true,
		ServerName: func() string {
			if u.Query().Get("host") != "" {
				return u.Query().Get("host")
			}
			return u.Query().Get("sni")
		}(),
		Flow: u.Query().Get("flow"),
	}

	log.Debugf("vless opt:%+v", opt)

	at, err := outbound.NewVless(opt)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	return NewProxyAdapter(adapter.NewProxy(at), opt)
}

func ParseLinkVmess(s string) (AdapterProxy, error) {
	base64Str, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(s, "vmess://"))
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	j, err := fastjson.ParseBytes(base64Str)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	var wsOpts outbound.WSOptions
	switch string(j.GetStringBytes("net")) {
	case "ws":
		wsOpts = outbound.WSOptions{
			Path: string(j.GetStringBytes("path")),
			Headers: map[string]string{
				"Host": string(j.GetStringBytes("host")),
			},
			MaxEarlyData:        0,
			EarlyDataHeaderName: "",
		}
	}

	opt := outbound.VmessOption{
		BasicOption: outbound.BasicOption{},
		Name:        string(j.GetStringBytes("ps")),
		Server: string(func() []byte {
			host := j.GetStringBytes("add")
			if string(host) != "" {
				return host
			}

			return j.GetStringBytes("host")
		}()),
		Port:    getInt(j, "port"),
		UUID:    string(j.GetStringBytes("id")),
		AlterID: getInt(j, "aid"),
		Cipher: string(func() []byte {
			scy := j.GetStringBytes("scy")
			if string(scy) != "" {
				return scy
			}

			return []byte("auto")
		}()),
		UDP:     true,
		Network: string(j.GetStringBytes("net")),
		TLS: func() bool {
			tls := j.Get("tls")
			switch tls.Type() {
			case fastjson.TypeString:
				return string(j.GetStringBytes("tls")) != ""
			case fastjson.TypeTrue:
				return true
			case fastjson.TypeFalse:
				return false
			default:
				log.Warnf("tls type is %s", tls.Type())
				return false
			}
		}(),
		SkipCertVerify: true,
		ServerName:     "",
		HTTPOpts:       outbound.HTTPOptions{},
		HTTP2Opts:      outbound.HTTP2Options{},
		GrpcOpts:       outbound.GrpcOptions{},
		WSOpts:         wsOpts,
	}

	log.Debugf("vmess opt:%+v", opt)

	at, err := outbound.NewVmess(opt)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	return NewProxyAdapter(adapter.NewProxy(at), opt)
}
