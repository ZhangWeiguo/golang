package easyserver

import (
	"fmt"
	"net"
	"runtime"
	"strconv"
	"sync"
	"time"
)

type EasyTcpServer struct {
	TType         TcpType
	Port          int
	Threads       int
	WriteBuffer   int
	ReadBuffer    int
	Timeout       time.Duration
	ReadTimeout   time.Duration
	WriteTimeout  time.Duration
	KeepAliveTime time.Duration
	Responser     func([]byte) []byte
	Logger        func(string)
	addr          *net.TCPAddr
	listener      *net.TCPListener
	wait          sync.WaitGroup
}

func (t *EasyTcpServer) Init() error {
	t.wait = sync.WaitGroup{}
	t.wait.Add(1)
	var err error
	if t.TType == TcpType("") {
		t.TType = TCP4
	}
	if t.Logger == nil {
		t.Logger = func(s string) {
			fmt.Println(s)
		}
	}
	if t.Responser == nil {
		t.Responser = func(s []byte) []byte {
			return []byte("OK")
		}
	}
	if t.Port <= 0 || t.Port > 65535 {
		t.Port = DEFAULT_PORT
	}
	if t.Threads <= 0 {
		t.Threads = runtime.NumCPU()
	}
	if t.WriteBuffer < MIN_WRITE_BUFFER {
		t.WriteBuffer = DEFAULT_WRITE_BUFFER
	}
	if t.ReadBuffer < MIN_READ_BUFFER {
		t.ReadBuffer = DEFAULT_READ_BUFFER
	}
	t.addr, err = net.ResolveTCPAddr(string(t.TType), "0.0.0.0:"+strconv.Itoa(t.Port))
	if err == nil {
		t.listener, err = net.ListenTCP(string(t.TType), t.addr)
		if err == nil {
			for i := 0; i < t.Threads; i++ {
				go t.listen()
			}
			t.Logger(fmt.Sprintf("TCP Serve Start Succ At Port %d", t.Port))
			t.Logger(fmt.Sprintf("TCP Serve Start %d Threads", t.Threads))
		} else {
			t.Logger(fmt.Sprintf("TCP Serve Start Fail: %s", err.Error()))
		}
	} else {
		t.Logger(fmt.Sprintf("TCP Serve Start Fail: %s", err.Error()))
	}
	t.wait.Wait()
	return err
}

func (t *EasyTcpServer) Close() (err error) {
	err = t.listener.Close()
	return err
}

func (t *EasyTcpServer) readFromTcp(conn *net.TCPConn) ([]byte, error) {
	readdata := make([]byte, t.ReadBuffer)
	read, err := conn.Read(readdata)
	if err == nil {
		t.Logger(fmt.Sprintf("Read %d From %s Succ: %s", read, conn.RemoteAddr().String(), string(readdata[0:read])))
	} else {
		t.Logger(fmt.Sprintf("Read Fail: %s", err.Error()))
	}
	return readdata, err
}

func (t *EasyTcpServer) writeToTcp(conn *net.TCPConn, writedata []byte) {
	write, err := conn.Write(writedata)
	if err == nil {
		t.Logger(fmt.Sprintf("Write %d To %s Succ: %s", write, conn.RemoteAddr().String(), string(writedata[0:write])))
	} else {
		t.Logger(fmt.Sprintf("Write Fail: %s", err.Error()))
	}
}

func (t *EasyTcpServer) listen() {
	for {
		conn, err := t.listener.AcceptTCP()
		if err == nil {
			t.setTime(conn)
			go t.serve(conn)
		} else {
			t.Logger(fmt.Sprintf("Accept Conn Fail: %s", err.Error()))
		}
	}
}

func (t *EasyTcpServer) setTime(conn *net.TCPConn) {
	if t.KeepAliveTime > 0 {
		_ = conn.SetKeepAlivePeriod(t.KeepAliveTime)
	}
	if t.Timeout > 0 {
		_ = conn.SetDeadline(time.Now().Add(t.Timeout))
	}
	if t.WriteTimeout > 0 {
		_ = conn.SetReadDeadline(time.Now().Add(t.ReadTimeout))
	}
	if t.ReadTimeout > 0 {
		_ = conn.SetWriteDeadline(time.Now().Add(t.WriteTimeout))
	}
	if t.WriteBuffer > 0 {
		_ = conn.SetWriteBuffer(t.WriteBuffer)
	}
	if t.ReadBuffer > 0 {
		_ = conn.SetReadBuffer(t.ReadBuffer)
	}
}

func (t *EasyTcpServer) serve(conn *net.TCPConn) {
	_ = conn.SetWriteBuffer(t.WriteBuffer)
	_ = conn.SetReadBuffer(t.ReadBuffer)
	var readdata, writedata []byte
	var err error
	for {
		readdata, err = t.readFromTcp(conn)
		if err == nil {
			writedata = t.Responser(readdata)
			t.writeToTcp(conn, writedata)
		} else {
			t.Logger(fmt.Sprintf("Read From TCP Fail: %s", err.Error()))
			break
		}
	}
}
