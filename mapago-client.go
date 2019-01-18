package main

import "fmt"
import "flag"
import "errors"
import "os"
import "sync"
import "time"
import "strconv"
import "github.com/protocollabs/mapago/control-plane/ctrl/client-protocols"
import "github.com/protocollabs/mapago/measurement-plane/tcp-throughput"
import "github.com/protocollabs/mapago/measurement-plane/udp-throughput"
import "github.com/protocollabs/mapago/control-plane/ctrl/shared"

var CTRL_PORT = 64321
var DEF_BUFFER_SIZE = 8096 * 8
var CONFIG_FILE = "conf.json"
var MSMT_STREAMS = 1

var msmtIdStorage map[string]string
var idStorage map[string]string
var msmtStorageInited = false
var idStorageInited = false
var seqNo uint64
var msmtStreams *int 
var msmtListenAddr *string
var msmtCallSize *int
var msmtUpdateTime *uint
var msmtStopTime *uint

func main() {
	ctrlProto := flag.String("ctrl-protocol", "tcp", "tcp, udp or udp_mcast")
	ctrlAddr := flag.String("ctrl-addr", "127.0.0.1", "localhost or userdefined addr")
	port := flag.Int("ctrl-port", CTRL_PORT, "port for interacting with control channel")
	callSize := flag.Int("call-size", DEF_BUFFER_SIZE, "control application buffer in bytes")
	msmtType := flag.String("msmt-type", "tcp-throughput", "tcp-throughput or udp-throughput")
	msmtStreams = flag.Int("msmt-streams", MSMT_STREAMS, "setting number of streams")
	msmtListenAddr = flag.String("msmt-listen-addr", "127.0.0.1", "localhost or userdefined addr")
	msmtCallSize = flag.Int("msmt-call-size", DEF_BUFFER_SIZE, "msmt application buffer in bytes")
	msmtUpdateTime = flag.Uint("msmt-update-interval", 2, "msmt update interval in seconds")
	msmtStopTime = flag.Uint("msmt-stop-interval", 10, "complete msmt time in seconds")

	flag.Parse()

	/*
	fmt.Println("mapago(c) - 2018")
	fmt.Println("Client side")
	fmt.Println("Control protocol:", *ctrlProto)
	fmt.Println("Control addr:", *ctrlAddr)
	fmt.Println("Control Port:", *port)
	fmt.Println("Call-Size: ", *callSize)
	fmt.Println("Msmt-type: ", *msmtType)
	fmt.Println("Msmt-Streams: ", *msmtStreams)
	fmt.Println("Msmt-Addr: ", *msmtListenAddr)
	fmt.Println("Msmt-CallSize: ", *msmtCallSize)
	fmt.Println("Update-Interval: ", *msmtUpdateTime)
	fmt.Println("Msmt-time: ", *msmtStopTime)
	*/

	if *ctrlProto == "tcp" {
		runTcpCtrlClient(*ctrlAddr, *port, *callSize, *msmtType)
	} else if *ctrlProto == "udp" {
		runUdpCtrlClient(*ctrlAddr, *port, *callSize, *msmtType)
	} else if *ctrlProto == "udp_mcast" {
		runUdpMcastCtrlClient(*ctrlAddr, *port, *callSize, *msmtType)
	} else {
		panic("tcp, udp or udp_mcast as ctrl-proto")
	}
}

