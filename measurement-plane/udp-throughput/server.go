package udpThroughput

import "fmt"
import "os"
import "net"
import "strconv"
import "github.com/protocollabs/mapago/control-plane/ctrl/shared"

type UdpThroughputMsmt struct {
	numStreams      int
	usedPorts       []int
	callSize        int
	listenAddr      string
	msmtId          string
	msmtInfoStorage map[string]*shared.MsmtInfoObj
	connStorage     map[string]*shared.UdpConn

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

	udpMsmt.msmtInfoStorage = make(map[string]*shared.MsmtInfoObj)
	udpMsmt.connStorage = make(map[string]*shared.UdpConn)

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
		stream := "stream" + strconv.Itoa(c)

		initTs := shared.ConvCurrDateToStr()
		msmtInfo := shared.MsmtInfoObj{Bytes: 0, FirstTs: initTs, LastTs: initTs}
		udpMsmt.msmtInfoStorage[stream] = &msmtInfo

		udpConn := shared.UdpConn{}
		udpMsmt.connStorage[stream] = &udpConn

		go udpMsmt.udpServerWorker(closeCh, heartbeatCh, startPort, c, &msmtInfo, &udpConn)
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

func (udpMsmt *UdpThroughputMsmt) udpServerWorker(closeCh <-chan interface{}, goHeartbeatCh chan<- bool, port int, streamIndex int, msmtInfo *shared.MsmtInfoObj, conn *shared.UdpConn) {
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
			udpMsmt.usedPorts = append(udpMsmt.usedPorts, port)
			conn.SrvSock = udpConn
			goHeartbeatCh <- true
			break
		}

		port++
	}

	message := make([]byte, udpMsmt.callSize, udpMsmt.callSize)

	for {
		bytes, cltAddr, error := udpConn.ReadFromUDP(message)
		if error != nil {

			if error.(*net.OpError).Err.Error() == "use of closed network connection" {
				// debug fmt.Println("\nClosed network detected! I am ignoring this")
				break
			}

			fmt.Printf("Udp server worker! Cannot read: %s\n", error)
			os.Exit(1)
		}

		if cltAddrExists == false {
			fmt.Println("\nConnection from: ", cltAddr)
			cltAddrExists = true
		}

		msmtInfo.Bytes += uint64(bytes)

		if fTsExists == false {
			fTs := shared.ConvCurrDateToStr()
			msmtInfo.FirstTs = fTs
			fTsExists = true
		}

		lTs := shared.ConvCurrDateToStr()
		msmtInfo.LastTs = lTs
	}
}

func (udpMsmt *UdpThroughputMsmt) CloseConn() {
	var mgmtData map[string]string
	var msmtData []shared.DataResultObj
	var combinedData shared.CombinedData

	for streamId, conn := range udpMsmt.connStorage {
		fmt.Println("\nUDP closing: ", streamId)
		conn.SrvSock.Close()
	}

	msmtReply := new(shared.ChMsmt2Ctrl)
	msmtReply.Status = "ok"

	mgmtData = make(map[string]string)
	mgmtData["msmtId"] = udpMsmt.msmtId
	mgmtData["msg"] = "all modules closed"
	combinedData.MgmtData = mgmtData

	for c := 1; c <= udpMsmt.numStreams; c++ {
		stream := "stream" + strconv.Itoa(c)

		msmtStruct := udpMsmt.msmtInfoStorage[stream]

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
		udpMsmt.msmt2CtrlCh <- *msmtReply
	}()
}

func (udpMsmt *UdpThroughputMsmt) GetMsmtInfo() {
	var msmtData []shared.DataResultObj

	msmtReply := new(shared.ChMsmt2Ctrl)
	msmtReply.Status = "ok"

	for c := 1; c <= udpMsmt.numStreams; c++ {
		stream := "stream" + strconv.Itoa(c)

		msmtStruct := udpMsmt.msmtInfoStorage[stream]

		dataElement := new(shared.DataResultObj)
		dataElement.Received_bytes = strconv.Itoa(int(msmtStruct.Bytes))
		dataElement.Timestamp_first = msmtStruct.FirstTs
		dataElement.Timestamp_last = msmtStruct.LastTs

		msmtData = append(msmtData, *dataElement)
		fmt.Println("\nmsmtData is: ", msmtData)
	}

	msmtReply.Data = msmtData

	go func() {
		udpMsmt.msmt2CtrlCh <- *msmtReply
	}()
}
