package easyserver

import (
	"fmt"
	"net"
	"net/http"
	"runtime"
	"sync"
	"time"
)

type EasyHttpServe struct {
	Port            int
	Threads         int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	Logger          func(string)
	EnableKeepAlive bool
	server          *http.Server
	handler         *http.ServeMux
	tcpListener     *net.TCPListener
	wait            sync.WaitGroup
}

func (eht *EasyHttpServe) Init() (err error) {
	eht.wait = sync.WaitGroup{}
	eht.wait.Add(1)
	if eht.ReadTimeout <= 0 {
		eht.ReadTimeout = DEFAULT_TIMEOUT
	}
	if eht.WriteTimeout <= 0 {
		eht.WriteTimeout = DEFAULT_TIMEOUT
	}
	if eht.Port <= 0 || eht.Port > 65535 {
		eht.Port = DEFAULT_PORT
	}
	if eht.Logger == nil {
		eht.Logger = func(s string) {
			fmt.Println(s)
		}
	}
	if eht.Threads <= 0 {
		eht.Threads = runtime.NumCPU()
	}
	var addr = fmt.Sprintf("0.0.0.0:%d", eht.Port)
	var tcpAddr *net.TCPAddr
	eht.handler = http.NewServeMux()
	eht.server = &http.Server{
		Addr:         addr,
		ReadTimeout:  eht.ReadTimeout,
		WriteTimeout: eht.WriteTimeout,
	}
	if eht.EnableKeepAlive {
		eht.server.SetKeepAlivesEnabled(true)
	}
	tcpAddr, err = net.ResolveTCPAddr("tcp", addr)
	if err == nil {
		eht.tcpListener, err = net.ListenTCP("tcp", tcpAddr)
		eht.Logger(fmt.Sprintf("Http Tcp Listen Succ: %s", addr))
	} else {
		eht.Logger(fmt.Sprintf("Http Tcp Listen Error: %v", err))
	}
	return err
}

// ToDo: 自定义正则路由实现
// ToDo: 区分Get/Post/Put等方法
func (eht *EasyHttpServe) AddRouter(router string, response func(http.ResponseWriter, *http.Request)) {
	eht.handler.HandleFunc(router, response)
}

func (eht *EasyHttpServe) Serve() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	eht.server.Handler = eht.handler
	for i := 0; i < eht.Threads; i++ {
		go eht.server.Serve(eht.tcpListener)
		eht.Logger(fmt.Sprintf("Http Server Serve Succ In %d Thread", i))
	}
	eht.wait.Wait()
}

func (eht *EasyHttpServe) Close() {
	err := eht.tcpListener.Close()
	if err == nil {
		err = eht.server.Close()
	}
	if err != nil {
		eht.Logger(fmt.Sprintf("Http Server Stoped Error: %v", err))
	} else {
		eht.Logger("Http Server Stoped Succ")
	}
	eht.wait.Done()
}
