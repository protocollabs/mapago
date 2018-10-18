package clientProtos

import "fmt"
import "net"
import "strconv"
import "os"
import "sync"

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
	tcpConnObj.connSock = tcpSock
	return tcpConnObj
}

func (tcp *TcpObj) Start() {
	fmt.Println("TcpObj start() called")
	var wg sync.WaitGroup

	rAddr := tcp.connAddr + ":" + strconv.Itoa(tcp.connPort)
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

	wg.Add(1)
	go tcp.handleTcpConn(tcpConn, wg)
	wg.Wait()
}

func (tcp *TcpObj) handleTcpConn(conn *net.TCPConn, wg sync.WaitGroup) {
	fmt.Println("\n\nGoroutine here!!!!")
	defer wg.Done()

	buf := make([]byte, tcp.connCallSize, tcp.connCallSize)
	buf = []byte(`{"Type":02,"Id":"0x1337","MsmtId":1337,"Seq":00}`)

	tcpConn := NewTcpConnObj(conn)
	defer tcpConn.connSock.Close()

	for {
		_, err := tcpConn.connSock.Write(buf)
		if err != nil {
			fmt.Printf("Cannot send!!! %s\n", err)
			os.Exit(1)
		}
	}
}

