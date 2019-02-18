package managementPlane

import "fmt"
import "os"
import "math/rand"
import "strings"
import "strconv"
import "github.com/protocollabs/mapago/control-plane/ctrl/shared"
import "github.com/protocollabs/mapago/measurement-plane/tcp-throughput"
import "github.com/protocollabs/mapago/measurement-plane/udp-throughput"
import "github.com/protocollabs/mapago/measurement-plane/quic-throughput"
import "github.com/protocollabs/mapago/measurement-plane/tcp-tls-throughput"

var msmtStorage map[string]*shared.MsmtStorageEntry
var mapInited = false
var startPort = 7000

func HandleMsmtStartReq(ctrlCh chan<- shared.ChMsmt2Ctrl, msmtStartReq *shared.DataObj, cltAddr string) {
	switch msmtStartReq.Measurement.Name {
	case "tcp-throughput":
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
	case "tcp-tls-throughput":
		msmtId := constructMsmtId(cltAddr)
		msmtCh := make(chan shared.ChMgmt2Msmt)

		if mapInited == false {
			msmtStorage = make(map[string]*shared.MsmtStorageEntry)
			mapInited = true
		}

		tcpMsmtObj := tcpTlsThroughput.NewTcpTlsThroughputMsmt(msmtCh, ctrlCh, msmtStartReq, msmtId, startPort)

		msmtEntry := new(shared.MsmtStorageEntry)
		msmtEntry.MsmtCh = msmtCh
		msmtEntry.MsmtObj = tcpMsmtObj
		msmtStorage[msmtId] = msmtEntry

		fmt.Println("\nmsmtStorage content: ", msmtStorage)

	case "udp-throughput":
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

	case "quic-throughput":
		msmtId := constructMsmtId(cltAddr)
		msmtCh := make(chan shared.ChMgmt2Msmt)

		if mapInited == false {
			msmtStorage = make(map[string]*shared.MsmtStorageEntry)
			mapInited = true
		}

		quicMsmt := quicThroughput.NewQuicThroughputMsmt(msmtCh, ctrlCh, msmtStartReq, msmtId, startPort)

		msmtEntry := new(shared.MsmtStorageEntry)
		msmtEntry.MsmtCh = msmtCh
		msmtEntry.MsmtObj = quicMsmt
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
	fmt.Println("\nmsmt storage Entry is: ", msmtEntry)

	switch msmstObj := msmtEntry.MsmtObj.(type) {
	case *tcpThroughput.TcpMsmtObj:
		msmstObj.CloseConn()
	// TODO: case *tcpTlsThroughput....
	case *tcpTlsThroughput.TcpTlsThroughputMsmt:
		msmstObj.CloseConn()
	case *udpThroughput.UdpThroughputMsmt:
		msmstObj.CloseConn()
	case *quicThroughput.QuicThroughputMsmt:
		msmstObj.CloseConn()
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
	// TODO: case *tcpTlsThroughput....
	case *tcpTlsThroughput.TcpTlsThroughputMsmt:
		msmstObj.GetMsmtInfo()
	case *udpThroughput.UdpThroughputMsmt:
		msmstObj.GetMsmtInfo()
	case *quicThroughput.QuicThroughputMsmt:
		msmstObj.GetMsmtInfo()
	default:
		fmt.Printf("Type assertion failed: Unknown msmt type")
		os.Exit(1)
	}
}
