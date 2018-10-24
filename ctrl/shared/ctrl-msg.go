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
	Type uint64
	Id string
	Seq uint64
	// TODO: check if time.Data can be handled
	Ts string
	Secret string
	Seq_rp uint64
	Arch string
	Os string
	Info string
	Status string
	Measurement_delay uint32
	Measurement_time_max uint32
	Padding string
	/*
	TODO: how to handle
	Modules []string
	Measurement
	ControlProtocol
	*/
}