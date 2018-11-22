package clientProtos

import "fmt"
import "net"
import "strconv"
import "os"
import "github.com/monfron/mapago/controlPlane/ctrl/shared"

// classes

type TcpObj struct {
	connName     string
	connAddr     string
	connPort     int
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

// Note: A better naming would be StartDiscoveryPhase
func (tcp *TcpObj) Start(jsonData []byte) *shared.DataObj {
	var repDataObj *shared.DataObj

	fmt.Println("TcpObj start() called")
	buf := make([]byte, tcp.connCallSize, tcp.connCallSize)

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

	tcpConnObj := NewTcpConnObj(tcpConn)
	defer tcpConnObj.connSock.Close()

	for {
		_, err := tcpConnObj.connSock.Write(jsonData)
		if err != nil {
			fmt.Printf("Cannot send!!! %s\n", err)
			os.Exit(1)
		}

		bytes, err := tcpConnObj.connSock.Read(buf)
		if err != nil {
			fmt.Printf("Cannot read!!!! msg: %s\n", err)
			os.Exit(1)
		}

		fmt.Println("\nClient read num bytes: ", bytes)
		repDataObj = shared.ConvJsonToDataStruct(buf[:bytes])

		if repDataObj.Type == shared.INFO_REPLY {
			break
		}
	}
	return repDataObj
}
