package clientProtos

import "net"
import "fmt"
import "strconv"
import "os"
import "github.com/monfron/mapago/ctrl/shared"

// classes

type UdpMcObj struct {
	connName     string
	connMcAddr   string // write to this address
	connPort     int
	connCallSize int
}

type UdpMcConnObj struct {
	connSock *net.UDPConn
	// connSrvAddr *net.UDPAddr Q: do we get the specific address of the server

	// TODO: The same socket could be used to receive additional UC
	// we dont need an additional socket
}

// Constructors

func NewUdpMcObj(name string, addr string, port int, callSize int) *UdpMcObj {
	udpMcObj := new(UdpMcObj)
	udpMcObj.connName = name
	udpMcObj.connMcAddr = addr // dest addr
	udpMcObj.connPort = port   // dest addr
	udpMcObj.connCallSize = callSize
	return udpMcObj
}

func NewUdpMcConnObj(udpSock *net.UDPConn) *UdpMcConnObj {
	UdpMcConnObj := new(UdpMcConnObj)
	// this is no special socket: could be used for UC and MC
	UdpMcConnObj.connSock = udpSock
	return UdpMcConnObj
}

func (udpMc *UdpMcObj) Start(jsonData []byte) *shared.DataObj {
	var repDataObj *shared.DataObj

	fmt.Println("UdpMcObj Start() here")
	buf := make([]byte, udpMc.connCallSize, udpMc.connCallSize)

	ip := net.ParseIP(udpMc.connMcAddr)
	if ip.IsLinkLocalMulticast() != true {
		fmt.Printf("\n Destination MC Addr is no ll mc addr!\n")
		os.Exit(1)
	}

	rAddr := udpMc.connMcAddr + ":" + strconv.Itoa(udpMc.connPort)
	udpMcAddr, err := net.ResolveUDPAddr("udp", rAddr)
	if err != nil {
		fmt.Printf("\nCannot resolve UDP MC: %s", err)
		os.Exit(1)
	}

	// TODO: What happens if nil changes? sniffer!
	udpConn, err := net.DialUDP("udp", nil, udpMcAddr)
	if err != nil {
		fmt.Printf("\nCannot dial UDP MC: %s", err)
		os.Exit(1)
	}

	udpConnObj := NewUdpMcConnObj(udpConn)
	defer udpConnObj.connSock.Close()

	for {
		// write as multicast
		_, err := udpConnObj.connSock.Write(jsonData)
		if err != nil {
			fmt.Printf("\nUDP MC Client: Cannot send multicast! %s", err)
			os.Exit(1)
		}

		// it will block here currently here: no write from SRV
		fmt.Println("\narrived right before read")

		// we could get answer as multicast or unicast
		bytes, err := udpConnObj.connSock.Read(buf)
		if err != nil {
			fmt.Printf("\nUDP MC Client: Cannot Read")
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
