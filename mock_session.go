package link

import (
	"fmt"
	"net"
	"time"
)

type MockSession struct {
	id       uint64
	mockConn MockConn
}

type MockConn struct {
	sendPacketChan chan []byte
}

func (this MockConn) RemoteAddr() net.Addr {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return addrs[1]
}

func (this MockConn) Read(b []byte) (n int, err error) {
	return 0, nil
}

func (this MockConn) Write(b []byte) (n int, err error) {
	return 0, nil
}

func (this MockConn) Close() error {
	return nil
}

func (this MockConn) LocalAddr() net.Addr {
	return this.RemoteAddr()
}

func (this MockConn) SetDeadline(t time.Time) error {
	return nil
}

func (this MockConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (this MockConn) SetWriteDeadline(t time.Time) error {
	return nil
}

func NewMockSession(id int) *MockSession {
	bytesChan := make(chan []byte, 100)
	mockConn := MockConn{
		sendPacketChan: bytesChan,
	}
	return &MockSession{
		id:       uint64(id),
		mockConn: mockConn,
	}
}

func (session *MockSession) Start() {
}

func (session *MockSession) Id() uint64 {
	return session.id
}

func (session *MockSession) Conn() net.Conn {
	return session.mockConn
}

func (session *MockSession) IsClosed() bool {
	return false
}

func (session *MockSession) SyncSendPacket(packet []byte) error {
	return nil
}

func (session *MockSession) Send(message Message, now time.Time) error {
	return nil
}
func (session *MockSession) SendBytes(data []byte, now time.Time) error {
	return nil
}
func (session *MockSession) ReadPacket() (data []byte, err error) { // this is for debug ,donot use this in product environment.
	return
}

func (session *MockSession) AddCloseCallback(handler interface{}, callback func()) {
}
func (session *MockSession) RemoveCloseCallback(handler interface{}) {

}
func (session *MockSession) Process(decoder Decoder) error {
	return nil

}

func (session *MockSession) Close() {
	close(session.mockConn.sendPacketChan)
}

func (session *MockSession) Read() (*InBuffer, error) {
	select {
	case <-time.After(time.Second * 2):
		return nil, nil
	case bytes := <-session.mockConn.sendPacketChan:
		inBuffer := &InBuffer{}
		inBuffer.Data = bytes
		return inBuffer, nil
	}
}

func (session *MockSession) GetState() (State interface{}) {
	return nil
}
func (session *MockSession) SetState(State interface{}) {

}
func (session *MockSession) GetLastRecvTime() time.Time {
	return time.Now()
}
func (session *MockSession) GetLastSendTime() time.Time {
	return time.Now()
}
func (session *MockSession) GetCreateTime() time.Time {
	return time.Now()
}
func (session *MockSession) PushToBuffer(message Message) error {
	return nil
}
func (session *MockSession) SendBufferedMessage(now time.Time) error {
	return nil
}
func (session *MockSession) AsyncSendBuffer(buffer *OutBuffer, timeout time.Duration) (w AsyncWork) {
	return
}
func (session *MockSession) SendNow(message Message) error {
	return session.Send(message, time.Now())
}
func (session *MockSession) SendDefault(message Message) error {
	return session.Send(message, zeroTime)
}
func (session *MockSession) SendBytesDefault(data []byte) error {
	return session.SendBytes(data, zeroTime)
}

func (session *MockSession) SendBytesNow(data []byte) error {
	return session.SendBytes(data, time.Now())
}
