package shared

// channel result

const (
	INFO_REQUEST              = 1
	INFO_REPLY                = 2
	MEASUREMENT_START_REQUEST = 3
	MEASUREMENT_START_REPLY   = 4
	MEASUREMENT_STOP_REQUEST  = 5
	MEASUREMENT_STOP_REPLY    = 6
	MEASUREMENT_INFO_REQUEST  = 7
	MEASUREMENT_INFO_REPLY    = 8
	TIME_DIFF_REQUEST         = 9
	TIME_DIFF_REPLY           = 10
	WARNING_ERR_MSG           = 255
)

const (
	IP4_LINK_LOCAL_MC = "224.0.0.1"
	IP6_LINK_LOCAL_MC = "FF02::1"
	CONTROL_PORT      = 64321
)

type ChResult struct {
	Json    []byte
	ConnObj ManageConn
}

/*
reply sent from the measurement module to control:
"i carried out your commanded duty XYZ, this is the corresponding reply"
*/
type ChMsmt2Ctrl struct {
	Status string
	Data   interface{}
}

/*
request sent from the management plane to a measurement module:
"measurement module do thing XYZ"
*/
type ChMgmt2Msmt struct {
	Cmd    string
	MsmtId string
}

/*
- NOTE: DISCONTINUED! This is now only used by the unfinished UDP msmt
- reply sent from within the (TCP/UDP/QUIC) measurement module
whats actually received: "i received 666 bytes"
*/
type ChMsmtResult struct {
	Bytes uint64
	Time  float64
}

type MsmtStorageEntry struct {
	MsmtCh chan ChMgmt2Msmt
	// this could be: tcpMsmtObj, udpMsmtObj, quicMsmtObj
	MsmtObj interface{}
}

type DataObj struct {
	Type                 uint64         `json:",omitempty"`
	Id                   string         `json:",omitempty"`
	Seq                  string         `json:",omitempty"`
	Measurement_id       string         `json:",omitempty"`
	Ts                   string         `json:",omitempty"`
	Secret               string         `json:",omitempty"`
	Seq_rp               string         `json:",omitempty"`
	Arch                 string         `json:",omitempty"`
	Os                   string         `json:",omitempty"`
	Info                 string         `json:",omitempty"`
	Status               string         `json:",omitempty"`
	Message              string         `json:",omitempty"`
	Measurement_delay    string         `json:",omitempty"`
	Measurement_time_max string         `json:",omitempty"`
	Padding              string         `json:",omitempty"`
	Modules              string         `json:",omitempty"`
	Measurement          MeasurementObj `json:",omitempty"`
	Data                 DataCollectionObj  `json:",omitempty"`
	
	/*
		TODO: how to handle
		ControlProtocol
	*/
}

type MeasurementObj struct {
	Name          string           `json:",omitempty"`
	Type          string           `json:",omitempty"`
	Configuration ConfigurationObj `json:",omitempty"`
}

type ConfigurationObj struct {
	Worker          string `json:",omitempty"`
	Payload_pattern string `json:",omitempty"`
	Port            string `json:",omitempty"`
	Rate            string `json:",omitempty"`
	Rate_burst      string `json:",omitempty"`
	Packet_length   string `json:",omitempty"`
	Dscp            string `json:",omitempty"`
	Ttl             string `json:",omitempty"`
	Output_format   string `json:",omitempty"`
	Tx_method       string `json:",omitempty"`

	// useful other fields?!
	Listen_addr string `json:",omitempty"`
	Call_size   string `json:",omitempty"`
}

type DataCollectionObj struct {
	DataElement []DataResultObj
}

type DataResultObj struct {
	Timestamp_first string `json:",omitempty"`
	Timestamp_last  string `json:",omitempty"`
	Received_bytes  string `json:",omitempty"`
}
