package adapter_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/darabuchi/nico/adapter"
	"gopkg.in/yaml.v3"
)

var m = map[string]any{}

func init() {
	err := yaml.Unmarshal([]byte(`{name: proxy, server: 18.v2-ray.cyou, port: 15018, type: vmess, uuid: 047184b7-6da2-3d3f-ac27-6a1a8701daf8, alterId: 2, cipher: auto, tls: false, network: ws, ws-opts: {path: /, headers: {Host: 18.v2-ray.cyou}}}`), &m)
	if err != nil {
		panic(err)
	}
}

func TestParseClash(t *testing.T) {
	p, err := adapter.ParseClash(m)
	if err != nil {
		t.Errorf("err:%v", err)
	}

	delay, err := p.URLTest(context.TODO(), "https://www.google.com")
	if err != nil {
		t.Errorf("err:%v", err)
	}

	t.Logf("delay is %d", delay)
}

func TestParseClashJson(t *testing.T) {
	buf, err := json.Marshal(m)
	if err != nil {
		t.Errorf("err:%v", err)
	}

	p, err := adapter.ParseClashJson(buf)
	if err != nil {
		t.Errorf("err:%v", err)
	}

	delay, err := p.URLTest(context.TODO(), "https://www.google.com")
	if err != nil {
		t.Errorf("err:%v", err)
	}

	t.Logf("delay is %d", delay)
}

func TestParseClashYaml(t *testing.T) {
	buf, err := yaml.Marshal(m)
	if err != nil {
		t.Errorf("err:%v", err)
	}

	p, err := adapter.ParseClashYaml(buf)
	if err != nil {
		t.Errorf("err:%v", err)
	}

	delay, err := p.URLTest(context.TODO(), "https://www.google.com")
	if err != nil {
		t.Errorf("err:%v", err)
	}

	t.Logf("delay is %d", delay)
}
