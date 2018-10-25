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

	dataObj := shared.DataObj{
							  Type: 0x01, 
							  Id: "1337",
							  Seq: 1337,
							  Ts: "2018-10-23T 16:31:13.720069",
							  Secret: "fancySecret"}

	json := convDataStructToJson(&dataObj)
	dataObRecovered := convJsonToDataStruct(json)

	fmt.Println("\ndataObj is: ", *dataObRecovered)
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
