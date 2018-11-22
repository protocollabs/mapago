package shared

type ManageConn interface {
	WriteAnswer([]byte)
	CloseConn()
}
