package serverProtos

import "fmt"
import "net"
import "strconv"
import "os"
import "github.com/monfron/mapago/ctrl/shared"
import "io"

// classes

type UdpObj struct {
	connName string
	connAddr string
	connPort int
	// do we really need that?
	connSock     *net.UDPConn
	connCallSize int
}

type UdpConnObj struct {
	connSock    *net.UDPConn
	connCltAddr *net.UDPAddr
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

func NewUdpConnObj(udpSock *net.UDPConn) *UdpConnObj {
	udpConnObj := new(UdpConnObj)
	udpConnObj.connSock = udpSock
	return udpConnObj
}

func (udpConn *UdpConnObj) WriteAnswer(answer []byte) {
	_, err := udpConn.connSock.WriteToUDP(answer, udpConn.connCltAddr)
	if err != nil {
		fmt.Printf("Cannot write to %s:%s", string(udpConn.connCltAddr.IP), strconv.Itoa(udpConn.connCltAddr.Port))
		os.Exit(1)
	}
}

func (udpConn *UdpConnObj) CloseConn() {
	err := udpConn.connSock.Close()
	if err != nil {
		fmt.Printf("Cannot close UDP conn: %s", err)
		os.Exit(1)
	}
}

func (udp *UdpObj) Start(ch chan<- shared.ChResult) {
	fmt.Println("UdpObj start() called")

	listenAddr := udp.connAddr + ":" + strconv.Itoa(udp.connPort)
	fmt.Printf("\nUDP Listening on: %s", listenAddr)

	udpAddr, err := net.ResolveUDPAddr("udp", listenAddr)
	if err != nil {
		fmt.Printf("\nCannot construct UDP addr %s\n", err)
		os.Exit(1)
	}

	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Printf("\nCannot listen on UDP addr %s\n", err)
		os.Exit(1)
	}

	udp.connSock = udpConn
	go udp.handleUdpConn(ch, udpConn)
}

func (udp *UdpObj) handleUdpConn(ch chan<- shared.ChResult, udpSock *net.UDPConn) {
	fmt.Println("\nhandleUdpConn goroutine called")

	buf := make([]byte, udp.connCallSize, udp.connCallSize)
	udpConn := NewUdpConnObj(udpSock)

	for {
		bytes, cltAddr, err := udpConn.connSock.ReadFromUDP(buf)

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
			fmt.Printf("\nCannot read from UDP! msg: %s\n", err)
			os.Exit(1)
		}

		fmt.Println("UDP Server read num bytes: ", bytes)
		fmt.Println("Request from UDP Client: ", cltAddr)
		udpConn.connCltAddr = cltAddr

		chRequest := new(shared.ChResult)
		chRequest.ConnObj = udpConn
		chRequest.Json = buf[:bytes]
		fmt.Printf("UDP: Sending into channel\n")
		ch <- *chRequest
	}
}
