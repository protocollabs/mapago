package serverProtos

import "fmt"
import "net"
import "strconv"
import "os"
import "github.com/monfron/mapago/ctrl/shared"
import "io"

// import "reflect"

const (
	LOOPBACK_IF = "lo"
	// predictable network interface naming
	ETH_IF = "enp0s31f6"
)

// classes

type UdpMcObj struct {
	connName   string
	connMcAddr string // listen to this address
	connPort   int
	// do we really need that?
	connMcSock   *net.UDPConn
	connCallSize int
}

type UdpMcConnObj struct {
	// ATM this is the multicast socket
	// this will be closed in server.go after we wrote the reply
	connMcSock *net.UDPConn

	// TODO: if we want to reply with UC we
	// have to open a UC socket for writing
	// consequences: the new UC socket has to be established within handleUdpConn
	// i.e. connUcSock *net.UDPConn

	// TODO: this should be the dest UC addr obtained via ReadFromUDP
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
	// TODO: Depending on the underlying socket
	// the reply will be a unicast or multicast reply
	// i.e. udpMcConn.connUcSock.Write(answer)
	fmt.Println("\nUDP MC: Dummy Write Answer!")
}

// TODO: If we have two Sockets => we have to close two aswell
// Q: can we do this in this single method or two separate?!
func (udpMcConn *UdpMcConnObj) CloseConn() {
	err := udpMcConn.connMcSock.Close()
	if err != nil {
		fmt.Printf("Cannot close UDP conn: %s", err)
		os.Exit(1)
	}
}

func (udpMc *UdpMcObj) Start(ch chan<- shared.ChResult) {
	fmt.Println("UdpMcObj start() called")

	listenAddr := udpMc.connMcAddr + ":" + strconv.Itoa(udpMc.connPort)
	fmt.Println("UDP MC Listening on: ", listenAddr)

	ip := net.ParseIP(udpMc.connMcAddr)

	// filtering link local multicast
	if ip.IsLinkLocalMulticast() != true {
		fmt.Printf("\nCannot listen! Addr is no ll mc addr!\n")
		os.Exit(1)
	}

	udpMcAddr, err := net.ResolveUDPAddr("udp", listenAddr)
	if err != nil {
		fmt.Printf("\nCannot construct UDP addr %s\n", err)
		os.Exit(1)
	}

	// TODO: What is better name or index?
	intf, err := net.InterfaceByName(LOOPBACK_IF)
	if err != nil {
		fmt.Printf("\nCould not determine interface by name %s", err)
		os.Exit(1)
	}
	// debug fmt.Println("\nSpec interface is: ", *intf)

	udpMcConn, err := net.ListenMulticastUDP("udp", intf, udpMcAddr)
	if err != nil {
		fmt.Printf("\nCannot listen on UDP addr %s\n", err)
		os.Exit(1)
	}

	udpMc.connMcSock = udpMcConn
	go udpMc.handleUdpConn(ch, udpMcConn)
}

func (udpMc *UdpMcObj) handleUdpConn(ch chan<- shared.ChResult, udpSock *net.UDPConn) {
	fmt.Println("\nHandle udp conn called!")

	buf := make([]byte, udpMc.connCallSize, udpMc.connCallSize)
	udpMcConn := NewUdpMcConnObj(udpSock)

	for {
		// TODO cltAddr => could be used for UC reply
		bytes, cltAddr, err := udpMcConn.connMcSock.ReadFromUDP(buf)

		// ok we have a error condition
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
			fmt.Printf("\nCannot read from UDP! msg: %s\n", err)
			os.Exit(1)
		}

		fmt.Println("\nUDP MC Server read num bytes: ", bytes)
		fmt.Println("\nMulticast Request from UDP Client: ", cltAddr)
		udpMcConn.connCltAddr = cltAddr

		chRequest := new(shared.ChResult)
		chRequest.ConnObj = udpMcConn
		chRequest.Json = buf[:bytes]
		fmt.Printf("UDP MC: Sending into channel\n")
		ch <- *chRequest
	}
}
