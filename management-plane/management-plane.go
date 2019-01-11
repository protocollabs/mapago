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
var startPort = 7000

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

		tcpMsmtObj := tcpThroughput.NewTcpMsmtObj(msmtCh, ctrlCh, msmtStartReq, msmtId, startPort)

		msmtEntry := new(shared.MsmtStorageEntry)
		msmtEntry.MsmtCh = msmtCh
		msmtEntry.MsmtObj = tcpMsmtObj
		msmtStorage[msmtId] = msmtEntry

		fmt.Println("\nmsmtStorage content: ", msmtStorage)

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

		udpMsmt := udpThroughput.NewUdpThroughputMsmt(msmtCh, ctrlCh, msmtStartReq, msmtId, startPort)

		msmtEntry := new(shared.MsmtStorageEntry)
		msmtEntry.MsmtCh = msmtCh
		msmtEntry.MsmtObj = udpMsmt
		msmtStorage[msmtId] = msmtEntry

		fmt.Println("\nmsmtStorage content: ", msmtStorage)
	}
}

func constructMsmtId(cltAddr string) string {
	// cut the port from clt_addr
	spltCltAddr := strings.Split(cltAddr, ":")
	msmtId := spltCltAddr[0] + "=" + strconv.Itoa(int(rand.Int31()))

	return msmtId
}

func HandleMsmtStopReq(msmtId string) {
	fmt.Printf("\nMsmtStopReq called!!")

	msmtEntry, exists := msmtStorage[msmtId]
	if exists == false {
		fmt.Printf("\nmsmtEntry for msmtId NOT in storage")
		os.Exit(1)
	}

	fmt.Println("\nhandle msmst stop req here")

	switch msmstObj := msmtEntry.MsmtObj.(type) {
	case *tcpThroughput.TcpMsmtObj:
		msmstObj.CloseConn()
	case *udpThroughput.UdpThroughputMsmt:
		msmstObj.CloseConn()
	// TODO	case *quicThroughput.QuicMsmtObj:
	default:
		fmt.Printf("Type assertion failed: Unknown msmt type")
		os.Exit(1)
	}
}

func HandleMsmtInfoReq(msmtId string) {
	fmt.Printf("\nMsmtInfoReq called!!")

	msmtEntry, exists := msmtStorage[msmtId]
	if exists == false {
		fmt.Printf("\nmsmtEntry for msmtId NOT in storage")
		os.Exit(1)
	}
	// debug
	fmt.Println("\n I found the appropriate msmt obj: ", msmtEntry)

	switch msmstObj := msmtEntry.MsmtObj.(type) {
	case *tcpThroughput.TcpMsmtObj:
		msmstObj.GetMsmtInfo()
	case *udpThroughput.UdpThroughputMsmt:
		msmstObj.GetMsmtInfo()
	// TODO	case *quicThroughput.QuicMsmtObj:
	default:
		fmt.Printf("Type assertion failed: Unknown msmt type")
		os.Exit(1)
	}
}
