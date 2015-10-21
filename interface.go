package link

import "net"
import "time"

type SessionAble interface {
	Id() uint64
	Conn() net.Conn
	IsClosed() bool

	AddCloseCallback(handler interface{}, callback func())
	RemoveCloseCallback(handler interface{})

	SendDefault(message Message) error
	SendNow(message Message) error
	Send(message Message, now time.Time) error

	SendBytesDefault(data []byte) error         //  this is not effective , please use  Send
	SendBytesNow(data []byte) error             //  this is not effective , please use  Send
	SendBytes(data []byte, now time.Time) error //  this is not effective , please use  Send

	ReadPacket() (data []byte, err error) // this is for debug ,donot use this in product environment.

	Process(decoder Decoder) error

	GetState() (State interface{})
	SetState(State interface{})
	GetLastRecvTime() time.Time
	GetLastSendTime() time.Time
	GetCreateTime() time.Time
	Close()
	AsyncSendBuffer(buffer *OutBuffer, timeout time.Duration) AsyncWork
}
