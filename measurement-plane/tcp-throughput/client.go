package tcpThroughput

import "sync"
import "net"
import "os"
import "strconv"
import "fmt"
import "github.com/monfron/mapago/control-plane/ctrl/shared"

var DEF_BUFFER_SIZE = 8096 * 8

func NewTcpMsmtClient(config shared.ConfigurationObj, msmtStartRep *shared.DataObj, wg *sync.WaitGroup, closeConnCh <-chan string) {
	lAddr := config.Listen_addr

	fmt.Println("\nTCP Destination listen addr: ", lAddr)
	serverPorts := shared.ConvStrToIntSlice(msmtStartRep.Measurement.Configuration.UsedPorts)

	for _, port := range serverPorts {
		listen := lAddr + ":" + strconv.Itoa(port)
		// debug fmt.Println("\nCommunicating with: ", listen)
		wg.Add(1)
		go tcpClientWorker(listen, wg, closeConnCh)
	}
}

func tcpClientWorker(addr string, wg *sync.WaitGroup, closeConnCh <-chan string) {
	buf := make([]byte, DEF_BUFFER_SIZE, DEF_BUFFER_SIZE)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		panic("dial")
	}

	for {
		select {
		case cmd := <-closeConnCh:
			if cmd == "close" {
				conn.Close()
				wg.Done()
				fmt.Println("\nClosing TCP connection")
				return
			} else {
				fmt.Printf("\nTcpClient worker did not understand cmd: %s", cmd)
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
