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
			args: "vmess://Y2hhY2hhMjAtcG9seTEzMDU6MTdmY2NjZjctMzRhZC0zYzRiLWI0N2MtNWEzOGU4OTk4M2M4QGJhaXlpbmdjdS5iYWljbS54eXo6NzQ2?remarks=%E9%A9%AC%E6%9D%A5%E8%A5%BF%E4%BA%9A%20%7C%20%E7%99%BD%E9%8A%80%20%7C%200.05X%20%7C%20%E5%8E%9F%E7%94%9F%E5%8A%A8%E6%80%81IP%20%7C%20D%20%7C&obfsParam=ssl.baiyu.click&path=/v2ray&obfs=websocket&tls=1&peer=ssl.baiyu.click",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := adapter.ParseV2ray(tt.args)
			if err != nil {
				t.Errorf("err:%v", err)
				return
			}

			t.Logf("%s", p.Sub4V2ray())
			t.Logf("%s", p.Sub4Clash())

			delay, err := p.URLTest(context.TODO(), "https://www.google.com")
			if err != nil {
				t.Errorf("err:%v", err)
				return
			}

			t.Logf("%s delay is %d", p.Name(), delay)
		})
	}
}
