package link

import (
	"container/list"
	"net"
	"sync/atomic"
	"time"

	"fmt"
	"sync"
)

var dialSessionId uint64

// The easy way to create a connection.
func Dial(network, address string) (*Session, error) {
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	id := atomic.AddUint64(&dialSessionId, 1)
	return NewSession(id, conn, DefaultProtocol, CLIENT_SIDE, DefaultSendChanSize, DefaultConnBufferSize)
}

// The easy way to create a connection with timeout setting.
func DialTimeout(network, address string, timeout time.Duration) (*Session, error) {
	conn, err := net.DialTimeout(network, address, timeout)
	if err != nil {
		return nil, err
	}
	id := atomic.AddUint64(&dialSessionId, 1)
	return NewSession(id, conn, DefaultProtocol, CLIENT_SIDE, DefaultSendChanSize, DefaultConnBufferSize)
}

type Decoder func(*InBuffer) error

// Session.
type Session struct {
	id uint64

	// About network
	conn     net.Conn
	protocol ProtocolState

	// About send and receive
	readMutex           sync.Mutex
	sendMutex           sync.Mutex
	asyncSendChan       chan asyncMessage
	asyncSendBufferChan chan asyncBuffer
	inBuffer            InBuffer
	outBuffer           OutBuffer
	outBufferMutex      sync.Mutex

	// About session close
	closeChan       chan int
	closeFlag       int32
	closeEventMutex sync.Mutex
	closeCallbacks  *list.List

	createTime   time.Time
	lastSendTime time.Time
	lastRecvTime time.Time
	// Put your session state here.
	State interface{}
}

func NewSession(id uint64, conn net.Conn, protocol Protocol, side ProtocolSide, sendChanSize int, readBufferSize int) (*Session, error) {
	if readBufferSize > 0 {
		conn = newBufferConn(conn, readBufferSize)
	}
	return newSession(id, conn, protocol, side, sendChanSize)
}

// Create a new session instance.
func newSession(id uint64, conn net.Conn, protocol Protocol, side ProtocolSide, sendChanSize int) (*Session, error) {
	protocolState, err := protocol.New(conn, side)
	if err != nil {
		return nil, err
	}

	session := &Session{
		id:                  id,
		conn:                conn,
		protocol:            protocolState,
		asyncSendChan:       make(chan asyncMessage, sendChanSize),
		asyncSendBufferChan: make(chan asyncBuffer, sendChanSize),
		inBuffer:            newInBuffer(),
		outBuffer:           newOutBuffer(),
		closeChan:           make(chan int),
		closeCallbacks:      list.New(),
		createTime:          time.Now(),
	}

	defer func() {
		if e := recover(); e != nil {
			fmt.Println("link.session.ERROR", e)
		}
	}()
	go session.sendLoop()

	return session, nil
}

// Get session id.
func (session *Session) Id() uint64 {
	return session.id
}

// Get session connection.
func (session *Session) Conn() net.Conn {
	return session.conn
}

// Check session is closed or not.
func (session *Session) IsClosed() bool {
	return atomic.LoadInt32(&session.closeFlag) != 0
}

// Close session.
func (session *Session) Close() {
	if atomic.CompareAndSwapInt32(&session.closeFlag, 0, 1) {
		session.conn.Close()

		// exit send loop and cancel async send
		close(session.closeChan)

		session.invokeCloseCallbacks()

		// session.inBuffer = nil
		// session.outBuffer = nil
	}
}

func (session *Session) GetLastSendTime() time.Time {
	return session.lastSendTime
}
func (session *Session) GetCreateTime() time.Time {
	return session.createTime
}

func (session *Session) GetLastRecvTime() time.Time {
	return session.lastRecvTime
}

var zeroTime time.Time

func (session *Session) SendNow(message Message) error {
	return session.Send(message, time.Now())
}
func (session *Session) SendDefault(message Message) error {
	return session.Send(message, zeroTime)
}
func (session *Session) SendBytesDefault(data []byte) error {
	return session.SendBytes(data, zeroTime)
}

func (session *Session) SendBytesNow(data []byte) error {
	return session.SendBytes(data, time.Now())
}

// do not use this ,use session.Send(Message,time.Time) ,this is not effective
//  pb->[]byte-> headerSize+[]byte
// if you implements your Message interface ,it directly pb->headerSize+[]byte
func (session *Session) SendBytes(data []byte, now time.Time) error {
	return session.Send(Bytes(data), now)
}

// Sync send a message. This method will block on IO.
func (session *Session) Send(message Message, now time.Time) error {
	session.outBufferMutex.Lock()
	defer session.outBufferMutex.Unlock()
	err := session.protocol.WriteToBuffer(&session.outBuffer, message)

	if err == nil {
		err = session.sendBuffer(&session.outBuffer)
	}

	session.outBuffer.reset()
	session.lastSendTime = now
	return err
}

