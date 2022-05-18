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
			args: "ssr://MjEzLjE4My41My4xNzc6OTAyNzpvcmlnaW46YWVzLTI1Ni1jZmI6cGxhaW46UlZoT00xTXpaVkZ3YWtVM1JVcDFPQT09Lz9ncm91cD1hSFIwY0hNNkx5OTJNbkpoZVhObExtTnZiUT09JnJlbWFya3M9OEorSHQvQ2ZoN3BmVWxWZjVMK0U1NzJYNXBhdjZJR1U2WUtt",
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
