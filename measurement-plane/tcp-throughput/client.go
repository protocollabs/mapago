package tcpThroughput

import "sync"
import "net"
import "os"
import "strconv"
import "fmt"
import "github.com/protocollabs/mapago/control-plane/ctrl/shared"

func NewTcpMsmtClient(config shared.ConfigurationObj, msmtStartRep *shared.DataObj, wg *sync.WaitGroup, closeConnCh <-chan string, callSize int, msmtByteStore map[string]*uint) {
	lAddr := config.Listen_addr
	serverPorts := shared.ConvStrToIntSlice(msmtStartRep.Measurement.Configuration.UsedPorts)

	// debug fmt.Println("\nbyte store is ", msmtByteStore)

	for i, port := range serverPorts {
		listen := lAddr + ":" + strconv.Itoa(port)
		stream := "stream" + strconv.Itoa(i + 1)
		
		// debug fmt.Println("\nstream is: ", stream)
		// debug fmt.Println("\nvalue is: ", msmtByteStore[stream])

		wg.Add(1)
		go tcpClientWorker(listen, wg, closeConnCh, callSize, msmtByteStore[stream])
	}
}

func tcpClientWorker(addr string, wg *sync.WaitGroup, closeConnCh <-chan string, callSize int, streamByteCtr *uint) {
	buf := make([]byte, callSize, callSize)
	
	// debug fmt.Println("\nbyte ctr is:", *streamByteCtr)


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

			bytes, err := conn.Write(buf)
			if err != nil {
				fmt.Printf("\nWrite error: %s", err)
				os.Exit(1)
			}

			*streamByteCtr = *streamByteCtr + uint(bytes)

		}
	}
}
