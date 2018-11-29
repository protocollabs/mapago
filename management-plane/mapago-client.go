package main

import "fmt"
import "flag"
// import "github.com/monfron/mapago/controlPlane/cmd/client"
import "github.com/monfron/mapago/control-plane"


var CTRL_PORT = 64321
var DEF_BUFFER_SIZE = 8096 * 8

func main() {
	ctrlProtoPtr := flag.String("ctrl-protocol", "tcp", "tcp, udp or udp_mcast")
	ctrlAddrPtr := flag.String("ctrl-addr", "127.0.0.1", "localhost or userdefined addr")
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
		controlPlane.RunTcpClient(*ctrlAddrPtr, *portPtr, *callSizePtr)
	} else if *ctrlProtoPtr == "udp" {
		controlPlane.RunUdpClient(*ctrlAddrPtr, *portPtr, *callSizePtr)
	} else if *ctrlProtoPtr == "udp_mcast" {
		controlPlane.RunUdpMcastClient(*ctrlAddrPtr, *portPtr, *callSizePtr)
	} else {
		panic("tcp, udp or udp_mcast as ctrl-proto")
	}
}
