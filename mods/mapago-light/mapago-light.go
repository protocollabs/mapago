package main 

import "fmt"
import "flag"
import "net"
import "reflect"

var CTRL_PORT = 64321

// interface

type IfAnswer interface {
	writeMsg([]byte) error
}

// classes

type TcpObj struct {
	connName string
	connObj *net.TCPConn
}

type UdpObj struct {
	connName string
	connObj *net.UDPConn
}

// struct for channel
type Result struct {
	json []byte
	connObj interface{}
}


// Implement interface 

// connects interface with class TcpObj and implement interface 
func (tcp *TcpObj) writeMsg(msg []byte) error {
	_, err := tcp.connObj.Write(msg)
	return err
} 

// connects interface with class UdpObj and implement interface 
func (udp *UdpObj) writeMsg(msg []byte) error {
	_, err := udp.connObj.Write(msg)
	return err
} 


// constructors
func NewTcpObj(ch chan<- Result, name string) *TcpObj {
	tcp := new(TcpObj)
	tcp.connName = name

	return tcp
}


func NewUdpObj(ch chan<- Result, name string) *UdpObj {
	udp := new(UdpObj)
	udp.connName = name
		
	return udp
}


func main() {
	modePtr := flag.String("mode", "server", "server or client")
	portPtr := flag.Int("port", CTRL_PORT, "port for interacting with control channel")
	protoPtr := flag.String("protocol", "udp_uc", "protocol for discovering mapago servers: udp_uc, udp_mc or tcp")
	addrPtr := flag.String("addr", "ipv4_uc", "target addr: ipv4_uc, ipv4_mc, ipv6_uc, ipv6_mc") 

	flag.Parse()

	fmt.Println("mapago-light(c) - 2018")
	fmt.Println("Mode:", *modePtr)
	fmt.Println("Port:", *portPtr)
	fmt.Println("Protocol:", *protoPtr)
	fmt.Println("Address:", *addrPtr)

	if *modePtr == "server" {
		run_server(*portPtr)
	} else if *modePtr == "client" {
		run_client(*portPtr, *addrPtr, *protoPtr)
	} else {
		panic("mode not supported! chose server or client")
	}
}

func run_server(port int) {
	fmt.Println("server handler dummy func")

	ch := make(chan Result)
	fmt.Println("ch is of type: ", reflect.TypeOf(ch))
}


func run_client(port int, addr string, proto string) {
	fmt.Println("client handler dummy func")
}