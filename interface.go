package link

import "net"

type SessionAble interface {
	Id() uint64
	Conn() net.Conn
	IsClosed() bool

	Send(message Message) error

	SendPacket(packet Packet) error

	Read() (*InBuffer, error)
	Close(reason interface{})
}
