package main

import "fmt"
import "flag"
import "time"
import "github.com/monfron/mapago/ctrl/serverProtos"
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

	tcpObj := serverProtos.NewTcpObj("TcpConn1")
	tcpObj.Start(ch)

	start := time.Now()
	for {
		fmt.Println("try to receive send something from channel")
		result := <- ch
		fmt.Println("Result: ", result)

		result.ConnObj.WriteAnswer([]byte("tcpStuffSent"))
				
		time.Sleep(2 * time.Millisecond)

		elapsed := time.Since(start)
		fmt.Println("Elapsed time recv channel: ", elapsed)
		start = time.Now()
	}
}