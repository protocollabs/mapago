package main

import "fmt"
import "flag"
import "errors"
import "os"
import "sync"
import "time"
import "github.com/monfron/mapago/control-plane/ctrl/client-protocols"
import "github.com/monfron/mapago/measurement-plane/tcp-throughput"
import "github.com/monfron/mapago/control-plane/ctrl/shared"


var CTRL_PORT = 64321
var DEF_BUFFER_SIZE = 8096 * 8
// TODO: param via cli to define path to config
var CONFIG_FILE = "conf.json"
// TODO: do this maybe as a config param
var MSMT_INFO_INTERVAL = 2
var MSMT_STOP_INTERVAL = 10
// content "MID":"msmt_type"
var msmtIdStorage map[string]string
var mapInited = false

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
		runTcpCtrlClient(*ctrlAddrPtr, *portPtr, *callSizePtr, *msmtTypePtr)
	} else if *ctrlProtoPtr == "udp" {
		runUdpCtrlClient(*ctrlAddrPtr, *portPtr, *callSizePtr, *msmtTypePtr)
	} else if *ctrlProtoPtr == "udp_mcast" {
		runUdpMcastCtrlClient(*ctrlAddrPtr, *portPtr, *callSizePtr, *msmtTypePtr)
	} else {
		panic("tcp, udp or udp_mcast as ctrl-proto")
	}
}

func runTcpCtrlClient(addr string, port int, callSize int, msmtType string) {
	tcpObj := clientProtos.NewTcpObj("TcpDiscoveryConn", addr, port, callSize)

	// TODO: build json "dummy" message
	reqDataObj := new(shared.DataObj)
	reqDataObj.Type = shared.INFO_REQUEST
	reqDataObj.Id = shared.ConstructId()
	reqDataObj.Seq = "0"
	reqDataObj.Ts = shared.ConvCurrDateToStr()
	reqDataObj.Secret = "fancySecret"
	reqJson := shared.ConvDataStructToJson(reqDataObj)
	// debug fmt.Printf("\nrequest JSON is: % s", reqJson)

	repDataObj := tcpObj.StartDiscovery(reqJson)
	fmt.Println("\nClient received Info_request: ", repDataObj)

	err := validateDiscovery(reqDataObj, repDataObj)
	if err != nil {
		fmt.Printf("TCP Discovery phase failed: %s\n", err)
		os.Exit(1)
	}

	if msmtType == "tcp-throughput" {
		sendTcpMsmtStartRequest(addr, port, callSize)
	} else if msmtType == "udp-throughput" {
		sendUdpMsmtStartRequest(addr, port, callSize)
	} else {
		panic("Measurement type not supported")
	}
}

// this starts the TCP throughput measurement
// underlying control channel is TCP based
func sendTcpMsmtStartRequest(addr string, port int, callSize int) {
	var wg sync.WaitGroup
	tcpObj := clientProtos.NewTcpObj("TcpThroughputMsmtConn", addr, port, callSize)

	// TODO: build json "dummy" message
	reqDataObj := new(shared.DataObj)
	reqDataObj.Type = shared.MEASUREMENT_START_REQUEST
	reqDataObj.Id = shared.ConstructId()
	reqDataObj.Seq = "1"
	reqDataObj.Secret = "fancySecret"
	reqDataObj.Measurement_delay = "666"
	reqDataObj.Measurement_time_max = "666"

	msmtObj := constructMeasurementObj("tcp-throughput", "module")
	reqDataObj.Measurement = *msmtObj

	reqJson := shared.ConvDataStructToJson(reqDataObj)
	// debug fmt.Printf("\nrequest JSON is: % s", reqJson)

	repDataObj := tcpObj.StartMeasurement(reqJson)
	fmt.Println("\n\nClient received (TCP) Measurement_Start_reply: ", repDataObj)

	if mapInited == false {
		msmtIdStorage = make(map[string]string)
		mapInited = true
	}

	msmtIdStorage[repDataObj.Measurement_id] = "tcp-throughput"

	// debug fmt.Println("\nWE ARE NOW READY TO START WITH THE TCP MSMT")

	tcpThroughput.NewTcpMsmtClient(msmtObj.Configuration, &wg)

	fmt.Println("\n\n---------- TCP MSMT is now running ---------- ")

	// TODO: Move that (when everything is rdy) to a separate func
	// TODO: we dont need a goroutine here imo: we dont execute anything else
	go func() {
		for {
			select {
			case <-tMsmtInfoReq.C:
				fmt.Println("\nIts time to send a Msmt_Info_Req")
				/*
				TODO: implement Msmt_info_req logic
				*/

				tMsmtInfoReq.Reset(time.Duration(MSMT_INFO_INTERVAL) * time.Second)

			case <-tMsmtStopReq.C:
				fmt.Println("\nIts time to finish the measurement! close all conns")
			
				// stop the msmt_info_timer: we do not need anymore values
				timerStopped := tMsmtInfoReq.Stop()
				if timerStopped {
					fmt.Printf("\nCall stopped msmt_info_req Timer!")
				} else {
					fmt.Printf("\nmsmt_info_req Timer already expired!")
				}

				/*
				TODO: implement Msmt_stop_req logic
				*/
			}
		}
	}()

	wg.Wait()
	fmt.Println("\nTcp workers are now finished")
}

