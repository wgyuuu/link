package link

import (
	"encoding/binary"
	"io"
	"math/rand"

	"github.com/0studio/link/util"
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

// Create a {outBuffer, N} protocol.
// The n means how many bytes of the outBuffer header.
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
	panic("unsupported outBuffer head size")
}

// The outBuffer spliting protocol like Erlang's {outBuffer, N}.
// Each outBuffer has a fix length outBuffer header to present outBuffer length.
type authProtocol struct {
	n             int
	c             int
	key           string
	bo            binary.ByteOrder
	encodeHead    func(message Message, msgSize int, out *OutBuffer)
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

func (p *authProtocol) New(v interface{}, _ ProtocolSide) (ProtocolState, error) {
	return p, nil
}

func (p *authProtocol) WriteToBuffer(buffer *OutBuffer, message Message) error {
	msgSize := message.Size()
	buffer.Prepare(p.n + p.c + msgSize)
	if p.MaxPacketSize > 0 && msgSize > p.MaxPacketSize {
		return PacketTooLargeForWriteError
	}
	p.EncodeAuth(buffer, message, msgSize)
	return buffer.WriteMessage(message)
}

func (p *authProtocol) Write(writer io.Writer, outBuffer *OutBuffer) error {
	if len(outBuffer.Data) == 0 || outBuffer.pos == 0 {
		return nil
	}

	if _, err := writer.Write(outBuffer.GetData()); err != nil {
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
	data := string(bytes[:len(bytes)-Encrypt_Len])
	encrtptString := string(bytes[len(bytes)-Encrypt_Len:])
	if md5String := util.Md5Encrypt(data); md5String[:Encrypt_Len] == encrtptString {
	} else if hexmd5String := util.HexMd5Encrypt(data); hexmd5String[:Encrypt_Len] == encrtptString {
	} else { // 验证失败
		return 0
	}
	// return size
	sizeData := bytes[:p.n]
	return p.decodeHead(sizeData)
}

func (p *authProtocol) EncodeAuth(buffer *OutBuffer, message Message, msgSize int) {
	encBuffer := NewOutBufferWithDefaultCap(p.c + 4)
	encBuffer.Prepare(p.c)
	p.encodeHead(message, msgSize, encBuffer)
	encBuffer.WriteUint32(Auth_Version, p.bo)
	encBuffer.WriteUint32(uint32(rand.Intn(msgSize)), p.bo)

	bytes := util.Md5Encrypt(string(encBuffer.GetData()))
	encData := util.LenString(Encrypt_Len, bytes)

	buffer.WriteString(string(encBuffer.GetData()))
	buffer.WriteString(string(encData))
}
