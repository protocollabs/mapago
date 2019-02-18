package quicThroughput

import "sync"
import "os"
import "strconv"
import "fmt"
import "crypto/tls"
import quic "github.com/lucas-clemente/quic-go"
import "github.com/protocollabs/mapago/control-plane/ctrl/shared"

func NewQuicMsmtClient(config shared.ConfigurationObj, msmtStartRep *shared.DataObj, wg *sync.WaitGroup, closeConnCh <-chan string, callSize int) {
	lAddr := config.Listen_addr
	serverPorts := shared.ConvStrToIntSlice(msmtStartRep.Measurement.Configuration.UsedPorts)

	for _, port := range serverPorts {
		listen := lAddr + ":" + strconv.Itoa(port)
		wg.Add(1)
		go quicClientWorker(listen, wg, closeConnCh, callSize)
	}
}

func quicClientWorker(addr string, wg *sync.WaitGroup, closeConnCh <-chan string, callSize int) {
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
				fmt.Printf("\nTcpClient worker did not understand cmd: %s", cmd)
				os.Exit(1)
			}
		default:
			// replace that with stream.Write(buf)
			_, err := stream.Write(buf)
			if err != nil {
				fmt.Printf("\nWrite error: %s", err)
				os.Exit(1)
			}
		}
	}
}
