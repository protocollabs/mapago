package main

import "fmt"
import "flag"
import "errors"
import "os"
import "sync"
import "time"
import "strconv"
import "strings"
import "math"
import "github.com/protocollabs/mapago/control-plane/ctrl/client-protocols"
import "github.com/protocollabs/mapago/measurement-plane/tcp-throughput"
import "github.com/protocollabs/mapago/measurement-plane/tcp-tls-throughput"
import "github.com/protocollabs/mapago/measurement-plane/udp-throughput"
import "github.com/protocollabs/mapago/measurement-plane/quic-throughput"
import "github.com/protocollabs/mapago/control-plane/ctrl/shared"

var CTRL_PORT = 64321
var DEF_BUFFER_SIZE = 8096 * 8
var BYTE_COUNT = 1 * uint(math.Pow(10, 6))
var CONFIG_FILE = "conf.json"
var MSMT_STREAMS = 1

var msmtIdStorage map[string]string
var idStorage map[string]string
var msmtStorageInited = false
var idStorageInited = false
var seqNo uint64
var streams *int
var serverAddr *string
var bufLength *int
var msmtUpdateTime *uint
var msmtTime *uint
var msmtTotalBytes *uint
var msmtDeadline *uint

func main() {
	ctrlProto := flag.String("control-protocol", "tcp", "tcp, udp or udp_mcast")
	ctrlAddr := flag.String("control-addr", "127.0.0.1", "localhost or userdefined addr")
	port := flag.Int("control-port", CTRL_PORT, "port for interacting with control channel")
	callSize := flag.Int("control-buffer-length", DEF_BUFFER_SIZE, "control application buffer in bytes")
	module := flag.String("module", "tcp-throughput", "tcp-throughput, tcp-tls-throughput, udp-throughput, quic-throughput")
	streams = flag.Int("streams", MSMT_STREAMS, "setting number of streams")
	serverAddr = flag.String("addr", "127.0.0.1", "localhost or userdefined addr")
	bufLength = flag.Int("buffer-length", DEF_BUFFER_SIZE, "msmt application buffer in bytes")
	msmtUpdateTime = flag.Uint("update-interval", 2, "msmt update interval in seconds")
	/* This should be removed when byte-termination for TCP-throughput works
	and UDP and QUIC can be adapted
	*/
	msmtTime = flag.Uint("time", 10, "complete msmt time in seconds")
	msmtTotalBytes = flag.Uint("bytes", BYTE_COUNT, "number of bytes sent by client")
	msmtDeadline = flag.Uint("deadline", 300, "deadline when msmt is regarded as failed")

	flag.Parse()

	/*
		fmt.Println("mapago(c) - 2018")
		fmt.Println("Client side")
		fmt.Println("Control protocol:", *ctrlProto)
		fmt.Println("Control addr:", *ctrlAddr)
		fmt.Println("Control Port:", *port)
		fmt.Println("Call-Size: ", *callSize)
		fmt.Println("module: ", *module)
		fmt.Println("Msmt-Streams: ", *streams)
		fmt.Println("Msmt-Addr: ", *serverAddr)
		fmt.Println("Msmt-CallSize: ", *bufLength)
		fmt.Println("Update-Interval: ", *msmtUpdateTime)
		fmt.Println("Msmt-time: ", *msmtTime)
	*/

	if *ctrlProto == "tcp" {
		runTcpCtrlClient(*ctrlAddr, *port, *callSize, *module)
	} else if *ctrlProto == "udp" {
		runUdpCtrlClient(*ctrlAddr, *port, *callSize, *module)
	} else if *ctrlProto == "udp_mcast" {
		runUdpMcastCtrlClient(*ctrlAddr, *port, *callSize, *module)
	} else {
		panic("tcp, udp or udp_mcast as ctrl-proto")
	}
}

