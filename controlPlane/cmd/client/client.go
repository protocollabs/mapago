package client

import "fmt"
import "os"
import "github.com/monfron/mapago/controlPlane/ctrl/clientProtos"
import "github.com/monfron/mapago/controlPlane/ctrl/shared"
import "errors"

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
	repDataObj := tcpObj.Start(reqJson)

	err := validateDiscovery(reqDataObj, repDataObj)
	if err != nil {
		fmt.Printf("TCP Discovery phase failed: %s\n", err)
		os.Exit(1)
	}
	// NEXT STEP Start Measurement
	// sendTcpMeasurementStartRequest(addr, port, callSize)
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
	// furthermore: Measurement_delay
	// further more: Measurement_time_max string

	reqJson := shared.ConvDataStructToJson(reqDataObj)
	// debug fmt.Printf("\nrequest JSON is: % s", reqJson)

	// Note: A better naming would be StartDiscoveryPhase()
	repDataObj := tcpObj.Start(reqJson)
	// WIP
	fmt.Println("repDataObj is: ", repDataObj)
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
	// sendUdpMeasurementStartRequest(addr, port, callSize)
}

func sendUdpMeasurementStartRequest(addr string, port int, callSize int) {
	udpObj := clientProtos.NewUdpObj("UdpMeasurementConn", addr, port, callSize)

	// TODO: build json "dummy" message
	reqDataObj := new(shared.DataObj)
	reqDataObj.Type = shared.MEASUREMENT_START_REQUEST
	reqDataObj.Id = shared.ConstructId()
	reqDataObj.Seq = "1"
	reqDataObj.Secret = "fancySecret"
	// furthermore: Measurement_delay
	// further more: Measurement_time_max string

	reqJson := shared.ConvDataStructToJson(reqDataObj)
	// debug fmt.Printf("\nrequest JSON is: % s", reqJson)

	// Note: A better naming would be StartDiscoveryPhase()
	repDataObj := udpObj.Start(reqJson)
	// WIP
	fmt.Println("repDataObj is : ", repDataObj)
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
