package adapter

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/Dreamacro/clash/adapter"
	"github.com/Dreamacro/clash/adapter/outbound"
	"github.com/Dreamacro/clash/constant"
	"github.com/darabuchi/log"
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

	opt := outbound.TrojanOption{
		BasicOption: outbound.BasicOption{
			Interface:   "",
			RoutingMark: 0,
		},
		Name:           u.Fragment,
		Server:         u.Hostname(),
		Password:       u.User.String(),
		ALPN:           nil,
		SNI:            u.Query().Get("sni"),
		SkipCertVerify: true,
		UDP:            true,
		Network:        "",
		GrpcOpts: outbound.GrpcOptions{
			GrpcServiceName: "",
		},
		WSOpts: outbound.WSOptions{
			Path:                "",
			Headers:             nil,
			MaxEarlyData:        0,
			EarlyDataHeaderName: "",
		},
	}

	opt.Port, _ = strconv.Atoi(u.Port())

	at, err := outbound.NewTrojan(opt)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	if opt.SNI == "" {
		opt.SNI = opt.Server
	}

	return adapter.NewProxy(at), nil
}
