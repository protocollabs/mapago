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

	// construct apriori
	ID = shared.ConstructId()
	OS = shared.DetectOs()
	ARCH = shared.DetectArch()
	MODULES = shared.ConvMapToStr(supportedModules())

	ch := make(chan shared.ChResult)

	tcpDiscObj := serverProtos.NewTcpObj("TcpConnDiscovery", lUcAddr, port, callSize)
	tcpDiscObj.Start(ch)

	udpDiscObj := serverProtos.NewUdpObj("UdpConnDiscovery", lUcAddr, port, callSize)
	udpDiscObj.Start(ch)

	// PROBLEM: Binding to MC_Addr:CtrlPort and UC_Addr:CtrlPort does not work
	// eventough we got a unique addressing going on
	udpMcDiscObj := serverProtos.NewUdpMcObj("UdpMcConnDiscovery", lMcAddr, 12346, callSize)
	// udpMcObj := serverProtos.NewUdpMcObj("UdpMcConn1", lMcAddr, port, callSize)
	udpMcDiscObj.Start(ch)

	// TODO: Make storage for reacting to multiple Measurement TCP conns
	var tcpMsmtObjStorage []*serverProtos.TcpObj

	for {
		request := <-ch
		reqDataObj := shared.ConvJsonToDataStruct(request.Json)

		switch reqDataObj.Type {
		case shared.INFO_REQUEST:
			go func() {
				// NOTE: we need "go" here: blocking calls
				repDataObj = constructInfoReply(reqDataObj)
				json := shared.ConvDataStructToJson(repDataObj)
				request.ConnObj.WriteAnswer(json)
				// deconstruct connection after discovery phase finished
				request.ConnObj.CloseConn()

				// TODO 0: open new tcp connection for controlling measurement phase
				// pay attention to scope!
				// NOT SURE HERE
				tcpMsmtObj := serverProtos.NewTcpObj("TcpConnMeasurement", lUcAddr, port, callSize)
				tcpMsmtObj.Start(ch)

				// NOTE: what happens if we got too many INFO_REQUESTs?
				// tcpObj is always overwritten
				// result: we can handle just a single measurement start request
				// save it in order to process several
				tcpMsmtObjStorage = append(tcpMsmtObjStorage, tcpMsmtObj)
			}()

		case shared.INFO_REPLY:
			fmt.Println("I received an INFO_REPLY, I am ignoring this")
			continue

		// interaction needed
		case shared.MEASUREMENT_START_REQUESTS:
			// NOTE: after (possible unreliable) discovery: we open reliable conn
			// for control communication
			go func() {

				// TODO 1: Mache rec_ch (empfange Solution)

				// TODO 2: main.handle_msmt_start_req(rec_ch, client_ip, proto, dataObj)

				// TODO 3: warte auf antwort von messungs modul

				// TODO 4: Speichere aufgemachte verbindung unter UID

				// TODO 5: schreibe über aufgemachte verbindunng raus
				fmt.Println("Construct MEASUREMENT_START_REP")
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
