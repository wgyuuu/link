package link

import "net"

type SessionAble interface {
	Id() uint64
	Conn() net.Conn
	IsClosed() bool
	Close()
	AddCloseCallback(handler interface{}, callback func())

	Send(message Message) error

	SendBytes(data []byte) error

	// Read() (*InBuffer, error)
}
