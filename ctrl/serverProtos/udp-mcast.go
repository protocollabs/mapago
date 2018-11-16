package serverProtos

import "fmt"
import "net"
import "strconv"
import "os"
import "github.com/monfron/mapago/ctrl/shared"
import "io"

// import "reflect"

const (
	LOOPBACK_IF = 1
	ETH_IF      = 2
)

// classes

type UdpMcObj struct {
	connName   string
	connMcAddr string
	connPort   int
	// do we really need that?
	connMcSock   *net.UDPConn
	connCallSize int
}

type UdpMcConnObj struct {
	// socket for receiving multicasts
	connMcSock *net.UDPConn
	// socket for sending multicasts
	connSock    *net.UDPConn
	connCltAddr *net.UDPAddr
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

func NewUdpMcConnObj(udpMcSock *net.UDPConn) *UdpMcConnObj {
	UdpMcConnObj := new(UdpMcConnObj)
	UdpMcConnObj.connMcSock = udpMcSock
	return UdpMcConnObj
}

func (udpMcConn *UdpMcConnObj) WriteAnswer(answer []byte) {

	rAddr := shared.IP4_LINK_LOCAL_MC + ":" + strconv.Itoa(shared.CONTROL_PORT)
	rUpdMcAddr, err := net.ResolveUDPAddr("udp", rAddr)
	if err != nil {
		fmt.Printf("\nUDP MC: Cannot resolve UDP MC: %s", err)
		os.Exit(1)
	}

	udpConn, err := net.DialUDP("udp", nil, rUpdMcAddr)
	if err != nil {
		fmt.Printf("\nUDP MC: Cannot dial UDP MC: %s", err)
		os.Exit(1)
	}

	udpMcConn.connSock = udpConn

	_, err = udpMcConn.connSock.Write(answer)
	if err != nil {
		fmt.Printf("\nUDP MC: Cannot send multicast! %s", err)
		os.Exit(1)
	}
}

func (udpMcConn *UdpMcConnObj) CloseConn() {
	err := udpMcConn.connMcSock.Close()
	if err != nil {
		fmt.Printf("UDP MC: Cannot close UDP conn: %s", err)
		os.Exit(1)
	}

	fmt.Println("\nSock for receiving MC closed")

	err = udpMcConn.connSock.Close()
	if err != nil {
		fmt.Printf("UDP MC: Cannot close UDP conn: %s", err)
		os.Exit(1)
	}

	fmt.Println("\nSock for sending MC closed")
}

func (udpMc *UdpMcObj) Start(ch chan<- shared.ChResult) {
	fmt.Println("UdpMcObj start() called")

	lAddr := udpMc.connMcAddr + ":" + strconv.Itoa(udpMc.connPort)
	fmt.Println("UDP MC Listening on: ", lAddr)

	ip := net.ParseIP(udpMc.connMcAddr)

	// filtering link local multicast
	if ip.IsLinkLocalMulticast() != true {
		fmt.Printf("\nUDP MC: Cannot listen! Addr is no ll mc addr!\n")
		os.Exit(1)
	}

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

	udpMc.connMcSock = udpMcConn
	go udpMc.handleUdpConn(ch, udpMcConn)
}

func (udpMc *UdpMcObj) handleUdpConn(ch chan<- shared.ChResult, udpMcSock *net.UDPConn) {
	fmt.Println("\nHandle udp conn called!")

	buf := make([]byte, udpMc.connCallSize, udpMc.connCallSize)
	udpMcConn := NewUdpMcConnObj(udpMcSock)

	for {
		bytes, cltAddr, err := udpMcConn.connMcSock.ReadFromUDP(buf)

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

		fmt.Println("\nMulticast Request from UDP Client: ", cltAddr)
		udpMcConn.connCltAddr = cltAddr

		chRequest := new(shared.ChResult)
		chRequest.ConnObj = udpMcConn
		chRequest.Json = buf[:bytes]
		fmt.Printf("UDP MC: Sending into channel\n")
		ch <- *chRequest
	}
}
