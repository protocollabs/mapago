package main

import "fmt"
import "flag"
import "github.com/monfron/mapago/controlPlane/cmd/server"

var CTRL_PORT = 64321
var DEF_BUFFER_SIZE = 8096 * 8

func main() {
	portPtr := flag.Int("port", CTRL_PORT, "port for interacting with control channel")
	callSizePtr := flag.Int("call-size", DEF_BUFFER_SIZE, "application buffer in bytes")
	lUcAddrPtr := flag.String("uc-listen-addr", "127.0.0.1", "unicast addr where to listen on")
	lMcAddrPtr := flag.String("mc-listen-addr", "224.0.0.1", "multicast addr where to listen on")

	flag.Parse()

	fmt.Println("mapago(c) - 2018")
	fmt.Println("Server side")
	fmt.Println("Unicast Listen-Addr:", *lUcAddrPtr)
	fmt.Println("Multicast Listen-Addr:", *lMcAddrPtr)
	fmt.Println("Port:", *portPtr)
	fmt.Println("Call-Size:", *callSizePtr)

	server.RunServer(*lUcAddrPtr, *lMcAddrPtr, *portPtr, *callSizePtr)
}
