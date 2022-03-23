package adapter

import (
	"encoding/json"
	"testing"

	"github.com/Dreamacro/clash/adapter"
	"gopkg.in/yaml.v3"
)

func TestParseClash(t *testing.T) {
	var m map[string]any
	err := yaml.Unmarshal([]byte(`{name: proxy, server: 127.0.0.1, port: 15018, type: vmess, uuid: 047184b7-6da2-3d3f-ac27-6a1a8701daf8, alterId: 2, cipher: auto, tls: false, network: ws, ws-opts: {path: /, headers: {Host: 127.0.0.1}}}`), &m)
	if err != nil {
		panic(err)
	}

	buf, err := json.Marshal(m)
	if err != nil {
		t.Errorf("err:%v", err)
	}

	err = json.Unmarshal(buf, &m)
	if err != nil {
		t.Errorf("err:%v", err)
	}

	_, err = adapter.ParseProxy(m)
	if err != nil {
		t.Errorf("err:%v", err)
	}
}
