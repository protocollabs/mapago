package main

import "fmt"
import "strconv"
import "os/exec"
import "os"
import "log"
import "flag"
import "time"

// check this
import "sync"

func main() {
	var wg sync.WaitGroup
	fmt.Println("Spawn topology here")

	serverPtr := flag.Int("numSrv", 1, "number of servers to spawn")
	clientPtr := flag.Int("numClt", 1, "number of clients to spawn")
	flag.Parse()

	wg.Add(2)
	go spawnServer(*serverPtr, &wg)
	go spawnClient(*clientPtr, &wg)
	wg.Wait()
}

func spawnServer(numSrv int, wg *sync.WaitGroup) {
	fmt.Printf("\nwant to create %d server", numSrv)
	nameArg0 := "-listen-addr"

	if numSrv < 1 {
		fmt.Println("Wrong val for spawning servers")
		os.Exit(1)
	}

	file, err := os.Create("serverSpawnErrorLog.txt")
	if err != nil {
		fmt.Println("Error while opening file")
		os.Exit(1)
	}

	defer file.Close()
	log.SetOutput(file)

	for c := 1; c <= numSrv; c++ {
		valArg0 := "127.0.0." + strconv.Itoa(c)
		fmt.Println("\nCreating server: ", valArg0)
		procObj := exec.Command("./server", nameArg0, valArg0)

		// returns stdout, stderr
		out, err := procObj.CombinedOutput()
		if err != nil {
			log.Printf("\nServer output dump: \n --------------------- \n %s \n ---------------------\n", string(out[:]))

			// TODO: perform "real" error handling (os.Exit(1)) here
			// Current problem: Client disconnects and server reads again => EOF
			// would throw err here and abort spawning, even though discovery passed
		}
	}
	fmt.Printf("\nSpawned %d servers", numSrv)
	defer wg.Done()
}

func spawnClient(numClt int, wg *sync.WaitGroup) {
	fmt.Printf("\nwant to create %d clients", numClt)
	nameArg0 := "-ctrl-addr"

	if numClt < 1 {
		fmt.Println("Wrong val for spawning clients")
		os.Exit(1)
	}

	file, err := os.Create("clientSpawnErrorLog.txt")
	if err != nil {
		fmt.Printf("Error while creating file")
		os.Exit(1)
	}

	defer file.Close()
	log.SetOutput(file)

	for c := 1; c <= numClt; c++ {
		time.Sleep(10 * time.Millisecond)
		valArg0 := "127.0.0." + strconv.Itoa(c)
		fmt.Println("\nConnecting to server: ", valArg0)
		procObj := exec.Command("./client", nameArg0, valArg0)

		out, err := procObj.CombinedOutput()
		if err != nil {
			log.Printf("\nServer output dump: \n --------------------- \n %s \n ---------------------\n", string(out[:]))
			os.Exit(1)
		}
	}
	fmt.Printf("\nSpawned %d clients", numClt)
	fmt.Printf("\nNo errors during %d discoveries", numClt)
	defer wg.Done()
}
