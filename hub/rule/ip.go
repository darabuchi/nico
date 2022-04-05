package rule

import (
	"fmt"
	"net"

	"github.com/Dreamacro/clash/constant"
	"github.com/darabuchi/nico/adapter"
)

type SrcIp struct {
	at adapter.AdapterType
	ip net.IP
}

func (p *SrcIp) Match(metadata *constant.Metadata) bool {
	return p.ip.Equal(metadata.SrcIP)
}

func (p *SrcIp) AdapterType() adapter.AdapterType {
	return p.at
}

func (p *SrcIp) Type() adapter.RuleType {
	return adapter.ScrIp
}

func (p *SrcIp) Export() adapter.RuleInfo {
	return export(p.Type(), p.Key(), p.AdapterType())
}

func (p *SrcIp) Key() string {
	return p.ip.String()
}

func NewSrcIp(ip string, at adapter.AdapterType) (adapter.Rule, error) {
	i := net.ParseIP(ip)
	if i == nil {
		return nil, fmt.Errorf("%s is not ip", ip)
	}

	p := &SrcIp{
		ip: i,
		at: at,
	}

	return p, nil
}

type DstIp struct {
	at adapter.AdapterType
	ip net.IP
}

func (p *DstIp) Match(metadata *constant.Metadata) bool {
	return p.ip.Equal(metadata.DstIP)
}

func (p *DstIp) AdapterType() adapter.AdapterType {
	return p.at
}

func (p *DstIp) Type() adapter.RuleType {
	return adapter.ScrIp
}

func (p *DstIp) Export() adapter.RuleInfo {
	return export(p.Type(), p.Key(), p.AdapterType())
}

func (p *DstIp) Key() string {
	return p.ip.String()
}

func NewDstIp(ip string, at adapter.AdapterType) (adapter.Rule, error) {
	i := net.ParseIP(ip)
	if i == nil {
		return nil, fmt.Errorf("%s is not ip", ip)
	}

	p := &DstIp{
		ip: i,
		at: at,
	}

	return p, nil
}
