package serverProtos

import "fmt"
import "time"
import "github.com/monfron/mapago/ctrl/shared"

// classes

type UdpObj struct {
	connName string
}

type UdpConnObj struct {
}

// Constructors

func NewUdpObj(name string) *UdpObj {
	udpObj := new(UdpObj)
	udpObj.connName = name
	return udpObj
}

func NewUdpConnObj() *UdpConnObj {
	udpConnObj := new(UdpConnObj)
	return udpConnObj
}

// TODO Interfaces
func (udpConn *UdpConnObj) WriteAnswer(answer []byte) {
	fmt.Println(answer)
}

// module start
func (udp *UdpObj) Start(ch chan<- shared.ChResult) {
	fmt.Println("UdpObj start() called")
	go udp.handleUdpConn(ch)
}

func (udp *UdpObj) handleUdpConn(ch chan<- shared.ChResult) {
	fmt.Println("handleUdpConn goroutine called")

	udpConn := NewUdpConnObj()
	fmt.Println("udp conn: ", *udpConn)

	start := time.Now()
	for {
		fmt.Println("\nUDP: Sending into channel")
		chReply := new(shared.ChResult)
		chReply.Json = []byte("UdpStuffReceived")
		chReply.ConnObj = udpConn

		ch <- *chReply

		time.Sleep(1 * time.Millisecond)
		elapsed := time.Since(start)
		fmt.Println("UDP: Elapsed time sending channel: ", elapsed)
		start = time.Now()
	}
}