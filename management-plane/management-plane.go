package managementPlane

import "fmt"
import "os"
import "math/rand"
import "strings"
import "strconv"
import "github.com/monfron/mapago/control-plane/ctrl/shared"

func HandleMsmtStartReq(recvCh chan<- shared.ChMsmt2Ctrl, msmtStartReq *shared.DataObj, cltAddr string) {
	fmt.Printf("\n\nclient addr is: % s", cltAddr)

	switch msmtStartReq.Measurement.Name {
	case "tcp-throughput":
		msmtId := constructMsmtId(cltAddr)
		fmt.Printf("\nMsmtId is: %s", msmtId)

	case "udp-throughput":
		fmt.Println("\nStarting UDP throughput module")

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
