package link

import "net"
import "time"

type SessionAble interface {
	Id() uint64
	Conn() net.Conn
	IsClosed() bool

	AddCloseCallback(handler interface{}, callback func())
	RemoveCloseCallback(handler interface{})

	Send(message Message, now time.Time) error
	ReadPacket() (data []byte, err error) // this is for debug ,donot use this in product environment.

	// push a message to buffer,call SendBufferedMessage to flush to client
	PushToBuffer(message Message) error
	SendBufferedMessage(now time.Time) error
	// SendBytes(data []byte, now time.Time) error
	Process(decoder Decoder) error

	GetState() (State interface{})
	SetState(State interface{})
	GetLastRecvTime() time.Time
	GetLastSendTime() time.Time
	GetCreateTime() time.Time
	Close()
}
