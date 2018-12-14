package tcpThroughput

import "fmt"
import "os"
import "net"
import "time"
import "strconv"
import "github.com/monfron/mapago/control-plane/ctrl/shared"

var UPDATE_INTERVAL = 5

type TcpMsmtObj struct {
	numStreams int
	// TODO: in the future the server should define which Port to use
	startPort  int
	callSize   int
	listenAddr string
	msmtId     string

	/*
		- this attribute can be used by start() to RECEIVE result from TcpServerWorker
		- i.e. => select => msmtResult := <-msmtResultCh
	*/
	msmtResultCh chan shared.ChMsmtResult

	/*
		- this attribute can be used by start() to RECEIVE cmd from managementplane
	*/
	msmtMgmt2MsmtCh <-chan shared.ChMgmt2Msmt

	/*
		- this attribute can be used by constructor and start() to SEND msgs to control plane
	*/
	msmt2CtrlCh chan<- shared.ChMsmt2Ctrl
}

func NewTcpMsmtObj(msmtCh <-chan shared.ChMgmt2Msmt, ctrlCh chan<- shared.ChMsmt2Ctrl, msmtStartReq *shared.DataObj, msmtId string) *TcpMsmtObj {
	var err error
	var msmtData map[string]string
	heartbeatCh := make(chan bool)
	resultCh := make(chan shared.ChMsmtResult)

	tcpMsmt := new(TcpMsmtObj)

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
	tcpMsmt.msmtResultCh = resultCh
	tcpMsmt.msmtMgmt2MsmtCh = msmtCh
	tcpMsmt.msmt2CtrlCh = ctrlCh

	fmt.Printf("\n\nTCP Msmt: Client wants %d workers to be started from start port %d!", tcpMsmt.numStreams, tcpMsmt.startPort)

	for c := 1; c <= tcpMsmt.numStreams; c++ {
		fmt.Printf("\n\nStarting stream %d on port %d", c, tcpMsmt.startPort)
		go tcpMsmt.tcpServerWorker(resultCh, heartbeatCh, tcpMsmt.startPort)
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

func (tcpMsmt *TcpMsmtObj) tcpServerWorker(resCh chan<- shared.ChMsmtResult, goHeartbeatCh chan<- bool, port int) {

	/*
		- we can not do that: listen := tcpMsmt.listenAddr + ":" + strconv.Itoa(tcpMsmt.startPort)
		- or we get a race condition :(
	*/
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

	/*
		- NOTE: We must not do "defer listener.Close()" here
		- this has to be done after msmt_stop_req incoming
	*/

	goHeartbeatCh <- true

	conn, error := listener.AcceptTCP()
	if error != nil {
		fmt.Printf("Cannot accept: %s\n", error)
		os.Exit(1)
	}

	/*
		- NOTE: We must not do "defer conn.Close()" here
		- this has to be done after msmt_stop_req incoming
	*/

	fmt.Printf("Connection from %s\n", conn.RemoteAddr())
	message := make([]byte, tcpMsmt.callSize, tcpMsmt.callSize)

	var bytes uint64 = 0
	start := time.Now()
	for {
		// this one is blocking
		n1, error := conn.Read(message)
		if error != nil {
			fmt.Printf("Cannot read: %s\n", error)
			os.Exit(1)
		}

		bytes += uint64(n1)

		elapsed := time.Since(start)
		if elapsed.Seconds() > float64(UPDATE_INTERVAL) {
			// this result is sent to start() => select => msmtResult := <-msmtResultCh
			resCh <- shared.ChMsmtResult{Bytes: bytes, Time: elapsed.Seconds()}
			start = time.Now()
			bytes = 0
		}
	}
}

func (tcpMsmt *TcpMsmtObj) Start() {
	numValCtr := 0
	var accumulated uint64
	var msmtData map[string]string

	go func() {
		for {
			// POSSIBLE BLOCKING CAUSE: select blocks until one of its cases can run
			select {
			case mgmtCmd := <-tcpMsmt.msmtMgmt2MsmtCh:
				fmt.Println("\nReceived Management Command: ", mgmtCmd.Cmd)

				switch mgmtCmd.Cmd {

				case "Msmt_close":
					fmt.Println("\nWe have to close TCP msmt module!")
					/*
						TODO10: we have to close the connection from here
						i.e. we use a select in the tcpServerWorker()
						with a channel we can communicate with that
					*/
					msmtReply := new(shared.ChMsmt2Ctrl)
					msmtReply.Status = "ok"

					msmtData = make(map[string]string)
					msmtData["msmtId"] = tcpMsmt.msmtId
					msmtData["msg"] = "all modules closed"
					msmtReply.Data = msmtData

					// TODO: we have to attach the Measurement Data
					tcpMsmt.msmt2CtrlCh <- *msmtReply

				case "Msmt_info":
					fmt.Println("\nTODO: We have to send TCP msmt info!")

				default:
					fmt.Printf("Unknown measurement module")
					os.Exit(1)
				}

			case msmtResult := <-tcpMsmt.msmtResultCh:
				numValCtr += 1
				accumulated += msmtResult.Bytes

				if numValCtr == tcpMsmt.numStreams {
					fmt.Printf("\nGot reply from all %d workers", tcpMsmt.numStreams)
					mbyte_sec := accumulated / (1000000 * uint64(UPDATE_INTERVAL))
					println("\nMByte/sec: ", mbyte_sec)
					// start next measurement burst
					accumulated = 0
					numValCtr = 0
				}
			}
		}
	}()
}
