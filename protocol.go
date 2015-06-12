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
}

// Protocol state.
type ProtocolState interface {
	// Packet a message.
	PrepareOutBuffer(buffer *OutBuffer, size int) error
	PrepareData(buffer *OutBuffer, message Message) error

	// Write a packet.
	Write(writer io.Writer, buffer *OutBuffer) error

	// Read a packet.
	Read(reader io.Reader, buffer *InBuffer) error
}

// Create a {packet, N} protocol.
// The n means how many bytes of the packet header.
// n must is 1、2、4 or 8.
func PacketN(n int, byteOrder ByteOrder, MaxPacketSize int) Protocol {
	switch n {
	case 1:
		switch byteOrder {
		case BigEndian:

			return packet1BE.setMaxPacketSize(MaxPacketSize)
		case LittleEndian:
			return packet1LE.setMaxPacketSize(MaxPacketSize)
		}
	case 2:
		switch byteOrder {
		case BigEndian:
			return packet2BE.setMaxPacketSize(MaxPacketSize)
		case LittleEndian:
			return packet2LE.setMaxPacketSize(MaxPacketSize)
		}
	case 4:
		switch byteOrder {
		case BigEndian:
			return packet4BE.setMaxPacketSize(MaxPacketSize)
		case LittleEndian:
			return packet4LE.setMaxPacketSize(MaxPacketSize)
		}
	case 8:
		switch byteOrder {
		case BigEndian:
			return packet8BE.setMaxPacketSize(MaxPacketSize)
		case LittleEndian:
			return packet8LE.setMaxPacketSize(MaxPacketSize)
		}
	}
	panic("unsupported packet head size")
}

// The packet spliting protocol like Erlang's {packet, N}.
// Each packet has a fix length packet header to present packet length.
type simpleProtocol struct {
	n             int
	bo            binary.ByteOrder
	encodeHead    func(message Message, out *OutBuffer)
	decodeHead    func([]byte) int
	MaxPacketSize int
}

func (p *simpleProtocol) setMaxPacketSize(MaxPacketSize int) *simpleProtocol {
	p.MaxPacketSize = MaxPacketSize
	return p

}

func newSimpleProtocol(n int, byteOrder binary.ByteOrder) *simpleProtocol {
	protocol := &simpleProtocol{
		n:  n,
		bo: byteOrder,
	}

	switch n {
	case 1:
		protocol.encodeHead = func(message Message, buffer *OutBuffer) {
			buffer.WriteUint8(uint8(message.OutBufferSize()))
		}
		protocol.decodeHead = func(buffer []byte) int {
			return int(buffer[0])
		}
	case 2:
		protocol.encodeHead = func(message Message, buffer *OutBuffer) {
			buffer.WriteUint16(uint16(message.OutBufferSize()), byteOrder)
		}
		protocol.decodeHead = func(buffer []byte) int {
			return int(byteOrder.Uint16(buffer))
		}
	case 4:
		protocol.encodeHead = func(message Message, buffer *OutBuffer) {
			buffer.WriteUint32(uint32(message.OutBufferSize()), byteOrder)
		}
		protocol.decodeHead = func(buffer []byte) int {
			return int(byteOrder.Uint32(buffer))
		}
	case 8:
		protocol.encodeHead = func(message Message, buffer *OutBuffer) {
			buffer.WriteUint64(uint64(message.OutBufferSize()), byteOrder)
		}
		protocol.decodeHead = func(buffer []byte) int {
			return int(byteOrder.Uint64(buffer))
		}
	default:
		panic("unsupported packet head size")
	}

	return protocol
}

func (p *simpleProtocol) New(v interface{}, _ ProtocolSide) (ProtocolState, error) {
	return p, nil
}

func (p *simpleProtocol) PrepareOutBuffer(buffer *OutBuffer, size int) error {
	buffer.Prepare(p.n + size)
	return nil
}
func (p *simpleProtocol) PrepareData(buffer *OutBuffer, message Message) error {
	if p.MaxPacketSize > 0 && message.OutBufferSize() > p.MaxPacketSize {
		return PacketTooLargeForWriteError
	}
	p.encodeHead(message, buffer)
	return buffer.WriteMessage(p, message)
}

func (p *simpleProtocol) Write(writer io.Writer, packet *OutBuffer) error {
	if _, err := writer.Write(packet.Data); err != nil {
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
	size := p.decodeHead(buffer.Data)
	if p.MaxPacketSize > 0 && size > p.MaxPacketSize {
		return PacketTooLargeforReadError
	}
	// body
	buffer.Prepare(size)
	if size == 0 {
		return nil
	}
	if _, err := io.ReadFull(reader, buffer.Data); err != nil {
		return err
	}
	return nil
}
