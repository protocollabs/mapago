package clientProtos

import "fmt"
import "net"
import "strconv"
import "os"
import "github.com/monfron/mapago/control-plane/ctrl/shared"

// classes

type TcpObj struct {
	connName     string
	connAddr     string
	connPort     int
	connCallSize int
}

type TcpConnObj struct {
	connSock *net.TCPConn
}

// constructors:
func NewTcpObj(name string, addr string, port int, callSize int) *TcpObj {
	tcpObj := new(TcpObj)
	tcpObj.connName = name
	tcpObj.connAddr = addr
	tcpObj.connPort = port
	tcpObj.connCallSize = callSize
	return tcpObj
}

func NewTcpConnObj(tcpSock *net.TCPConn) *TcpConnObj {
	tcpConnObj := new(TcpConnObj)
	tcpConnObj.connSock = tcpSock
	return tcpConnObj
}

func (tcp *TcpObj) StartDiscovery(jsonData []byte) *shared.DataObj {
	var repDataObj *shared.DataObj

	fmt.Println("TcpObj StartDiscovery() called")
	buf := make([]byte, tcp.connCallSize, tcp.connCallSize)

	rAddr := tcp.connAddr + ":" + strconv.Itoa(tcp.connPort)
	rTcpAddr, err := net.ResolveTCPAddr("tcp", rAddr)
	if err != nil {
		fmt.Printf("Cannot parse \"%s\": %s\n", rAddr, err)
		os.Exit(1)
	}

	tcpConn, err := net.DialTCP("tcp", nil, rTcpAddr)
	if err != nil {
		fmt.Printf("Cannot dial \"%s\": %s\n", rAddr, err)
		os.Exit(1)
	}

	tcpConnObj := NewTcpConnObj(tcpConn)
	defer tcpConnObj.connSock.Close()

	for {
		_, err := tcpConnObj.connSock.Write(jsonData)
		if err != nil {
			fmt.Printf("Cannot send!!! %s\n", err)
			os.Exit(1)
		}

		bytes, err := tcpConnObj.connSock.Read(buf)
		if err != nil {
			fmt.Printf("Cannot read!!!! msg: %s\n", err)
			os.Exit(1)
		}

		fmt.Println("\nClient read num bytes: ", bytes)
		repDataObj = shared.ConvJsonToDataStruct(buf[:bytes])

		if repDataObj.Type == shared.INFO_REPLY {
			break
		}
	}
	return repDataObj
}

func (tcp *TcpObj) StartMeasurement(jsonData []byte) *shared.DataObj {
	var repDataObj *shared.DataObj

	fmt.Println("TcpObj StartMeasurement() called")
	buf := make([]byte, tcp.connCallSize, tcp.connCallSize)

	rAddr := tcp.connAddr + ":" + strconv.Itoa(tcp.connPort)
	rTcpAddr, err := net.ResolveTCPAddr("tcp", rAddr)
	if err != nil {
		fmt.Printf("Cannot parse \"%s\": %s\n", rAddr, err)
		os.Exit(1)
	}

	tcpConn, err := net.DialTCP("tcp", nil, rTcpAddr)
	if err != nil {
		fmt.Printf("Cannot dial \"%s\": %s\n", rAddr, err)
		os.Exit(1)
	}

	tcpConnObj := NewTcpConnObj(tcpConn)
	defer tcpConnObj.connSock.Close()

	for {
		_, err := tcpConnObj.connSock.Write(jsonData)
		if err != nil {
			fmt.Printf("Cannot send!!! %s\n", err)
			os.Exit(1)
		}

		bytes, err := tcpConnObj.connSock.Read(buf)
		if err != nil {
			fmt.Printf("Cannot read!!!! msg: %s\n", err)
			os.Exit(1)
		}

		fmt.Println("\nClient read num bytes: ", bytes)
		repDataObj = shared.ConvJsonToDataStruct(buf[:bytes])

		if repDataObj.Type == shared.MEASUREMENT_START_REPLY {
			fmt.Printf("\nClient received an Measurement_Start_Reply!!!")
			break
		}
	}
	return repDataObj
}

func (tcp *TcpObj) StopMeasurement(jsonData []byte) *shared.DataObj {
	var repDataObj *shared.DataObj

	fmt.Println("TcpObj StopMeasurement() called")
	buf := make([]byte, tcp.connCallSize, tcp.connCallSize)

	rAddr := tcp.connAddr + ":" + strconv.Itoa(tcp.connPort)
	rTcpAddr, err := net.ResolveTCPAddr("tcp", rAddr)
	if err != nil {
		fmt.Printf("Cannot parse \"%s\": %s\n", rAddr, err)
		os.Exit(1)
	}

	tcpConn, err := net.DialTCP("tcp", nil, rTcpAddr)
	if err != nil {
		fmt.Printf("Cannot dial \"%s\": %s\n", rAddr, err)
		os.Exit(1)
	}

	tcpConnObj := NewTcpConnObj(tcpConn)
	defer tcpConnObj.connSock.Close()

	for {
		_, err := tcpConnObj.connSock.Write(jsonData)
		if err != nil {
			fmt.Printf("Cannot send!!! %s\n", err)
			os.Exit(1)
		}

		// NOTE: We will block here until we get a msmst_stop_reply
		bytes, err := tcpConnObj.connSock.Read(buf)
		if err != nil {
			fmt.Printf("Cannot read!!!! msg: %s\n", err)
			os.Exit(1)
		}

		fmt.Println("\nClient read num bytes: ", bytes)
		repDataObj = shared.ConvJsonToDataStruct(buf[:bytes])

		if repDataObj.Type == shared.MEASUREMENT_STOP_REPLY {
			fmt.Printf("\nClient received an Measurement_Stop_Reply!!!")
			break
		}
	}
	return repDataObj
}

func (tcp *TcpObj) GetMeasurementInfo(jsonData []byte) *shared.DataObj {
	var repDataObj *shared.DataObj

	fmt.Println("TcpObj GetMeasurementInfo() called")
	buf := make([]byte, tcp.connCallSize, tcp.connCallSize)

	rAddr := tcp.connAddr + ":" + strconv.Itoa(tcp.connPort)
	rTcpAddr, err := net.ResolveTCPAddr("tcp", rAddr)
	if err != nil {
		fmt.Printf("Cannot parse \"%s\": %s\n", rAddr, err)
		os.Exit(1)
	}

	tcpConn, err := net.DialTCP("tcp", nil, rTcpAddr)
	if err != nil {
		fmt.Printf("Cannot dial \"%s\": %s\n", rAddr, err)
		os.Exit(1)
	}

	tcpConnObj := NewTcpConnObj(tcpConn)
	defer tcpConnObj.connSock.Close()

	for {
		_, err := tcpConnObj.connSock.Write(jsonData)
		if err != nil {
			fmt.Printf("Cannot send!!! %s\n", err)
			os.Exit(1)
		}

		// NOTE: We will block here until we get a msmst_info_reply
		bytes, err := tcpConnObj.connSock.Read(buf)
		if err != nil {
			fmt.Printf("Cannot read!!!! msg: %s\n", err)
			os.Exit(1)
		}

		fmt.Println("\nClient read num bytes: ", bytes)
		repDataObj = shared.ConvJsonToDataStruct(buf[:bytes])

		if repDataObj.Type == shared.MEASUREMENT_INFO_REPLY {
			fmt.Printf("\nClient received an Measurement_Info_Reply!!!")
			break
		}
	}
	return repDataObj
}
