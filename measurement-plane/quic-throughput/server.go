package quicThroughput

import "fmt"
import "os"
import "strconv"
import "crypto/tls"
import "crypto/rsa"
import "crypto/rand"
import "crypto/x509"
import "math/big"
import "encoding/pem"
import "io"
import quic "github.com/lucas-clemente/quic-go"
import "github.com/protocollabs/mapago/control-plane/ctrl/shared"
import "strings"

type QuicThroughputMsmt struct {
	numStreams      int
	usedPorts       []int
	callSize        int
	listenAddr      string
	msmtId          string
	msmtInfoStorage map[string]*shared.MsmtInfoObj
	connStorage     map[string]*shared.QuicConn

	/*
		- this attribute can be used by start() to RECEIVE cmd from managementplane
	*/
	msmtMgmt2MsmtCh <-chan shared.ChMgmt2Msmt

	/*
		- this attribute can be used by constructor and start() to SEND msgs to control plane
	*/
	msmt2CtrlCh chan<- shared.ChMsmt2Ctrl

	/*
		- channel for closing connection
	*/
	closeConnCh chan interface{}
}

func NewQuicThroughputMsmt(msmtCh <-chan shared.ChMgmt2Msmt, ctrlCh chan<- shared.ChMsmt2Ctrl, msmtStartReq *shared.DataObj, msmtId string, startPort int) *QuicThroughputMsmt {
	var err error
	var msmtData map[string]string
	heartbeatCh := make(chan bool)
	closeCh := make(chan interface{})
	quicMsmt := new(QuicThroughputMsmt)

	quicMsmt.msmtInfoStorage = make(map[string]*shared.MsmtInfoObj)
	quicMsmt.connStorage = make(map[string]*shared.QuicConn)
	fmt.Println("\nClient QUIC request is: ", msmtStartReq)

	quicMsmt.numStreams, err = strconv.Atoi(msmtStartReq.Measurement.Configuration.Worker)
	if err != nil {
		fmt.Printf("\nCannot convert worker value: %s", err)
		os.Exit(1)
	}

	quicMsmt.callSize, err = strconv.Atoi(msmtStartReq.Measurement.Configuration.Call_size)
	if err != nil {
		fmt.Printf("\nCannot convert callSize value: %s", err)
		os.Exit(1)
	}

	quicMsmt.msmtId = msmtId
	quicMsmt.listenAddr = msmtStartReq.Measurement.Configuration.Listen_addr
	quicMsmt.msmtMgmt2MsmtCh = msmtCh
	quicMsmt.msmt2CtrlCh = ctrlCh
	quicMsmt.closeConnCh = closeCh

	for c := 1; c <= quicMsmt.numStreams; c++ {
		stream := "stream" + strconv.Itoa(c)

		initTs := shared.ConvCurrDateToStr()
		msmtInfo := shared.MsmtInfoObj{Bytes: 0, FirstTs: initTs, LastTs: initTs}
		quicMsmt.msmtInfoStorage[stream] = &msmtInfo

		quicConns := shared.QuicConn{}
		quicMsmt.connStorage[stream] = &quicConns

		go quicMsmt.quicServerWorker(closeCh, heartbeatCh, startPort, c, &msmtInfo, &quicConns)
	}

	for c := 1; c <= quicMsmt.numStreams; c++ {
		heartbeat := <-heartbeatCh
		if heartbeat != true {
			panic("quic_server_worker goroutine not ok")
		}
	}

	fmt.Println("\n\nPorts used by quic module: ", quicMsmt.usedPorts)

	msmtReply := new(shared.ChMsmt2Ctrl)
	msmtReply.Status = "ok"

	msmtData = make(map[string]string)
	msmtData["msmtId"] = quicMsmt.msmtId
	msmtData["msg"] = "all modules running"
	msmtData["usedPorts"] = shared.ConvIntSliceToStr(quicMsmt.usedPorts)

	msmtReply.Data = msmtData

	/*
		NOTE: I dont get why exactly we need to perform this as a goroutine
		- we use unbuffered channels so chan<- and <-chan should wait for each other
	*/
	go func() {
		ctrlCh <- *msmtReply
	}()

	return quicMsmt
}

