package main

import "fmt"
import "flag"
import "github.com/monfron/mapago/ctrl/server/tcp"
import "github.com/monfron/mapago/ctrl/shared"

var CTRL_PORT = 64321

func main() {
	portPtr := flag.Int("port", CTRL_PORT, "port for interacting with control channel")

	flag.Parse()

	fmt.Println("mapago(c) - 2018")
	fmt.Println("Server side")
	fmt.Println("Port:", *portPtr)

	run_server(*portPtr)
}

func run_server(port int) {
	ch := make(chan shared.ChResult)
	fmt.Println(ch)

	tcpObj := tcp_server.NewTcpObj("TcpConn1")
	tcpObj.Start(ch)

	// WIP...
}