package adapter

import (
	"encoding/json"

	"github.com/Dreamacro/clash/adapter"
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
