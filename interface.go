package link

type SessionAble interface {
	Id() uint64
	Conn() net.Conn
	IsClosed() bool

	Send(message Message) error

	SendPacket(packet []byte) error
}

