package adapter

import (
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"

	"github.com/Dreamacro/clash/adapter/outbound"
	"github.com/Dreamacro/clash/constant"
	"github.com/darabuchi/log"
	"github.com/elliotchance/pie/pie"
	"gopkg.in/yaml.v3"
)

type Proxy interface {
	constant.Proxy

	Sub4V2ray() string
	Sub4Clash() string
	UniqueId() string
}

type ProxyAdapter struct {
	constant.Proxy

	opt map[string]any

	uniqueId string

	name string
}

func NewProxyAdapter(adapter constant.Proxy, opt interface{}) (Proxy, error) {
	p := &ProxyAdapter{
		Proxy: adapter,
		name:  adapter.Name(),
	}

	switch v := opt.(type) {
	case map[string]any:
		p.opt = v
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
		return nil, fmt.Errorf("know %s", reflect.TypeOf(opt))
	}

	delete(p.opt, "Name")

	var updateUnionId func(opt interface{}) string
	updateUnionId = func(value interface{}) string {
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
			log.Panicf("know type:%v,value:%v", reflect.TypeOf(value), value)
		}

		return unionId
	}

	p.uniqueId = updateUnionId(p.opt)

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
		log.Panicf("know type %s", p.Type())
	}

	u.RawQuery = query.Encode()

	return u.String()
}

func (p *ProxyAdapter) UniqueId() string {
	return p.uniqueId
}

func (p *ProxyAdapter) Clone() Proxy {
	np := &ProxyAdapter{
		Proxy:    p.Proxy,
		opt:      p.opt,
		uniqueId: p.uniqueId,
	}

	return np
}
