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

	err = ex.AddNodeByV2rayLink("trojan://nykzeZ-virza6-kagmyf@teoeu.com:443#ShareCentre")
	if err != nil {
		log.Errorf("err:%v", err)
		return
	}

	select {}
}
