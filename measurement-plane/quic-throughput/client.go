package quicThroughput

import "sync"
import "os"
import "strconv"
import "fmt"
import "crypto/tls"
import "math"
import quic "github.com/lucas-clemente/quic-go"
import "github.com/protocollabs/mapago/control-plane/ctrl/shared"

func NewQuicMsmtClient(config shared.ConfigurationObj, msmtStartRep *shared.DataObj, wg *sync.WaitGroup, closeConnCh <-chan string, callSize int, sentStreamBytes map[string]*uint, msmtTotalBytes uint) {
	lAddr := config.Listen_addr
	serverPorts := shared.ConvStrToIntSlice(msmtStartRep.Measurement.Configuration.UsedPorts)
	workers, err := strconv.ParseUint(config.Worker, 10, 32)
	if err != nil {
		fmt.Printf("\n Parseuint error: %s", err)
		os.Exit(1)
	}

	// we need to ceil if the byte count per stream is uneven => or we cant reach the threshold
	StreamBytes := uint(math.Ceil(float64(msmtTotalBytes) / float64(workers)))

	for i, port := range serverPorts {
		listen := lAddr + ":" + strconv.Itoa(port)
		stream := "stream" + strconv.Itoa(i+1)

		wg.Add(1)
		go quicClientWorker(listen, wg, closeConnCh, uint(callSize), sentStreamBytes[stream], StreamBytes)
	}
}

func quicClientWorker(addr string, wg *sync.WaitGroup, closeConnCh <-chan string, callSize uint, sentStreamBytes *uint, streamBytes uint) {
	buf := make([]byte, callSize, callSize)

	tlsConf := tls.Config{InsecureSkipVerify: true}

	session, err := quic.DialAddr(addr, &tlsConf, nil)
	if err != nil {
		panic("dial")
	}

	stream, err := session.OpenStreamSync()
	if err != nil {
		panic("openStream")
	}

	for {
		select {
		case cmd := <-closeConnCh:
			if cmd == "close" {
				session.Close()
				wg.Done()
				return
			} else {
				fmt.Printf("\nQuicClient worker did not understand cmd: %s", cmd)
				os.Exit(1)
			}
		default:
			// sent as long as "stream threshold" not reached
			// case a) send whole callSize
			if streamBytes >= callSize {
				bytes, err := stream.Write(buf)

				if err != nil {
					fmt.Printf("\nWrite error: %s", err)
					os.Exit(1)
				}

				// update per stream counter
				streamBytes -= uint(bytes)
				// update stream counter reference for mapago-client => determine when its done
				*sentStreamBytes = *sentStreamBytes + uint(bytes)

				// case b) last bytes to send are not a "full" buffer
			} else if streamBytes < callSize && streamBytes > 0 {
				buf = make([]byte, streamBytes, streamBytes)
				bytes, err := stream.Write(buf)

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
