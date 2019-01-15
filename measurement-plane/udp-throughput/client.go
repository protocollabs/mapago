package udpThroughput

import "sync"
import "net"
import "os"
import "strconv"
import "fmt"
import "github.com/protocollabs/mapago/control-plane/ctrl/shared"

var DEF_BUFFER_SIZE = 8096 * 8

func NewUdpMsmtClient(config shared.ConfigurationObj, msmtStartRep *shared.DataObj, wg *sync.WaitGroup, closeConnCh <-chan string) {
	lAddr := config.Listen_addr

	fmt.Println("\nUDP Destination listen addr: ", lAddr)
	serverPorts := shared.ConvStrToIntSlice(msmtStartRep.Measurement.Configuration.UsedPorts)

	for _, port := range serverPorts {
		listen := lAddr + ":" + strconv.Itoa(port)
		fmt.Println("\nCommunicating with: ", listen)
		wg.Add(1)
		go udpClientWorker(listen, wg, closeConnCh)
	}
}

func udpClientWorker(addr string, wg *sync.WaitGroup, closeConnCh <-chan string) {
	buf := make([]byte, DEF_BUFFER_SIZE, DEF_BUFFER_SIZE)
	conn, err := net.Dial("udp", addr)
	if err != nil {
		panic("dial")
	}

	for {
		select {
		case cmd := <-closeConnCh:
			if cmd == "close" {
				conn.Close()
				wg.Done()
				fmt.Println("\nClosing UDP connection")
				return
			} else {
				fmt.Printf("\nudpClient worker did not understand cmd: %s", cmd)
				os.Exit(1)
			}
		default:
			_, err := conn.Write(buf)

			if err != nil {
				fmt.Printf("\nWrite error: %s", err)
				os.Exit(1)
			}
		}
	}
}
