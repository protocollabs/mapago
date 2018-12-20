package tcpThroughput

import "fmt"
import "os"
import "net"
import "strconv"
import "sync"
import "github.com/monfron/mapago/control-plane/ctrl/shared"

var UPDATE_INTERVAL = 5

type TcpMsmtObj struct {
	numStreams int
	// TODO: in the future the server should define which Port to use
	startPort  int
	callSize   int
	listenAddr string
	msmtId     string
	byteStorage map[string]uint64
	byteStorageMutex sync.RWMutex

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

func NewTcpMsmtObj(msmtCh <-chan shared.ChMgmt2Msmt, ctrlCh chan<- shared.ChMsmt2Ctrl, msmtStartReq *shared.DataObj, msmtId string) *TcpMsmtObj {
	var err error
	var msmtData map[string]string
	heartbeatCh := make(chan bool)
	closeCh := make(chan interface{})
	tcpMsmt := new(TcpMsmtObj)
	
	tcpMsmt.byteStorage = make(map[string]uint64)
	
	fmt.Println("\nClient request is: ", msmtStartReq)

	tcpMsmt.numStreams, err = strconv.Atoi(msmtStartReq.Measurement.Configuration.Worker)
	if err != nil {
		fmt.Printf("\nCannot convert worker value: %s", err)
		os.Exit(1)
	}

	tcpMsmt.startPort, err = strconv.Atoi(msmtStartReq.Measurement.Configuration.Port)
	if err != nil {
		fmt.Printf("\nCannot convert port value: %s", err)
		os.Exit(1)
	}

	tcpMsmt.callSize, err = strconv.Atoi(msmtStartReq.Measurement.Configuration.Call_size)
	if err != nil {
		fmt.Printf("\nCannot convert callSize value: %s", err)
		os.Exit(1)
	}

	// TODO: this should be the id sent to client
	tcpMsmt.msmtId = msmtId
	tcpMsmt.listenAddr = msmtStartReq.Measurement.Configuration.Listen_addr
	tcpMsmt.msmtMgmt2MsmtCh = msmtCh
	tcpMsmt.msmt2CtrlCh = ctrlCh
	tcpMsmt.closeConnCh = closeCh

	fmt.Printf("\n\nTCP Msmt: Client wants %d workers to be started from start port %d!", tcpMsmt.numStreams, tcpMsmt.startPort)

	for c := 1; c <= tcpMsmt.numStreams; c++ {
		fmt.Printf("\n\nStarting stream %d on port %d", c, tcpMsmt.startPort)
		go tcpMsmt.tcpServerWorker(closeCh, heartbeatCh, tcpMsmt.startPort, c)
		tcpMsmt.startPort++
	}

	for c := 1; c <= tcpMsmt.numStreams; c++ {
		heartbeat := <-heartbeatCh
		if heartbeat != true {
			panic("tcp_server_worker goroutine not ok")
		}
	}

	fmt.Println("\n\nGoroutines ok: We are save to send a reply!")

	// send reply to control plane
	msmtReply := new(shared.ChMsmt2Ctrl)
	msmtReply.Status = "ok"

	msmtData = make(map[string]string)
	msmtData["msmtId"] = tcpMsmt.msmtId
	msmtData["msg"] = "all modules running"
	msmtReply.Data = msmtData

	/*
		NOTE: I dont get why exactly we need to perform this as a goroutine
		- we use unbuffered channels so chan<- and <-chan should wait for each other
	*/
	go func() {
		ctrlCh <- *msmtReply
	}()

	return tcpMsmt
}

func (tcpMsmt *TcpMsmtObj) tcpServerWorker(closeCh <-chan interface{}, goHeartbeatCh chan<- bool, port int, streamIndex int) {
	stream := "stream" + strconv.Itoa(streamIndex)
	fmt.Printf("\n%s is here", stream)

	listen := tcpMsmt.listenAddr + ":" + strconv.Itoa(port)

	addr, error := net.ResolveTCPAddr("tcp", listen)
	if error != nil {
		fmt.Printf("Cannot parse \"%s\": %s\n", listen, error)
		goHeartbeatCh <- false
		os.Exit(1)
	}

	fmt.Println("\nlistening on addr : ", addr)

	listener, error := net.ListenTCP("tcp", addr)
	if error != nil {
		fmt.Printf("Cannot listen: %s\n", error)
		goHeartbeatCh <- false
		os.Exit(1)
	}

	goHeartbeatCh <- true

	conn, error := listener.AcceptTCP()
	if error != nil {
		fmt.Printf("Cannot accept: %s\n", error)
		os.Exit(1)
	}

	fmt.Printf("Connection from %s\n", conn.RemoteAddr())
	message := make([]byte, tcpMsmt.callSize, tcpMsmt.callSize)

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
			listener.Close()
			conn.Close()
			return
		default:
			bytes, error := conn.Read(message)
			if error != nil {
				fmt.Printf("Cannot read: %s\n", error)
				os.Exit(1)
			}

			tcpMsmt.writeByteStorage(stream, uint64(bytes))

		}
	}
}

func (tcpMsmt *TcpMsmtObj) writeByteStorage(stream string, bytes uint64) {
	tcpMsmt.byteStorageMutex.Lock()
	tcpMsmt.byteStorage[stream] = tcpMsmt.byteStorage[stream] + bytes
	tcpMsmt.byteStorageMutex.Unlock()
}

// we can precise which key to address
func (tcpMsmt *TcpMsmtObj) readByteStorage(stream string) uint64{
	tcpMsmt.byteStorageMutex.RLock()
	bytes := tcpMsmt.byteStorage[stream]
	tcpMsmt.byteStorageMutex.RUnlock()

	return bytes
}


func (tcpMsmt *TcpMsmtObj) CloseConn() {
	var msmtData map[string]string

	for i := 0; i < tcpMsmt.numStreams; i++ {
		tcpMsmt.closeConnCh <- "close"
	}

	msmtReply := new(shared.ChMsmt2Ctrl)
	msmtReply.Status = "ok"

	msmtData = make(map[string]string)
	msmtData["msmtId"] = tcpMsmt.msmtId
	msmtData["msg"] = "all modules closed"
	msmtReply.Data = msmtData

	// TODO: we have to attach the final Measurement Data
	go func() {
		tcpMsmt.msmt2CtrlCh <- *msmtReply
	}()
}

func (tcpMsmt *TcpMsmtObj) GetMsmtInfo() {
	var msmtData []shared.DataResultObj

	msmtReply := new(shared.ChMsmt2Ctrl)
	msmtReply.Status = "ok"

	// prepare msmtReply.Data
	for c := 1; c <= tcpMsmt.numStreams; c++ {
		stream := "stream" + strconv.Itoa(c)
		bytes := tcpMsmt.readByteStorage(stream)

		// create single data element: 
		// i.e. read stream bytes, timestamp calculation
		dataElement := new(shared.DataResultObj)
		dataElement.Received_bytes = strconv.Itoa(int(bytes))

		// these are dummy values
		dataElement.Timestamp_first = shared.ConvCurrDateToStr()
		dataElement.Timestamp_last = shared.ConvCurrDateToStr()

		// add single DataElement to slice
		msmtData = append(msmtData, *dataElement)
		fmt.Println("\nmsmtData is: ", msmtData)
	}

	msmtReply.Data = msmtData

	go func() {
		tcpMsmt.msmt2CtrlCh <- *msmtReply
	}()
}
