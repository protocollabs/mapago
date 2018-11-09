package serverProtos

import "fmt"
import "net"
import "strconv"
import "os"
import "github.com/monfron/mapago/ctrl/shared"

// classes

type TcpObj struct {
	connName     string
	connAddr     string
	connSrvSock  *net.TCPListener
	connPort     int
	connCallSize int
}

type TcpConnObj struct {
	connAcceptSock *net.TCPConn
}

// constructors:
func NewTcpObj(name string, laddr string, port int, callSize int) *TcpObj {
	tcpObj := new(TcpObj)
	tcpObj.connName = name
	tcpObj.connAddr = laddr
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

	listenAddr := tcp.connAddr + ":" + strconv.Itoa(tcp.connPort)

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

	go tcp.handleTcpConn(ch)
}

func (tcp *TcpObj) handleTcpConn(ch chan<- shared.ChResult) {
	fmt.Println("handleTcpConn goroutine called")

	buf := make([]byte, tcp.connCallSize, tcp.connCallSize)

	// note: block until accepted conn established
	tcpAccepted, err := tcp.connSrvSock.AcceptTCP()
	if err != nil {
		fmt.Printf("Cannot accept: %s\n", err)
		os.Exit(1)
	}

	tcpConn := NewTcpConnObj(tcpAccepted)

	defer tcp.connSrvSock.Close()
	// its ok here: this will be executed AFTER client send
	// conn teardown => i.e. for loop break
	defer tcpConn.connAcceptSock.Close()

	for {
		bytes, err := tcpConn.connAcceptSock.Read(buf)

		if err != nil {
			fmt.Printf("Cannot read!!!! msg: %s\n", err)
			os.Exit(1)
		}

		fmt.Println("TCP Server read num bytes: ", bytes)

		chRequest := new(shared.ChResult)
		chRequest.ConnObj = tcpConn
		chRequest.Json = buf[:bytes]
		fmt.Printf("TCP: Sending into channel\n")
		ch <- *chRequest
	}
}