func (session *Session) sendBuffer(buffer *OutBuffer) error {
	session.sendMutex.Lock()
	defer session.sendMutex.Unlock()

	return session.protocol.Write(session.conn, buffer)
}

// Process one request.
func (session *Session) ProcessOnce(decoder Decoder) error {
	session.readMutex.Lock()
	defer session.readMutex.Unlock()

	err := session.protocol.Read(session.conn, &session.inBuffer)
	if err != nil {
		session.inBuffer.reset()
		session.Close()
		return err
	}
	session.lastRecvTime = time.Now()

	err = decoder(&session.inBuffer)
	session.inBuffer.reset()

	return nil
}

// Process request.
func (session *Session) Process(decoder Decoder) error {
	for {
		if err := session.ProcessOnce(decoder); err != nil {
			return err
		}
	}
}

func (session *Session) ReadPacket() (data []byte, err error) {
	// [Warning]:do not use this, use session.Process
	// session.Read() just for debug
	session.readMutex.Lock()
	defer session.readMutex.Unlock()

	err = session.protocol.Read(session.conn, &session.inBuffer)
	if err != nil {
		session.inBuffer.reset()
		session.Close()
		return
	}
	session.lastRecvTime = time.Now()

	// this is slow
	data = make([]byte, len(session.inBuffer.Data))
	copy(data, session.inBuffer.Data)
	session.inBuffer.reset()

	return
}

// Async work.
type AsyncWork struct {
	c <-chan error
}

// Wait work done. Returns error when work failed.
func (aw AsyncWork) Wait() error {
	return <-aw.c
}

type asyncMessage struct {
	C chan<- error
	M Message
}

type asyncBuffer struct {
	C chan<- error
	B *OutBuffer
}

// Loop and transport responses.
func (session *Session) sendLoop() {
	for {
		select {
		case buffer := <-session.asyncSendBufferChan:
			buffer.C <- session.sendBuffer(buffer.B)
			// buffer.B.broadcastFree()
		case message := <-session.asyncSendChan:
			message.C <- session.Send(message.M, time.Now())
		case <-session.closeChan:
			return
		}
	}
}

// Async send a message.
func (session *Session) AsyncSend(message Message, timeout time.Duration) AsyncWork {
	c := make(chan error, 1)
	if session.IsClosed() {
		c <- SendToClosedError
	} else {
		select {
		case session.asyncSendChan <- asyncMessage{c, message}:
		default:
			if timeout == 0 {
				session.Close()
				c <- AsyncSendTimeoutError
			} else {
				go func() {
					select {
					case session.asyncSendChan <- asyncMessage{c, message}:
					case <-session.closeChan:
						c <- SendToClosedError
					case <-time.After(timeout):
						session.Close()
						c <- AsyncSendTimeoutError
					}
				}()
			}
		}
	}
	return AsyncWork{c}
}

// Async send a packet.
func (session *Session) AsyncSendBuffer(buffer *OutBuffer, timeout time.Duration) AsyncWork {
	c := make(chan error, 1)
	if session.IsClosed() {
		c <- SendToClosedError
	} else {
		select {
		case session.asyncSendBufferChan <- asyncBuffer{c, buffer}:
		default:
			if timeout == 0 {
				session.Close()
				c <- AsyncSendTimeoutError
			} else {
				go func() {
					select {
					case session.asyncSendBufferChan <- asyncBuffer{c, buffer}:
					case <-session.closeChan:
						c <- SendToClosedError
					case <-time.After(timeout):
						session.Close()
						c <- AsyncSendTimeoutError
					}
				}()
			}
		}
	}
	return AsyncWork{c}
}

type closeCallback struct {
	Handler interface{}
	Func    func()
}

// Add close callback.
func (session *Session) AddCloseCallback(handler interface{}, callback func()) {
	if session.IsClosed() {
		return
	}

	session.closeEventMutex.Lock()
	defer session.closeEventMutex.Unlock()

	session.closeCallbacks.PushBack(closeCallback{handler, callback})
}

// Remove close callback.
func (session *Session) RemoveCloseCallback(handler interface{}) {
	if session.IsClosed() {
		return
	}

	session.closeEventMutex.Lock()
	defer session.closeEventMutex.Unlock()

	for i := session.closeCallbacks.Front(); i != nil; i = i.Next() {
		if i.Value.(closeCallback).Handler == handler {
			session.closeCallbacks.Remove(i)
			return
		}
	}
}

// Dispatch close event.
func (session *Session) invokeCloseCallbacks() {
	session.closeEventMutex.Lock()
	defer session.closeEventMutex.Unlock()

	for i := session.closeCallbacks.Front(); i != nil; i = i.Next() {
		callback := i.Value.(closeCallback)
		callback.Func()
	}
}

func (session *Session) GetState() (State interface{}) {
	// Get your session state here.
	return session.State
}
func (session *Session) SetState(State interface{}) {
	// Put your session state here.
	session.State = State
}
