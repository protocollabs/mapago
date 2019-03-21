package quicThroughput

import "sync"
import "os"
import "strconv"
import "fmt"
import "crypto/tls"
import "math"
import "strings"
import quic "github.com/lucas-clemente/quic-go"
import "github.com/protocollabs/mapago/control-plane/ctrl/shared"

func NewQuicMsmtClient(config shared.ConfigurationObj, msmtStartRep *shared.DataObj, wg *sync.WaitGroup, closeConnCh <-chan string, callSize int, msmtTotalBytes uint) {
	lAddr := config.Listen_addr
	serverPorts := shared.ConvStrToIntSlice(msmtStartRep.Measurement.Configuration.UsedPorts)
	workers, err := strconv.ParseUint(config.Worker, 10, 32)
	if err != nil {
		fmt.Printf("\n Parseuint error: %s", err)
		os.Exit(1)
	}

	// we need to ceil if the byte count per stream is uneven => or we cant reach the threshold
	StreamBytes := uint(math.Ceil(float64(msmtTotalBytes) / float64(workers)))
	/*
		fmt.Println("\nTotal bytes: ", msmtTotalBytes)
		fmt.Println("\nbytes per stream: ", StreamBytes)
		fmt.Println("\ntotal bytes over all streams", StreamBytes * uint(workers))
	*/

	for _, port := range serverPorts {
		listen := lAddr + ":" + strconv.Itoa(port)
		wg.Add(1)
		// i is for debugging
		go quicClientWorker(listen, wg, closeConnCh, uint(callSize), StreamBytes)
	}
}

func quicClientWorker(addr string, wg *sync.WaitGroup, closeConnCh <-chan string, callSize uint, streamBytes uint) {
	buf := make([]byte, callSize, callSize)

	tlsConf := tls.Config{InsecureSkipVerify: true}

	session, err := quic.DialAddr(addr, &tlsConf, nil)
	if err != nil {
		// Catch: handshake timeout
		return
	}

	stream, err := session.OpenStreamSync()
	if err != nil {
		fmt.Println("\nOpenStreamSync() err is: ", err)
		os.Exit(1)
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
			if streamBytes >= callSize {
				bytes, err := stream.Write(buf)

				if err != nil {
					errStr := strings.TrimSpace(err.Error())
					if errStr == "NO_ERROR" {
						continue
					}

					if strings.Contains(errStr, ":") {
						index := strings.IndexByte(err.Error(), ':')
						errStr = strings.TrimSpace(errStr[index+1:])

						if errStr == "No recent network activity" {
							// ok serious error we have to leave our "write-for-loop"
							session.Close()
							wg.Done()
							return
						}
					}

					os.Exit(1)
				}

				streamBytes -= uint(bytes)

				// case b) last bytes to send are not a "full" buffer
			} else if streamBytes < callSize && streamBytes > 0 {
				buf = make([]byte, streamBytes, streamBytes)
				bytes, err := stream.Write(buf)

				if err != nil {
					errStr := strings.TrimSpace(err.Error())
					if errStr == "NO_ERROR" {
						continue
					}

					if strings.Contains(errStr, ":") {
						index := strings.IndexByte(err.Error(), ':')
						errStr = strings.TrimSpace(errStr[index+1:])

						if errStr == "No recent network activity" {
							// ok serious error we have to leave our "write-for-loop"
							session.Close()
							wg.Done()
							return
						}
					}

					os.Exit(1)
				}

				streamBytes -= uint(bytes)

				// case c): Default (streamBytes == 0 => enough sent) => Do nothing: Wait for channels
			}
		}
	}

}
