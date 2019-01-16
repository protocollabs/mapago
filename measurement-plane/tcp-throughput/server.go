package tcpThroughput

import "fmt"
import "os"
import "net"
import "strconv"
import "sync"
import "github.com/protocollabs/mapago/control-plane/ctrl/shared"

var UPDATE_INTERVAL = 5

type TcpMsmtObj struct {
	numStreams          int
	usedPorts           []int
	callSize            int
	listenAddr          string
	msmtId              string
	byteStorage         map[string]uint64
	byteStorageMutex    sync.RWMutex
	fTsStorage          map[string]string
	fTsStorageMutex     sync.RWMutex
	lTsStorage          map[string]string
	lTsStorageMutex     sync.RWMutex
	tcpConnStorage      map[string]*net.TCPConn
	tcpConnStorageMutex sync.RWMutex

	listenerStorage      map[string]*net.TCPListener
	listenerStorageMutex sync.RWMutex

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

func NewTcpMsmtObj(msmtCh <-chan shared.ChMgmt2Msmt, ctrlCh chan<- shared.ChMsmt2Ctrl, msmtStartReq *shared.DataObj, msmtId string, startPort int) *TcpMsmtObj {
	var err error
	var msmtData map[string]string
	heartbeatCh := make(chan bool)
	closeCh := make(chan interface{})
	tcpMsmt := new(TcpMsmtObj)

	tcpMsmt.byteStorage = make(map[string]uint64)
	tcpMsmt.fTsStorage = make(map[string]string)
	tcpMsmt.lTsStorage = make(map[string]string)
	tcpMsmt.tcpConnStorage = make(map[string]*net.TCPConn)
	tcpMsmt.listenerStorage = make(map[string]*net.TCPListener)

	fmt.Println("\nClient TCP request is: ", msmtStartReq)

	tcpMsmt.numStreams, err = strconv.Atoi(msmtStartReq.Measurement.Configuration.Worker)
	if err != nil {
		fmt.Printf("\nCannot convert worker value: %s", err)
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

	for c := 1; c <= tcpMsmt.numStreams; c++ {
		go tcpMsmt.tcpServerWorker(closeCh, heartbeatCh, startPort, c)
	}

	for c := 1; c <= tcpMsmt.numStreams; c++ {
		heartbeat := <-heartbeatCh
		if heartbeat != true {
			panic("tcp_server_worker goroutine not ok")
		}
	}

	fmt.Println("\n\nPorts used by TCP module: ", tcpMsmt.usedPorts)

	msmtReply := new(shared.ChMsmt2Ctrl)
	msmtReply.Status = "ok"

	msmtData = make(map[string]string)
	msmtData["msmtId"] = tcpMsmt.msmtId
	msmtData["msg"] = "all modules running"
	msmtData["usedPorts"] = shared.ConvIntSliceToStr(tcpMsmt.usedPorts)

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
	var listener *net.TCPListener
	fTsExists := false
	stream := "stream" + strconv.Itoa(streamIndex)
	fmt.Printf("\n%s is here", stream)

	for {
		listen := tcpMsmt.listenAddr + ":" + strconv.Itoa(port)

		addr, error := net.ResolveTCPAddr("tcp", listen)
		if error != nil {
			fmt.Printf("Cannot parse \"%s\": %s\n", listen, error)
			goHeartbeatCh <- false
			os.Exit(1)
		}

		listener, error = net.ListenTCP("tcp", addr)
		if error == nil {
			// debug fmt.Printf("\nCan listen on addr: %s\n", listen)
			tcpMsmt.usedPorts = append(tcpMsmt.usedPorts, port)

			tcpMsmt.writeListenerStorage(stream, listener)

			goHeartbeatCh <- true
			break
		}

		port++
	}

	conn, error := listener.AcceptTCP()
	if error != nil {
		fmt.Printf("Cannot accept: %s\n", error)
		os.Exit(1)
	}

	fmt.Printf("Connection from %s\n", conn.RemoteAddr())
	// The accept socket must be saved  or we get a concurrent write race condition
	tcpMsmt.writeTcpConnStorage(stream, conn)

	message := make([]byte, tcpMsmt.callSize, tcpMsmt.callSize)

	for {
		bytes, error := conn.Read(message)
		if error != nil {

			// differ cases of error
			if error.(*net.OpError).Err.Error() == "use of closed network connection" {
				// debug fmt.Println("\nTCP Closed network detected! I am ignoring this")
				break
			}

			fmt.Printf("TCP server worker! Cannot read: %s\n", error)
			os.Exit(1)
		}

		tcpMsmt.writeByteStorage(stream, uint64(bytes))

		// maybe we could also do this when receing the actual data
		if fTsExists == false {
			fTs := shared.ConvCurrDateToStr()
			tcpMsmt.writefTsStorage(stream, fTs)
			fTsExists = true
		}

		lTs := shared.ConvCurrDateToStr()
		tcpMsmt.writelTsStorage(stream, lTs)
	}
}

/*
	// 2. create go func to read asynchronously
	go func(readCh chan<- int) {
		for {
			bytes, error := conn.Read(message)
			if error != nil {

				// differ cases of error
				if error.(*net.OpError).Err.Error() == "use of closed network connection" {
					fmt.Println("\nTCP Closed network detected! I am ignoring this")
					break
				}

				fmt.Printf("TCP server worker! Cannot read: %s\n", error)
				os.Exit(1)
			}

			readCh <- bytes
		}
	}(readCh)

	// 3. for loop and select
	for {
		select {
		// ok, there is data to read!
		case bytes := <-readCh:
			tcpMsmt.writeByteStorage(stream, uint64(bytes))

			// maybe we could also do this when receing the actual data
			if fTsExists == false {
				fTs := shared.ConvCurrDateToStr()
				tcpMsmt.writefTsStorage(stream, fTs)
				fTsExists = true
			}

			lTs := shared.ConvCurrDateToStr()
			tcpMsmt.writelTsStorage(stream, lTs)

		// ok, i received a close cmd: tear the socket down
		// PROBLEM: the asyncronous goroutine still reads from the socket
		// result: read from closed connection
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
			fmt.Printf("\nClosing udpConn for stream %s", stream)
			listener.Close()
			conn.Close()

			return
		}
	}
*/

func (tcpMsmt *TcpMsmtObj) writefTsStorage(stream string, ts string) {
	tcpMsmt.fTsStorageMutex.Lock()
	tcpMsmt.fTsStorage[stream] = ts
	tcpMsmt.fTsStorageMutex.Unlock()
}

func (tcpMsmt *TcpMsmtObj) readfTsStorage(stream string) string {
	ts := tcpMsmt.fTsStorage[stream]

	return ts
}

func (tcpMsmt *TcpMsmtObj) writelTsStorage(stream string, ts string) {
	tcpMsmt.lTsStorageMutex.Lock()
	tcpMsmt.lTsStorage[stream] = ts
	tcpMsmt.lTsStorageMutex.Unlock()
}

func (tcpMsmt *TcpMsmtObj) readlTsStorage(stream string) string {
	ts := tcpMsmt.lTsStorage[stream]

	return ts
}

func (tcpMsmt *TcpMsmtObj) writeByteStorage(stream string, bytes uint64) {
	tcpMsmt.byteStorageMutex.Lock()
	tcpMsmt.byteStorage[stream] = tcpMsmt.byteStorage[stream] + bytes
	tcpMsmt.byteStorageMutex.Unlock()
}

// we can precise which key to address
func (tcpMsmt *TcpMsmtObj) readByteStorage(stream string) uint64 {
	bytes := tcpMsmt.byteStorage[stream]

	return bytes
}

func (tcpMsmt *TcpMsmtObj) writeTcpConnStorage(stream string, acceptSock *net.TCPConn) {
	tcpMsmt.tcpConnStorageMutex.Lock()
	tcpMsmt.tcpConnStorage[stream] = acceptSock
	tcpMsmt.tcpConnStorageMutex.Unlock()
}

func (tcpMsmt *TcpMsmtObj) writeListenerStorage(stream string, serverSock *net.TCPListener) {
	tcpMsmt.listenerStorageMutex.Lock()
	tcpMsmt.listenerStorage[stream] = serverSock
	tcpMsmt.listenerStorageMutex.Unlock()
}

func (tcpMsmt *TcpMsmtObj) CloseConn() {
	var msmtData map[string]string

	/*
		for i := 0; i < tcpMsmt.numStreams; i++ {
			tcpMsmt.closeConnCh <- "close"
		}
	*/

	for streamId, acceptSock := range tcpMsmt.tcpConnStorage {
		fmt.Println("\nTCP accept sock closing: ", streamId)
		acceptSock.Close()
	}

	for streamId, srvSock := range tcpMsmt.listenerStorage {
		fmt.Println("\nTCP srv sock closing: ", streamId)
		srvSock.Close()
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
		dataElement.Timestamp_first = tcpMsmt.readfTsStorage(stream)
		dataElement.Timestamp_last = tcpMsmt.readlTsStorage(stream)

		// add single DataElement to slice
		msmtData = append(msmtData, *dataElement)
		// debug fmt.Println("\nmsmtData is: ", msmtData)
	}

	msmtReply.Data = msmtData

	go func() {
		tcpMsmt.msmt2CtrlCh <- *msmtReply
	}()
}
