package rule

import (
	"sync"

	"github.com/Dreamacro/clash/constant"
	"github.com/darabuchi/log"
	"github.com/darabuchi/nico/adapter"
	"github.com/darabuchi/nico/config"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

var ar = NewAdapterRule()

func Match(metadata *constant.Metadata) adapter.AdapterType {
	return ar.Match(metadata)
}

func AddRules(rules ...adapter.Rule) *AdapterRule {
	return ar.AddRule(rules...)
}

func Sync() {
	ar.Sync()
}

func GetAdapterRule() *AdapterRule {
	return ar
}

type AdapterRule struct {
	lock    sync.RWMutex
	ruleMap map[string]adapter.Rule

	c *viper.Viper
}

func NewAdapterRule() *AdapterRule {
	p := &AdapterRule{
		ruleMap: map[string]adapter.Rule{},
	}

	value := config.Get("rule")
	if value != nil {
		b, err := yaml.Marshal(value)
		if err != nil {
			log.Errorf("err:%v", err)
		} else {
			var l []adapter.RuleInfo
			err = yaml.Unmarshal(b, &l)
			if err != nil {
				log.Errorf("err:%v", err)
			} else {
				for _, info := range l {
					r, err := NewRule(info)
					if err != nil {
						log.Errorf("err:%v", err)
					} else {
						p.addRule(r)
					}
				}
			}
		}
	}

	return p
}

func (p *AdapterRule) AddRule(rules ...adapter.Rule) *AdapterRule {
	p.addRule(rules...)
	return p
}

func (p *AdapterRule) addRule(rules ...adapter.Rule) {
	p.lock.Lock()
	defer p.lock.Unlock()

	for _, rule := range rules {
		if _, ok := p.ruleMap[rule.Key()]; !ok {
			ex := rule.Export()
			log.Infof("add rule %s,%s,%s", ex.Rule, ex.Payload, ex.Adapter)

			p.ruleMap[rule.Key()] = rule
		}
	}
}

func (p *AdapterRule) Match(metadata *constant.Metadata) adapter.AdapterType {
	p.lock.RLock()
	defer p.lock.RUnlock()

	for _, rule := range p.ruleMap {
		if rule.Match(metadata) {
			return rule.AdapterType()
		}
	}

	return adapter.Direct
}

func (p *AdapterRule) Sync() {
	p.lock.RLock()
	defer p.lock.RUnlock()
	var l []adapter.RuleInfo

	for _, rule := range p.ruleMap {
		l = append(l, rule.Export())
	}

	config.Set("rule", l)
}
