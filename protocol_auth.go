package link

import (
	"encoding/binary"
	"fmt"
	"github.com/0studio/link/util"
	"io"
	"math/rand"
)

var (
	Auth_Version  uint32 = 1
	Version_Len   int    = 4
	Random_Len    int    = 4
	Encrypt_Len   int    = 4
	authPacket1BE        = newAuthProtocol(1, 12, BigEndian)
	authPacket1LE        = newAuthProtocol(1, 12, LittleEndian)
	authPacket2BE        = newAuthProtocol(2, 12, BigEndian)
	authPacket2LE        = newAuthProtocol(2, 12, LittleEndian)
	authPacket4BE        = newAuthProtocol(4, 12, BigEndian)
	authPacket4LE        = newAuthProtocol(4, 12, LittleEndian)
	authPacket8BE        = newAuthProtocol(8, 12, BigEndian)
	authPacket8LE        = newAuthProtocol(8, 12, LittleEndian)
)

// Create a {packet, N} protocol.
// The n means how many bytes of the packet header.
// n must is 1、2、4 or 8.
func AuthPacketN(n int, authKey string, byteOrder ByteOrder, MaxPacketSize int) Protocol {
	switch n {
	case 1:
		switch byteOrder {
		case BigEndian:
			return authPacket1BE.setMaxPacketSize(MaxPacketSize, authKey)
		case LittleEndian:
			return authPacket1LE.setMaxPacketSize(MaxPacketSize, authKey)
		}
	case 2:
		switch byteOrder {
		case BigEndian:
			return authPacket2BE.setMaxPacketSize(MaxPacketSize, authKey)
		case LittleEndian:
			return authPacket2LE.setMaxPacketSize(MaxPacketSize, authKey)
		}
	case 4:
		switch byteOrder {
		case BigEndian:
			return authPacket4BE.setMaxPacketSize(MaxPacketSize, authKey)
		case LittleEndian:
			return authPacket4LE.setMaxPacketSize(MaxPacketSize, authKey)
		}
	case 8:
		switch byteOrder {
		case BigEndian:
			return authPacket8BE.setMaxPacketSize(MaxPacketSize, authKey)
		case LittleEndian:
			return authPacket8LE.setMaxPacketSize(MaxPacketSize, authKey)
		}
	}
	panic("unsupported packet head size")
}

// The packet spliting protocol like Erlang's {packet, N}.
// Each packet has a fix length packet header to present packet length.
type authProtocol struct {
	n             int
	c             int
	key           string
	bo            binary.ByteOrder
	encodeHead    func(message Message, out *OutBuffer)
	decodeHead    func([]byte) int
	MaxPacketSize int
}

func (p *authProtocol) setMaxPacketSize(MaxPacketSize int, authKey string) *authProtocol {
	p.MaxPacketSize = MaxPacketSize
	p.key = authKey
	return p
}

func newAuthProtocol(n, c int, byteOrder binary.ByteOrder) *authProtocol {
	protocol := &authProtocol{
		n:  n,
		c:  c,
		bo: byteOrder,
	}

	switch n {
	case 1:
		protocol.encodeHead = func(message Message, buffer *OutBuffer) {
			buffer.WriteUint8(uint8(message.Size()))
		}
		protocol.decodeHead = func(buffer []byte) int {
			return int(buffer[0])
		}
	case 2:
		protocol.encodeHead = func(message Message, buffer *OutBuffer) {
			buffer.WriteUint16(uint16(message.Size()), protocol.bo)
		}
		protocol.decodeHead = func(buffer []byte) int {
			return int(protocol.bo.Uint16(buffer))
		}
	case 4:
		protocol.encodeHead = func(message Message, buffer *OutBuffer) {
			buffer.WriteUint32(uint32(message.Size()), protocol.bo)
		}
		protocol.decodeHead = func(buffer []byte) int {
			return int(protocol.bo.Uint32(buffer))
		}
	case 8:
		protocol.encodeHead = func(message Message, buffer *OutBuffer) {
			buffer.WriteUint64(uint64(message.Size()), protocol.bo)
		}
		protocol.decodeHead = func(buffer []byte) int {
			return int(protocol.bo.Uint64(buffer))
		}
	default:
		panic("unsupported packet head size")
	}

	return protocol
}

func (p *authProtocol) New(v interface{}, _ ProtocolSide) (ProtocolState, error) {
	return p, nil
}

func (p *authProtocol) WriteToBuffer(buffer *OutBuffer, message Message) error {
	msgSize := message.Size()
	buffer.Prepare(p.n + p.c + msgSize)
	if p.MaxPacketSize > 0 && msgSize > p.MaxPacketSize {
		return PacketTooLargeForWriteError
	}
	p.EncodeAuth(buffer, message)
	return buffer.WriteMessage(message)
}

func (p *authProtocol) Write(writer io.Writer, packet *OutBuffer) error {
	if len(packet.Data) == 0 || packet.pos == 0 {
		return nil
	}

	if _, err := writer.Write(packet.Data[0:packet.pos]); err != nil {
		return err
	}
	return nil
}

func (p *authProtocol) Read(reader io.Reader, buffer *InBuffer) error {
	// head
	buffer.Prepare(p.n + p.c)
	if _, err := io.ReadFull(reader, buffer.Data); err != nil {
		return err
	}
	if buffer.Data == nil || len(buffer.Data) == 0 {
		return PacketTooLargeforReadError
	}
	size := p.DecodeAuth(buffer.Data)
	if p.MaxPacketSize > 0 && size > p.MaxPacketSize {
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

func (p *authProtocol) DecodeAuth(bytes []byte) int {
	if len(bytes) < (p.n + p.c) {
		return 0
	}
	// check
	decData := util.Md5Encrypt(string(bytes[:len(bytes)-Encrypt_Len]))
	fmt.Println("================", bytes[len(bytes)-Encrypt_Len:], ">>>", []byte(decData[:Encrypt_Len]))
	if string(bytes[len(bytes)-Encrypt_Len:]) != decData[:Encrypt_Len] {
		fmt.Println("----------->>>>>>")
		return 0
	}
	// return size
	sizeData := bytes[:p.n]
	return p.decodeHead(sizeData)
}

func (p *authProtocol) EncodeAuth(buffer *OutBuffer, message Message) {
	encBuffer := newOutBufferWithDefaultCap(p.c + 4)
	encBuffer.Prepare(p.c + Version_Len + Random_Len)
	p.encodeHead(message, encBuffer)
	encBuffer.WriteUint32(Auth_Version, p.bo)
	encBuffer.WriteUint32(uint32(rand.Intn(message.Size())), p.bo)

	bytes := util.Md5Encrypt(string(encBuffer.GetData()))
	encData := util.LenString(Encrypt_Len, bytes)

	p.encodeHead(message, buffer)
	buffer.WriteString(string(encBuffer.GetData()))
	buffer.WriteString(string(encData))
}
