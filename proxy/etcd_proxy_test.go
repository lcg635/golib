package proxy

import (
	"log"
	"testing"
	"time"
)

type TestRes struct {
	Port string `json:"port"`
}

func TestEtcdProxy(t *testing.T) {
	opt := DefaultHealthCheckOption()
	proxy, err := NewHttpProxyBaseOnEtcd("test", "/gateways/genetv-test", 2*time.Second, opt)
	if err != nil {
		t.Error(err)
		return
	}

	tt := 0
	timer := time.NewTimer(1 * time.Second)
	for {
		<-timer.C
		tt = tt + 1
		res := new(TestRes)
		err := proxy.Get(res, "/", nil)
		if err != nil {
			log.Println(err)
		}
		t.Log(res)
		if tt == 1000 {
			return
		}
		timer.Reset(1 * time.Second)
	}
}
