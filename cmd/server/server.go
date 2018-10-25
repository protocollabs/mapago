package main

import "fmt"
import "flag"
import "encoding/json"
import "encoding/binary"
import "os"
// WIP removed during JSON test import "github.com/monfron/mapago/ctrl/serverProtos"
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

	/* WIP removed during JSON test
	tcpObj := serverProtos.NewTcpObj("TcpConn1", port, callSize)
	tcpObj.Start(ch)
	*/

	/* WIP: disabled for reduced complexity
	udpObj := serverProtos.NewUdpObj("UdpConn1")
	udpObj.Start(ch)
	*/

	dummyJson := constructDummyJson()
	dataObj := convJsonToDataStruct(dummyJson)
	switch dataObj.Type {
	case shared.INFO_REQUEST:
		fmt.Println("INFO_REQUEST")
	case shared.MEASUREMENT_START_REQUESTS:
		fmt.Println("MEASUREMENT_START_REQUESTS")
	case shared.MEASUREMENT_STOP_REQUEST:
		fmt.Println("MEASUREMENT_STOP_REQUEST")
	case shared.MEASUREMENT_INFO_REQUEST:
		fmt.Println("MEASUREMENT_INFO_REQUEST")
	case shared.TIME_DIFF_REQUEST:
		fmt.Println("TIME_DIFF_REQUEST")
	case shared.WARNING_ERR_MSG:
		fmt.Println("WARNING_ERR_MSG")
	default:
		fmt.Printf("Unknown type")
		os.Exit(1)

	}

	/* WIP removed during JSON test
	for {
		result := <- ch
		fmt.Println("Server received from client: ", result)

		// TODO: here will go the conJsonToDataStruct in the future
		dataObj := conJsonToDataStruct(result.Json)

		result.ConnObj.WriteAnswer([]byte("ServerReply"))
	}
	*/
}

func convJsonToDataStruct(jsonData []byte) *shared.DataObj {
	fmt.Printf("\nProcessing the following data: % x", jsonData)

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
	fmt.Println("convert data struct to json (i.e. prepare outgoing msg")
	return []byte("ShutUpGo")
}




// this is a shrinked version of convDataStructToJson
// use it to play around...
// note "Json" is not precise: includes type field aswell
func constructDummyJson() []byte {
	var resB []byte
	// construct DataObj that will be the actual JSON
	// in this case construct info req:
	dataObj := shared.DataObj{Id: "1337",
							  Seq: 1337,
							  Ts: "2018-10-23T 16:31:13.720069",
							  Secret: "fancySecret"}

	jsonB, err := json.Marshal(dataObj)
	if err != nil {
		fmt.Printf("Cannot Marshal %s\n", err)
		os.Exit(1)
	}

	// now take a byte slice add the type field 0x01
	typeB := make([]byte, 2)
	binary.BigEndian.PutUint16(typeB[0:2], 0x01)

	// add type field to result
	resB = append(resB, typeB...)

	// add json field to result
	resB = append(resB, jsonB...)

	return resB
}