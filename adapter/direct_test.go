package adapter

import (
	"testing"
	"time"
	
	"github.com/darabuchi/log"
)

func TestNewProxyDirect(t *testing.T) {
	direct := NewProxyDirect()
	p, err := NewProxyAdapter(direct, map[string]any{
		"type": "direct",
	})
	if err != nil {
		log.Errorf("err:%v", err)
		return
	}
	
	resp, err := p.Get("https://www.baidu.com", time.Second*5, map[string]string{})
	if err != nil {
		log.Errorf("err:%v", err)
		return
	}
	log.Info(string(resp))
}
