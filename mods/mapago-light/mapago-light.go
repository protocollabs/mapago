package main 

import "fmt"
import "flag"

var CTRL_PORT = 64321

func main() {
	modePtr := flag.String("mode", "server", "server or client")
	portPtr := flag.Int("port", CTRL_PORT, "port for interacting with control channel")
	protoPtr := flag.String("protocol", "udp_uc", "protocol for discovering mapago servers: udp_uc, udp_mc or tcp")
	addrPtr := flag.String("addr", "ipv4_uc", "target addr: ipv4_uc, ipv4_mc, ipv6_uc, ipv6_mc") 

	flag.Parse()

	fmt.Println("mapago-light(c) - 2018")
	fmt.Println("Mode:", *modePtr)
	fmt.Println("Port:", *portPtr)
	fmt.Println("Protocol:", *protoPtr)
	fmt.Println("Address:", *addrPtr)

	if *modePtr == "server" {
		run_server(*portPtr)
	} else if *modePtr == "client" {
		run_client(*portPtr, *addrPtr, *protoPtr)
	} else {
		panic("mode not supported! chose server or client")
	}
}

func run_server(port int) {
	c := make(chan []byte)
	protos := supported_disco_protos()

	fmt.Println("supported protos: ", protos)

	// listen for incoming discovery messages
	for _, discoProto := range protos {
		go listen_info_req(discoProto, port, c)
	}

	// process incoming discovery messages
	// (receive via chan, JSON decoding etc.)
	for i := 0; i < len(protos); i++ {
		jsonInfoReq := <- c
		fmt.Println("info req JSON data: ", jsonInfoReq)

		// TODO: JSON parser func processing jsonInfoReq

		// TODO: create_info_reply()

		// TODO: send_info_reply() (as unicast)
	}
}

func supported_disco_protos() []string {
	var discoProtos []string
	discoProtos = append(discoProtos, "udp4_uc", "udp6_uc")
	discoProtos = append(discoProtos, "udp4_mc", "udp6_mc")
	discoProtos = append(discoProtos, "tcp4_uc", "tcp6_uc")
	// placeholder for possible QUIC support

	return discoProtos
}

func listen_info_req(discoProto string, port int, c chan<- []byte) {
	// TODO: udp4_uc handler func (i.e. listen, parse and return val)

	// TODO: udp6_uc handler func (i.e. listen, parse and return val)

	// TODO: udp4_mc handler func (i.e. listen, parse and return val)

	// TODO: udp6_mc handler func (i.e. listen, parse and return val)

	// TODO: tcp4_uc handler func (i.e. listen, parse and return val)

	// TODO: tcp6_uc handler func (i.e. listen, parse and return val)

	// dummy channel "return value"
	jsonBytes := []byte(discoProto)
	c <- jsonBytes
}



func run_client(port int, addr string, proto string) {
	fmt.Println("client handler dummy func")
}