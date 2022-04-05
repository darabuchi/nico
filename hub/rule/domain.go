package rule

import (
	"github.com/Dreamacro/clash/constant"
	"github.com/darabuchi/nico/adapter"
)

type Domain struct {
	at     adapter.AdapterType
	domain string
}

func (p *Domain) Match(metadata *constant.Metadata) bool {
	return p.domain == metadata.Host
}

func (p *Domain) AdapterType() adapter.AdapterType {
	return p.at
}

func (p *Domain) Type() adapter.RuleType {
	return adapter.Domain
}

func (p *Domain) Export() adapter.RuleInfo {
	return export(p.Type(), p.Key(), p.AdapterType())
}

func (p *Domain) Key() string {
	return p.domain
}

func NewDomain(domain string, at adapter.AdapterType) (adapter.Rule, error) {
	return &Domain{
		at:     at,
		domain: domain,
	}, nil
}
