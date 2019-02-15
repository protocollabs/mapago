package tcpTlsThroughput

import "sync"
import "os"
import "strconv"
import "fmt"
import "crypto/tls"
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
	certPath := "/src/github.com/protocollabs/mapago/measurement-plane/tcp-tls-throughput/certs"
	goPath := os.Getenv("GOPATH")
	cltPemPath := goPath + certPath + "/client.pem"
	cltKeyPath := goPath + certPath + "/client.key"

	cert, error := tls.LoadX509KeyPair(cltPemPath, cltKeyPath)
	if error != nil {
		fmt.Printf("Cannot loadkeys: %s\n", error)
		os.Exit(1)
	}

	// accept any certificate presented by server
	config := tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}

	conn, err := tls.Dial("tcp", addr, &config)
	if err != nil {
		panic("dial")
	}

	// TODO: check connection state

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
