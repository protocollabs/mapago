package serverProtos

import "fmt"
import "net"
import "strconv"
import "os"
import "github.com/monfron/mapago/ctrl/shared"

// classes

type TcpObj struct {
	connName string
	connSrvSock *net.TCPListener
	connPort int
	connCallSize int
}

type TcpConnObj struct {
	connAcceptSock *net.TCPConn
}

// constructors:
func NewTcpObj(name string, port int, callSize int) *TcpObj {
	tcpObj := new(TcpObj)
	tcpObj.connName = name
	tcpObj.connPort = port
	tcpObj.connCallSize = callSize
	return tcpObj
}

func NewTcpConnObj(tcpAcceptSock *net.TCPConn) *TcpConnObj {
	tcpConnObj := new(TcpConnObj)
	tcpConnObj.connAcceptSock = tcpAcceptSock
	return tcpConnObj
}

func (tcpConn *TcpConnObj) WriteAnswer(answer []byte) {
	_, err := tcpConn.connAcceptSock.Write(answer)
	if err != nil {
		fmt.Printf("Cannot send/write %s\n", err)
		os.Exit(1)
	}
}

// methods
func (tcp *TcpObj) Start(ch chan<- shared.ChResult) {
	fmt.Println("TcpObj start() called")

	listenAddr := "[::]:" + strconv.Itoa(tcp.connPort)

	tcpAddr, err := net.ResolveTCPAddr("tcp", listenAddr)
	if err != nil {
		fmt.Printf("Cannot parse \"%s\": %s\n", listenAddr, err)
		os.Exit(1)
	}

	tcpListener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		fmt.Printf("Cannot Listen \"%s\"", err)
		os.Exit(1)
	}

	tcp.connSrvSock = tcpListener

	// note: block until accepted conn established
	tcpConn, err := tcpListener.AcceptTCP()
	if err != nil {
		fmt.Printf("Cannot accept: %s\n", err)
		os.Exit(1)
	}

	go tcp.handleTcpConn(ch, tcpConn)
}

func (tcp *TcpObj) handleTcpConn(ch chan<- shared.ChResult, tcpAccepted *net.TCPConn) {
	fmt.Println("handleTcpConn goroutine called")

	buf := make([]byte, tcp.connCallSize, tcp.connCallSize)
	tcpConn := NewTcpConnObj(tcpAccepted)

	defer tcp.connSrvSock.Close()
	defer tcpConn.connAcceptSock.Close()

	for {
		_, err := tcpConn.connAcceptSock.Read(buf)

		if err != nil {
			fmt.Printf("Cannot read!!!! msg: %s\n", err)
			os.Exit(1)
		}

		chRequest := new(shared.ChResult)
		chRequest.ConnObj = tcpConn
		chRequest.Json = buf

		fmt.Printf("Sending into channel\n")
		ch <- *chRequest
	}
}
