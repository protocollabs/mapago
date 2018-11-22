package clientProtos

import "net"
import "fmt"
import "strconv"
import "os"
import "github.com/monfron/mapago/ctrl/shared"
import "io"

const (
	LOOPBACK_IF = 1
	ETH_IF = 2
)

// classes

type UdpMcObj struct {
	connName     string
	connMcAddr   string
	connPort     int
	connCallSize int
}

type UdpMcConnObj struct {
	connSock *net.UDPConn
}

// Constructors

func NewUdpMcObj(name string, addr string, port int, callSize int) *UdpMcObj {
	udpMcObj := new(UdpMcObj)
	udpMcObj.connName = name
	udpMcObj.connMcAddr = addr
	udpMcObj.connPort = port
	udpMcObj.connCallSize = callSize
	return udpMcObj
}

func NewUdpMcConnObj(udpSock *net.UDPConn) *UdpMcConnObj {
	UdpMcConnObj := new(UdpMcConnObj)
	UdpMcConnObj.connSock = udpSock
	return UdpMcConnObj
}

func (udpMc *UdpMcObj) Start(jsonData []byte) *shared.DataObj {
	var repDataObj *shared.DataObj

	fmt.Println("UdpMcObj Start() here")
	buf := make([]byte, udpMc.connCallSize, udpMc.connCallSize)

	ip := net.ParseIP(udpMc.connMcAddr)
	if ip.IsLinkLocalMulticast() != true {
		fmt.Printf("\nUDP MC: Destination MC Addr is no ll mc addr!\n")
		os.Exit(1)
	}

	rAddr := udpMc.connMcAddr + ":" + strconv.Itoa(udpMc.connPort)
	rUdpMcAddr, err := net.ResolveUDPAddr("udp", rAddr)
	if err != nil {
		fmt.Printf("\nUDP MC: Cannot resolve UDP MC addr: %s", err)
		os.Exit(1)
	}

	udpConn, err := net.DialUDP("udp", nil, rUdpMcAddr)
	if err != nil {
		fmt.Printf("\nUDP MC: Cannot dial UDP MC: %s", err)
		os.Exit(1)
	}

	udpConnObj := NewUdpMcConnObj(udpConn)
	defer udpConnObj.connSock.Close()

	for {
		_, err := udpConnObj.connSock.Write(jsonData)
		if err != nil {
			fmt.Printf("\nUDP MC: Cannot send multicast! %s", err)
			os.Exit(1)
		}

		/* OPTION 1 "Req: Multicast, Rep: Unicast": read from socket where multicast was sent
		STATUS: Doesnt work...
		bytes, err := udpConnObj.connSock.Read(buf)
		if err != nil {
			fmt.Printf("\nUDP MC Client: Cannot Read")
			os.Exit(1)
		}
		fmt.Printf("Client read num bytes: %d", bytes)
		repDataObj = shared.ConvJsonToDataStruct(buf[:bytes])
		*/

		/* OPTION 2 "Req: Multicast, Rep: Multicast" */
		lAddr :=  shared.IP4_LINK_LOCAL_MC + ":" + strconv.Itoa(shared.CONTROL_PORT)
		lUdpMcAddr, err := net.ResolveUDPAddr("udp", lAddr)
		if err != nil {
			fmt.Printf("\nUDP MC: Cannot construct UDP addr %s\n", err)
			os.Exit(1)
		}
		intf, err := net.InterfaceByIndex(ETH_IF)
		if err != nil {
			fmt.Printf("\nUDP MC: Could not determine interface by name %s", err)
			os.Exit(1)
		}

		udpMcConn, err := net.ListenMulticastUDP("udp", intf, lUdpMcAddr)
		if err != nil {
			fmt.Printf("\nUDP MC: Cannot listen on UDP addr %s\n", err)
			os.Exit(1)
		}

		bytes, srvAddr, err := udpMcConn.ReadFromUDP(buf)

		// ok we have a error condition
		if err != nil {
			if err == io.EOF {
				fmt.Println("\nUDP MC: EOF detected")
				break
			}

			if err.(*net.OpError).Err.Error() == "use of closed network connection" {
				fmt.Println("\nUDP MC: Closed network detected!")
				break
			}

			// something different serious...
			fmt.Printf("\nUDP MC: Cannot read from UDP! msg: %s\n", err)
			os.Exit(1)
		}
		fmt.Println("\nUDP MC: Connected with srv: ", srvAddr)
		repDataObj = shared.ConvJsonToDataStruct(buf[:bytes])
	
		if repDataObj.Type == shared.INFO_REPLY {
			break
		}
	}
	return repDataObj
}
