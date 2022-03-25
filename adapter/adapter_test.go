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
		{
			name: "",
			args: `trojan://E73EAC23-83C8-AC21-E491-A6F3FC7C1FE0@gzdata1.tencentcdn.xyz:12133?allowinsecure=1&sni=douyincdn.com&mux=0&ws=0&wspath=&wshost=&ss=0&ssmethod=aes-128-gcm&sspasswd=&group=#%F0%9F%87%B8%F0%9F%87%AC%20%E6%96%B0%E5%8A%A0%E5%9D%A103`,
		},
		{
			name: "",
			args: `trojan://E73EAC23-83C8-AC21-E491-A6F3FC7C1FE0@gzdata1.tencentcdn.xyz:12133?allowinsecure=1&sni=douyincdn.com&mux=0&ws=0&wspath=&wshost=&ss=0&ssmethod=aes-128-gcm&sspasswd=&group=#%F0%9F%87%B8%F0%9F%87%AC%20%E6%96%B0%E5%8A%A0%E5%9D%A103`,
		},
		{
			name: "",
			args: "vmess://ew0KICAidiI6ICIyIiwNCiAgInBzIjogIijwn4e68J+HuOe+juWbvSnml6Llt7Log4zlvbHlgL7lpKnkuIvvvIzkvZXlv4XovazouqvkubHoirPljY7jgIIiLA0KICAiYWRkIjogImR5Y3hoLm1sIiwNCiAgInBvcnQiOiAiMjg1MiIsDQogICJpZCI6ICJjMGJlYzNhNS1iZWM0LTRiYzQtZmE1OS1kNzVhNmZlYmE0MDUiLA0KICAiYWlkIjogIjAiLA0KICAic2N5IjogImF1dG8iLA0KICAibmV0IjogInRjcCIsDQogICJ0eXBlIjogInZtZXNzIiwNCiAgImhvc3QiOiAiZHljeGgubWwiLA0KICAicGF0aCI6ICIiLA0KICAidGxzIjogIiIsDQogICJzbmkiOiAiIiwNCiAgImFscG4iOiAiIg0KfQ==",
		},
		{
			name: "",
			args: "vmess://ew0KICAidiI6ICIyIiwNCiAgInBzIjogIijwn4e48J+HrOaWsOWKoOWdoSnlhbblrp7vvIznlLXohJHmuLjmiI/ku47lsI/lsLHorq3nu4PkvaDvvIzmiopCb3Nz5b2T5L2c6Ieq5bex5pyA5aSn55qE5pWM5Lq644CCIiwNCiAgImFkZCI6ICI1Mi4xNjMuODkuMjUzIiwNCiAgInBvcnQiOiAiMzUyNTIiLA0KICAiaWQiOiAiODRhNjk1MTItNzM2NS00ODdmLWIyYWMtZDFkNWM0ZTA1NzUwIiwNCiAgImFpZCI6ICIwIiwNCiAgInNjeSI6ICJhdXRvIiwNCiAgIm5ldCI6ICJ0Y3AiLA0KICAidHlwZSI6ICJ2bWVzcyIsDQogICJob3N0IjogIjUyLjE2My44OS4yNTMiLA0KICAicGF0aCI6ICIiLA0KICAidGxzIjogIiIsDQogICJzbmkiOiAiIiwNCiAgImFscG4iOiAiIg0KfQ==",
		},
		{
			name: "",
			args: "trojan://283695dc-fcc8-11ea-8684-f23c913c8d2b@us2.tcpbbr.net:443?security=tls&type=tcp&headerType=none#(%f0%9f%87%ba%f0%9f%87%b8%e7%be%8e%e5%9b%bd)%e6%88%91%e6%83%b3%e9%87%8d%e6%96%b0%e8%ae%a4%e8%af%86%e4%bd%a0%ef%bc%8c%e4%bb%8e%e4%bd%a0%e5%8f%ab%e4%bb%80%e4%b9%88%e5%90%8d%e5%ad%97%e5%bc%80%e5%a7%8b%e3%80%82%e4%bd%a0%e5%8f%ab%e4%bb%80%e4%b9%88%e6%9d%a5%e7%9d%80%ef%bc%9f",
		},
		{
			name: "",
			args: "vmess://ew0KICAidiI6ICIyIiwNCiAgInBzIjogIijwn4et8J+HsOmmmea4rynniLHnrJHnmoTlpbPnlJ/ov5DmsJTkuI3kvJrlpKrlt67jgILor7Tlrp7or53vvIzlpoLmnpzkuIDkuKrlpbPnlJ/ov5DmsJTkuIDnm7TkuI3lpb3vvIzmiJHkuI3nn6XpgZPlpbnmgI7kuYjnrJHlvpflh7rmnaXjgIIiLA0KICAiYWRkIjogIjE1LnYyLXJheS5jeW91IiwNCiAgInBvcnQiOiAiMTUwMTUiLA0KICAiaWQiOiAiMDQ3MTg0YjctNmRhMi0zZDNmLWFjMjctNmExYTg3MDFkYWY4IiwNCiAgImFpZCI6ICIyIiwNCiAgInNjeSI6ICJhdXRvIiwNCiAgIm5ldCI6ICJ3cyIsDQogICJ0eXBlIjogInZtZXNzIiwNCiAgImhvc3QiOiAiMTUudjItcmF5LmN5b3UiLA0KICAicGF0aCI6ICIvIiwNCiAgInRscyI6ICIiLA0KICAic25pIjogIiIsDQogICJhbHBuIjogIiINCn0=",
		},
		{
			name: "",
			args: "vmess://ew0KICAidiI6ICIyIiwNCiAgInBzIjogIijwn4et8J+HsOmmmea4rynpg73or7Tlp5DmvILkuq7vvIzlhbblrp7pg73mmK/lpoblh7rmnaXnmoQiLA0KICAiYWRkIjogIjE4LnYyLXJheS5jeW91IiwNCiAgInBvcnQiOiAiMTUwMTgiLA0KICAiaWQiOiAiMDQ3MTg0YjctNmRhMi0zZDNmLWFjMjctNmExYTg3MDFkYWY4IiwNCiAgImFpZCI6ICIyIiwNCiAgInNjeSI6ICJhdXRvIiwNCiAgIm5ldCI6ICJ3cyIsDQogICJ0eXBlIjogInZtZXNzIiwNCiAgImhvc3QiOiAiMTgudjItcmF5LmN5b3UiLA0KICAicGF0aCI6ICIiLA0KICAidGxzIjogIiIsDQogICJzbmkiOiAiIiwNCiAgImFscG4iOiAiIg0KfQ==",
		},
		{
			name: "",
			args: "trojan://73730588-b569-3569-a42c-7f51318a171b@gz-iplc-kl.klee.store:11159?sni=gz-iplc-kl.klee.store#%F0%9F%87%B0%F0%9F%87%B7%20V3%E9%9F%A9%E5%9B%BDA11%E4%B8%A82x%E4%B8%A8IPLC%E4%B8%93%E7%BA%BF%22%22%22",
		},
		{
			name: "",
			args: "vmess://ew0KICAidiI6ICIyIiwNCiAgInBzIjogIuOAkOWbuuWumklQ44CR5b635Zu9MDEt6Kej6ZSB5rWB5aqS5L2TIiwNCiAgImFkZCI6ICJvcmFjbGVjbG91ZC5oYW93YW53YW4ueHl6IiwNCiAgInBvcnQiOiAiNDQzIiwNCiAgImlkIjogIjU4MTFiMzIxLWEzYjAtNGJiZi05YmUzLTM3YjVmZjhiZTVjYSIsDQogICJhaWQiOiAiMCIsDQogICJzY3kiOiAiYXV0byIsDQogICJuZXQiOiAidGNwIiwNCiAgInR5cGUiOiAibm9uZSIsDQogICJob3N0IjogIiIsDQogICJwYXRoIjogIiIsDQogICJ0bHMiOiAidGxzIiwNCiAgInNuaSI6ICIiLA0KICAiYWxwbiI6ICIiDQp9",
		},
		{
			name: "",
			args: "vmess://ew0KICAidiI6ICIyIiwNCiAgInBzIjogIuOAkOWbuuWumklQ44CR5rOV5Zu9MDEt6Kej6ZSB5rWB5aqS5L2TIiwNCiAgImFkZCI6ICI1MS4xNTkuNi4yMjEiLA0KICAicG9ydCI6ICI1OTI2MiIsDQogICJpZCI6ICJjYjNiMjI2Yy1hMjdhLTRlZWUtYWQ3ZC05NzM4Y2UzNjkxM2IiLA0KICAiYWlkIjogIjY0IiwNCiAgInNjeSI6ICJhdXRvIiwNCiAgIm5ldCI6ICJ0Y3AiLA0KICAidHlwZSI6ICJub25lIiwNCiAgImhvc3QiOiAiIiwNCiAgInBhdGgiOiAiIiwNCiAgInRscyI6ICIiLA0KICAic25pIjogIiIsDQogICJhbHBuIjogIiINCn0=",
		},
		{
			name: "",
			args: "vmess://ew0KICAidiI6ICIyIiwNCiAgInBzIjogIuOAkOWbuuWumklQ44CR6Z+p5Zu9MDQt6Kej6ZSB5rWB5aqS5L2TIiwNCiAgImFkZCI6ICJkb3dubG9hZC5mcmFua2llNDk0Lm9ubGluZSIsDQogICJwb3J0IjogIjQ0MyIsDQogICJpZCI6ICJlNmEwMGYwNi0xNjhmLTRiYmUtYjc3My00MmM4NDhkNzRjZDUiLA0KICAiYWlkIjogIjAiLA0KICAic2N5IjogImF1dG8iLA0KICAibmV0IjogIndzIiwNCiAgInR5cGUiOiAibm9uZSIsDQogICJob3N0IjogIiIsDQogICJwYXRoIjogIi9mcmFua2llNDkiLA0KICAidGxzIjogInRscyIsDQogICJzbmkiOiAiIiwNCiAgImFscG4iOiAiIg0KfQ==",
		},
		{
			name: "",
			args: "vmess://ew0KICAidiI6ICIyIiwNCiAgInBzIjogIuOAkOWbuuWumklQ44CR5pel5pysMDUt6Kej6ZSB5rWB5aqS5L2TIiwNCiAgImFkZCI6ICJqcDA1LnZpcC12MnJheS5tbCIsDQogICJwb3J0IjogIjI1MTQwIiwNCiAgImlkIjogIjlkZGIxMWNjLWEzZmEtNGNhYS05YmY1LTdhM2NjNWJhY2E4MyIsDQogICJhaWQiOiAiMCIsDQogICJzY3kiOiAiYXV0byIsDQogICJuZXQiOiAidGNwIiwNCiAgInR5cGUiOiAibm9uZSIsDQogICJob3N0IjogIiIsDQogICJwYXRoIjogIiIsDQogICJ0bHMiOiAiIiwNCiAgInNuaSI6ICIiLA0KICAiYWxwbiI6ICIiDQp9",
		},
		{
			name: "",
			args: "trojan://NAN2005shan@nafei.itomato.top:443?security=tls&type=tcp&headerType=none#%e3%80%90%e5%9b%ba%e5%ae%9aIP%e3%80%91%e7%be%8e%e5%9b%bd25-%e8%a7%a3%e9%94%81%e6%b5%81%e5%aa%92%e4%bd%93",
		},
		{
			name: "",
			args: "vless://b03241a5-f3c5-4750-eba5-3cd4967e4dd7@fuqing.tk:443?encryption=none&security=tls&sni=fuqing.tk&type=ws&host=fuqing.tk&path=%2fdate2021#%e3%80%90%e5%9b%ba%e5%ae%9aIP%e3%80%91%e6%97%a5%e6%9c%ac12-%e8%a7%a3%e9%94%81%e6%b5%81%e5%aa%92%e4%bd%93",
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
