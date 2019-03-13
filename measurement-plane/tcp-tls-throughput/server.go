package tcpTlsThroughput

import "fmt"
import "os"
import "net"
import "io"
import "strconv"
import "crypto/tls"
import "github.com/protocollabs/mapago/control-plane/ctrl/shared"

type TcpTlsThroughputMsmt struct {
	numStreams int
	usedPorts  []int
	callSize   int
	listenAddr string
	msmtId     string
	// TODO: Maybe this is then not valid anymore!!!
	msmtInfoStorage map[string]*shared.MsmtInfoObj
	connStorage     map[string]*shared.TcpTlsConnObj

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

func NewTcpTlsThroughputMsmt(msmtCh <-chan shared.ChMgmt2Msmt, ctrlCh chan<- shared.ChMsmt2Ctrl, msmtStartReq *shared.DataObj, msmtId string, startPort int) *TcpTlsThroughputMsmt {
	var err error
	var msmtData map[string]string
	heartbeatCh := make(chan bool)
	closeCh := make(chan interface{})
	tcpMsmt := new(TcpTlsThroughputMsmt)

	tcpMsmt.msmtInfoStorage = make(map[string]*shared.MsmtInfoObj)
	tcpMsmt.connStorage = make(map[string]*shared.TcpTlsConnObj)
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

		initTs := shared.ConvCurrDateToStr()
		msmtInfo := shared.MsmtInfoObj{Bytes: 0, FirstTs: initTs, LastTs: initTs}
		tcpMsmt.msmtInfoStorage[stream] = &msmtInfo

		tcpConns := shared.TcpTlsConnObj{}
		tcpMsmt.connStorage[stream] = &tcpConns

		go tcpMsmt.tcpTlsServerWorker(closeCh, heartbeatCh, startPort, c, &msmtInfo, &tcpConns)
	}

	for c := 1; c <= tcpMsmt.numStreams; c++ {
		heartbeat := <-heartbeatCh
		if heartbeat != true {
			panic("tcp_server_worker goroutine not ok")
		}
	}

	fmt.Println("\n\nPorts used by TCP TLS module: ", tcpMsmt.usedPorts)

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

// TODO: most of the TLS changes happen here
func (tcpMsmt *TcpTlsThroughputMsmt) tcpTlsServerWorker(closeCh <-chan interface{}, goHeartbeatCh chan<- bool, port int, streamIndex int, msmtInfoPtr *shared.MsmtInfoObj, connPtr *shared.TcpTlsConnObj) {
	var listener net.Listener
	fTsExists := false
	srvCertPath := "/src/github.com/protocollabs/mapago/measurement-plane/tcp-tls-throughput/certs"
	stream := "stream" + strconv.Itoa(streamIndex)
	fmt.Printf("\n%s is here", stream)

	goPath := os.Getenv("GOPATH")
	srvPemPath := goPath + srvCertPath + "/server.pem"
	srvKeyPath := goPath + srvCertPath + "/server.key"

	// new
	cert, error := tls.LoadX509KeyPair(srvPemPath, srvKeyPath)
	if error != nil {
		fmt.Printf("Cannot loadkeys: %s\n", error)
		os.Exit(1)
	}

	// new
	config := tls.Config{Certificates: []tls.Certificate{cert}}

	for {
		lAddr := tcpMsmt.listenAddr + ":" + strconv.Itoa(port)

		listener, error = tls.Listen("tcp", lAddr, &config)
		if error == nil {
			tcpMsmt.usedPorts = append(tcpMsmt.usedPorts, port)
			// SrvSock is net.Listener
			connPtr.SrvSock = listener

			goHeartbeatCh <- true
			break
		}

		port++
	}

	// AcceptTCP returns *net.TCPConn, Accept returns net.Conn
	conn, error := listener.Accept()
	if error != nil {
		fmt.Printf("Cannot accept: %s\n", error)
		os.Exit(1)
	}

	fmt.Printf("Connection from %s\n", conn.RemoteAddr())
	connPtr.AcceptSock = conn

	// TODO: We could also check tls.ConnectionsState

	message := make([]byte, tcpMsmt.callSize, tcpMsmt.callSize)

	for {
		bytes, err := conn.Read(message)
		if err != nil {
			if err == io.EOF {
				break
			}

			if err.(*net.OpError).Err.Error() == "use of closed network connection" {
				break
			}

			// something different serious...
			os.Exit(1)
		}

		msmtInfoPtr.Bytes += uint64(bytes)

		if fTsExists == false {
			fTs := shared.ConvCurrDateToStr()
			msmtInfoPtr.FirstTs = fTs
			fTsExists = true
		}

		lTs := shared.ConvCurrDateToStr()
		msmtInfoPtr.LastTs = lTs
	}
}

func (tcpMsmt *TcpTlsThroughputMsmt) CloseConn() {
	var mgmtData map[string]string
	var msmtData []shared.DataResultObj
	var combinedData shared.CombinedData

	for c, tcpConns := range tcpMsmt.connStorage {
		fmt.Println("\nClosing stream: ", c)
		tcpConns.AcceptSock.Close()
		tcpConns.SrvSock.Close()
	}

	msmtReply := new(shared.ChMsmt2Ctrl)
	msmtReply.Status = "ok"

	mgmtData = make(map[string]string)
	mgmtData["msmtId"] = tcpMsmt.msmtId
	mgmtData["msg"] = "all modules closed"

	combinedData.MgmtData = mgmtData

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

	combinedData.MsmtData = msmtData
	msmtReply.Data = combinedData

	go func() {
		tcpMsmt.msmt2CtrlCh <- *msmtReply
	}()
}

func (tcpMsmt *TcpTlsThroughputMsmt) GetMsmtInfo() {
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
