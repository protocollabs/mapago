package server

import "fmt"
import "os"
import "github.com/monfron/mapago/controlPlane/ctrl/serverProtos"
import "github.com/monfron/mapago/controlPlane/ctrl/shared"

var CTRL_PORT = 64321
var DEF_BUFFER_SIZE = 8096 * 8
var ID string
var ARCH string
var OS string
var MODULES string

func RunServer(lUcAddr string, lMcAddr string, port int, callSize int) {
	var repDataObj *shared.DataObj
	// debug: count info_reqs
	info_req_ctr := 0

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
				info_req_ctr++
				fmt.Println("Number of received info request: ", info_req_ctr)

				repDataObj = constructInfoReply(reqDataObj)
				json := shared.ConvDataStructToJson(repDataObj)
				request.ConnObj.WriteAnswer(json)
				// ATM: This closes only the TCP ACCEPT sock
				// the UDP (server) socket is untouched 
				request.ConnObj.CloseConn()

				/* 
				My idea was to create several tcpObjs during runtime,
				instead of "reconfiguring" the existing tcpObj every time
				tcpMsmtObj := serverProtos.NewTcpObj("TcpConnMeasurement", lUcAddr, port, callSize)
				tcpMsmtObj.Start(ch)
				*/

				// Be ready for receiving another control message via TCP...
				go tcpObj.HandleTcpConn(ch)
			}()

		case shared.INFO_REPLY:
			fmt.Println("I received an INFO_REPLY, I am ignoring this")
			continue

		case shared.MEASUREMENT_START_REQUEST:
			go func() {
				fmt.Println("\n------------- Measurement Start Request -------------")

				// TODO 1: Mache rec_ch (empfange Solution)
				rec_ch := make(chan shared.ChMsmt2Ctrl)
				fmt.Println(rec_ch)

				// TODO 2: main.handle_msmt_start_req(rec_ch, client_ip, proto, dataObj)

				// TODO 3: warte auf antwort von messungs modul

				// TODO 4: Speichere aufgemachte verbindung unter UID

				// TODO 5: schreibe über aufgemachte verbindunng raus
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

func supportedModules() map[string]string {
	fmt.Println("\nConstructing supported modules")
	supportedMods := make(map[string]string)
	supportedMods["udp-goodput"] = "no-support"
	supportedMods["tcp-goodput"] = "no-support"
	return supportedMods
}
