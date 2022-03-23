package adapter

import (
	"github.com/Dreamacro/clash/constant"
)

type Proxy interface {
	constant.Proxy
	ToClash() string
	ToV2ray() string
}
