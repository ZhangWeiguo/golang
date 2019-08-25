package easyserver

import (
	"fmt"
	"net"
	"runtime"
	"sync"
)

type EasyUdpServer struct {
	UType       UdpType
	Port        int
	Threads     int
	Responser   func([]byte) []byte
	Logger      func(string)
	WriteBuffer int
	ReadBuffer  int
	listener    *net.UDPConn
	wait        sync.WaitGroup
}

func (u *EasyUdpServer) Init() error {
	u.wait = sync.WaitGroup{}
	u.wait.Add(1)
	if u.Logger == nil {
		u.Logger = func(s string) {
			fmt.Println(s)
		}
	}
	if u.Responser == nil {
		u.Responser = func(s []byte) []byte {
			return []byte("OK")
		}
	}
	if u.Port <= 0 || u.Port > 65535 {
		u.Port = DEFAULT_PORT
	}
	if u.Threads <= 0 {
		u.Threads = runtime.NumCPU()
	}
	if u.WriteBuffer < MIN_WRITE_BUFFER && u.WriteBuffer >= 0 {
		u.WriteBuffer = DEFAULT_WRITE_BUFFER
	}
	if u.ReadBuffer < MIN_READ_BUFFER && u.ReadBuffer >= 0 {
		u.ReadBuffer = DEFAULT_READ_BUFFER
	}
	var err error
	u.listener, err = getUdpListener(string(u.UType), u.Port)
	if err == nil {
		if u.ReadBuffer >= 0 {
			_ = u.listener.SetReadBuffer(u.ReadBuffer)
		}
		if u.WriteBuffer >= 0 {
			_ = u.listener.SetWriteBuffer(u.WriteBuffer)
		}
		u.Logger(fmt.Sprintf("UDP Serve Start Succ At Port %d", u.Port))
		u.Logger(fmt.Sprintf("UDP Serve Start %d Threads", u.Threads))
		for i := 0; i < u.Threads; i++ {
			go u.listen()
		}
	} else {
		u.Logger(fmt.Sprintf("UDP Serve Start Fail: %s", err.Error()))
	}
	u.wait.Wait()
	return err
}

func (u *EasyUdpServer) listen() {
	for {
		readdata, remoteAddr, err := u.readFromUdp()
		if err == nil {
			go u.serve(readdata, remoteAddr)
		}
	}
}

func (u *EasyUdpServer) serve(readdata []byte, remoteAddr *net.UDPAddr) {
	senddata := u.Responser(readdata)
	u.writeToUdp(remoteAddr, senddata)
}

func (u *EasyUdpServer) readFromUdp() ([]byte, *net.UDPAddr, error) {
	var msg string
	readdata := make([]byte, u.ReadBuffer)
	read, remoteAddr, err := u.listener.ReadFromUDP(readdata)
	if err == nil {
		msg = fmt.Sprintf("Read %d From %s Succ: %s", read, udpaddr2str(remoteAddr), string(readdata[0:read]))
		readdata = readdata[0:read]
	} else {
		msg = fmt.Sprintf("Read Fail: %s", err.Error())
	}

	u.Logger(msg)
	return readdata, remoteAddr, err
}

func (u *EasyUdpServer) writeToUdp(remoteAddr *net.UDPAddr, send []byte) {
	var msg string
	write, err := u.listener.WriteToUDP(send, remoteAddr)
	if err == nil {
		msg = fmt.Sprintf("Write %d To %s SUcc: %s", write, udpaddr2str(remoteAddr), string(send[0:write]))
	} else {
		msg = fmt.Sprintf("Write Fail: %s", err.Error())
	}
	u.Logger(msg)
}

func (u *EasyUdpServer) Close() {
	u.listener.Close()
}

func getUdpListener(proto string, port int) (*net.UDPConn, error) {
	listener, err := net.ListenUDP(proto, &net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: port,
	})
	return listener, err
}

func udpaddr2str(add *net.UDPAddr) (s string) {
	return fmt.Sprintf("%v:%v", add.IP, add.Port)
}
