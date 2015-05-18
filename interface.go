package link

import "net"
import "time"

type SessionAble interface {
	Id() uint64
	Conn() net.Conn
	IsClosed() bool

	AddCloseCallback(handler interface{}, callback func())

	Send(message Message, now time.Time) error
	ReadPacket() (data []byte, err error) // this is for debug ,donot use this in product environment.

	SendBytes(data []byte, now time.Time) error
	Process(decoder Decoder) error

	Close()
}
