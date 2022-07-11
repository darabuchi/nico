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
			args: "vmess://eyJhZGQiOiIxMDMuMTY3LjE5Ny4yMDAiLCJ2IjoyLCJwcyI6Iui2iuWNl18xIiwicG9ydCI6IjgwIiwiaWQiOiJkOTdlZTczNy03YmMwLTRkY2EtYWNmMS04NTEzNWVkNGQ5N2UiLCJhaWQiOiIwIiwic2N5IjoiYXV0byIsIm5ldCI6IndzIiwidHlwZSI6IiIsImhvc3QiOiJIVlBOLmFrYW1haXplZC5uZXQiLCJ0bHMiOiIiLCJwYXRoIjoiLyJ9",
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

func TestYaml(t *testing.T) {
	s := `{
        "tls": false,
        "udp": true,
        "name": "_油管：全网最强白嫖 311",
        "port": 443,
        "type": "vmess",
        "uuid": "c0156451-4efb-45e2-84fc-8d315c4650db",
        "cipher": "auto",
        "server": "51.81.223.29",
        "alterId": 32,
        "unique_id": "dcf1b6a6d135577995a79c9a4145dc2903d7fb13b79023400a7fd76cd1e044c0",
        "skip-cert-verify": true
    }`
	
	var m map[string]any
	err := json.Unmarshal([]byte(s), &m)
	if err != nil {
	    t.Errorf("err:%v", err)
	    return
	}
	
	buf, err := json.Marshal(m)
	if err != nil {
	    t.Errorf("err:%v", err)
	    return
	}
	
	t.Log(string(buf))
}
