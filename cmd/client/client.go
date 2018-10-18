package main

import "fmt"
import "flag"
import "github.com/monfron/mapago/ctrl/clientProtos"

var CTRL_PORT = 64321
var DEF_BUFFER_SIZE = 8096 * 8

func main() {
	ctrlProtoPtr := flag.String("ctrl-protocol", "tcp", "tcp, udp or udp_mcast")
	ctrlAddrPtr := flag.String("ctrl-addr", "127.0.0.1", "localhost or userdefined addr" )
	portPtr := flag.Int("port", CTRL_PORT, "port for interacting with control channel")
	callSizePtr := flag.Int("call-size", DEF_BUFFER_SIZE, "application buffer in bytes")

	flag.Parse()

	fmt.Println("mapago(c) - 2018")
	fmt.Println("Client side")
	fmt.Println("Control protocol:", *ctrlProtoPtr)
	fmt.Println("Control addr:", *ctrlAddrPtr)
	fmt.Println("Port:", *portPtr)
	fmt.Println("Call-Size: ", *callSizePtr)

	if *ctrlProtoPtr == "tcp" {
		run_tcp_client(*ctrlAddrPtr, *portPtr, *callSizePtr)
	} else if *ctrlProtoPtr == "udp" {
		run_udp_client(*ctrlAddrPtr, *portPtr, *callSizePtr)
	} else if *ctrlProtoPtr == "udp_mcast" {
		run_udp_mcast_client(*ctrlAddrPtr, *portPtr, *callSizePtr)
	} else {
		panic("tcp, udp or udp_mcast as ctrl-proto")
	}
}

func run_tcp_client(addr string, port int, callSize int) {
	// TODO: we need a channel here aswell in the future
	// use case: we receive a server response. using the server response
	// we can determine what next to do. i.e. info rep => do msmt start req etc.
	tcpObj := clientProtos.NewTcpObj("TcpConn1", addr, port, callSize)
	tcpObj.Start()

	// TODO: wait for answer from server and then decide next communication steps
	// not necessary for toy server
}

func run_udp_client(addr string, port int, callSize int) {
	fmt.Println("DUMMY udp module called")
}

func run_udp_mcast_client(addr string, port int, callSize int) {
	fmt.Println("DUMMY udp mcast module called")
}