package tcpTlsThroughput

import "sync"
import "net"
import "os"
import "strconv"
import "fmt"
import "github.com/protocollabs/mapago/control-plane/ctrl/shared"

func NewTcpTlsMsmtClient(config shared.ConfigurationObj, msmtStartRep *shared.DataObj, wg *sync.WaitGroup, closeConnCh <-chan string, callSize int) {
	lAddr := config.Listen_addr
	serverPorts := shared.ConvStrToIntSlice(msmtStartRep.Measurement.Configuration.UsedPorts)

	for _, port := range serverPorts {
		listen := lAddr + ":" + strconv.Itoa(port)
		wg.Add(1)
		go tcpTlsClientWorker(listen, wg, closeConnCh, callSize)
	}
}

func tcpTlsClientWorker(addr string, wg *sync.WaitGroup, closeConnCh <-chan string, callSize int) {
	buf := make([]byte, callSize, callSize)
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