// this starts the UDP throughput measurement
// underlying control channel is TCP based
func sendUdpMsmtStartRequest(addr string, port int, callSize int) {
	tcpObj := clientProtos.NewTcpObj("UdpThroughputMsmtConn", addr, port, callSize)

	// TODO: build json "dummy" message
	reqDataObj := new(shared.DataObj)
	reqDataObj.Type = shared.MEASUREMENT_START_REQUEST
	reqDataObj.Id = shared.ConstructId()
	reqDataObj.Seq = "1"
	reqDataObj.Secret = "fancySecret"
	reqDataObj.Measurement_delay = "666"
	reqDataObj.Measurement_time_max = "666"

	msmtObj := constructMeasurementObj("udp-throughput", "module")
	reqDataObj.Measurement = *msmtObj

	reqJson := shared.ConvDataStructToJson(reqDataObj)
	// debug fmt.Printf("\nrequest JSON is: % s", reqJson)

	repDataObj := tcpObj.StartMeasurement(reqJson)
	fmt.Println("\n\nClient received (UDP) Measurement_Start_reply: ", repDataObj)

	if mapInited == false {
		msmtIdStorage = make(map[string]string)
		mapInited = true
	}

	msmtIdStorage[repDataObj.Measurement_id] = "udp-throughput"
	fmt.Println("\nWE ARE NOW READY TO START WITH THE UDP MSMT")

	// TODO: UdpThroughput call

}

func constructMeasurementObj(name string, msmtType string) *shared.MeasurementObj {
	MsmtObj := new(shared.MeasurementObj)
	MsmtObj.Name = name
	MsmtObj.Type = msmtType

	confObj := shared.ConstructConfiguration(CONFIG_FILE)
	MsmtObj.Configuration = *confObj
	return MsmtObj
}

func runUdpCtrlClient(addr string, port int, callSize int, msmtType string) {
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

	if msmtType == "tcp-throughput" {
		sendTcpMsmtStartRequest(addr, port, callSize)
	} else if msmtType == "udp-throughput" {
		sendUdpMsmtStartRequest(addr, port, callSize)
	} else {
		panic("Measurement type not supported")
	}
}

func runUdpMcastCtrlClient(addr string, port int, callSize int, msmtType string) {
	udpMcObj := clientProtos.NewUdpMcObj("UdpMcConn1", addr, port, callSize)

	// TODO: build json "dummy" message
	reqDataObj := new(shared.DataObj)
	reqDataObj.Type = shared.INFO_REQUEST
	reqDataObj.Id = shared.ConstructId()
	reqDataObj.Seq = "0"
	reqDataObj.Ts = shared.ConvCurrDateToStr()
	reqDataObj.Secret = "fancySecret"

	reqJson := shared.ConvDataStructToJson(reqDataObj)
	repDataObj := udpMcObj.Start(reqJson)

	err := validateDiscovery(reqDataObj, repDataObj)
	if err != nil {
		fmt.Printf("UDP MC Discovery phase failed: %s\n", err)
		os.Exit(1)
	}

	if msmtType == "tcp-throughput" {
		sendTcpMsmtStartRequest(addr, port, callSize)
	} else if msmtType == "udp-throughput" {
		sendUdpMsmtStartRequest(addr, port, callSize)
	} else {
		panic("Measurement type not supported")
	}
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

