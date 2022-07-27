package adapter_test

import (
	"testing"
	"time"
	
	"github.com/darabuchi/log"
	"github.com/darabuchi/nico/adapter"
)

func TestNewProxyAdapter(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name, args string
	}{
		{
			name: "",
			args: "trojan://28b31550-9aae-40f5-9511-b3d7e475e3fc@s2.apifly.uk:34501?security=tls&sni=sni.apifly.uk&type=tcp&headerType=none#%F0%9F%87%AD%F0%9F%87%B0%20%E9%A6%99%E6%B8%AF%20%E8%B4%9F%E8%BD%BD%E5%9D%87%E8%A1%A1%2002\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := adapter.ParseV2ray(tt.args)
			if err != nil {
				t.Errorf("err:%v", err)
				return
			}
			
			_, err = p.Get("https://www.google.com", time.Second*5, map[string]string{})
			if err != nil {
				t.Errorf("err:%v", err)
				return
			}
			
			log.Info(p.GetTotalDownload())
			log.Info(p.GetTotalUpload())
		})
	}
}