func runTcpCtrlClient(addr string, port int, callSize int, msmtType string) {
	tcpObj := clientProtos.NewTcpObj("TcpDiscoveryConn", addr, port, callSize)
	seqNo = shared.ConstructSeqNo()

	if idStorageInited == false {
		idStorage = make(map[string]string)
		idStorageInited = true
	}

	reqDataObj := new(shared.DataObj)
	reqDataObj.Type = shared.INFO_REQUEST

	_, ok := idStorage["host-uuid"]
	if ok != true {
		idStorage["host-uuid"] = shared.ConstructId()
	}

	reqDataObj.Id = idStorage["host-uuid"]

	reqDataObj.Seq = strconv.FormatUint(seqNo, 10)
	reqDataObj.Ts = shared.ConvCurrDateToStr()
	reqDataObj.Secret = "fancySecret"
	reqJson := shared.ConvDataStructToJson(reqDataObj)
	repDataObj := tcpObj.StartDiscovery(reqJson)

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

	reqDataObj := new(shared.DataObj)
	reqDataObj.Type = shared.MEASUREMENT_START_REQUEST
		
	if val, ok := idStorage["host-uuid"]; ok {
		reqDataObj.Id = val
	} else {
		fmt.Println("\nFound not the id")
	}
	
	seqNo++
	reqDataObj.Seq = strconv.FormatUint(seqNo, 10)
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

	if msmtStorageInited == false {
		msmtIdStorage = make(map[string]string)
		msmtStorageInited = true
	}

	msmtIdStorage["tcp-throughput1"] = repDataObj.Measurement_id


	tcpThroughput.NewTcpMsmtClient(msmtObj.Configuration, repDataObj, &wg, closeConnCh)

	
	manageTcpMsmt(addr, port, callSize, &wg, closeConnCh, numWorker)
}


func manageTcpMsmt(addr string, port int, callSize int, wg *sync.WaitGroup, closeConnCh chan<- string, workers int) {
	tMsmtInfoReq := time.NewTimer(time.Duration(*msmtUpdateTime) * time.Second)
	/* TODO: nicht zeitgetriggert sonder daten getriggert
	wenn letzte daten erhalten => close conns
	würde Probleme geben bei starker packet verlustrate
	wir sagen 10s und während dessen empfangen wir nichts
	dann bauen wir schon verbindung ab => client ist immer noch im retransmit
	*/
	tMsmtStopReq := time.NewTimer(time.Duration(*msmtStopTime) * time.Second)

	for {
		select {
		case <-tMsmtInfoReq.C:
			sendTcpMsmtInfoRequest(addr, port, callSize)
			tMsmtInfoReq.Reset(time.Duration(*msmtUpdateTime) * time.Second)

		case <-tMsmtStopReq.C:
			tMsmtInfoReq.Stop()

			for i := 0; i < workers; i++ {
				closeConnCh<- "close"
			}

			wg.Wait()
			
			// all connections are now terminated: server should shut down aswell
			sendTcpMsmtStopRequest(addr, port, callSize)

			return
		}
	}
}

func sendTcpMsmtInfoRequest(addr string, port int, callSize int) {
	tcpObj := clientProtos.NewTcpObj("TcpThroughputMsmtInfoReqConn", addr, port, callSize)

	reqDataObj := new(shared.DataObj)
	reqDataObj.Type = shared.MEASUREMENT_INFO_REQUEST
	
	if val, ok := idStorage["host-uuid"]; ok {
		reqDataObj.Id = val
	} else {
		fmt.Println("\nFound not the id")
	}
	
	seqNo++
	reqDataObj.Seq = strconv.FormatUint(seqNo, 10)
	reqDataObj.Secret = "fancySecret"

	if val, ok := msmtIdStorage["tcp-throughput1"]; ok {
		reqDataObj.Measurement_id = val
	} else {
		fmt.Println("\nFound not the measurement id for tcp throughput")
	}

	reqJson := shared.ConvDataStructToJson(reqDataObj)
	// debug fmt.Printf("\nmsmt stop request JSON is: % s", reqJson)

	msmtInfoRep := tcpObj.GetMeasurementInfo(reqJson)
	fmt.Println(msmtInfoRep)
}


