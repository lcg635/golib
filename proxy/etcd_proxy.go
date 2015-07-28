package proxy

import (
	"koogroup/lib/log"
	"time"

	"github.com/coreos/go-etcd/etcd"
)

var etcdProxyLogger = log.DefaultLogger().WithField("tag", "etcd_proxy")

type HttpProxyBaseOnEtcd struct {
	*HttpProxy
	confChangeReceiver chan *etcd.Response
	etcdClient         *etcd.Client
	etcdKey            string
}

func NewHttpProxyBaseOnEtcd(etcdClient *etcd.Client, name, etcdKey string, timeout time.Duration, opt *HealthCheckOption) (*HttpProxyBaseOnEtcd, error) {
	res, err := etcdClient.Get(etcdKey, false, false)
	if err != nil {
		return nil, err
	}

	servers := FetchServersFromEtcd(res)
	httpProxy := NewHttpProxy(name, servers, timeout, opt)

	proxy := &HttpProxyBaseOnEtcd{
		HttpProxy:          httpProxy,
		confChangeReceiver: make(chan *etcd.Response),
		etcdClient:         etcdClient,
		etcdKey:            etcdKey,
	}

	//循环监控是否配置有变化
	go proxy.WatchServersConfig()
	//接收配置变化并更新代理的服务器信息
	go proxy.LoopHandleServersConfigChanged()

	return proxy, nil
}

//检查服务器配置是否变更
func (this *HttpProxyBaseOnEtcd) WatchServersConfig() {
	_, err := this.etcdClient.Watch(this.etcdKey, 0, true, this.confChangeReceiver, this.stopC)
	if err != nil {
		etcdProxyLogger.Error("etcd watch ", this.etcdKey, err)
	}
}

//服务器配置变更时更新服务器信息
func (this *HttpProxyBaseOnEtcd) LoopHandleServersConfigChanged() {
	for {
		select {
		case response := <-this.confChangeReceiver:
			etcdProxyLogger.Info(response)
			res, err := this.etcdClient.Get(this.etcdKey, false, false)
			if err == nil {
				servers := FetchServersFromEtcd(res)
				etcdProxyLogger.Info("proxy ", this.Name, " reinit :", servers)
				this.ReInitServers(servers)
			} else {
				etcdProxyLogger.Error("get proxy servers ", err.Error())
			}
		case <-this.stopC:
			return
		}
	}
}

//停止服务器
func (this *HttpProxyBaseOnEtcd) Stop() {
	this.HttpProxy.Stop()
	this.stopC <- true
}

func FetchServersFromEtcd(response *etcd.Response) []string {
	servers := []string{}
	if response.Node.Dir == true {
		tmp := make(map[string]string)
		for _, node := range response.Node.Nodes {
			tmp[node.Value] = node.Value
		}
		for _, v := range tmp {
			servers = append(servers, v)
		}
	} else {
		servers = append(servers, response.Node.Value)
	}
	return servers
}
