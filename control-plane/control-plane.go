package controlPlane

import "fmt"
import "os"
import "github.com/protocollabs/mapago/control-plane/ctrl/server-protocols"
import "github.com/protocollabs/mapago/control-plane/ctrl/shared"
import "github.com/protocollabs/mapago/management-plane"

var CTRL_PORT = 64321
var DEF_BUFFER_SIZE = 8096 * 8
var ID string
var ARCH string
var OS string
var MODULES string

func RunServer(lUcAddr string, lMcAddr string, port int, callSize int) {
	var repDataObj *shared.DataObj
	var recvChStorage map[string]chan shared.ChMsmt2Ctrl

	// construct apriori
	ID = shared.ConstructId()
	OS = shared.DetectOs()
	ARCH = shared.DetectArch()
	MODULES = shared.ConvMapToStr(supportedModules())

	ch := make(chan shared.ChResult)
	recvChStorage = make(map[string]chan shared.ChMsmt2Ctrl)

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

				// DISCUSS: we dont need go tcpObj.HandleTcpConn(ch)
				tcpObj.HandleTcpConn(ch)

			}()

		case shared.INFO_REPLY:
			fmt.Println("I received an INFO_REPLY, I am ignoring this")
			continue

		case shared.MEASUREMENT_START_REQUEST:
			go func() {
				fmt.Println("\n------------- MEASUREMENT_START_REQUEST -------------")
				/*
					DISCUSS:
					- recvCh is a local variable saved on the stack of this goroutine
					- using this recvCh we have to also receive the reply from the *associated*
					msmt module when for example a Msmt Stop Req or Msmt Info Req
					- i.e. the recvCh has to be available in the scopes of the other goroutines aswell
					- so we cannot say we store this recvCh everytime in a more "global" variable
					or the value is overwritten everytime a new Msmt_start_req comes in
					- we could say instead we store the recvCh in a map, indexed by the msmt_id (currenetly done)
				*/
				recvCh := make(chan shared.ChMsmt2Ctrl)

				clientIp := request.ConnObj.DetectRemoteAddr()
				// RFC: atm the server listen port is advertised by the client
				managementPlane.HandleMsmtStartReq(recvCh, reqDataObj, clientIp.String())

				/*
					POSSIBLE BLOCKING CAUSE
				*/
				msmtReply := <-recvCh

				// at this point all systems are started
				repDataObj = constructMsmtStartReply(reqDataObj, msmtReply)

				recvChStorage[repDataObj.Measurement_id] = recvCh
				fmt.Println("\nRecvChStorage: ", recvChStorage)

				json := shared.ConvDataStructToJson(repDataObj)
				request.ConnObj.WriteAnswer(json)

				request.ConnObj.CloseConn()
				// DISCUSS: we dont need go tcpObj.HandleTcpConn(ch)
				tcpObj.HandleTcpConn(ch)

			}()

		// interaction needed
		case shared.MEASUREMENT_STOP_REQUEST:
			go func() {
				fmt.Println("\n------------- MEASUREMENT_STOP_REQUEST -------------")

				recvCh, exists := recvChStorage[reqDataObj.Measurement_id]
				if exists == false {
					fmt.Printf("\nrecvChEntry NOT in storage")
					os.Exit(1)
				}

				managementPlane.HandleMsmtStopReq(reqDataObj.Measurement_id)

				msmtReply := <-recvCh

				repDataObj = constructMsmtStopReply(reqDataObj, msmtReply)

				json := shared.ConvDataStructToJson(repDataObj)
				request.ConnObj.WriteAnswer(json)

				// Remove recvCh from recvChStorage

				request.ConnObj.CloseConn()
				tcpObj.HandleTcpConn(ch)
			}()

		// interaction needed
		case shared.MEASUREMENT_INFO_REQUEST:
			go func() {
				fmt.Println("\n------------- MEASUREMENT_INFO_REQUEST -------------")

				// debug fmt.Println("\nrequest is: ", reqDataObj)

				recvCh, exists := recvChStorage[reqDataObj.Measurement_id]
				if exists == false {
					fmt.Printf("\nrecvChEntry NOT in storage")
					os.Exit(1)
				}

				managementPlane.HandleMsmtInfoReq(reqDataObj.Measurement_id)

				msmtReply := <-recvCh

				repDataObj = constructMsmtInfoReply(reqDataObj, msmtReply)
				// debug fmt.Println("\nMsmt Info Reply is: ", repDataObj)

				json := shared.ConvDataStructToJson(repDataObj)
				request.ConnObj.WriteAnswer(json)

				request.ConnObj.CloseConn()
				tcpObj.HandleTcpConn(ch)
			}()

		// not handled atm
		case shared.TIME_DIFF_REQUEST:
			fmt.Println("\nConstruct TIME_DIFF_REP")

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
	fmt.Println("\n------------- Constructing INFO_REPLY -------------")

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
	fmt.Println("\n------------- Constructing MSMT_START_REPLY -------------")

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
	repDataObj.Measurement.Configuration.UsedPorts = msmtData["usedPorts"]

	return repDataObj
}

func constructMsmtStopReply(reqDataObj *shared.DataObj, msmtRep shared.ChMsmt2Ctrl) *shared.DataObj {
	fmt.Println("\n------------- Constructing MSMT_STOP_REPLY -------------")

	repDataObj := new(shared.DataObj)
	repDataObj.Type = shared.MEASUREMENT_STOP_REPLY
	repDataObj.Status = msmtRep.Status
	repDataObj.Id = ID
	repDataObj.Seq_rp = reqDataObj.Seq

	combined, ok := msmtRep.Data.(shared.CombinedData)
	if ok == false {
		fmt.Printf("Type assertion failed: Looking for combined data %t", ok)
		os.Exit(1)
	}

	repDataObj.Measurement_id = combined.MgmtData["msmtId"]
	repDataObj.Message = combined.MgmtData["msg"]
	repDataObj.Data.DataElements = combined.MsmtData

	return repDataObj
}

func constructMsmtInfoReply(reqDataObj *shared.DataObj, msmtRep shared.ChMsmt2Ctrl) *shared.DataObj {
	fmt.Println("\n------------- Constructing MSMT_INFO_REPLY -------------")

	repDataObj := new(shared.DataObj)
	repDataObj.Type = shared.MEASUREMENT_INFO_REPLY
	repDataObj.Status = msmtRep.Status
	repDataObj.Id = ID
	repDataObj.Seq_rp = reqDataObj.Seq

	msmtData, ok := msmtRep.Data.([]shared.DataResultObj)
	if ok == false {
		fmt.Printf("Type assertion failed: Looking for []shared.DataResultObj %t", ok)
		os.Exit(1)
	}

	repDataObj.Measurement_id = reqDataObj.Measurement_id
	repDataObj.Data.DataElements = msmtData

	// TODO: Add the timestamps
	return repDataObj
}

func supportedModules() map[string]string {
	fmt.Println("\nConstructing supported modules")
	supportedMods := make(map[string]string)
	supportedMods["udp-goodput"] = "supported"
	supportedMods["tcp-goodput"] = "supported"
	supportedMods["quic-goodput"] = "not-supported"
	// TODO: Bring that up to data
	return supportedMods
}