// this stops the TCP throughput measurement
// underlying control channel is TCP based
func sendTcpMsmtStopRequest(addr string, port int, callSize int) {
	tcpObj := clientProtos.NewTcpObj("TcpThroughputMsmtStopReqConn", addr, port, callSize)

	reqDataObj := new(shared.DataObj)
	reqDataObj.Type = shared.MEASUREMENT_STOP_REQUEST
	
	if val, ok := idStorage["host-uuid"]; ok {
		reqDataObj.Id = val
	} else {
		fmt.Println("\nFound not the id")
	}
	
	seqNo++
	reqDataObj.Seq = strconv.FormatUint(seqNo, 10)
	reqDataObj.Secret = "fancySecret"

	if val, ok := msmtIdStorage["tcp-throughput1"]; ok {
		reqDataObj.Measurement_id = val
	} else {
		fmt.Println("\nFound not the measurement id for tcp throughput")
	}

	reqJson := shared.ConvDataStructToJson(reqDataObj)
	// debug fmt.Printf("\nmsmt stop request JSON is: % s", reqJson)

	tcpObj.StopMeasurement(reqJson)
}


// this starts the UDP throughput measurement
// underlying control channel is TCP based
func sendUdpMsmtStartRequest(addr string, port int, callSize int) {
	var wg sync.WaitGroup
	closeConnCh := make(chan string)
	tcpObj := clientProtos.NewTcpObj("UdpThroughputMsmtConn", addr, port, callSize)

	reqDataObj := new(shared.DataObj)
	reqDataObj.Type = shared.MEASUREMENT_START_REQUEST
	
	if val, ok := idStorage["host-uuid"]; ok {
		reqDataObj.Id = val
	} else {
		fmt.Println("\nFound not the id")
	}

	seqNo++
	reqDataObj.Seq = strconv.FormatUint(seqNo, 10)
	reqDataObj.Secret = "fancySecret"
	reqDataObj.Measurement_delay = "666"
	reqDataObj.Measurement_time_max = "666"

	msmtObj := constructMeasurementObj("udp-throughput", "module")
	reqDataObj.Measurement = *msmtObj

	numWorker, err := strconv.Atoi(reqDataObj.Measurement.Configuration.Worker)
	if err != nil {
		fmt.Printf("Could not parse Workers: %s\n", err)
		os.Exit(1)
	}

	reqJson := shared.ConvDataStructToJson(reqDataObj)

	repDataObj := tcpObj.StartMeasurement(reqJson)

	if msmtStorageInited == false {
		msmtIdStorage = make(map[string]string)
		msmtStorageInited = true
	}

	msmtIdStorage["udp-throughput1"] = repDataObj.Measurement_id

	udpThroughput.NewUdpMsmtClient(msmtObj.Configuration, repDataObj, &wg, closeConnCh)

	manageUdpMsmt(addr, port, callSize, &wg, closeConnCh, numWorker)
}


func manageUdpMsmt(addr string, port int, callSize int, wg *sync.WaitGroup, closeConnCh chan<- string, workers int) {
	tMsmtInfoReq := time.NewTimer(time.Duration(*msmtUpdateTime) * time.Second)
	/* TODO: nicht zeitgetriggert sonder daten getriggert
	wenn letzte daten erhalten => close conns
	würde Probleme geben bei starker packet verlustrate
	wir sagen 10s und während dessen empfangen wir nichts
	dann bauen wir schon verbindung ab => client ist immer noch im retransmit
	*/
	tMsmtStopReq := time.NewTimer(time.Duration(*msmtStopTime) * time.Second)

	for {
		select {
		case <-tMsmtInfoReq.C:
			sendUdpMsmtInfoRequest(addr, port, callSize)

			tMsmtInfoReq.Reset(time.Duration(*msmtUpdateTime) * time.Second)

		case <-tMsmtStopReq.C:
			tMsmtInfoReq.Stop()

			// NOTED: optional we could first send a msmt stop request
			// wait until the server sockets are down
			// and then close our own
			// sendUdpMsmtStopRequest(addr, port, callSize)

			for i := 0; i < workers; i++ {
				closeConnCh<- "close"
			}

			wg.Wait()
			sendUdpMsmtStopRequest(addr, port, callSize)
			return
		}
	}

}