func (quicMsmt *QuicThroughputMsmt) quicServerWorker(closeCh <-chan interface{}, goHeartbeatCh chan<- bool, port int, streamIndex int, msmtInfoPtr *shared.MsmtInfoObj, connPtr *shared.QuicConn) {
	var listener quic.Listener
	var err error
	fTsExists := false
	stream := "stream" + strconv.Itoa(streamIndex)
	fmt.Printf("\n%s is here", stream)

	tlsConf := create_tls_config()

	for {
		listen := quicMsmt.listenAddr + ":" + strconv.Itoa(port)

		listener, err = quic.ListenAddr(listen, tlsConf, nil)
		if err == nil {
			// debug fmt.Printf("\nCan listen on addr: %s\n", listen)
			quicMsmt.usedPorts = append(quicMsmt.usedPorts, port)
			connPtr.SrvSock = listener
			goHeartbeatCh <- true
			break
		}

		port++
	}

	sess, err := listener.Accept()
	if err != nil {
		fmt.Printf("Accept() err is: %s\n", err)

		errStr := strings.TrimSpace(err.Error())
		if errStr == "server closed" {
			return
		}

		fmt.Println("\nUnknown accept() error! exiting!")
		os.Exit(1)
	}

	connPtr.AcceptSock = sess

	quicStream, err := sess.AcceptStream()
	if err != nil {
		fmt.Printf("AcceptStream() err is: %s\n", err)
		os.Exit(1)
	}

	message := make([]byte, quicMsmt.callSize, quicMsmt.callSize)

	for {
		bytes, err := quicStream.Read(message)
		if err != nil {

			if err == io.EOF {
				break
			}

			errStr := strings.TrimSpace(err.Error())
			if errStr == "NO_ERROR" {
				// ok quic-go bug there...just continue
				continue
			}

			if strings.Contains(errStr, ":") {
				index := strings.IndexByte(err.Error(), ':')
				errStr = strings.TrimSpace(errStr[index+1:])

				if errStr == "No recent network activity" {
					break
				}
				// no additional cases atm
			}

			fmt.Println("\nUnknown read error! exiting!")
			os.Exit(1)
		}

		msmtInfoPtr.Bytes += uint64(bytes)

		if fTsExists == false {
			fTs := shared.ConvCurrDateToStr()
			msmtInfoPtr.FirstTs = fTs
			fTsExists = true
		}

		lTs := shared.ConvCurrDateToStr()
		msmtInfoPtr.LastTs = lTs
	}
}

func (quicMsmt *QuicThroughputMsmt) CloseConn() {
	var mgmtData map[string]string
	var msmtData []shared.DataResultObj
	var combinedData shared.CombinedData

	for _, quicConns := range quicMsmt.connStorage {
		quicConns.SrvSock.Close()

		if quicConns.AcceptSock != nil {
			err := quicConns.AcceptSock.Close()
			if err != nil {
				fmt.Printf("Close() err is: %s\n", err)
				os.Exit(1)
			}
		}
	}

	msmtReply := new(shared.ChMsmt2Ctrl)
	msmtReply.Status = "ok"

	mgmtData = make(map[string]string)
	mgmtData["msmtId"] = quicMsmt.msmtId
	mgmtData["msg"] = "all modules closed"

	combinedData.MgmtData = mgmtData

	for c := 1; c <= quicMsmt.numStreams; c++ {
		stream := "stream" + strconv.Itoa(c)

		msmtStruct := quicMsmt.msmtInfoStorage[stream]

		dataElement := new(shared.DataResultObj)
		dataElement.Received_bytes = strconv.Itoa(int(msmtStruct.Bytes))
		dataElement.Timestamp_first = msmtStruct.FirstTs
		dataElement.Timestamp_last = msmtStruct.LastTs

		msmtData = append(msmtData, *dataElement)
		fmt.Println("\nmsmtData is: ", msmtData)
	}

	combinedData.MsmtData = msmtData
	msmtReply.Data = combinedData

	go func() {
		quicMsmt.msmt2CtrlCh <- *msmtReply
	}()
}

func (quicMsmt *QuicThroughputMsmt) GetMsmtInfo() {
	var msmtData []shared.DataResultObj

	fmt.Println("\nPreparing quic msmt info req")

	msmtReply := new(shared.ChMsmt2Ctrl)
	msmtReply.Status = "ok"

	// prepare msmtReply.Data
	for c := 1; c <= quicMsmt.numStreams; c++ {
		stream := "stream" + strconv.Itoa(c)

		msmtStruct := quicMsmt.msmtInfoStorage[stream]

		dataElement := new(shared.DataResultObj)
		dataElement.Received_bytes = strconv.Itoa(int(msmtStruct.Bytes))
		dataElement.Timestamp_first = msmtStruct.FirstTs
		dataElement.Timestamp_last = msmtStruct.LastTs

		msmtData = append(msmtData, *dataElement)
		fmt.Println("\nmsmtData is: ", msmtData)
	}

	msmtReply.Data = msmtData

	go func() {
		quicMsmt.msmt2CtrlCh <- *msmtReply
	}()
}

// we could also move that to common-funcs => improvement
func create_tls_config() *tls.Config {
	// 1. generate KEY: generate 1024 bit key using RNG
	pKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic("generateRsa")
	}

	// 2. x509 CERT template
	certTemplate := x509.Certificate{SerialNumber: big.NewInt(1)}

	// 3. create self-signed x509 certificate => DER used for binary encoded certs
	certDER, err := x509.CreateCertificate(rand.Reader, &certTemplate, &certTemplate, &pKey.PublicKey, pKey)
	if err != nil {
		panic("generateX509DER")
	}

	// 4. encode key in PEM
	pKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(pKey)})

	// 5. encode certDER in PEM
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	// 6. create tls cert
	tlsCert, err := tls.X509KeyPair(certPEM, pKeyPEM)
	if err != nil {
		fmt.Println("generateTlsCert")
	}

	return &tls.Config{Certificates: []tls.Certificate{tlsCert}}
}
