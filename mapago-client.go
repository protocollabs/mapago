package main

import "fmt"
import "flag"
import "errors"
import "os"
import "sync"
import "time"
import "strconv"
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
// Content "ID":"Hostname=UUID"
var idStorage map[string]string
var msmtStorageInited = false
var idStorageInited = false

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

	if idStorageInited == false {
		idStorage = make(map[string]string)
		idStorageInited = true
	}

	// TODO: Still some fields are HARDCODED
	reqDataObj := new(shared.DataObj)
	reqDataObj.Type = shared.INFO_REQUEST

	idStorage["tcp-id"] = shared.ConstructId()
	reqDataObj.Id = idStorage["tcp-id"]

	reqDataObj.Seq = "0"
	reqDataObj.Ts = shared.ConvCurrDateToStr()
	reqDataObj.Secret = "fancySecret"
	reqJson := shared.ConvDataStructToJson(reqDataObj)
	// debug fmt.Printf("\ninfo request JSON is: % s", reqJson)

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
	closeConnCh := make(chan string)
	tcpObj := clientProtos.NewTcpObj("TcpThroughputMsmtStartReqConn", addr, port, callSize)

	// TODO: Still some fields are HARDCODED
	reqDataObj := new(shared.DataObj)
	reqDataObj.Type = shared.MEASUREMENT_START_REQUEST
		
	if val, ok := idStorage["tcp-id"]; ok {
		reqDataObj.Id = val
	} else {
		fmt.Println("\nFound not the id")
	}
	
	reqDataObj.Seq = "1"
	reqDataObj.Secret = "fancySecret"
	reqDataObj.Measurement_delay = "666"
	reqDataObj.Measurement_time_max = "666"

	msmtObj := constructMeasurementObj("tcp-throughput", "module")
	reqDataObj.Measurement = *msmtObj
	
	numWorker, err := strconv.Atoi(reqDataObj.Measurement.Configuration.Worker)
	if err != nil {
		fmt.Printf("Could not parse Workers: %s\n", err)
		os.Exit(1)
	}
	

	reqJson := shared.ConvDataStructToJson(reqDataObj)
	// debug fmt.Printf("\nmsmt start request JSON is: % s", reqJson)

	repDataObj := tcpObj.StartMeasurement(reqJson)
	fmt.Println("\n\nClient received (TCP) Measurement_Start_reply: ", repDataObj)

	if msmtStorageInited == false {
		msmtIdStorage = make(map[string]string)
		msmtStorageInited = true
	}

	msmtIdStorage["tcp-throughput1"] = repDataObj.Measurement_id

	// debug fmt.Println("\nWE ARE NOW READY TO START WITH THE TCP MSMT")
	tcpThroughput.NewTcpMsmtClient(msmtObj.Configuration, &wg, closeConnCh)

	fmt.Println("\n\n---------- TCP MSMT is now running ---------- ")
	
	manageTcpMsmt(addr, port, callSize, &wg, closeConnCh, numWorker)
}


func manageTcpMsmt(addr string, port int, callSize int, wg *sync.WaitGroup, closeConnCh chan<- string, workers int) {
	tMsmtInfoReq := time.NewTimer(time.Duration(MSMT_INFO_INTERVAL) * time.Second)
	/* TODO: nicht zeitgetriggert sonder daten getriggert
	wenn letzte daten erhalten => close conns
	würde Probleme geben bei starker packet verlustrate
	wir sagen 10s und während dessen empfangen wir nichts
	dann bauen wir schon verbindung ab => client ist immer noch im retransmit
	*/
	tMsmtStopReq := time.NewTimer(time.Duration(MSMT_STOP_INTERVAL) * time.Second)

	for {
		select {
		case <-tMsmtInfoReq.C:
			fmt.Println("\nIts time to send a Msmt_Info_Req")
			/*
			- TODO: implement Msmt_info_req logic
			- i.e. construct Msmt_info_Req Message
			- send an wait for reply 
			*/
			tMsmtInfoReq.Reset(time.Duration(MSMT_INFO_INTERVAL) * time.Second)

		case <-tMsmtStopReq.C:
			fmt.Println("\nIts time to finish the measurement! close all conns")
		
			tMsmtInfoReq.Stop()
			sendTcpMsmtStopRequest(addr, port, callSize)

			for i := 0; i < workers; i++ {
				closeConnCh<- "close"
			}
			
			wg.Wait()
			fmt.Println("\nAll tcp workers are now finished")
			return
		}
	}
}


// this stops the TCP throughput measurement
// underlying control channel is TCP based
func sendTcpMsmtStopRequest(addr string, port int, callSize int) {
	tcpObj := clientProtos.NewTcpObj("TcpThroughputMsmtStopReqConn", addr, port, callSize)

	// TODO: build json "dummy" message
	reqDataObj := new(shared.DataObj)
	reqDataObj.Type = shared.MEASUREMENT_STOP_REQUEST
	
	if val, ok := idStorage["tcp-id"]; ok {
		reqDataObj.Id = val
	} else {
		fmt.Println("\nFound not the id")
	}
	
	reqDataObj.Seq = "2"
	reqDataObj.Secret = "fancySecret"

	if val, ok := msmtIdStorage["tcp-throughput1"]; ok {
		reqDataObj.Measurement_id = val
	} else {
		fmt.Println("\nFound not the measurement id for tcp throughput")
	}

	reqJson := shared.ConvDataStructToJson(reqDataObj)
	// debug fmt.Printf("\nmsmt stop request JSON is: % s", reqJson)

	repDataObj := tcpObj.StopMeasurement(reqJson)
	fmt.Println("\n\nClient received (TCP) Measurement_Stop_reply: ", repDataObj)
}


// this starts the UDP throughput measurement
// underlying control channel is TCP based
func sendUdpMsmtStartRequest(addr string, port int, callSize int) {
	tcpObj := clientProtos.NewTcpObj("UdpThroughputMsmtConn", addr, port, callSize)

	// TODO: Still some fields are HARDCODED
	reqDataObj := new(shared.DataObj)
	reqDataObj.Type = shared.MEASUREMENT_START_REQUEST
	
	if val, ok := idStorage["udp-id"]; ok {
		reqDataObj.Id = val
	} else {
		fmt.Println("\nFound not the id")
	}

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

	if msmtStorageInited == false {
		msmtIdStorage = make(map[string]string)
		msmtStorageInited = true
	}

	msmtIdStorage[repDataObj.Measurement_id] = "udp-throughput"
	fmt.Println("\nWE ARE NOW READY TO START WITH THE UDP MSMT")

	/* TODO: 
	- UdpThroughput call
	- UDP oop approach
	*/
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

	if idStorageInited == false {
		idStorage = make(map[string]string)
		idStorageInited = true
	}

	// TODO: Still some fields are HARDCODED
	reqDataObj := new(shared.DataObj)
	reqDataObj.Type = shared.INFO_REQUEST

	idStorage["udp-id"] = shared.ConstructId()
	reqDataObj.Id = idStorage["udp-id"]

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

	if idStorageInited == false {
		idStorage = make(map[string]string)
		idStorageInited = true
	}

	// TODO: build json "dummy" message
	reqDataObj := new(shared.DataObj)
	reqDataObj.Type = shared.INFO_REQUEST

	idStorage["mcast-id"] = shared.ConstructId()
	reqDataObj.Id = idStorage["mcast-id"]

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

