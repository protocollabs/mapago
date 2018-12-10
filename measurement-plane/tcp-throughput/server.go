package tcpThroughput

import "fmt"
import "os"
import "net"
import "time"
import "strconv"
import "github.com/monfron/mapago/control-plane/ctrl/shared"

var UPDATE_INTERVAL = 5

func NewTcpMsmtServer(msmtCh <-chan shared.ChMgmt2Msmt, ctrlCh chan<- shared.ChMsmt2Ctrl, msmtStartReq *shared.DataObj) {
	var msmtData map[string]string
	var countWorkers int
	var accumulated uint64
	numValCtr := 0
	msmtResultCh := make(chan shared.ChMsmtResult)
	goHeartbeatCh := make(chan bool)

	for {
		// POSSIBLE BLOCKING CAUSE: select blocks until one of its cases can run
		select {
		case mgmtCmd := <-msmtCh:
			fmt.Println("\nReceived Management Command: ", mgmtCmd.Cmd)

			switch mgmtCmd.Cmd {
			case "Msmt_start":
				/*
					This will be startTcp()
				*/

				fmt.Println("\nClient request is: ", msmtStartReq)

				numWorkers, err := strconv.Atoi(msmtStartReq.Measurement.Configuration.Worker)
				if err != nil {
					fmt.Printf("\nCannot convert worker value: %s", err)
					os.Exit(1)
				}

				countWorkers = numWorkers

				startPort, err := strconv.Atoi(msmtStartReq.Measurement.Configuration.Port)
				if err != nil {
					fmt.Printf("\nCannot convert port value: %s", err)
					os.Exit(1)
				}

				callSize, err := strconv.Atoi(msmtStartReq.Measurement.Configuration.Call_size)
				if err != nil {
					fmt.Printf("\nCannot convert callSize value: %s", err)
					os.Exit(1)
				}

				lAddr := msmtStartReq.Measurement.Configuration.Listen_addr

				fmt.Printf("\nTCP Msmt: Client wants %d workers to be started from start port %d!", numWorkers, startPort)

				for c := 1; c <= numWorkers; c++ {
					fmt.Printf("\n\nStarting worker %d on port %d", c, startPort)
					go tcpServerWorker(msmtResultCh, goHeartbeatCh, startPort, callSize, lAddr)
					startPort++
				}

				/*
					Handle Misconfiguration:
					Wait for OK from goroutine, then send reply to control plane
					true = goroutine ok, false = " not (will be basically not be send => os.Exit comes in place)
					=> we can guarentee that a correct reply is sent to control plane
				*/
				for c := 1; c <= numWorkers; c++ {
					heartbeat := <-goHeartbeatCh
					if heartbeat != true {
						panic("tcp_server_worker goroutine not ok")
					}
				}

				fmt.Println("\nGoroutines ok: We are save to send a reply!")
				/*
					end of startTcp()
				*/

				// SEND REPLY TO CONTROL PLANE
				msmtReply := new(shared.ChMsmt2Ctrl)
				msmtReply.Status = "ok"

				msmtData = make(map[string]string)
				msmtData["msmtId"] = mgmtCmd.MsmtId
				// This could be used to define a more detailed error msg
				msmtData["msg"] = "all modules running"
				msmtReply.Data = msmtData
				ctrlCh <- *msmtReply

			case "Msmt_close":
				fmt.Println("\nTODO: We have to close TCP msmt module!")

			case "Msmt_info":
				fmt.Println("\nTODO: We have to send TCP msmt info!")

			default:
				fmt.Printf("Unknown measurement module")
				os.Exit(1)
			}

		case msmtResult := <-msmtResultCh:
			numValCtr += 1
			accumulated += msmtResult.Bytes

			if numValCtr == countWorkers {
				fmt.Printf("\nGot reply from all %d workers", countWorkers)
				mbyte_sec := accumulated / (1000000 * uint64(UPDATE_INTERVAL))
				println("\nMByte/sec: ", mbyte_sec)
				// start next measurement burst
				accumulated = 0
				numValCtr = 0
			}
		}
	}
}

func tcpServerWorker(c chan<- shared.ChMsmtResult, goHeartbeatCh chan<- bool, port int, bufSize int, lAddr string) {
	listen := lAddr + ":" + strconv.Itoa(port)
	addr, error := net.ResolveTCPAddr("tcp", listen)
	if error != nil {
		fmt.Printf("Cannot parse \"%s\": %s\n", listen, error)
		goHeartbeatCh <- false
		os.Exit(1)
	}
	listener, error := net.ListenTCP("tcp", addr)
	if error != nil {
		fmt.Printf("Cannot listen: %s\n", error)
		goHeartbeatCh <- false
		os.Exit(1)
	}
	defer listener.Close()

	goHeartbeatCh <- true

	// This one is blocking
	conn, error := listener.AcceptTCP()
	if error != nil {
		fmt.Printf("Cannot accept: %s\n", error)
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Printf("Connection from %s\n", conn.RemoteAddr())
	message := make([]byte, bufSize, bufSize)

	var bytes uint64 = 0
	start := time.Now()
	for {
		// this one is blocking
		n1, error := conn.Read(message)
		if error != nil {
			fmt.Printf("Cannot read: %s\n", error)
			os.Exit(1)
		}

		bytes += uint64(n1)

		elapsed := time.Since(start)
		if elapsed.Seconds() > float64(UPDATE_INTERVAL) {
			c <- shared.ChMsmtResult{Bytes: bytes, Time: elapsed.Seconds()}
			start = time.Now()
			bytes = 0
		}
	}
}
