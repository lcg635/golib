package log

import (
	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/negroni"
	"net/http"
	"sync"
	"time"
)

var (
	apiMonitor *ApiMonitor
)

type ApiMonitor struct {
	AppName  string
	Server   string
	Logger   *logrus.Logger
	Entries  map[string]*ApiMonitorEndpoint
	Interval time.Duration //定时上报的间隔
	mu       sync.Mutex
}

func NewApiMonitor(appName, server string, interval time.Duration) *ApiMonitor {
	if apiMonitor == nil {
		apiMonitor = &ApiMonitor{
			AppName:  appName,
			Server:   server,
			Logger:   logrus.New(),
			Interval: interval,
			Entries:  make(map[string]*ApiMonitorEndpoint),
		}
	}
	return apiMonitor
}

func (this *ApiMonitor) GetEndpoint(requestUri string) *ApiMonitorEndpoint {
	endpoint, ok := this.Entries[requestUri]

	if !ok {
		this.mu.Lock()
		endpoint = &ApiMonitorEndpoint{
			RequestURI: requestUri,
			Enable:     true,
			CallCount:  0,
			TotalTime:  0,
		}
		this.Entries[requestUri] = endpoint
		this.mu.Unlock()
	}
	return endpoint
}

func (this *ApiMonitor) ReportAll() {
	this.Logger.Infoln("**************************")
	for _, endpoint := range this.Entries {
		if endpoint.CallCount == 0 {
			continue
		}
		avgNanosecond := endpoint.TotalTime / endpoint.CallCount / 1000
		this.Logger.Infoln(endpoint.RequestURI, endpoint.CallCount, avgNanosecond, "ms")
	}
	this.Logger.Infoln("--------------------------")
}

func (this *ApiMonitor) LoopReport() {
	timer := time.NewTimer(this.Interval)
	go func() {
		for {
			timer.Reset(this.Interval)
			<-timer.C
			this.ReportAll()
		}
	}()
}

type ApiMonitorEndpoint struct {
	RequestURI string
	CallCount  int64
	TotalTime  int64
	Enable     bool
	mu         sync.Mutex
}

func (this *ApiMonitorEndpoint) AddCallCount(count int64, since int64) {
	this.mu.Lock()
	this.CallCount = this.CallCount + count
	this.TotalTime = this.TotalTime + since
	this.mu.Unlock()
}

/**
 * 监控api请求情况的日志中间件
 */
func NewApiMonitorLoggerMiddleware(appName, server string) negroni.HandlerFunc {
	apiMonitor := NewApiMonitor(appName, server, 60*time.Second)
	apiMonitor.LoopReport()

	return func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		start := time.Now()
		next(rw, r)
		apiMonitor.GetEndpoint(r.RequestURI).AddCallCount(1, int64(time.Since(start)))
	}
}

/*
type ApiMonitorHook struct {
	server string
}

func (this *ApiMonitorHook) Levels() []Level {

}

func (this *ApiMonitorHook) Fire(*Entry) error {

}

func NewApiMonitorLogger(name, server string) *logrus.Logger {
	hook := &ApiMonitorHook{server: server}

}*/
