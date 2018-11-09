package main

import "fmt"
import "flag"
import "os"
import "github.com/monfron/mapago/ctrl/clientProtos"
import "github.com/monfron/mapago/ctrl/shared"
import "errors"

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
		runTcpClient(*ctrlAddrPtr, *portPtr, *callSizePtr)
	} else if *ctrlProtoPtr == "udp" {
		runUdpClient(*ctrlAddrPtr, *portPtr, *callSizePtr)
	} else if *ctrlProtoPtr == "udp_mcast" {
		runUdpMcastClient(*ctrlAddrPtr, *portPtr, *callSizePtr)
	} else {
		panic("tcp, udp or udp_mcast as ctrl-proto")
	}
}

func runTcpClient(addr string, port int, callSize int) {
	// TODO: we need a channel here aswell in the future
	// use case: we receive a server response. using the server response
	// we can determine what next to do. i.e. info rep => do msmt start req etc.
	tcpObj := clientProtos.NewTcpObj("TcpConn1", addr, port, callSize)

	// TODO: build json "dummy" message
	reqDataObj := new(shared.DataObj)
	reqDataObj.Type = shared.INFO_REQUEST
	reqDataObj.Id = shared.ConstructId()
	reqDataObj.Seq = "0"
	reqDataObj.Ts = shared.ConvCurrDateToStr()
	reqDataObj.Secret = "fancySecret"

	reqJson := shared.ConvDataStructToJson(reqDataObj)
	// debug fmt.Printf("\nrequest JSON is: % s", reqJson)

	// Note: A better naming would be StartDiscoveryPhase()
	repDataObj := tcpObj.Start(reqJson)

	err := validateDiscovery(reqDataObj, repDataObj)
	if err != nil {
		fmt.Printf("TCP Discovery phase failed: %s\n", err)
		os.Exit(1)
	}
	// NEXT STEP Start Measurement
}

func runUdpClient(addr string, port int, callSize int) {
	udpObj := clientProtos.NewUdpObj("UdpConn1", addr, port, callSize)

	// TODO: build json "dummy" message
	reqDataObj := new(shared.DataObj)
	reqDataObj.Type = shared.INFO_REQUEST
	reqDataObj.Id = shared.ConstructId()
	reqDataObj.Seq = "0"
	reqDataObj.Ts = shared.ConvCurrDateToStr()
	reqDataObj.Secret = "fancySecret"

	reqJson := shared.ConvDataStructToJson(reqDataObj)
	repDataObj := udpObj.Start(reqJson)

	err := validateDiscovery(reqDataObj, repDataObj)
	if err != nil {
		fmt.Printf("UDP Discovery phase failed: %s\n", err)
		os.Exit(1)
	}
}

func runUdpMcastClient(addr string, port int, callSize int) {
	fmt.Println("DUMMY udp mcast module called")
}

func validateDiscovery(req *shared.DataObj, rep *shared.DataObj) error {
	if rep.Type != shared.INFO_REPLY {
		return errors.New("Received message is not INFO_REPLY")
	}

	if rep.Seq_rp != req.Seq {
		return errors.New("Wrong INFO_REQUEST handled by srv")
	}

	fmt.Println("\nDiscovery phase finished. Connected to: ", rep.Id)
	return nil
}
