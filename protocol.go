package link

import (
	"encoding/binary"
	"io"
)

var (
	BigEndian    = ByteOrder(binary.BigEndian)
	LittleEndian = ByteOrder(binary.LittleEndian)

	packet1BE = newSimpleProtocol(1, BigEndian)
	packet1LE = newSimpleProtocol(1, LittleEndian)
	packet2BE = newSimpleProtocol(2, BigEndian)
	packet2LE = newSimpleProtocol(2, LittleEndian)
	packet4BE = newSimpleProtocol(4, BigEndian)
	packet4LE = newSimpleProtocol(4, LittleEndian)
	packet8BE = newSimpleProtocol(8, BigEndian)
	packet8LE = newSimpleProtocol(8, LittleEndian)
)

type ByteOrder binary.ByteOrder

type ProtocolSide int

const (
	SERVER_SIDE ProtocolSide = 1
	CLIENT_SIDE ProtocolSide = 2
)

// Packet protocol.
type Protocol interface {
	// Create protocol state.
	// New(net.Conn) for session protocol state.
	// New(*Server) for server protocol state.
	// New(*Channel) for channel protocol state.
	// If the protocol need handshake for connection initialization.
	// Do it in Protocol.New() and returns nil and a error when handshake failed.
	New(interface{}, ProtocolSide) (ProtocolState, error)
	// test
	DecodeAuth([]byte) int
	EncodeAuth(buf *OutBuffer, msg Message, messageSize int)
}

// Protocol state.
type ProtocolState interface {
	// Packet a message.
	WriteToBuffer(buffer *OutBuffer, message Message) error

	// Write a outBuffer.
	Write(writer io.Writer, buffer *OutBuffer) error

	// Read a outBuffer.
	Read(reader io.Reader, buffer *InBuffer) error
}

// Create a {outBuffer, N} protocol.
// The n means how many bytes of the outBuffer header.
// n must is 1、2、4 or 8.
func PacketN(n int, byteOrder ByteOrder, maxPacketReadSize, maxPacketWriteSize int) Protocol {
	switch n {
	case 1:
		switch byteOrder {
		case BigEndian:

			return packet1BE.setMaxPacketSize(maxPacketReadSize, maxPacketWriteSize)
		case LittleEndian:
			return packet1LE.setMaxPacketSize(maxPacketReadSize, maxPacketWriteSize)
		}
	case 2:
		switch byteOrder {
		case BigEndian:
			return packet2BE.setMaxPacketSize(maxPacketReadSize, maxPacketWriteSize)
		case LittleEndian:
			return packet2LE.setMaxPacketSize(maxPacketReadSize, maxPacketWriteSize)
		}
	case 4:
		switch byteOrder {
		case BigEndian:
			return packet4BE.setMaxPacketSize(maxPacketReadSize, maxPacketWriteSize)
		case LittleEndian:
			return packet4LE.setMaxPacketSize(maxPacketReadSize, maxPacketWriteSize)
		}
	case 8:
		switch byteOrder {
		case BigEndian:
			return packet8BE.setMaxPacketSize(maxPacketReadSize, maxPacketWriteSize)
		case LittleEndian:
			return packet8LE.setMaxPacketSize(maxPacketReadSize, maxPacketWriteSize)
		}
	}
	panic("unsupported outBuffer head size")
}

// The outBuffer spliting protocol like Erlang's {outBuffer, N}.
// Each outBuffer has a fix length outBuffer header to present outBuffer length.
type simpleProtocol struct {
	n                  int
	bo                 binary.ByteOrder
	encodeHead         func(message Message, msgSize int, out *OutBuffer)
	decodeHead         func([]byte) int
	maxPacketReadSize  int
	maxPacketWriteSize int
}

func (p *simpleProtocol) setMaxPacketSize(maxPacketReadSize, maxPacketWriteSize int) *simpleProtocol {
	p.maxPacketReadSize = maxPacketReadSize
	p.maxPacketWriteSize = maxPacketWriteSize
	return p

}

func newSimpleProtocol(n int, byteOrder binary.ByteOrder) *simpleProtocol {
	protocol := &simpleProtocol{
		n:  n,
		bo: byteOrder,
	}

	switch n {
	case 1:
		protocol.encodeHead = func(message Message, msgSize int, buffer *OutBuffer) {
			buffer.WriteUint8(uint8(msgSize))
		}
		protocol.decodeHead = func(buffer []byte) int {
			return int(buffer[0])
		}
	case 2:
		protocol.encodeHead = func(message Message, msgSize int, buffer *OutBuffer) {
			buffer.WriteUint16(uint16(msgSize), protocol.bo)
		}
		protocol.decodeHead = func(buffer []byte) int {
			return int(protocol.bo.Uint16(buffer))
		}
	case 4:
		protocol.encodeHead = func(message Message, msgSize int, buffer *OutBuffer) {
			buffer.WriteUint32(uint32(msgSize), protocol.bo)
		}
		protocol.decodeHead = func(buffer []byte) int {
			return int(protocol.bo.Uint32(buffer))
		}
	case 8:
		protocol.encodeHead = func(message Message, msgSize int, buffer *OutBuffer) {
			buffer.WriteUint64(uint64(msgSize), protocol.bo)
		}
		protocol.decodeHead = func(buffer []byte) int {
			return int(protocol.bo.Uint64(buffer))
		}
	default:
		panic("unsupported outBuffer head size")
	}

	return protocol
}

func (p *simpleProtocol) New(v interface{}, _ ProtocolSide) (ProtocolState, error) {
	return p, nil
}

func (p *simpleProtocol) WriteToBuffer(buffer *OutBuffer, message Message) error {
	msgSize := message.Size()
	buffer.Prepare(p.n + msgSize)
	if p.maxPacketWriteSize > 0 && msgSize > p.maxPacketWriteSize {
		return PacketTooLargeForWriteError
	}
	p.EncodeAuth(buffer, message, msgSize)
	return buffer.WriteMessage(message)
}

func (p *simpleProtocol) Write(writer io.Writer, outBuffer *OutBuffer) error {
	if len(outBuffer.Data) == 0 || outBuffer.pos == 0 {
		return nil
	}

	if _, err := writer.Write(outBuffer.GetData()); err != nil {
		return err
	}
	return nil
}

func (p *simpleProtocol) Read(reader io.Reader, buffer *InBuffer) error {
	// head
	buffer.Prepare(p.n)
	if _, err := io.ReadFull(reader, buffer.Data); err != nil {
		return err
	}
	if buffer.Data == nil || len(buffer.Data) == 0 {
		return PacketTooLargeforReadError
	}
	size := p.DecodeAuth(buffer.Data)
	if p.maxPacketReadSize > 0 && size > p.maxPacketReadSize {
		return PacketTooLargeforReadError
	}
	// body
	if size == 0 {
		return nil
	}
	buffer.Prepare(size)
	if _, err := io.ReadFull(reader, buffer.Data); err != nil {
		return err
	}
	return nil
}

func (p *simpleProtocol) EncodeAuth(buffer *OutBuffer, message Message, msgSize int) {
	p.encodeHead(message, msgSize, buffer)
}

func (p *simpleProtocol) DecodeAuth(bytes []byte) int {
	return p.decodeHead(bytes)
}
