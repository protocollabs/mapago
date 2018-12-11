package managementPlane

import "fmt"
import "os"
import "math/rand"
import "strings"
import "strconv"
import "github.com/monfron/mapago/control-plane/ctrl/shared"
import "github.com/monfron/mapago/measurement-plane/tcp-throughput"
import "github.com/monfron/mapago/measurement-plane/udp-throughput"

var msmtStorage map[string]chan shared.ChMgmt2Msmt
var mapInited = false

func HandleMsmtStartReq(ctrlCh chan<- shared.ChMsmt2Ctrl, msmtStartReq *shared.DataObj, cltAddr string) {
	switch msmtStartReq.Measurement.Name {
	case "tcp-throughput":
		// TODO: differ in constructing msmtId
		msmtId := constructMsmtId(cltAddr)
		msmtCh := make(chan shared.ChMgmt2Msmt)

		if mapInited == false {
			msmtStorage = make(map[string]chan shared.ChMgmt2Msmt)
			mapInited = true
		}

		msmtStorage[msmtId] = msmtCh
		fmt.Println("\nmsmtStorage content: ", msmtStorage)

		tcpMsmtObj := tcpThroughput.NewTcpMsmtObj(msmtCh, ctrlCh, msmtStartReq, msmtId)
		fmt.Println("\nConstructor constructed tcp msmt object: ", tcpMsmtObj)

		// TODO: tcpThroughput.Start() etc.

	case "udp-throughput":
		// TODO: MOVE COMMON STUFF UP
		msmtId := constructMsmtId(cltAddr)
		msmtCh := make(chan shared.ChMgmt2Msmt)

		if mapInited == false {
			msmtStorage = make(map[string]chan shared.ChMgmt2Msmt)
			mapInited = true
		}

		msmtStorage[msmtId] = msmtCh
		fmt.Println("\nmsmtStorage content: ", msmtStorage)

		/*
			POSSIBLE BLOCKING CAUSE
			we have to call it via goroutine asynchronously
			or we stay within the for loop and block on the channel
			and cannot receive anything else
		*/
		go udpThroughput.NewUdpMsmt(msmtCh, ctrlCh, msmtStartReq)

		/*
			POSSIBLE BLOCKING CAUSE
			send blocks until corresponding read is called
			PROBLEM: this function is blocked => the callee is blocked aswell
			=> connClose() and HandleConn() cannot be called => no further requests
		*/

		mgmtCmd := new(shared.ChMgmt2Msmt)
		mgmtCmd.Cmd = "Msmt_start"
		mgmtCmd.MsmtId = msmtId
		msmtCh <- *mgmtCmd

	case "quic-throughput":
		fmt.Println("\nStarting QUIC throughput module")

	case "udp-ping":
		fmt.Println("\nStarting UDP ping module")

	default:
		fmt.Printf("Unknown measurement module")
		os.Exit(1)
	}
}

func constructMsmtId(cltAddr string) string {
	// cut the port from clt_addr
	spltCltAddr := strings.Split(cltAddr, ":")
	msmtId := spltCltAddr[0] + "=" + strconv.Itoa(int(rand.Int31()))

	return msmtId
}
