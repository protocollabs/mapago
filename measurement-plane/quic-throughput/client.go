package quicThroughput

import "sync"
import "os"
import "strconv"
import "fmt"
import "crypto/tls"
import quic "github.com/lucas-clemente/quic-go"
import "github.com/protocollabs/mapago/control-plane/ctrl/shared"

var DEF_BUFFER_SIZE = 8096 * 8

func NewQuicMsmtClient(config shared.ConfigurationObj, msmtStartRep *shared.DataObj, wg *sync.WaitGroup, closeConnCh <-chan string) {
	lAddr := config.Listen_addr
	serverPorts := shared.ConvStrToIntSlice(msmtStartRep.Measurement.Configuration.UsedPorts)

	for _, port := range serverPorts {
		listen := lAddr + ":" + strconv.Itoa(port)
		wg.Add(1)
		go quicClientWorker(listen, wg, closeConnCh)
	}
}

func quicClientWorker(addr string, wg *sync.WaitGroup, closeConnCh <-chan string) {
	buf := make([]byte, DEF_BUFFER_SIZE, DEF_BUFFER_SIZE)

	// create tls config
	tlsConf := tls.Config{InsecureSkipVerify: true}

	// replace that with quic.DialAddr
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
				// replace that with session.Close()
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