func runTcpCtrlClient(addr string, port int, callSize int, module string) {
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

	if module == "tcp-throughput" {
		sendTcpMsmtStartRequest(addr, port, callSize)
	} else if module == "tcp-tls-throughput" {
		sendTcpTlsMsmtStartRequest(addr, port, callSize)
	} else if module == "udp-throughput" {
		sendUdpMsmtStartRequest(addr, port, callSize)
	} else if module == "quic-throughput" {
		sendQuicMsmtStartRequest(addr, port, callSize)
	} else {
		panic("Measurement type not supported")
	}
}

func sendTcpTlsMsmtStartRequest(addr string, port int, callSize int) {
	var wg sync.WaitGroup
	closeConnCh := make(chan string)
	tcpObj := clientProtos.NewTcpObj("TcpTlsMsmtStartReqConn", addr, port, callSize)

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

	// CHECK THIS
	msmtObj := constructMeasurementObj("tcp-tls-throughput", "module")
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

	msmtIdStorage["tcp-tls-throughput1"] = repDataObj.Measurement_id
	tcpTlsThroughput.NewTcpTlsMsmtClient(msmtObj.Configuration, repDataObj, &wg, closeConnCh, *bufLength)
	manageTcpTlsMsmt(addr, port, callSize, &wg, closeConnCh, numWorker)
}

func manageTcpTlsMsmt(addr string, port int, callSize int, wg *sync.WaitGroup, closeConnCh chan<- string, workers int) {
	tMsmtInfoReq := time.NewTimer(time.Duration(*msmtUpdateTime) * time.Second)
	/* TODO: nicht zeitgetriggert sonder daten getriggert
	wenn letzte daten erhalten => close conns
	würde Probleme geben bei starker packet verlustrate
	wir sagen 10s und während dessen empfangen wir nichts
	dann bauen wir schon verbindung ab => client ist immer noch im retransmit
	*/
	tMsmtStopReq := time.NewTimer(time.Duration(*msmtTime) * time.Second)

	for {
		select {
		case <-tMsmtInfoReq.C:
			sendTcpTlsMsmtInfoRequest(addr, port, callSize)
			tMsmtInfoReq.Reset(time.Duration(*msmtUpdateTime) * time.Second)

		case <-tMsmtStopReq.C:
			tMsmtInfoReq.Stop()

			for i := 0; i < workers; i++ {
				closeConnCh <- "close"
			}

			wg.Wait()

			// all connections are now terminated: server should shut down aswell
			sendTcpTlsMsmtStopRequest(addr, port, callSize)
			return
		}
	}
}

func sendTcpTlsMsmtInfoRequest(addr string, port int, callSize int) {
	tcpObj := clientProtos.NewTcpObj("TcpTlsMsmtInfoReqConn", addr, port, callSize)

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

	if val, ok := msmtIdStorage["tcp-tls-throughput1"]; ok {
		reqDataObj.Measurement_id = val
	} else {
		fmt.Println("\nFound not the measurement id for tcp tls throughput")
	}

	reqJson := shared.ConvDataStructToJson(reqDataObj)
	// debug fmt.Printf("\nmsmt stop request JSON is: % s", reqJson)

	msmtInfoRep := tcpObj.GetMeasurementInfo(reqJson)
	prepareOutput(msmtInfoRep)
}

// this stops the QUIC throughput measurement
// underlying control channel is TCP based
func sendTcpTlsMsmtStopRequest(addr string, port int, callSize int) {
	tcpObj := clientProtos.NewTcpObj("TcpTlsMsmtStopReqConn", addr, port, callSize)

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

	if val, ok := msmtIdStorage["tcp-tls-throughput1"]; ok {
		reqDataObj.Measurement_id = val
	} else {
		fmt.Println("\nFound not the measurement id for tcp tls throughput")
	}

	reqJson := shared.ConvDataStructToJson(reqDataObj)
	// debug fmt.Printf("\nmsmt stop request JSON is: % s", reqJson)

	msmtStopRep := tcpObj.StopMeasurement(reqJson)
	prepareOutput(msmtStopRep)
}

