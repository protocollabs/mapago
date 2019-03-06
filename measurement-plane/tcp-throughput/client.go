package tcpThroughput

import "sync"
import "net"
import "os"
import "strconv"
import "fmt"
import "github.com/protocollabs/mapago/control-plane/ctrl/shared"

func NewTcpMsmtClient(config shared.ConfigurationObj, msmtStartRep *shared.DataObj, wg *sync.WaitGroup, closeConnCh <-chan string, callSize int, sentStreamBytes map[string]*uint, msmtTotalBytes uint) {
	lAddr := config.Listen_addr
	serverPorts := shared.ConvStrToIntSlice(msmtStartRep.Measurement.Configuration.UsedPorts)
	workers, err  := strconv.ParseUint(config.Worker, 10, 32)
	if err != nil {
		fmt.Printf("\n Parseuint error: %s", err)
		os.Exit(1)
	}

	streamBytes := msmtTotalBytes / uint(workers) 
	// debug fmt.Println("\nevery stream has to sent: ", streamBytes)

	for i, port := range serverPorts {
		listen := lAddr + ":" + strconv.Itoa(port)
		stream := "stream" + strconv.Itoa(i + 1)

		wg.Add(1)
		go tcpClientWorker(listen, wg, closeConnCh, uint(callSize), sentStreamBytes[stream], streamBytes)
	}
}

func tcpClientWorker(addr string, wg *sync.WaitGroup, closeConnCh <-chan string, callSize uint, sentStreamBytes *uint, streamBytes uint) {
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
			// sent as long as "stream threshold" not reached
			// case a) send whole callSize
			if streamBytes >= callSize  {
				bytes, err := conn.Write(buf)
	
				if err != nil {
					fmt.Printf("\nWrite error: %s", err)
					os.Exit(1)
				}

				// update per stream counter
				streamBytes -= uint(bytes)
				// update stream counter reference for mapago-client => determine when its done
				*sentStreamBytes = *sentStreamBytes + uint(bytes)
			
			// case b) last bytes to send are not a "full" buffer
			} else if (streamBytes < callSize && streamBytes > 0) {
				buf = make([]byte, streamBytes, streamBytes)
				bytes, err := conn.Write(buf)
	
				if err != nil {
					fmt.Printf("\nWrite error: %s", err)
					os.Exit(1)
				}

				// update per stream counter
				streamBytes -= uint(bytes)
				// update stream counter reference for mapago-client => determine when its done
				*sentStreamBytes = *sentStreamBytes + uint(bytes)

			// case c): Default (streamBytes == 0 => enough sent) => Do nothing: Wait for channels
			}
		}
	}
}