package shared

// channel result

const (
	INFO_REQUEST = 1
	INFO_REPLY = 2
	MEASUREMENT_START_REQUESTS = 3
	MEASUREMENT_START_REPLY = 4
	MEASUREMENT_STOP_REQUEST = 5
	MEASUREMENT_STOP_REPLY = 6
	MEASUREMENT_INFO_REQUEST = 7
	MEASUREMENT_INFO_REPLY = 8
	TIME_DIFF_REQUEST = 9
	TIME_DIFF_REPLY = 10
	WARNING_ERR_MSG = 255
)

type ChResult struct {
	Json []byte
	ConnObj WriterAnswer
}

type DataObj struct {
	// type is currently 1 byte in type field: 0x02 etc.
	// not part of the actual json blob...so could be string too
	Type uint64 `json:",omitempty"`
	Id string `json:",omitempty"`
	Seq string `json:",omitempty"`
	// TODO: check if time.Data can be handled
	Ts string `json:",omitempty"`
	Secret string `json:",omitempty"`
	Seq_rp string `json:",omitempty"`
	Arch string `json:",omitempty"`
	Os string `json:",omitempty"`
	Info string `json:",omitempty"`
	Status string `json:",omitempty"`
	Measurement_delay string `json:",omitempty"`
	Measurement_time_max string `json:",omitempty"`
	Padding string `json:",omitempty"`
	Modules string
	/*
	TODO: how to handle
	Measurement
	ControlProtocol
	*/
}