func sendQuicMsmtStartRequest(addr string, port int, callSize int) {
	var sentStreamBytes map[string]*uint
	var wg sync.WaitGroup
	closeConnCh := make(chan string)
	tcpObj := clientProtos.NewTcpObj("QuicMsmtStartReqConn", addr, port, callSize)

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

	// STOPPED
	msmtObj := constructMeasurementObj("quic-throughput", "module")
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

	msmtIdStorage["quic-throughput1"] = repDataObj.Measurement_id
	
	sentStreamBytes = make(map[string]*uint)
	numStreams, _ := strconv.Atoi(msmtObj.Configuration.Worker)
	for c := 1; c <= numStreams; c++ {
		stream := "stream" + strconv.Itoa(c)
		streamBytes := uint(0)
		sentStreamBytes[stream] = &streamBytes
	}

	quicThroughput.NewQuicMsmtClient(msmtObj.Configuration, repDataObj, &wg, closeConnCh, *bufLength, sentStreamBytes, *msmtTotalBytes)

	manageQuicMsmt(addr, port, callSize, &wg, closeConnCh, numWorker, sentStreamBytes)
}

func manageQuicMsmt(addr string, port int, callSize int, wg *sync.WaitGroup, closeConnCh chan<- string, workers int, sentStreamBytes map[string]*uint) {
	tMsmtInfoReq := time.NewTimer(time.Duration(*msmtUpdateTime) * time.Second)
	/* TODO: nicht zeitgetriggert sonder daten getriggert
	wenn letzte daten erhalten => close conns
	würde Probleme geben bei starker packet verlustrate
	wir sagen 10s und während dessen empfangen wir nichts
	dann bauen wir schon verbindung ab => client ist immer noch im retransmit
	*/
	tMsmtStopReq := time.NewTimer(time.Duration(*msmtTime) * time.Second)

	for {
		select {
		case <-tMsmtInfoReq.C:
			sendQuicMsmtInfoRequest(addr, port, callSize)
			tMsmtInfoReq.Reset(time.Duration(*msmtUpdateTime) * time.Second)

		case <-tMsmtStopReq.C:
			tMsmtInfoReq.Stop()

			for i := 0; i < workers; i++ {
				closeConnCh <- "close"
			}

			wg.Wait()

			// all connections are now terminated: server should shut down aswell
			sendQuicMsmtStopRequest(addr, port, callSize)
			return
		}
	}
}

func sendQuicMsmtInfoRequest(addr string, port int, callSize int) {
	tcpObj := clientProtos.NewTcpObj("QuicMsmtInfoReqConn", addr, port, callSize)

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

	if val, ok := msmtIdStorage["quic-throughput1"]; ok {
		reqDataObj.Measurement_id = val
	} else {
		fmt.Println("\nFound not the measurement id for quic throughput")
	}

	reqJson := shared.ConvDataStructToJson(reqDataObj)
	// debug fmt.Printf("\nmsmt stop request JSON is: % s", reqJson)

	msmtInfoRep := tcpObj.GetMeasurementInfo(reqJson)
	prepareOutput(msmtInfoRep)
}

// this stops the QUIC throughput measurement
// underlying control channel is TCP based
func sendQuicMsmtStopRequest(addr string, port int, callSize int) {
	tcpObj := clientProtos.NewTcpObj("QuicMsmtStopReqConn", addr, port, callSize)

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

	if val, ok := msmtIdStorage["quic-throughput1"]; ok {
		reqDataObj.Measurement_id = val
	} else {
		fmt.Println("\nFound not the measurement id for quic throughput")
	}

	reqJson := shared.ConvDataStructToJson(reqDataObj)
	// debug fmt.Printf("\nmsmt stop request JSON is: % s", reqJson)

	msmtStopRep := tcpObj.StopMeasurement(reqJson)
	prepareOutput(msmtStopRep)
}

