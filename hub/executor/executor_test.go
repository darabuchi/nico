package executor

import (
	"testing"

	"github.com/darabuchi/log"
)

func TestExecutor(t *testing.T) {
	log.SetLevel(log.InfoLevel)
	ex := NewExecutor()
	err := ex.Listen("7891")
	if err != nil {
		log.Errorf("err:%v", err)
		return
	}

	err = ex.AddNodeByV2rayLink("trojan://E73EAC23-83C8-AC21-E491-A6F3FC7C1FE0@shdata1.ourdvsss.xyz:12122?security=tls&sni=douyincdn.com&type=tcp&headerType=none#%f0%9f%87%ba%f0%9f%87%b8+%e7%be%8e%e5%9b%bd02")
	if err != nil {
		log.Errorf("err:%v", err)
		return
	}

	select {}
}
