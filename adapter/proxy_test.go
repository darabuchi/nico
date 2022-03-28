package adapter_test

import (
	"context"
	"testing"

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
			args: "ss://YWVzLTI1Ni1nY206UENubkg2U1FTbmZvUzI3@134.195.196.51:8090#%E5%8A%A0%E6%8B%BF%E5%A4%A7_Tg%E9%A2%91%E9%81%93%3Ahttps%3A%2F%2Ft.me%2Fbpjzx2_3",
		},
		{
			name: "",
			args: "trojan://8422215be9bb456c8bf81e9566e6da76@1cc50d1541c.cc5d5.cf:443?security=tls&type=tcp&headerType=none#%e5%9c%88%e9%87%8f",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := adapter.ParseV2ray(tt.args)
			if err != nil {
				t.Errorf("err:%v", err)
				return
			}

			t.Logf("clash:%s", p.Sub4Clash())

			t.Logf("v2ray:%s", p.Sub4V2ray())

			delay, err := p.URLTest(context.TODO(), "https://www.google.com")
			if err != nil {
				t.Errorf("err:%v", err)
				return
			}

			t.Logf("%s delay is %d", p.Name(), delay)
		})
	}
}