// this starts the TCP throughput measurement
// underlying control channel is TCP based
func sendTcpMsmtStartRequest(addr string, port int, callSize int) {
	// bytes per stream
	var sentStreamBytes map[string]*uint
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

	sentStreamBytes = make(map[string]*uint)
	numStreams, _ := strconv.Atoi(msmtObj.Configuration.Worker)
	for c := 1; c <= numStreams; c++ {
		stream := "stream" + strconv.Itoa(c)
		streamBytes := uint(0)
		sentStreamBytes[stream] = &streamBytes
	}

	tcpThroughput.NewTcpMsmtClient(msmtObj.Configuration, repDataObj, &wg, closeConnCh, *bufLength, sentStreamBytes, *msmtTotalBytes)

	manageTcpMsmt(addr, port, callSize, &wg, closeConnCh, numWorker, sentStreamBytes)
}

func manageTcpMsmt(addr string, port int, callSize int, wg *sync.WaitGroup, closeConnCh chan<- string, workers int, sentStreamBytes map[string]*uint) {
	tMsmtInfoReq := time.NewTimer(time.Duration(*msmtUpdateTime) * time.Second)
	tDeadline := time.NewTimer(time.Duration(*msmtDeadline) * time.Second)

	var currSrvBytes uint
	var lastSrvBytes uint

	for {
		select {
		case <-tMsmtInfoReq.C:
			sendTcpMsmtInfoRequest(addr, port, callSize, &currSrvBytes)
			tMsmtInfoReq.Reset(time.Duration(*msmtUpdateTime) * time.Second)

		// TODO: Invalidate data sent to test-sequencer or it will plot
		case <-tDeadline.C:
			fmt.Println("\nDeadline fired")
			tDeadline.Stop()
			for i := 0; i < workers; i++ {
				closeConnCh <- "close"
			}

			wg.Wait()

			// all connections are now terminated: server should shut down aswell
			sendTcpMsmtStopRequest(addr, port, callSize)
			return

		default:
			// reset timer as long a) msmtTotalBytes not reached OR b) msmt info reply changes
			if doneSending(sentStreamBytes) == false || noUpdates(&lastSrvBytes, &currSrvBytes) == false {
				// fmt.Println("\nReseting timeout")
				tDeadline.Reset(time.Duration(*msmtDeadline) * time.Second)

			}

			if msmtFinished(&currSrvBytes) {
				tDeadline.Stop()

				for i := 0; i < workers; i++ {
					closeConnCh <- "close"
				}

				wg.Wait()

				// all connections are now terminated: server should shut down aswell
				sendTcpMsmtStopRequest(addr, port, callSize)
				return
			}
		}
	}
}

// purpose: checks if server received all bytes sent by client
// called by: manage(Tcp|Udp|Quic)Msmt
func msmtFinished(currSrvBytes *uint) bool {
	if *currSrvBytes >= *msmtTotalBytes {
		// debug fmt.Println("\nServer received everything from client: Msmt finished")
		return true
	}

	return false
}

// purpose: checks if sum of all "stream bytes" (i.e. bytes sent per stream) matches the parameter msmtTotalBytes => does clt sent all bytes?
// called by: manage(Tcp|Udp|Quic)Msmt
func doneSending(sentStreamBytes map[string]*uint) bool {
	var sumBytes uint
	done := false

	// iterate over streams
	for _, val := range sentStreamBytes {
		sumBytes += *val

		if sumBytes >= *msmtTotalBytes {
			done = true
			// debug fmt.Println("\nEnough bytes sent")
			break
		}
	}
	return done
}

// purpose: compare results from MsmtInfoRep => does byte count change? => influence deadline timer
// called by: manage(Tcp|Udp|Quic)Msmt
func noUpdates(lastSrvBytes *uint, currSrvBytes *uint) bool {
	if *currSrvBytes > *lastSrvBytes {
		// "new" threshold
		*lastSrvBytes = *currSrvBytes
		return false
	}

	return true
}

