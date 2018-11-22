package main

import "fmt"
import "flag"
import "github.com/monfron/mapago/controlPlane/cmd/server"

var CTRL_PORT = 64321
var DEF_BUFFER_SIZE = 8096 * 8

func main() {
	portPtr := flag.Int("port", CTRL_PORT, "port for interacting with control channel")
	callSizePtr := flag.Int("call-size", DEF_BUFFER_SIZE, "application buffer in bytes")
	lAddrPtr := flag.String("listen-addr", "[::]", "addr where to listen on")

	flag.Parse()

	fmt.Println("mapago(c) - 2018")
	fmt.Println("Server side")
	fmt.Println("Listen-Addr:", *lAddrPtr)
	fmt.Println("Port:", *portPtr)
	fmt.Println("Call-Size:", *callSizePtr)

	server.RunServer(*lAddrPtr, *portPtr, *callSizePtr)
}
