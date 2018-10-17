package clientProtos

import "fmt"
import "net"
import "strconv"
import "os"
import "reflect"
import "github.com/monfron/mapago/ctrl/shared"

// classes

type TcpObj struct {
	connName string
	connAddr string
	connPort int
	connCallSize int
}

type TcpConnObj struct {
	connSock *net.TCPConn
}

// constructors:
func NewTcpObj(name string, addr string, port int, callSize int) *TcpObj {
	tcpObj := new(TcpObj)
	tcpObj.connName = name
	tcpObj.connAddr = addr
	tcpObj.connPort = port
	tcpObj.connCallSize = callSize
	return tcpObj
}

func NewTcpConnObj(tcpSock *net.TCPConn) *TcpConnObj {
	tcpConnObj := new(TcpConnObj)
	// when TcpConnObj is created,
	// initialize it with accepted Socket
	tcpConnObj.connSock = tcpSock
	return tcpConnObj
}

func (tcp *TcpObj) Start(ch chan<- shared.ChResult) {
	fmt.Println("TcpObj start() called")

	rAddr := tcp.connAddr + ":" + strconv.Itoa(tcp.connPort)
	fmt.Println("FullAddr is: ", rAddr)

	rTcpAddr, err := net.ResolveTCPAddr("tcp", rAddr)
	if err != nil {
		fmt.Printf("Cannot parse \"%s\": %s\n", rAddr, err)
		os.Exit(1)
	}

	tcpConn, err := net.DialTCP("tcp", nil, rTcpAddr)
	if err != nil {
		fmt.Printf("Cannot dial \"%s\": %s\n", rAddr, err)
		os.Exit(1)
	}

	//go tcp.handleTcpConn(ch, tcpConn)
	tcp.handleTcpConn(ch, tcpConn)
}

func (tcp *TcpObj) handleTcpConn(ch chan<- shared.ChResult, conn *net.TCPConn) {
	fmt.Println("\n\nGoroutine here!!!!")

	buf := make([]byte, tcp.connCallSize, tcp.connCallSize)

	tcpConn := NewTcpConnObj(conn)
	fmt.Println("tcpConn is of type: ", reflect.TypeOf(tcpConn))

	for {
		_, err := tcpConn.connSock.Write([]byte("testest!!!"))
		if err != nil {
			fmt.Printf("Cannot send!!! %s\n", err)
			os.Exit(1)
		}
		fmt.Println("sending!!!")
	}

	defer tcpConn.connSock.Close()
	fmt.Println("buf is of type: ", reflect.TypeOf(buf))
	fmt.Println("tcpConn is: ", tcpConn)
}

