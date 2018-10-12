package serverProtos

import "fmt"
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
// separate interface declaration and implementation?

// methods

func (tcp *TcpObj) Start(ch chan<- shared.ChResult) {
	fmt.Println("TcpObj start() called!!!!!")
	go tcp.handleTcpConn(ch)

}

func (tcp *TcpObj) handleTcpConn(ch chan<- shared.ChResult) {
	fmt.Println("handleTcpConn goroutine called")
}