// purpose: counts bytes of all streams received by MsmstInfoRep => used for evaluation expectation value
// called by: send(Tcp|Udp|Quic)MsmtInfoRequest
func countCurrSrvBytes(msmtInfoRep *shared.DataObj) uint {
	var currSrvByte uint

	for _, dataElement := range msmtInfoRep.Data.DataElements {
		bytes, err := strconv.ParseUint(dataElement.Received_bytes, 10, 0)

		if err != nil {
			fmt.Printf("Could not parse bytes: %s\n", err)
			os.Exit(1)
		}

		currSrvByte += uint(bytes)
	}
	return currSrvByte
}

func sendTcpMsmtInfoRequest(addr string, port int, callSize int, currSrvBytes *uint) {
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
	*currSrvBytes = countCurrSrvBytes(msmtInfoRep)

	prepareOutput(msmtInfoRep)
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

	msmtStopRep := tcpObj.StopMeasurement(reqJson)
	prepareOutput(msmtStopRep)
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

	udpThroughput.NewUdpMsmtClient(msmtObj.Configuration, repDataObj, &wg, closeConnCh, *bufLength)

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
	tMsmtStopReq := time.NewTimer(time.Duration(*msmtTime) * time.Second)

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
				closeConnCh <- "close"
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
	prepareOutput(msmtInfoRep)
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
	msmtStopRep := tcpObj.StopMeasurement(reqJson)
	prepareOutput(msmtStopRep)
}

func constructMeasurementObj(name string, module string) *shared.MeasurementObj {
	MsmtObj := new(shared.MeasurementObj)
	MsmtObj.Name = name
	MsmtObj.Type = module

	// we dont need a configuration anymore
	//	confObj := shared.ConstructConfiguration(CONFIG_FILE)
	confObj := new(shared.ConfigurationObj)
	confObj.Worker = strconv.Itoa(*streams)
	confObj.Listen_addr = *serverAddr
	confObj.Call_size = strconv.Itoa(*bufLength)

	MsmtObj.Configuration = *confObj
	return MsmtObj
}

func runUdpCtrlClient(addr string, port int, callSize int, module string) {
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

	if module == "tcp-throughput" {
		sendTcpMsmtStartRequest(addr, port, callSize)
	} else if module == "tcp-tls-throughput" {
		sendTcpTlsMsmtStartRequest(addr, port, callSize)
	} else if module == "udp-throughput" {
		sendUdpMsmtStartRequest(addr, port, callSize)
	} else if module == "quic-throughput" {
		sendQuicMsmtStartRequest(addr, port, callSize)
	} else {
		panic("Measurement type not supported")
	}
}

func runUdpMcastCtrlClient(addr string, port int, callSize int, module string) {
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

	if module == "tcp-throughput" {
		sendTcpMsmtStartRequest(addr, port, callSize)
	} else if module == "tcp-tls-throughput" {
		sendTcpTlsMsmtStartRequest(addr, port, callSize)
	} else if module == "udp-throughput" {
		sendUdpMsmtStartRequest(addr, port, callSize)
	} else if module == "quic-throughput" {
		sendQuicMsmtStartRequest(addr, port, callSize)
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

func prepareOutput(msmtInfoRep *shared.DataObj) {
	var evaluationStr strings.Builder

	// 1. add prefix
	evaluationStr.WriteString("[ ")

	// 2. sweep through data elements
	for index, dataElement := range msmtInfoRep.Data.DataElements {
		// 3. write ts-start
		evaluationStr.WriteString("{ \"ts-start\" : \"")
		evaluationStr.WriteString(dataElement.Timestamp_first)
		evaluationStr.WriteString("\", ")

		// 4. write ts-end
		evaluationStr.WriteString("\"ts-end\" : \"")
		evaluationStr.WriteString(dataElement.Timestamp_last)
		evaluationStr.WriteString("\", ")

		// 5. write bytes
		evaluationStr.WriteString("\"bytes\" : \"")
		evaluationStr.WriteString(dataElement.Received_bytes)
		evaluationStr.WriteString("\"}")

		if index != len(msmtInfoRep.Data.DataElements)-1 {
			evaluationStr.WriteString(", ")
		}
	}

	// 6. add suffix
	evaluationStr.WriteString(" ]")

	fmt.Println(evaluationStr.String())
}
