package link

import "net"

type SessionAble interface {
	Id() uint64
	Conn() net.Conn
	IsClosed() bool

	AddCloseCallback(handler interface{}, callback func())

	Send(message Message) error
	ReadPacket() (data []byte, err error) // this is for debug ,donot use this in product environment.

	SendBytes(data []byte) error
	Process(decoder Decoder) error

	Close()
}
