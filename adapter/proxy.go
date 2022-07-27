package adapter

import (
	"io"
	"net/http"
	"net/url"
	"time"
	
	"github.com/Dreamacro/clash/constant"
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
	case constant.Hysteria:
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
	
	GetTotalUpload() uint64
	GetTotalDownload() uint64
}

//go:generate pie ProxyList.*
type ProxyList []AdapterProxy
