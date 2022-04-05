package rule

import (
	"fmt"

	"github.com/darabuchi/log"
	"github.com/darabuchi/nico/adapter"
	"gopkg.in/yaml.v3"
)

func ParseRule(s string) (adapter.Rule, error) {
	var rule adapter.RuleInfo
	err := yaml.Unmarshal([]byte(s), rule)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	return NewRule(rule)
}

func NewRule(rule adapter.RuleInfo) (adapter.Rule, error) {
	switch rule.Rule {
	case adapter.Domain.String():
		return NewDomain(rule.Payload, adapter.ParseAdapterType(rule.Adapter))
	case adapter.ScrIp.String():
		return NewSrcIp(rule.Payload, adapter.ParseAdapterType(rule.Adapter))
	case adapter.DstIp.String():
		return NewDstIp(rule.Payload, adapter.ParseAdapterType(rule.Adapter))

	default:
		return nil, fmt.Errorf("unknow rule type %s", rule.Rule)
	}
}

func export(ruleType adapter.RuleType, payload string, at adapter.AdapterType) adapter.RuleInfo {
	return adapter.RuleInfo{
		Rule:    ruleType.String(),
		Payload: payload,
		Adapter: at.String(),
	}
}
