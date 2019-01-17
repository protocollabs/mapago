package tcpThroughput

import "fmt"
import "os"
import "net"
import "strconv"
import "github.com/protocollabs/mapago/control-plane/ctrl/shared"

var UPDATE_INTERVAL = 5

type TcpMsmtObj struct {
	numStreams      int
	usedPorts       []int
	callSize        int
	listenAddr      string
	msmtId          string
	msmtInfoStorage map[string]*shared.MsmtInfoObj
	// we dont need a ptr to be precisely
	connStorage map[string]*shared.TcpConnObj

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

	tcpMsmt.msmtInfoStorage = make(map[string]*shared.MsmtInfoObj)
	tcpMsmt.connStorage = make(map[string]*shared.TcpConnObj)
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

	tcpMsmt.msmtId = msmtId
	tcpMsmt.listenAddr = msmtStartReq.Measurement.Configuration.Listen_addr
	tcpMsmt.msmtMgmt2MsmtCh = msmtCh
	tcpMsmt.msmt2CtrlCh = ctrlCh
	tcpMsmt.closeConnCh = closeCh

	for c := 1; c <= tcpMsmt.numStreams; c++ {
		stream := "stream" + strconv.Itoa(c)

		msmtInfo := shared.MsmtInfoObj{}
		tcpMsmt.msmtInfoStorage[stream] = &msmtInfo

		tcpConns := shared.TcpConnObj{}
		tcpMsmt.connStorage[stream] = &tcpConns

		go tcpMsmt.tcpServerWorker(closeCh, heartbeatCh, startPort, c, &msmtInfo, &tcpConns)
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

func (tcpMsmt *TcpMsmtObj) tcpServerWorker(closeCh <-chan interface{}, goHeartbeatCh chan<- bool, port int, streamIndex int, msmtInfoPtr *shared.MsmtInfoObj, connPtr *shared.TcpConnObj) {
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
			connPtr.SrvSock = listener
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
	connPtr.AcceptSock = conn

	message := make([]byte, tcpMsmt.callSize, tcpMsmt.callSize)

	for {
		bytes, error := conn.Read(message)
		if error != nil {
			if error.(*net.OpError).Err.Error() == "use of closed network connection" {
				// debug fmt.Println("\nTCP Closed network detected! I am ignoring this")
				break
			}

			fmt.Printf("TCP server worker! Cannot read: %s\n", error)
			os.Exit(1)
		}

		msmtInfoPtr.Bytes = uint64(bytes)

		if fTsExists == false {
			fTs := shared.ConvCurrDateToStr()
			msmtInfoPtr.FirstTs = fTs
			fTsExists = true
		}

		lTs := shared.ConvCurrDateToStr()
		msmtInfoPtr.LastTs = lTs
	}
}

func (tcpMsmt *TcpMsmtObj) CloseConn() {
	var msmtData map[string]string

	for c, tcpConns := range tcpMsmt.connStorage {
		fmt.Println("\nClosing stream: ", c)
		tcpConns.AcceptSock.Close()
		tcpConns.SrvSock.Close()
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

		msmtStruct := tcpMsmt.msmtInfoStorage[stream]

		dataElement := new(shared.DataResultObj)
		dataElement.Received_bytes = strconv.Itoa(int(msmtStruct.Bytes))
		dataElement.Timestamp_first = msmtStruct.FirstTs
		dataElement.Timestamp_last = msmtStruct.LastTs

		msmtData = append(msmtData, *dataElement)
		fmt.Println("\nmsmtData is: ", msmtData)
	}

	msmtReply.Data = msmtData

	go func() {
		tcpMsmt.msmt2CtrlCh <- *msmtReply
	}()
}
