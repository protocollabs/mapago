package udpThroughput

import "fmt"
import "os"
import "github.com/monfron/mapago/control-plane/ctrl/shared"

func NewUdpMsmt(msmtCh <-chan shared.ChMgmt2Msmt, ctrlCh chan<- shared.ChMsmt2Ctrl, msmtStartReq *shared.DataObj) {
	var msmtData map[string]string
	msmtResultCh := make(chan shared.ChMsmtResult)

	// select call
	for {
		// POSSIBLE BLOCKING CAUSE: select blocks until one of its cases can block
		select {
		case mgmtCmd := <-msmtCh:
			fmt.Println("\nReceived Management Command: ", mgmtCmd.Cmd)

			switch mgmtCmd.Cmd {
			case "Msmt_start":
				fmt.Println("\nTODO: We have to start UDP msmt module!")

				// TODO: DO THE UDP MODULE SETUP: PORTS etc.

				// SEND REPLY TO CONTROL PLANE
				msmtReply := new(shared.ChMsmt2Ctrl)
				msmtReply.Status = "ok"

				msmtData = make(map[string]string)
				msmtData["msmtId"] = mgmtCmd.MsmtId
				// This could be used to define a more detailed error msg
				msmtData["msg"] = "all modules running"
				msmtReply.Data = msmtData
				ctrlCh <- *msmtReply

			case "Msmt_close":
				fmt.Println("\nTODO: We have to close UDP msmt module!")

			case "Msmt_info":
				fmt.Println("\nTODO: We have to send UDP msmt info!")

			default:
				fmt.Printf("Unknown measurement module")
				os.Exit(1)
			}

		case msmtResult := <-msmtResultCh:
			/*
				- NOTE: the measurement Result could be sent
				at first from trxer
				- we could modify trxer to be a stand alone package
				- func UDP_server(threads int, bufSize int) of trxer
				is then basically our startUDP()
				- within startUDP we access go trxer.UDP_server_worker(c chan<- measurement, port int, bufSize int) {
				- this will send then stuff to msmtResult := <-msmtResultCh:
			*/
			fmt.Println("\nReceived Measurement result!!!")
			fmt.Println("Received bytes: ", msmtResult.Bytes)

		}

	}

}
