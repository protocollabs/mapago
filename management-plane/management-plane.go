package managementPlane

import "fmt"
import "os"
import "math/rand"
import "strings"
import "strconv"
import "github.com/monfron/mapago/control-plane/ctrl/shared"
import "github.com/monfron/mapago/measurement-plane/tcp-throughput"
import "github.com/monfron/mapago/measurement-plane/udp-throughput"

var msmtStorage map[string]*shared.MsmtStorageEntry
var mapInited = false

func HandleMsmtStartReq(ctrlCh chan<- shared.ChMsmt2Ctrl, msmtStartReq *shared.DataObj, cltAddr string) {
	switch msmtStartReq.Measurement.Name {
	case "tcp-throughput":
		/*
			- TODO: differ in constructing msmtId
			- i.e. send only ID not hostaddr
		*/
		msmtId := constructMsmtId(cltAddr)
		msmtCh := make(chan shared.ChMgmt2Msmt)

		if mapInited == false {
			msmtStorage = make(map[string]*shared.MsmtStorageEntry)
			mapInited = true
		}

		tcpMsmtObj := tcpThroughput.NewTcpMsmtObj(msmtCh, ctrlCh, msmtStartReq, msmtId)

		msmtEntry := new(shared.MsmtStorageEntry)
		msmtEntry.MsmtCh = msmtCh
		msmtEntry.MsmtObj = tcpMsmtObj
		msmtStorage[msmtId] = msmtEntry

		fmt.Println("\nmsmtStorage content: ", msmtStorage)

		tcpMsmtObj.Start()

	case "udp-throughput":
		/*
			- TODO: MOVE COMMON STUFF UP
		*/
		msmtId := constructMsmtId(cltAddr)
		msmtCh := make(chan shared.ChMgmt2Msmt)

		if mapInited == false {
			msmtStorage = make(map[string]*shared.MsmtStorageEntry)
			mapInited = true
		}

		msmtEntry := new(shared.MsmtStorageEntry)
		msmtEntry.MsmtCh = msmtCh
		// cannot be used ATM : WIP msmtEntry.MsmtObj = nil
		msmtStorage[msmtId] = msmtEntry
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

/*
- TODO: manager sendet nicht über channel in tcpworker
- ruft einfach tcpThorughput.stop() auf
- dieser sendet dann an TcpWorker über Channel
*/
func HandleMsmtStopReq(msmtId string) {
	fmt.Printf("\nMsmtStartReq called!!")

	msmtEntry, exists := msmtStorage[msmtId]
	if exists == false {
		fmt.Printf("\nmsmtEntry for msmtId NOT in storage")
		os.Exit(1)
	}

	switch msmstObj := msmtEntry.MsmtObj.(type) {
	case *tcpThroughput.TcpMsmtObj:
		msmstObj.CloseConn()
	// TODO	case *udpThroughput.UdpMsmtObj:
	// TODO	case *quicThroughput.QuicMsmtObj:
	default:
		fmt.Printf("Type assertion failed: Unknown msmt type")
		os.Exit(1)
	}
}
