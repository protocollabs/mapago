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
	portPtr := flag.Int("ctrl-port", CTRL_PORT, "port for interacting with control channel")
	callSizePtr := flag.Int("call-size", DEF_BUFFER_SIZE, "application buffer in bytes")
	msmtTypePtr := flag.String("msmt-type", "tcp-throughput", "tcp-throughput or udp-throughput")

	flag.Parse()

	fmt.Println("mapago(c) - 2018")
	fmt.Println("Client side")
	fmt.Println("Control protocol:", *ctrlProtoPtr)
	fmt.Println("Control addr:", *ctrlAddrPtr)
	fmt.Println("Control Port:", *portPtr)
	fmt.Println("Call-Size: ", *callSizePtr)
	fmt.Println("Msmt-type: ", *msmtTypePtr)


	if *ctrlProtoPtr == "tcp" {
		controlPlane.RunTcpCtrlClient(*ctrlAddrPtr, *portPtr, *callSizePtr, *msmtTypePtr)
	} else if *ctrlProtoPtr == "udp" {
		controlPlane.RunUdpCtrlClient(*ctrlAddrPtr, *portPtr, *callSizePtr, *msmtTypePtr)
	} else if *ctrlProtoPtr == "udp_mcast" {
		controlPlane.RunUdpMcastCtrlClient(*ctrlAddrPtr, *portPtr, *callSizePtr, *msmtTypePtr)
	} else {
		panic("tcp, udp or udp_mcast as ctrl-proto")
	}
}
