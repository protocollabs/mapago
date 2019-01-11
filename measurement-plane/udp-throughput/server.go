package udpThroughput

import "fmt"
import "os"
import "net"
import "sync"
import "strconv"
import "github.com/monfron/mapago/control-plane/ctrl/shared"

var UPDATE_INTERVAL = 5

type UdpThroughputMsmt struct {
	numStreams       int
	usedPorts        []int
	callSize         int
	listenAddr       string
	msmtId           string
	byteStorage      map[string]uint64
	byteStorageMutex sync.RWMutex
	fTsStorage       map[string]string
	fTsStorageMutex  sync.RWMutex
	lTsStorage       map[string]string
	lTsStorageMutex  sync.RWMutex

	/*
		- this attribute can be used by start() to RECEIVE cmd from managementplane
	*/
	msmtMgmt2MsmtCh <-chan shared.ChMgmt2Msmt

	/*
		- this attribute can be used by constructor and start() to SEND msgs to control plane
	*/
	msmt2CtrlCh chan<- shared.ChMsmt2Ctrl

	/*
		- channel for closing connection
	*/
	closeConnCh chan interface{}
}

func NewUdpThroughputMsmt(msmtCh <-chan shared.ChMgmt2Msmt, ctrlCh chan<- shared.ChMsmt2Ctrl, msmtStartReq *shared.DataObj, msmtId string, startPort int) *UdpThroughputMsmt {
	var err error
	var msmtData map[string]string
	heartbeatCh := make(chan bool)
	closeCh := make(chan interface{})
	udpMsmt := new(UdpThroughputMsmt)

	udpMsmt.byteStorage = make(map[string]uint64)
	udpMsmt.fTsStorage = make(map[string]string)
	udpMsmt.lTsStorage = make(map[string]string)

	fmt.Println("\nClient UDP request is: ", msmtStartReq)

	udpMsmt.numStreams, err = strconv.Atoi(msmtStartReq.Measurement.Configuration.Worker)
	if err != nil {
		fmt.Printf("\nCannot convert worker value: %s", err)
		os.Exit(1)
	}

	udpMsmt.callSize, err = strconv.Atoi(msmtStartReq.Measurement.Configuration.Call_size)
	if err != nil {
		fmt.Printf("\nCannot convert callSize value: %s", err)
		os.Exit(1)
	}

	// TODO: this should be the id sent to client
	udpMsmt.msmtId = msmtId
	udpMsmt.listenAddr = msmtStartReq.Measurement.Configuration.Listen_addr
	udpMsmt.msmtMgmt2MsmtCh = msmtCh
	udpMsmt.msmt2CtrlCh = ctrlCh
	udpMsmt.closeConnCh = closeCh

	for c := 1; c <= udpMsmt.numStreams; c++ {
		go udpMsmt.udpServerWorker(closeCh, heartbeatCh, startPort, c)
	}

	for c := 1; c <= udpMsmt.numStreams; c++ {
		heartbeat := <-heartbeatCh
		if heartbeat != true {
			panic("udp_server_worker goroutine not ok")
		}
	}

	fmt.Println("\n\nPorts used by udp module: ", udpMsmt.usedPorts)

	msmtReply := new(shared.ChMsmt2Ctrl)
	msmtReply.Status = "ok"

	msmtData = make(map[string]string)
	msmtData["msmtId"] = udpMsmt.msmtId
	msmtData["msg"] = "all modules running"
	msmtData["usedPorts"] = shared.ConvIntSliceToStr(udpMsmt.usedPorts)

	msmtReply.Data = msmtData

	/*
		NOTE: I dont get why exactly we need to perform this as a goroutine
		- we use unbuffered channels so chan<- and <-chan should wait for each other
	*/
	go func() {
		ctrlCh <- *msmtReply
	}()

	return udpMsmt
}

