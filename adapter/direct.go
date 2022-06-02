package adapter

import (
	"context"
	
	"github.com/Dreamacro/clash/adapter/outbound"
	"github.com/Dreamacro/clash/constant"
)

type ProxyDirect struct {
	*outbound.Direct
}

func (p *ProxyDirect) DialUDP(metadata *constant.Metadata) (constant.PacketConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), constant.DefaultUDPTimeout)
	defer cancel()
	return p.ListenPacketContext(ctx, metadata)
}

func (p *ProxyDirect) Alive() bool {
	return true
}

func (p *ProxyDirect) DelayHistory() []constant.DelayHistory {
	return nil
}

func (p *ProxyDirect) LastDelay() uint16 {
	return 0
}

func (p *ProxyDirect) URLTest(ctx context.Context, url string) (uint16, error) {
	return 0, nil
}

func (p *ProxyDirect) Dial(metadata *constant.Metadata) (constant.Conn, error) {
	conn, err := p.Direct.DialContext(context.TODO(), metadata)
	return conn, err
}

func NewProxyDirect() constant.Proxy {
	return &ProxyDirect{
		Direct: outbound.NewDirect(),
	}
}
