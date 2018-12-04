package controlPlane

import "fmt"
import "os"
import "errors"
import "github.com/monfron/mapago/control-plane/ctrl/server-protocols"
import "github.com/monfron/mapago/control-plane/ctrl/client-protocols"
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

/* Client functionality */

func RunTcpClient(addr string, port int, callSize int) {
	// TODO: we need a channel here aswell in the future
	// use case: we receive a server response. using the server response
	// we can determine what next to do. i.e. info rep => do msmt start req etc.
	tcpObj := clientProtos.NewTcpObj("TcpDiscoveryConn", addr, port, callSize)

	// TODO: build json "dummy" message
	reqDataObj := new(shared.DataObj)
	reqDataObj.Type = shared.INFO_REQUEST
	reqDataObj.Id = shared.ConstructId()
	reqDataObj.Seq = "0"
	reqDataObj.Ts = shared.ConvCurrDateToStr()
	reqDataObj.Secret = "fancySecret"
	reqJson := shared.ConvDataStructToJson(reqDataObj)
	// debug fmt.Printf("\nrequest JSON is: % s", reqJson)

	// Note: A better naming would be StartDiscoveryPhase()
	repDataObj := tcpObj.StartDiscovery(reqJson)

	err := validateDiscovery(reqDataObj, repDataObj)
	if err != nil {
		fmt.Printf("TCP Discovery phase failed: %s\n", err)
		os.Exit(1)
	}
	sendTcpMeasurementStartRequest(addr, port, callSize)
}

// Dummy part
func sendTcpMeasurementStartRequest(addr string, port int, callSize int) {
	tcpObj := clientProtos.NewTcpObj("TcpMeasurementConn", addr, port, callSize)

	// TODO: build json "dummy" message
	reqDataObj := new(shared.DataObj)
	reqDataObj.Type = shared.MEASUREMENT_START_REQUEST
	reqDataObj.Id = shared.ConstructId()
	reqDataObj.Seq = "1"
	reqDataObj.Secret = "fancySecret"
	reqDataObj.Measurement_delay = "666"
	reqDataObj.Measurement_time_max = "666"

	msmtObj := constructMeasurementObj("tcp-throughput", "module")
	reqDataObj.Measurement = *msmtObj

	reqJson := shared.ConvDataStructToJson(reqDataObj)
	// debug fmt.Printf("\nrequest JSON is: % s", reqJson)

	repDataObj := tcpObj.StartMeasurement(reqJson)

	// TODO: We have to save the received Measurement_id etc.
	fmt.Println("\nrepDataObj is: ", repDataObj)
}

func constructMeasurementObj(name string, msmt_type string) *shared.MeasurementObj {
	MsmtObj := new(shared.MeasurementObj)
	MsmtObj.Name = name
	MsmtObj.Type = msmt_type

	confObj := constructConfiguration()
	MsmtObj.Configuration = *confObj

	return MsmtObj
}

func constructConfiguration() *shared.ConfigurationObj {
	ConfObj := new(shared.ConfigurationObj)
	// TODO: these are currently dummy params => useful params 
	ConfObj.Config_param1 = "FancyConfigParam1"
	ConfObj.Config_param2 = "FancyConfigParam2"
	ConfObj.Config_param3 = "FancyConfigParam3"

	return ConfObj
}

func RunUdpClient(addr string, port int, callSize int) {
	udpObj := clientProtos.NewUdpObj("UdpConn1", addr, port, callSize)

	// TODO: build json "dummy" message
	reqDataObj := new(shared.DataObj)
	reqDataObj.Type = shared.INFO_REQUEST
	reqDataObj.Id = shared.ConstructId()
	reqDataObj.Seq = "0"
	reqDataObj.Ts = shared.ConvCurrDateToStr()
	reqDataObj.Secret = "fancySecret"

	reqJson := shared.ConvDataStructToJson(reqDataObj)
	repDataObj := udpObj.Start(reqJson)

	err := validateDiscovery(reqDataObj, repDataObj)
	if err != nil {
		fmt.Printf("UDP Discovery phase failed: %s\n", err)
		os.Exit(1)
	}

	// Discovery phase is send as UDP/TCP/Mcast
	// but Measurement phase is TCP
	sendTcpMeasurementStartRequest(addr, port, callSize)
}

func RunUdpMcastClient(addr string, port int, callSize int) {
	udpMcObj := clientProtos.NewUdpMcObj("UdpMcConn1", addr, port, callSize)

	// TODO: build json "dummy" message
	reqDataObj := new(shared.DataObj)
	reqDataObj.Type = shared.INFO_REQUEST
	reqDataObj.Id = shared.ConstructId()
	reqDataObj.Seq = "0"
	reqDataObj.Ts = shared.ConvCurrDateToStr()
	reqDataObj.Secret = "fancySecret"

	reqJson := shared.ConvDataStructToJson(reqDataObj)
	repDataObj := udpMcObj.Start(reqJson)

	err := validateDiscovery(reqDataObj, repDataObj)
	if err != nil {
		fmt.Printf("UDP MC Discovery phase failed: %s\n", err)
		os.Exit(1)
	}

	// Discovery phase is send as UDP/TCP/Mcast
	// but Measurement phase is TCP
	sendTcpMeasurementStartRequest(addr, port, callSize)
}

func validateDiscovery(req *shared.DataObj, rep *shared.DataObj) error {
	if rep.Type != shared.INFO_REPLY {
		return errors.New("Received message is not INFO_REPLY")
	}

	if rep.Seq_rp != req.Seq {
		return errors.New("Wrong INFO_REQUEST handled by srv")
	}

	fmt.Println("\nDiscovery phase finished. Connected to: ", rep.Id)
	return nil
}
