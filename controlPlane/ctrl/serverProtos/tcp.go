package serverProtos

import "fmt"
import "net"
import "strconv"
import "os"
import "github.com/monfron/mapago/controlPlane/ctrl/shared"
import "io"

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

func (tcpConn *TcpConnObj) CloseConn() {
	err := tcpConn.connAcceptSock.Close()
	if err != nil {
		fmt.Printf("Cannot close conn %s\n", err)
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

	for {
		bytes, err := tcpConn.connAcceptSock.Read(buf)

		// ok we have error condition
		if err != nil {
			if err == io.EOF {
				fmt.Println("\nEOF detected")
				break
			}

			if err.(*net.OpError).Err.Error() == "use of closed network connection" {
				fmt.Println("\nClosed network detected!")
				break
			}

			// something different serious...
			fmt.Printf("\nCannot read from TCP! msg: %s\n", err)
			os.Exit(1)
		}

		fmt.Println("\nTCP Server read num bytes: ", bytes)

		chRequest := new(shared.ChResult)
		chRequest.ConnObj = tcpConn
		chRequest.Json = buf[:bytes]

		fmt.Printf("\nTCP: Sending into channel\n")
		ch <- *chRequest
	}
}
