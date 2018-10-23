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

	constructDummyJson()

	/* WIP removed during JSON test
	for {
		result := <- ch
		fmt.Println("Server received from client: ", result)

		result.ConnObj.WriteAnswer([]byte("ServerReply"))
	}
	*/
}

func convJsonToDataStruct(jsonData []byte) *shared.DataObj {
	fmt.Println("convert json to data struct (i.e. process incoming msg")
	dataObj := new(shared.DataObj)
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

	fmt.Println("json encoded version of dataObj: ", string(jsonB))

	// now take a byte slice add the type field 0x01
	// needed: memory representation in big endian!
	fmt.Println("length of json Bytes: ", len(jsonB))

	b := make([]byte, len(jsonB) + 1)

	binary.BigEndian.PutUint16(b[0:], 0x01)

	fmt.Println("length of b: ", len(b))
	fmt.Printf("content of b: % x", b)


	return []byte("ShutUpGo")
}