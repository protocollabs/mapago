package serverProtos

import "fmt"
import "time"
import "github.com/monfron/mapago/ctrl/shared"

// classes

type TcpObj struct {
	connName string
}

type TcpConnObj struct {
}

// constructors:
func NewTcpObj(name string) *TcpObj {
	tcpObj := new(TcpObj)
	tcpObj.connName = name
	return tcpObj
}

func NewTcpConnObj() *TcpConnObj {
	tcpConnObj := new(TcpConnObj)
	return tcpConnObj
}

// TODO implement interface from shared code
func (tcpConn *TcpConnObj) WriteAnswer(answer []byte) {
	fmt.Println(answer)
}

// methods

func (tcp *TcpObj) Start(ch chan<- shared.ChResult) {
	fmt.Println("TcpObj start() called")
	go tcp.handleTcpConn(ch)

}

func (tcp *TcpObj) handleTcpConn(ch chan<- shared.ChResult) {
	fmt.Println("handleTcpConn goroutine called")

	tcpConn := NewTcpConnObj()
	fmt.Println("tcp conn: ", *tcpConn)

	start := time.Now()
	for {
		fmt.Println("\nTCP: Sending into channel")
		chReply := new(shared.ChResult)
		chReply.DummyJson = "tcpStuffReceived"
		chReply.ConnObj = tcpConn

		ch <- *chReply

		time.Sleep(1 * time.Millisecond)
		elapsed := time.Since(start)
		fmt.Println("TCP: Elapsed time sending channel: ", elapsed)
		start = time.Now()
	}
}
