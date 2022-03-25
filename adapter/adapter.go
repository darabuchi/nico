package adapter

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/Dreamacro/clash/adapter"
	"github.com/Dreamacro/clash/adapter/outbound"
	"github.com/Dreamacro/clash/constant"
	"github.com/darabuchi/log"
	"github.com/valyala/fastjson"
	"gopkg.in/yaml.v3"
)

func ParseClash(m map[string]any) (constant.Proxy, error) {
	p, err := adapter.ParseProxy(m)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	return p, nil
}

func ParseClashJson(s []byte) (constant.Proxy, error) {
	var m map[string]any
	err := json.Unmarshal(s, &m)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	return ParseClash(m)
}

func ParseClashYaml(s []byte) (constant.Proxy, error) {
	var m map[string]any
	err := yaml.Unmarshal(s, &m)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	return ParseClash(m)
}

func ParseV2ray(s string) (constant.Proxy, error) {
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
	default:
		return nil, fmt.Errorf("know scheme")
	}
}

func ParseLinkTrojan(s string) (constant.Proxy, error) {
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
		Network:        "",
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

	return adapter.NewProxy(at), nil
}

func ParseLinkVless(s string) (constant.Proxy, error) {
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

	return adapter.NewProxy(at), nil
}

func ParseLinkVmess(s string) (constant.Proxy, error) {
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
				"host": string(j.GetStringBytes("host")),
			},
			MaxEarlyData:        0,
			EarlyDataHeaderName: "",
		}
	}

	opt := outbound.VmessOption{
		BasicOption: outbound.BasicOption{},
		Name:        string(j.GetStringBytes("ps")),
		Server: string(func() []byte {
			host := j.GetStringBytes("host")
			if string(host) != "" {
				return host
			}

			return j.GetStringBytes("add")
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

	return adapter.NewProxy(at), nil
}
