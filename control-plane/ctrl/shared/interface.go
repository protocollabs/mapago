package shared

import "net"

type ManageConn interface {
	WriteAnswer([]byte)
	CloseConn()
	DetectRemoteAddr() net.Addr
}
