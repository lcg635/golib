package comm

import (
	"github.com/rcrowley/goagain"
	"log"
	"net"
	"os"
	"syscall"
	"time"
)

/**
 * 服务器启动者
 */
type ServerStarter func(l net.Listener)

/**
 * 服务器关闭者
 */
type ServerExiter func(l net.Listener, sig syscall.Signal)

/**
 * 运行一个可以平滑重启的服务器
 */
func RunZeroDowntimeRestartServer(address string, serve ServerStarter, exiter ServerExiter) {
	l, err := goagain.Listener()

	if nil != err {
		l, err = net.Listen("tcp", address)
		if nil != err {
			log.Fatalln(err)
		}
		log.Println("listening on", l.Addr())
		go serve(l)
	} else {
		log.Println("resuming listening on", l.Addr())
		go serve(l)
		if err := goagain.Kill(); nil != err {
			log.Fatalln(err)
		}
	}

	sig, err := goagain.Wait(l)
	if nil != err {
		log.Println(sig, err)
	}
	exiter(l, sig)
}

/**
 * 一个简单的服务器关闭机制
 */
func DefaultGracefulExiter(timeout time.Duration, pidPath string) ServerExiter {
	return func(l net.Listener, sig syscall.Signal) {
		// In this case, we'll simply stop listening and wait one second.
		if err := l.Close(); nil != err {
			log.Fatalln(err)
		}

		log.Println(sig)
		if sig != syscall.SIGUSR2 && sig != syscall.SIGQUIT {
			_, err := os.Stat(pidPath)
			if err == nil {
				os.Remove(pidPath)
			}
		}

		time.Sleep(timeout)
	}
}
