package easyserver

import (
	"net"
	"strconv"
)

type EasyUdpClient struct {
	UType  UdpType
	Host   string
	Port   int
	Logger func(string)
	conn   net.Conn
}

func (u *EasyUdpClient) Init() (err error) {
	u.conn, err = net.Dial(string(u.UType), u.Host+":"+strconv.Itoa(u.Port))
	if err != nil {
		u.Logger("UDP Client Conn Fail")
	} else {
		u.Logger("UDP CLient Conn Succ")
	}
	return err
}

func (u *EasyUdpClient) Close() (err error) {
	err = u.conn.Close()
	return err
}

func (u *EasyUdpClient) Send(msg []byte) (s []byte, err error) {
	var read int
	get := make([]byte, 4096)
	_, err = u.conn.Write(msg)
	if err == nil {
		u.Logger("Send Msg Succ: " + string(msg))
		read, err = u.conn.Read(get)
		s = get[0:read]
	} else {
		u.Logger("Send Msg Fail: " + err.Error())
	}
	return s, err
}
