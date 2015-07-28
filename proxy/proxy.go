package proxy

import (
	"errors"
	"log"
	"net/url"
	"sync"
	"time"
)

type Proxy interface {
	Post(interface{}, string, url.Values) error
	Get(interface{}, string, url.Values) error
	ReInitServers([]string)
	Stop()
}

type HttpProxy struct {
	sync.Mutex
	Name              string
	curServerIndex    int
	serverLength      int
	servers           []*HttpProxyServer
	timeout           time.Duration
	healthCheckOption *HealthCheckOption
	healthCheckTimer  *time.Timer
	stopC             chan bool
	stoped            bool
}

func NewHttpProxy(name string, servers []string, timeout time.Duration, healthCheckOption *HealthCheckOption) *HttpProxy {
	proxy := &HttpProxy{
		Name:              name,
		curServerIndex:    0,
		serverLength:      0,
		servers:           nil,
		timeout:           timeout,
		healthCheckOption: healthCheckOption,
		healthCheckTimer:  time.NewTimer(healthCheckOption.ReconnectInterval),
		stoped:            false,
		stopC:             make(chan bool),
	}

	proxy.ReInitServers(servers)
	go proxy.LoopHealthCheck()

	return proxy
}

/**
 * ROUND ROBIN 方式获取server
 */
func (this *HttpProxy) GetServer() (*HttpProxyServer, error) {
	if this.stoped == true {
		return nil, errors.New("proxy " + this.Name + " stoped")
	}

	var result *HttpProxyServer
	this.Lock()
	for i := 0; i < this.serverLength; i++ {
		server := this.servers[this.curServerIndex]
		this.curServerIndex = (this.curServerIndex + 1) % this.serverLength
		if server.failed == true {
			continue
		}
		result = server
		break
	}
	this.Unlock()

	if result == nil {
		return nil, errors.New("no available " + this.Name + " server")
	}

	return result, nil
}

func (this *HttpProxy) Post(result interface{}, api string, values url.Values) error {
	server, err := this.GetServer()
	if err != nil {
		return err
	}

	err = HttpPost(result, server.url+api, values, this.timeout)
	if err != nil {
		server.IncrErrorTimes()
		return err
	}

	return nil
}

func (this *HttpProxy) Get(result interface{}, api string, values url.Values) error {
	server, err := this.GetServer()
	if err != nil {
		return err
	}

	err = HttpGet(result, server.url+api, values, this.timeout)
	if err != nil {
		server.IncrErrorTimes()
		return err
	}

	return nil
}

func (this *HttpProxy) ReInitServers(servers []string) {
	httpProxyServers := []*HttpProxyServer{}
	for _, server := range servers {
		httpProxyServers = append(httpProxyServers, &HttpProxyServer{
			url:           server,
			errorTimes:    0,
			maxErrorTimes: this.healthCheckOption.MaxErrorTimes,
			failed:        false,
		})
	}
	this.Lock()
	this.serverLength = len(httpProxyServers)
	this.curServerIndex = 0
	this.servers = httpProxyServers
	this.Unlock()
}

func (this *HttpProxy) LoopHealthCheck() {
	for {
		select {
		case <-this.healthCheckTimer.C:
			this.Lock()
			for _, server := range this.servers {
				if server.failed == false {
					server.errorTimes = 0
					continue
				}
				result := &HttpProxyServerStatus{}
				err := HttpGet(result, server.url+this.healthCheckOption.HealCheckApi, nil, this.timeout)
				if err != nil {
					log.Println(err)
					continue
				}
				if result.status == "ok" {
					server.failed = false
				}
			}
			this.Unlock()
			this.healthCheckTimer.Reset(this.healthCheckOption.ReconnectInterval)
		case <-this.stopC:
			return
		}
	}
}

func (this *HttpProxy) Stop() {
	this.healthCheckTimer.Stop()
	this.Lock()
	this.stoped = true
	this.Unlock()
}

type HttpProxyServer struct {
	sync.Mutex
	url           string
	errorTimes    int
	maxErrorTimes int
	failed        bool
}

func (this *HttpProxyServer) IncrErrorTimes() {
	this.Lock()
	this.errorTimes = this.errorTimes + 1
	if this.errorTimes > this.maxErrorTimes {
		this.failed = true
	}
	this.Unlock()
}

type HttpProxyServerStatus struct {
	status string
}

type HealthCheckOption struct {
	ReconnectInterval time.Duration
	HealCheckApi      string
	MaxErrorTimes     int
}

func DefaultHealthCheckOption() *HealthCheckOption {
	return &HealthCheckOption{
		ReconnectInterval: 5 * time.Second,
		HealCheckApi:      "/status",
		MaxErrorTimes:     5,
	}
}
