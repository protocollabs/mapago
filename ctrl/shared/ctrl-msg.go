package shared

// channel result

type ChResult struct {
	Json []byte
	ConnObj WriterAnswer
}

type DataObj struct {
	// type is currently 1 byte in type field: 0x02 etc.
	Type uint64
	Id string
	MsmtId uint64
	Seq uint64
	/* not mandatory atm:
	Timestamp time.Date
	Secret string
	SeqRp uint64
	Modules []string => could be also [][]string
	Os string
	Arch string
	Measurement [][]string
	ControlProtocol [][]string
	*/
}