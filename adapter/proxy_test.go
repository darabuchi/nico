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
			args: "vmess://ew0KICAidiI6ICIyIiwNCiAgInBzIjogIuaXpeacrDE3Leino+mUgea1geWqkuS9k+OAkOWbuuWumklQ44CRIiwNCiAgImFkZCI6ICJqcC0xNy54cmF5dmlwLmNmIiwNCiAgInBvcnQiOiAiMzE5NDUiLA0KICAiaWQiOiAiZmZiNmI0NTUtNGFjMC00MzM4LWJmZTUtM2U2YjVkNzBhNzdkIiwNCiAgImFpZCI6ICIwIiwNCiAgInNjeSI6ICJhdXRvIiwNCiAgIm5ldCI6ICJ0Y3AiLA0KICAidHlwZSI6ICJub25lIiwNCiAgImhvc3QiOiAiIiwNCiAgInBhdGgiOiAiIiwNCiAgInRscyI6ICIiLA0KICAic25pIjogIiINCn0=",
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
