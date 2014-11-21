package link

import "net"

type SessionAble interface {
	Id() uint64
	Conn() net.Conn
	IsClosed() bool

	Send(message Message) error

	SendPacket(OutMessage) error

	Read() (InMessage, error)
	Close(reason interface{})
}
