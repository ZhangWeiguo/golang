package easyserver

type TcpType string
type UdpType string

const (
	TCP  = TcpType("tcp")
	TCP4 = TcpType("tcp4")
	TCP6 = TcpType("tcp6")
	UDP  = UdpType("udp")
	UDP4 = UdpType("udp4")
	UDP6 = UdpType("udp6")
)
const (
	DEFAULT_PORT         = 8080
	MIN_WRITE_BUFFER     = 16
	MIN_READ_BUFFER      = 16
	DEFAULT_WRITE_BUFFER = 4028
	DEFAULT_READ_BUFFER  = 4028
	DEFAULT_TIMEOUT      = 300000000000 // NanoSecond
)
