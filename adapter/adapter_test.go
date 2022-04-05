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
		return
	}

	delay, err := p.URLTest(context.TODO(), "https://www.google.com")
	if err != nil {
		t.Errorf("err:%v", err)
		return
	}

	t.Logf("delay is %d", delay)
}

func TestParseClashJson(t *testing.T) {
	buf, err := json.Marshal(m)
	if err != nil {
		t.Errorf("err:%v", err)
		return
	}

	p, err := adapter.ParseClashJson(buf)
	if err != nil {
		t.Errorf("err:%v", err)
		return
	}

	delay, err := p.URLTest(context.TODO(), "https://www.google.com")
	if err != nil {
		t.Errorf("err:%v", err)
		return
	}

	t.Logf("delay is %d", delay)
}

func TestParseClashYaml(t *testing.T) {
	buf, err := yaml.Marshal(m)
	if err != nil {
		t.Errorf("err:%v", err)
		return
	}

	p, err := adapter.ParseClashYaml(buf)
	if err != nil {
		t.Errorf("err:%v", err)
		return
	}

	delay, err := p.URLTest(context.TODO(), "https://www.google.com")
	if err != nil {
		t.Errorf("err:%v", err)
		return
	}

	t.Logf("delay is %d", delay)
}

func TestParseV2ray(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name, args string
	}{
		{
			name: "",
			args: "trojan://E73EAC23-83C8-AC21-E491-A6F3FC7C1FE0@shdata1.ourdvsss.xyz:12122?security=tls&sni=douyincdn.com&type=tcp&headerType=none#%f0%9f%87%ba%f0%9f%87%b8+%e7%be%8e%e5%9b%bd02",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := adapter.ParseV2ray(tt.args)
			if err != nil {
				t.Errorf("err:%v", err)
				return
			}

			delay, err := p.URLTest(context.TODO(), "https://www.google.com")
			if err != nil {
				t.Errorf("err:%v", err)
				return
			}

			t.Logf("%s delay is %d", p.Name(), delay)
		})
	}
}
