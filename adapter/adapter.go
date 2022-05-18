package adapter

import (
	"encoding/json"
	"errors"
	"net/url"
	"strconv"
	"strings"

	"github.com/Dreamacro/clash/adapter"
	"github.com/Dreamacro/clash/adapter/outbound"
	"github.com/darabuchi/log"
	"github.com/darabuchi/utils"
	"github.com/elliotchance/pie/pie"
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
	case "ssr":
		return ParseLinkSSR(s)
	default:
		return nil, ErrUnsupportedType
	}
}

func ParseLinkSSR(s string) (AdapterProxy, error) {
	urlStr := Base64Decode(strings.TrimPrefix(s, "ssr://"))
	params := strings.Split(urlStr, `:`)
	if len(params) != 6 {
		//
		return nil, errors.New("invalid ssr url")
	}

	port, _ := strconv.Atoi(params[1])

	protocol := params[2]
	obfs := params[4]
	cipher := params[3]

	suffix := strings.Split(params[5], "/?")
	if len(suffix) != 2 {
		return nil, errors.New("invalid ssr url")
	}

	password := Base64Decode(suffix[0])

	m, err := url.ParseQuery(suffix[1])
	if err != nil {
		return nil, err
	}

	var obfsParam, protocolParam, name string
	for k, v := range m {
		de := Base64Decode(v[0])
		switch k {
		case "obfsparam":
			obfsParam = de
		case "protoparam":
			protocolParam = de
		case "remarks":
			name = de
		}
	}

	if protocol == "origin" && obfs == "plain" {
		switch cipher {
		case "aes-128-gcm", "aes-192-gcm", "aes-256-gcm",
			"aes-128-cfb", "aes-192-cfb", "aes-256-cfb",
			"aes-128-ctr", "aes-192-ctr", "aes-256-ctr",
			"rc4-md5", "chacha20", "chacha20-ietf", "xchacha20",
			"chacha20-ietf-poly1305", "xchacha20-ietf-poly1305":
			// opt := outbound.ShadowSocksOption{
			// 	BasicOption: outbound.BasicOption{},
			// 	Name:        name,
			// 	Server:      params[0],
			// 	Port:        port,
			// 	Password:    password,
			// 	Cipher:      cipher,
			// 	UDP:         false,
			// 	Plugin:      "",
			// 	PluginOpts:  nil,
			// }
			return nil, errors.New("invalid ssr url")
		}
	}

	opt := outbound.ShadowSocksROption{
		BasicOption:   outbound.BasicOption{},
		Name:          name,
		Server:        params[0],
		Port:          port,
		Password:      password,
		Cipher:        cipher,
		Obfs:          obfs,
		ObfsParam:     obfsParam,
		Protocol:      protocol,
		ProtocolParam: protocolParam,
		UDP:           true,
	}

	log.Debugf("ssr opt:%+v", opt)

	at, err := outbound.NewShadowSocksR(opt)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	return NewProxyAdapter(adapter.NewProxy(at), opt)
}

func ParseLinkSS(s string) (AdapterProxy, error) {
	var urlStr string
	var fragment string
	bu, err := url.Parse(s)
	if err != nil {
		log.Errorf("err:%v", err)
		urlStr = "ss://" + Base64Decode(strings.TrimPrefix(s, "ss://"))
	} else {
		fragment = bu.Fragment
		bu.Fragment = ""
		urlStr = "ss://" + Base64Decode(strings.TrimPrefix(bu.String(), "ss://"))
	}

	log.Debugf("urlStr:%s", urlStr)

	u, err := url.Parse(urlStr)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	if fragment == "" {
		fragment = u.Fragment
	}

	port, _ := strconv.Atoi(u.Port())

	var cipher, password string
	// 对username解析
	userStr := Base64Decode(u.User.String())

	log.Debugf("userStr:%s", userStr)

	userSplit := strings.Split(userStr, ":")
	if len(userSplit) > 0 {
		cipher = userSplit[0]
	}

	if len(userSplit) > 1 {
		password = userSplit[1]
	}

	opt := outbound.ShadowSocksOption{
		BasicOption: outbound.BasicOption{},
		Name:        fragment,
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
	var opt outbound.VmessOption
	base64Str := Base64Decode(strings.TrimPrefix(s, "vmess://"))
	m, err := utils.NewMapWithJson([]byte(base64Str))
	if err != nil {
		log.Errorf("err:%v", err)

		u, err := url.Parse(s)
		if err != nil {
			log.Errorf("err:%v", err)
			return nil, err
		} else {
			u.Host = Base64Decode(u.Host)
		}

		urlStr, err := url.QueryUnescape(u.String())
		if err != nil {
			log.Errorf("err:%v", err)
			return nil, err
		}

		u, err = url.Parse(urlStr)
		if err != nil {
			log.Errorf("err:%v", err)
			return nil, err
		}

		var wsOpts outbound.WSOptions
		var network string
		switch u.Query().Get("obfs") {
		case "websocket":
			network = "ws"
			wsOpts = outbound.WSOptions{
				Path: u.Query().Get("path"),
				Headers: map[string]string{
					"Host": u.Query().Get("peer"),
				},
				MaxEarlyData:        0,
				EarlyDataHeaderName: "",
			}
		}

		opt = outbound.VmessOption{
			BasicOption: outbound.BasicOption{},
			Name:        u.Query().Get("remarks"),
			Server:      u.Hostname(),
			Port:        utils.ToInt(u.Port()),
			UUID: func() string {
				pwd, ok := u.User.Password()
				if ok {
					return pwd
				}
				return u.User.Username()
			}(),
			AlterID: 0,
			Cipher: func() string {
				_, ok := u.User.Password()
				if ok {
					return u.User.Username()
				}
				return ""
			}(),
			UDP:            true,
			Network:        network,
			TLS:            utils.ToBool(u.Query().Get("tls")),
			SkipCertVerify: true,
			ServerName:     "",
			HTTPOpts:       outbound.HTTPOptions{},
			HTTP2Opts:      outbound.HTTP2Options{},
			GrpcOpts:       outbound.GrpcOptions{},
			WSOpts:         wsOpts,
		}

	} else {
		var wsOpts outbound.WSOptions
		switch m.GetString("net") {
		case "ws":
			wsOpts = outbound.WSOptions{
				Path: m.GetString("path"),
				Headers: map[string]string{
					"Host": m.GetString("host"),
				},
				MaxEarlyData:        0,
				EarlyDataHeaderName: "",
			}
		}

		opt = outbound.VmessOption{
			BasicOption: outbound.BasicOption{},
			Name:        m.GetString("ps"),
			Server: func() string {
				if m.GetString("add") != "" {
					return m.GetString("add")
				}

				return m.GetString("host")
			}(),
			Port:    m.GetInt("port"),
			UUID:    m.GetString("id"),
			AlterID: m.GetInt("aid"),
			Cipher: func() string {
				if m.GetString("scy") != "" {
					return m.GetString("scy")
				}

				return "auto"
			}(),
			UDP:     true,
			Network: m.GetString("net"),
			TLS: func() bool {
				val, err := m.Get("tls")
				if err != nil {
					return false
				}

				switch utils.CheckValueType(val) {
				case utils.ValueString:
					return m.GetString("tls") != ""
				case utils.ValueBool:
					return m.GetBool("tls")
				default:
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
	}

	log.Debugf("vmess opt:%+v", opt)

	at, err := outbound.NewVmess(opt)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	return NewProxyAdapter(adapter.NewProxy(at), opt)
}
