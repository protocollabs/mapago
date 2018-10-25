package main

import "fmt"
import "flag"
import "encoding/json"
import "encoding/binary"
import "os"
import "github.com/monfron/mapago/ctrl/serverProtos"
import "github.com/monfron/mapago/ctrl/shared"

var CTRL_PORT = 64321
var DEF_BUFFER_SIZE = 8096 * 8

func main() {
	portPtr := flag.Int("port", CTRL_PORT, "port for interacting with control channel")
	callSizePtr := flag.Int("call-size", DEF_BUFFER_SIZE, "application buffer in bytes")


	flag.Parse()

	fmt.Println("mapago(c) - 2018")
	fmt.Println("Server side")
	fmt.Println("Port:", *portPtr)
	fmt.Println("Call-Size:", *callSizePtr)

	runServer(*portPtr, *callSizePtr)
}

func runServer(port int, callSize int) {
	ch := make(chan shared.ChResult)
	fmt.Println(ch)

	tcpObj := serverProtos.NewTcpObj("TcpConn1", port, callSize)
	tcpObj.Start(ch)

	/* WIP: disabled for reduced complexity
	udpObj := serverProtos.NewUdpObj("UdpConn1")
	udpObj.Start(ch)
	*/

	for {
		request := <- ch
		fmt.Printf("Server received from client: % x", request)

		repDataObj := new(shared.DataObj)
		// TODO we have to cut the received JSON
		// (JSON is for example only 76 bytes
		// but application buffer read is larger)
		// or rsult in unmarshaling error => HARDCODED ATM
		reqDataObj := convJsonToDataStruct(request.Json[:76])

		switch reqDataObj.Type {
		case shared.INFO_REQUEST:
			fmt.Println("Construct INFO_REP")
			// POSSIBLE AS A SEPARATE FUNC
			// i.e. constructInfoReply(repDataObj) etc.
			// not yet for complexity
			repDataObj.Type = shared.INFO_REPLY
			repDataObj.Id = "fancyId"
			repDataObj.Seq_rp = reqDataObj.Seq
			// repDataObj.modules
			// repDataObj.Arch
			// repDataObj.Os
			repDataObj.Info = "fancyInfo"
		case shared.MEASUREMENT_START_REQUESTS:
			fmt.Println("Construct MEASUREMENT_START_REP")

		case shared.MEASUREMENT_STOP_REQUEST:
			fmt.Println("Construct MEASUREMENT_STOP_REP")

		case shared.MEASUREMENT_INFO_REQUEST:
			fmt.Println("Construct MEASUREMENT_INFO_REP")

		case shared.TIME_DIFF_REQUEST:
			fmt.Println("Construct TIME_DIFF_REP")

		case shared.WARNING_ERR_MSG:
			fmt.Println("WARNING_ERR_MSG")

		default:
			fmt.Printf("Unknown type")
			os.Exit(1)
		}

		json := convDataStructToJson(repDataObj)
		request.ConnObj.WriteAnswer(json)
	}
}

func convJsonToDataStruct(jsonData []byte) *shared.DataObj {
	fmt.Printf("\n Converting json data: % x", jsonData)

	dataObj := new(shared.DataObj)

	// FIXED
	// extract type field and add to struct
	typeField := jsonData[0:2]
	dataObj.Type = uint64(binary.BigEndian.Uint16(typeField))

	jsonB :=  jsonData[2:]
	err := json.Unmarshal(jsonB, dataObj)
	if err != nil {
		fmt.Printf("Cannot Unmarshal %s\n", err)
		os.Exit(1)
	}

	return dataObj
}

func convDataStructToJson(data *shared.DataObj) []byte {
	fmt.Println("\nConverting datastruct ", *data)

	var resB []byte

	// construct type field
	typeB := make([]byte, 2)
	binary.BigEndian.PutUint16(typeB[0:2], uint16(data.Type))

	// ignore Type
	data.Type = 0

	jsonB, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("Cannot Marshal %s\n", err)
		os.Exit(1)
	}

	resB = append(resB, typeB...)
	resB = append(resB, jsonB...)
	return resB
}