func sendUdpMsmtInfoRequest(addr string, port int, callSize int) {
	tcpObj := clientProtos.NewTcpObj("UdpThroughputMsmtInfoReqConn", addr, port, callSize)

	reqDataObj := new(shared.DataObj)
	reqDataObj.Type = shared.MEASUREMENT_INFO_REQUEST
	
	if val, ok := idStorage["host-uuid"]; ok {
		reqDataObj.Id = val
	} else {
		fmt.Println("\nFound not the id")
	}
	
	seqNo++
	reqDataObj.Seq = strconv.FormatUint(seqNo, 10)
	reqDataObj.Secret = "fancySecret"

	if val, ok := msmtIdStorage["udp-throughput1"]; ok {
		reqDataObj.Measurement_id = val
	} else {
		fmt.Println("\nFound not the measurement id for udp throughput")
	}

	reqJson := shared.ConvDataStructToJson(reqDataObj)
	msmtInfoRep := tcpObj.GetMeasurementInfo(reqJson)
	fmt.Println(msmtInfoRep)
}

func sendUdpMsmtStopRequest(addr string, port int, callSize int) {
	tcpObj := clientProtos.NewTcpObj("UdpThroughputMsmtStopReqConn", addr, port, callSize)

	reqDataObj := new(shared.DataObj)
	reqDataObj.Type = shared.MEASUREMENT_STOP_REQUEST
	
	if val, ok := idStorage["host-uuid"]; ok {
		reqDataObj.Id = val
	} else {
		fmt.Println("\nFound not the id")
	}
	
	seqNo++
	reqDataObj.Seq = strconv.FormatUint(seqNo, 10)
	reqDataObj.Secret = "fancySecret"

	if val, ok := msmtIdStorage["udp-throughput1"]; ok {
		reqDataObj.Measurement_id = val
	} else {
		fmt.Println("\nFound not the measurement id for udp throughput")
	}

	reqJson := shared.ConvDataStructToJson(reqDataObj)
	tcpObj.StopMeasurement(reqJson)
}

func constructMeasurementObj(name string, msmtType string) *shared.MeasurementObj {
	MsmtObj := new(shared.MeasurementObj)
	MsmtObj.Name = name
	MsmtObj.Type = msmtType

	// we dont need a configuration anymore
//	confObj := shared.ConstructConfiguration(CONFIG_FILE)
	confObj := new(shared.ConfigurationObj)
	confObj.Worker = strconv.Itoa(*msmtStreams)
	confObj.Listen_addr = *msmtListenAddr
	confObj.Call_size = strconv.Itoa(*msmtCallSize)


	MsmtObj.Configuration = *confObj
	return MsmtObj
}

func runUdpCtrlClient(addr string, port int, callSize int, msmtType string) {
	udpObj := clientProtos.NewUdpObj("UdpConn1", addr, port, callSize)
	seqNo = shared.ConstructSeqNo()

	if idStorageInited == false {
		idStorage = make(map[string]string)
		idStorageInited = true
	}

	reqDataObj := new(shared.DataObj)
	reqDataObj.Type = shared.INFO_REQUEST

	_, ok := idStorage["host-uuid"]
	if ok != true {
		idStorage["host-uuid"] = shared.ConstructId()
	}

	reqDataObj.Id = idStorage["host-uuid"]

	reqDataObj.Seq = strconv.FormatUint(seqNo, 10)
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
	seqNo = shared.ConstructSeqNo()

	if idStorageInited == false {
		idStorage = make(map[string]string)
		idStorageInited = true
	}

	reqDataObj := new(shared.DataObj)
	reqDataObj.Type = shared.INFO_REQUEST

	idStorage["mcast-id"] = shared.ConstructId()
	reqDataObj.Id = idStorage["mcast-id"]

	reqDataObj.Seq = strconv.FormatUint(seqNo, 10)
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

	return nil
}

