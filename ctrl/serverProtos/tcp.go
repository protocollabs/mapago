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
	// when TcpConnObj is created,
	// initialise it with accepted Socket
	tcpConnObj.connAcceptSock = tcpAcceptSock
	return tcpConnObj
}

// TODO implement interface from shared code
func (tcpConn *TcpConnObj) WriteAnswer(answer []byte) {
	fmt.Println(answer)
}

// methods
func (tcp *TcpObj) Start(ch chan<- shared.ChResult) {
	fmt.Println("TcpObj start() called")

	listenAddr := "[::]:" + strconv.Itoa(tcp.connPort)
	// fmt.Println("Listening on: ", listenAddr)

	tcpAddr, err := net.ResolveTCPAddr("tcp", listenAddr)
	if err != nil {
		fmt.Printf("Cannot parse \"%s\": %s\n", listenAddr, err)
		os.Exit(1)
	}

	// fmt.Println("tcpAddr is: ", tcpAddr)
	tcpListener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		fmt.Printf("Cannot Listen \"%s\"", err)
		os.Exit(1)
	}
	defer tcpListener.Close()

	tcp.connSrvSock = tcpListener

	// until here: block until accepted conn established
	tcpConn, err := tcpListener.AcceptTCP()
	if err != nil {
		fmt.Printf("Cannot accept: %s\n", err)
		os.Exit(1)
	}
	defer tcpConn.Close()

	go tcp.handleTcpConn(ch, tcpConn)
}

func (tcp *TcpObj) handleTcpConn(ch chan<- shared.ChResult, tcpAccepted *net.TCPConn) {
	fmt.Println("handleTcpConn goroutine called")

	buf := make([]byte, tcp.connCallSize, tcp.connCallSize)
	tcpConn := NewTcpConnObj(tcpAccepted)

	for {
		n, err := tcpConn.connAcceptSock.Read(buf)

		if err != nil {
			fmt.Printf("Cannot read msg: %s\n", err)
			os.Exit(1)
		}
		// debug output, will be removed
		fmt.Printf("read %i", n, " bytes ")

		chRequest := new(shared.ChResult)
		chRequest.ConnObj = tcpConn
		chRequest.Json = buf

		fmt.Printf("Sending into channel\n")
		ch <- *chRequest
	}
}
