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

func TestParseV2ray(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name, args string
	}{
		{
			name: "",
			args: `trojan://2f61436da667601d@d80d756.ga.gladns.com:3384?peer=n2.gladns.com#GLaDOS-GEOIP-HK-01-Free`,
		},
		{
			name: "",
			args: `trojan://2f61436da667601d@d80d756.ga.gladns.com:3384#GLaDOS-GEOIP-HK-01-Free`,
		},
		{
			name: "",
			args: `trojan://E73EAC23-83C8-AC21-E491-A6F3FC7C1FE0@gzdata1.tencentcdn.xyz:12132?allowInsecure=1&sni=douyincdn.com#%F0%9F%87%B8%F0%9F%87%AC%20%E6%96%B0%E5%8A%A0%E5%9D%A102`,
		},
		{
			name: "",
			args: `trojan://E73EAC23-83C8-AC21-E491-A6F3FC7C1FE0@shdata1.ourdvsss.xyz:12122?allowInsecure=1&sni=douyincdn.com#%F0%9F%87%BA%F0%9F%87%B8%20%E7%BE%8E%E5%9B%BD02`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := adapter.ParseV2ray(tt.args)
			if err != nil {
				t.Errorf("err:%v", err)
			}

			delay, err := p.URLTest(context.TODO(), "https://www.google.com")
			if err != nil {
				t.Errorf("err:%v", err)
			}

			t.Logf("%s delay is %d", p.Name(), delay)
		})
	}
}
