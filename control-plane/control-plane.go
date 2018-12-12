package controlPlane

import "fmt"
import "os"
import "github.com/monfron/mapago/control-plane/ctrl/server-protocols"
import "github.com/monfron/mapago/control-plane/ctrl/shared"
import "github.com/monfron/mapago/management-plane"

var CTRL_PORT = 64321
var DEF_BUFFER_SIZE = 8096 * 8
var ID string
var ARCH string
var OS string
var MODULES string

func RunServer(lUcAddr string, lMcAddr string, port int, callSize int) {
	var repDataObj *shared.DataObj

	// construct apriori
	ID = shared.ConstructId()
	OS = shared.DetectOs()
	ARCH = shared.DetectArch()
	MODULES = shared.ConvMapToStr(supportedModules())

	ch := make(chan shared.ChResult)

	tcpObj := serverProtos.NewTcpObj("TcpConn", lUcAddr, port, callSize)
	tcpObj.Start(ch)

	udpObj := serverProtos.NewUdpObj("UdpConn", lUcAddr, port, callSize)
	udpObj.Start(ch)

	// PROBLEM: Binding to MC_Addr:CtrlPort and UC_Addr:CtrlPort does not work
	// eventough we got a unique addressing going on
	udpMcObj := serverProtos.NewUdpMcObj("UdpMcConn", lMcAddr, 12346, callSize)
	udpMcObj.Start(ch)

	for {
		request := <-ch
		reqDataObj := shared.ConvJsonToDataStruct(request.Json)

		switch reqDataObj.Type {
		case shared.INFO_REQUEST:
			go func() {
				fmt.Println("\n ------------- Info Request received ------------- ")

				repDataObj = constructInfoReply(reqDataObj)
				json := shared.ConvDataStructToJson(repDataObj)
				request.ConnObj.WriteAnswer(json)
				// ATM: This closes only the TCP ACCEPT sock
				// the UDP (server) socket is untouched
				request.ConnObj.CloseConn()

				// Be ready for receiving another control message via TCP...
				go tcpObj.HandleTcpConn(ch)
			}()

		case shared.INFO_REPLY:
			fmt.Println("I received an INFO_REPLY, I am ignoring this")
			continue

		case shared.MEASUREMENT_START_REQUEST:
			go func() {
				fmt.Println("\n------------- Measurement Start Request -------------")

				recvCh := make(chan shared.ChMsmt2Ctrl)
				clientIp := request.ConnObj.DetectRemoteAddr()
				managementPlane.HandleMsmtStartReq(recvCh, reqDataObj, clientIp.String())

				/*
					POSSIBLE BLOCKING CAUSE
				*/
				msmtReply := <-recvCh

				// at this point all systems are started
				repDataObj = constructMsmtStartReply(reqDataObj, msmtReply)
				json := shared.ConvDataStructToJson(repDataObj)
				request.ConnObj.WriteAnswer(json)

				request.ConnObj.CloseConn()
				go tcpObj.HandleTcpConn(ch)
			}()

		// interaction needed
		case shared.MEASUREMENT_STOP_REQUEST:
			fmt.Println("Construct MEASUREMENT_STOP_REP")

		// interaction needed
		case shared.MEASUREMENT_INFO_REQUEST:
			fmt.Println("Construct MEASUREMENT_INFO_REP")

		// not handled atm
		case shared.TIME_DIFF_REQUEST:
			fmt.Println("Construct TIME_DIFF_REP")

		// not handled atm
		case shared.WARNING_ERR_MSG:
			fmt.Println("WARNING_ERR_MSG")

		default:
			fmt.Printf("Unknown type")
			os.Exit(1)
		}
	}
}

func constructInfoReply(reqDataObj *shared.DataObj) *shared.DataObj {
	fmt.Println("\nConstructing INFO_REP")

	// construct ID
	repDataObj := new(shared.DataObj)
	repDataObj.Type = shared.INFO_REPLY
	repDataObj.Id = ID
	// TODO: THIS IS HARDCODED ATM
	repDataObj.Seq = "0"
	repDataObj.Seq_rp = reqDataObj.Seq
	// maparo STD: timestamp replied untouched by server
	reqDataObj.Ts = reqDataObj.Ts
	repDataObj.Modules = MODULES
	repDataObj.Arch = ARCH
	repDataObj.Os = OS
	repDataObj.Info = "fancyInfo"
	return repDataObj
}

func constructMsmtStartReply(reqDataObj *shared.DataObj, msmtRep shared.ChMsmt2Ctrl) *shared.DataObj {
	fmt.Println("\nConstructing MSMT_START_REP")

	repDataObj := new(shared.DataObj)
	repDataObj.Type = shared.MEASUREMENT_START_REPLY
	repDataObj.Status = msmtRep.Status
	repDataObj.Id = ID
	repDataObj.Seq_rp = reqDataObj.Seq

	msmtData, ok := msmtRep.Data.(map[string]string)
	if ok == false {
		fmt.Printf("Type assertion failed: Looking for map %t", ok)
		os.Exit(1)
	}

	repDataObj.Measurement_id = msmtData["msmtId"]
	repDataObj.Message = msmtData["msg"]

	return repDataObj
}

func supportedModules() map[string]string {
	fmt.Println("\nConstructing supported modules")
	supportedMods := make(map[string]string)
	supportedMods["udp-goodput"] = "no-support"
	supportedMods["tcp-goodput"] = "no-support"
	return supportedMods
}
