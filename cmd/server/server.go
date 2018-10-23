package main

import "fmt"
import "flag"
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

	run_server(*portPtr, *callSizePtr)
}

func run_server(port int, callSize int) {
	ch := make(chan shared.ChResult)
	fmt.Println(ch)

	tcpObj := serverProtos.NewTcpObj("TcpConn1", port, callSize)
	tcpObj.Start(ch)

	/* WIP: disabled for reduced complexity
	udpObj := serverProtos.NewUdpObj("UdpConn1")
	udpObj.Start(ch)
	*/

	for {
		result := <- ch
		fmt.Println("Server received from client: ", result)

		/* WIP: disabled for reduced complexity
		dataObj := shared.NewDataObj()
		fmt.Println("dataObj is: ", dataObj)
		dataObj.TransformJson(result.Json)
		fmt.Println("Val of current data obj: ", dataObj)
		*/

		result.ConnObj.WriteAnswer([]byte("ServerReply"))
	}
}