func (udpMsmt *UdpThroughputMsmt) udpServerWorker(closeCh <-chan interface{}, goHeartbeatCh chan<- bool, port int, streamIndex int) {
	var udpConn *net.UDPConn
	fTsExists := false
	cltAddrExists := false
	stream := "stream" + strconv.Itoa(streamIndex)
	fmt.Printf("\n%s is here", stream)

	for {
		listen := udpMsmt.listenAddr + ":" + strconv.Itoa(port)

		udpAddr, error := net.ResolveUDPAddr("udp", listen)
		if error != nil {
			fmt.Printf("Cannot parse \"%s\": %s\n", listen, error)
			goHeartbeatCh <- false
			os.Exit(1)
		}

		udpConn, error = net.ListenUDP("udp", udpAddr)
		if error == nil {
			// debug fmt.Printf("\nCan listen on addr: %s\n", listen)
			udpMsmt.usedPorts = append(udpMsmt.usedPorts, port)
			goHeartbeatCh <- true
			break
		}

		// debug fmt.Printf("\nCannot listen on addr: %s\n", listen)
		port++
	}

	// TODO maybe that does not work at that point
	// fmt.Printf("Connection from %s\n", udpConn.RemoteAddr())
	message := make([]byte, udpMsmt.callSize, udpMsmt.callSize)

	for {
		select {
		case data := <-closeCh:
			cmd, ok := data.(string)
			if ok == false {
				fmt.Printf("Type assertion failed: Looking for string %t", ok)
				os.Exit(1)
			}

			if cmd != "close" {
				fmt.Printf("Wrong cmd: Looking for close cmd")
				os.Exit(1)
			}
			// debug:
			fmt.Println("\nClosing udpConn!")
			udpConn.Close()
			return
		default:
			bytes, cltAddr, error := udpConn.ReadFromUDP(message)
			if error != nil {
				fmt.Printf("Cannot read: %s\n", error)
				os.Exit(1)
			}

			if cltAddrExists == false {
				fmt.Println("\nConnection from: ", cltAddr)
				cltAddrExists = true
			}

			udpMsmt.writeByteStorage(stream, uint64(bytes))

			if fTsExists == false {
				fTs := shared.ConvCurrDateToStr()
				udpMsmt.writefTsStorage(stream, fTs)
				fTsExists = true
			}

			lTs := shared.ConvCurrDateToStr()
			udpMsmt.writelTsStorage(stream, lTs)
		}
	}
}

func (udpMsmt *UdpThroughputMsmt) writefTsStorage(stream string, ts string) {
	udpMsmt.fTsStorageMutex.Lock()
	udpMsmt.fTsStorage[stream] = ts
	udpMsmt.fTsStorageMutex.Unlock()
}

func (udpMsmt *UdpThroughputMsmt) readfTsStorage(stream string) string {
	udpMsmt.fTsStorageMutex.RLock()
	ts := udpMsmt.fTsStorage[stream]
	udpMsmt.fTsStorageMutex.RUnlock()

	return ts
}

func (udpMsmt *UdpThroughputMsmt) writelTsStorage(stream string, ts string) {
	udpMsmt.lTsStorageMutex.Lock()
	udpMsmt.lTsStorage[stream] = ts
	udpMsmt.lTsStorageMutex.Unlock()
}

func (udpMsmt *UdpThroughputMsmt) readlTsStorage(stream string) string {
	udpMsmt.lTsStorageMutex.RLock()
	ts := udpMsmt.lTsStorage[stream]
	udpMsmt.lTsStorageMutex.RUnlock()

	return ts
}

func (udpMsmt *UdpThroughputMsmt) writeByteStorage(stream string, bytes uint64) {
	udpMsmt.byteStorageMutex.Lock()
	udpMsmt.byteStorage[stream] = udpMsmt.byteStorage[stream] + bytes
	udpMsmt.byteStorageMutex.Unlock()
}

// we can precise which key to address
func (udpMsmt *UdpThroughputMsmt) readByteStorage(stream string) uint64 {
	udpMsmt.byteStorageMutex.RLock()
	bytes := udpMsmt.byteStorage[stream]
	udpMsmt.byteStorageMutex.RUnlock()

	return bytes
}

func (udpMsmt *UdpThroughputMsmt) CloseConn() {
	var msmtData map[string]string

	for i := 0; i < udpMsmt.numStreams; i++ {
		udpMsmt.closeConnCh <- "close"
	}

	msmtReply := new(shared.ChMsmt2Ctrl)
	msmtReply.Status = "ok"

	msmtData = make(map[string]string)
	msmtData["msmtId"] = udpMsmt.msmtId
	msmtData["msg"] = "all modules closed"
	msmtReply.Data = msmtData

	// TODO: we have to attach the final Measurement Data
	go func() {
		udpMsmt.msmt2CtrlCh <- *msmtReply
	}()
}

func (udpMsmt *UdpThroughputMsmt) GetMsmtInfo() {
	var msmtData []shared.DataResultObj

	msmtReply := new(shared.ChMsmt2Ctrl)
	msmtReply.Status = "ok"

	// prepare msmtReply.Data
	for c := 1; c <= udpMsmt.numStreams; c++ {
		stream := "stream" + strconv.Itoa(c)
		bytes := udpMsmt.readByteStorage(stream)

		// create single data element:
		// i.e. read stream bytes, timestamp calculation
		dataElement := new(shared.DataResultObj)
		dataElement.Received_bytes = strconv.Itoa(int(bytes))
		dataElement.Timestamp_first = udpMsmt.readfTsStorage(stream)
		dataElement.Timestamp_last = udpMsmt.readlTsStorage(stream)

		// add single DataElement to slice
		msmtData = append(msmtData, *dataElement)
		// debug fmt.Println("\nmsmtData is: ", msmtData)
	}

	msmtReply.Data = msmtData

	go func() {
		udpMsmt.msmt2CtrlCh <- *msmtReply
	}()
}
