package tcpThroughput

import "sync"
import "net"
import "os"
import "strconv"
import "fmt"
import "github.com/monfron/mapago/control-plane/ctrl/shared"

var DEF_BUFFER_SIZE = 8096 * 8

func NewTcpMsmtClient(config shared.ConfigurationObj) {
	var wg sync.WaitGroup
	lAddr := config.Listen_addr

	port, err := strconv.Atoi(config.Port)
	if err != nil {
		fmt.Printf("\nCannot convert port value: %s", err)
		os.Exit(1)
	}

	numThreads, err := strconv.Atoi(config.Worker)
	if err != nil {
		fmt.Printf("\nCannot convert thread value: %s", err)
		os.Exit(1)
	}

	fmt.Println("\nSpawning num threads:", numThreads)
	fmt.Println("\nStarting port:", port)
	fmt.Println("\nDestination listen addr: ", lAddr)

	for i := 0; i < numThreads; i++ {
		listen := lAddr + ":" + strconv.Itoa(port)
		fmt.Println("\nCommunicating with: ", listen)
		wg.Add(1)
		go tcpClientWorker(listen, &wg)
		port += 1
	}
	wg.Wait()
}

func tcpClientWorker(addr string, wg *sync.WaitGroup) {

	defer wg.Done()
	buf := make([]byte, DEF_BUFFER_SIZE, DEF_BUFFER_SIZE)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		panic("dial")
	}
	defer conn.Close()

	for {
		_, err := conn.Write(buf)
		if err != nil {
			panic("write")
		}
	}
}
