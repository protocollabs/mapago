package clientProtos

import "net"
import "fmt"
import "strconv"
import "os"
import "github.com/monfron/mapago/control-plane/ctrl/shared"

// classes

type UdpObj struct {
	connName     string
	connAddr     string
	connPort     int
	connCallSize int
}

type UdpConnObj struct {
	connSock *net.UDPConn
	// not needed atm, but necessary for future conns (msmt start)
	connSrvAddr *net.UDPAddr
}

// Constructors

func NewUdpObj(name string, addr string, port int, callSize int) *UdpObj {
	udpObj := new(UdpObj)
	udpObj.connName = name
	udpObj.connAddr = addr
	udpObj.connPort = port
	udpObj.connCallSize = callSize
	return udpObj
}

func NewUdpConnObj(udpSock *net.UDPConn, udpAddr *net.UDPAddr) *UdpConnObj {
	udpConnObj := new(UdpConnObj)
	udpConnObj.connSock = udpSock
	udpConnObj.connSrvAddr = udpAddr
	return udpConnObj
}

func (udp *UdpObj) Start(jsonData []byte) *shared.DataObj {
	var repDataObj *shared.DataObj

	fmt.Println("UdpObj Start() here")
	buf := make([]byte, udp.connCallSize, udp.connCallSize)

	rAddr := udp.connAddr + ":" + strconv.Itoa(udp.connPort)
	udpAddr, err := net.ResolveUDPAddr("udp", rAddr)
	if err != nil {
		fmt.Printf("\nCannot resolve UDP: %s", err)
		os.Exit(1)
	}

	udpConn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		fmt.Printf("\n Cannot dial UDP: %s", err)
		os.Exit(1)
	}

	udpConnObj := NewUdpConnObj(udpConn, udpAddr)
	defer udpConnObj.connSock.Close()

	for {
		_, err := udpConnObj.connSock.Write(jsonData)
		if err != nil {
			fmt.Printf("\nUDP Client: Cannot send! %s", err)
			os.Exit(1)
		}

		// TODO: handle srvAddr != *udpConnObj.connSrvAddr
		// i.e. we receive a info reply from a different server

		bytes, err := udpConnObj.connSock.Read(buf)
		if err != nil {
			fmt.Printf("\nUDP Client: Cannot Read")
			os.Exit(1)
		}
		fmt.Printf("Client read num bytes: %d", bytes)
		repDataObj = shared.ConvJsonToDataStruct(buf[:bytes])

		if repDataObj.Type == shared.INFO_REPLY {
			break
		}
	}
	return repDataObj
}
