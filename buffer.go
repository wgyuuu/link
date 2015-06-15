package link

import (
	"encoding/binary"
	"github.com/0studio/link/buffer"
	"io"
	"math"
	"unicode/utf8"
)

var (
	DefaultInBuffSize  = 128
	DefaultOutBuffSize = 128
	globalPool         = newBufferPool(DefaultInBuffSize, DefaultOutBuffSize)
)

type bufferPool struct {
	inBufferMgr  *buffer.BufferPoolMgr
	outBufferMgr *buffer.BufferPoolMgr
}

func newBufferPool(defaultInBufferSize, defaultOutBufferSize int) *bufferPool {
	return &bufferPool{
		inBufferMgr:  buffer.NewBufferPoolMgr(defaultInBufferSize),
		outBufferMgr: buffer.NewBufferPoolMgr(defaultOutBufferSize),
	}
}

func (pool *bufferPool) PutOutDataBuffer(data []byte) {
	pool.outBufferMgr.Put(data)

}

func (pool *bufferPool) GetOutDataBuffer(size int) (data []byte) {
	return pool.outBufferMgr.Get(size)
}
func (pool *bufferPool) PutInDataBuffer(data []byte) {
	pool.inBufferMgr.Put(data)
}

func (pool *bufferPool) GetInDataBuffer(size int) (data []byte) {
	return pool.inBufferMgr.Get(size)

}

// Incomming message buffer.
type InBuffer struct {
	Data    []byte // Buffer data.
	ReadPos int    // Read position.
}

func newInBuffer() *InBuffer {
	return &InBuffer{Data: globalPool.GetInDataBuffer(DefaultInBuffSize)}
}

func (in *InBuffer) reset() {
	in.ReadPos = 0
	globalPool.PutInDataBuffer(in.Data)
	in.Data = nil
}

// Prepare buffer for next message.
// This method is for custom protocol only.
// Dont' use it in application logic.
func (in *InBuffer) Prepare(size int) {
	if in.Data == nil {
		in.Data = globalPool.GetInDataBuffer(size)
	}

	if cap(in.Data) < size {
		if len(in.Data) != 0 {
			globalPool.PutInDataBuffer(in.Data)
		}

		in.Data = globalPool.GetInDataBuffer(size)
	} else {
		if len(in.Data) != size {
			in.Data = in.Data[0:size]
		}
	}
}

// Slice some bytes from buffer.
func (in *InBuffer) Slice(n int) []byte {
	r := in.Data[in.ReadPos : in.ReadPos+n]
	in.ReadPos += n
	return r
}

// Implement io.Reader interface
func (in *InBuffer) Read(b []byte) (int, error) {
	if in.ReadPos == len(in.Data) {
		return 0, io.EOF
	}
	n := len(b)
	if n+in.ReadPos > len(in.Data) {
		n = len(in.Data) - in.ReadPos
	}
	copy(b, in.Data[in.ReadPos:])
	in.ReadPos += n
	return n, nil
}

// Read some bytes from buffer.
func (in *InBuffer) ReadBytes(n int) []byte {
	x := make([]byte, n, n)
	copy(x, in.Slice(n))
	return x
}

// Read a string from buffer.
func (in *InBuffer) ReadString(n int) string {
	return string(in.Slice(n))
}

// Read a rune from buffer.
func (in *InBuffer) ReadRune() rune {
	x, size := utf8.DecodeRune(in.Data[in.ReadPos:])
	in.ReadPos += size
	return x
}

// Read a uint8 value from buffer.
func (in *InBuffer) ReadUint8() uint8 {
	return uint8(in.Slice(1)[0])
}

// Read a uint16 value from buffer using little endian byte order.
func (in *InBuffer) ReadUint16LE() uint16 {
	return binary.LittleEndian.Uint16(in.Slice(2))
}

// Read a uint16 value from buffer using big endian byte order.
func (in *InBuffer) ReadUint16BE() uint16 {
	return binary.BigEndian.Uint16(in.Slice(2))
}

// Read a uint32 value from buffer using little endian byte order.
func (in *InBuffer) ReadUint32LE() uint32 {
	return binary.LittleEndian.Uint32(in.Slice(4))
}

// Read a uint32 value from buffer using big endian byte order.
func (in *InBuffer) ReadUint32BE() uint32 {
	return binary.BigEndian.Uint32(in.Slice(4))
}

// Read a uint64 value from buffer using little endian byte order.
func (in *InBuffer) ReadUint64LE() uint64 {
	return binary.LittleEndian.Uint64(in.Slice(8))
}

// Read a uint64 value from buffer using big endian byte order.
func (in *InBuffer) ReadUint64BE() uint64 {
	return binary.BigEndian.Uint64(in.Slice(8))
}

// Read a float32 value from buffer using little endian byte order.
func (in *InBuffer) ReadFloat32LE() float32 {
	return math.Float32frombits(in.ReadUint32LE())
}

// Read a float32 value from buffer using big endian byte order.
func (in *InBuffer) ReadFloat32BE() float32 {
	return math.Float32frombits(in.ReadUint32BE())
}

// Read a float64 value from buffer using little endian byte order.
func (in *InBuffer) ReadFloat64LE() float64 {
	return math.Float64frombits(in.ReadUint64LE())
}

// Read a float64 value from buffer using big endian byte order.
func (in *InBuffer) ReadFloat64BE() float64 {
	return math.Float64frombits(in.ReadUint64BE())
}

// ReadVarint reads an encoded signed integer from buffer and returns it as an int64.
func (in *InBuffer) ReadVarint() int64 {
	v, n := binary.Varint(in.Data[in.ReadPos:])
	in.ReadPos += n
	return v
}

// ReadUvarint reads an encoded unsigned integer from buffer and returns it as a uint64.
func (in *InBuffer) ReadUvarint() uint64 {
	v, n := binary.Uvarint(in.Data[in.ReadPos:])
	in.ReadPos += n
	return v
}

// Outgoing message buffer.
type OutBuffer struct {
	Data []byte // Buffer data.
	pos  int
}

func newOutBuffer() *OutBuffer {
	return &OutBuffer{Data: globalPool.GetOutDataBuffer(DefaultOutBuffSize)}
}

func newOutBufferWithDefaultCap(cap int) *OutBuffer {
	return &OutBuffer{Data: make([]byte, 0, cap)}
}

func (out *OutBuffer) reset() {
	out.pos = 0
	globalPool.PutOutDataBuffer(out.Data)

	// out.Data = out.Data[0:0]
	out.Data = nil
}
func (out *OutBuffer) IsEmpty() bool {
	return len(out.Data)-out.pos <= 0
}

// Prepare for next message.
// This method is for custom protocol only.
// Don't use it in application logic.
func (out *OutBuffer) Prepare(size int) {
	if out.Data == nil {
		out.pos = 0
		out.Data = globalPool.GetOutDataBuffer(size)
	}

	pos := out.pos

	if cap(out.Data)-pos < size {
		data := globalPool.GetOutDataBuffer(pos + size)
		if out.pos > 0 && len(out.Data) > 0 {
			copy(data, out.Data[0:out.pos])
		}
		oldData := out.Data
		globalPool.PutOutDataBuffer(oldData)
		out.Data = data
	} else {
		out.Data = out.Data[0 : pos+size]
	}
}
func (out *OutBuffer) GetContainer() (data []byte) {
	if out.Data == nil {
		out.Data = globalPool.GetOutDataBuffer(DefaultOutBuffSize)
		out.pos = 0
	}

	data = out.Data[out.pos:]
	return
}

// 	you should call out.Prepare(message.Size()) first
func (out *OutBuffer) WriteMessage(message Message) (err error) {
	var n int
	n, err = message.MarshalTo(out.GetContainer())
	out.pos += n
	return
}

// Write a uint8 value into buffer.
func (out *OutBuffer) WriteUint8(v uint8) bool {
	container := out.GetContainer()
	if len(container) < 1 {
		return false
	}

	container[0] = byte(v)
	out.pos += 1
	return true
}

func (out *OutBuffer) WriteUint16(v uint16, order binary.ByteOrder) bool {
	container := out.GetContainer()
	if len(container) < 2 {
		return false
	}

	order.PutUint16(out.GetContainer(), v)
	out.pos += 2
	return true
}

// Write a uint16 value into buffer using little endian byte order.
func (out *OutBuffer) WriteUint16LE(v uint16) bool {
	container := out.GetContainer()
	if len(container) < 2 {
		return false
	}

	binary.LittleEndian.PutUint16(out.GetContainer(), v)
	out.pos += 2
	return true
}

// Write a uint16 value into buffer using big endian byte order.
func (out *OutBuffer) WriteUint16BE(v uint16) bool {
	container := out.GetContainer()
	if len(container) < 2 {
		return false
	}

	binary.BigEndian.PutUint16(out.GetContainer(), v)
	out.pos += 2
	return true
}
func (out *OutBuffer) WriteUint32(v uint32, order binary.ByteOrder) bool {
	container := out.GetContainer()
	if len(container) < 4 {
		return false
	}
	order.PutUint32(out.GetContainer(), v)
	out.pos += 4
	return true
}

// Write a uint32 value into buffer using little endian byte order.
func (out *OutBuffer) WriteUint32LE(v uint32) bool {
	container := out.GetContainer()
	if len(container) < 4 {
		return false
	}

	binary.LittleEndian.PutUint32(out.GetContainer(), v)
	out.pos += 4
	return true
}

// Write a uint32 value into buffer using big endian byte order.
func (out *OutBuffer) WriteUint32BE(v uint32) bool {
	container := out.GetContainer()
	if len(container) < 4 {
		return false
	}

	binary.BigEndian.PutUint32(out.GetContainer(), v)
	out.pos += 4
	return true
}

func (out *OutBuffer) WriteUint64(v uint64, order binary.ByteOrder) bool {
	container := out.GetContainer()
	if len(container) < 8 {
		return false
	}

	order.PutUint64(out.GetContainer(), v)
	out.pos += 8
	return true
}

// Write a uint64 value into buffer using little endian byte order.
func (out *OutBuffer) WriteUint64LE(v uint64) bool {
	container := out.GetContainer()
	if len(container) < 8 {
		return false
	}

	binary.LittleEndian.PutUint64(out.GetContainer(), v)
	out.pos += 8
	return true
}

// Write a uint64 value into buffer using big endian byte order.
func (out *OutBuffer) WriteUint64BE(v uint64) bool {
	container := out.GetContainer()
	if len(container) < 8 {
		return false
	}

	binary.BigEndian.PutUint64(out.GetContainer(), v)
	out.pos += 8
	return true
}
func (out *OutBuffer) WriteString(s string) bool {
	container := out.GetContainer()
	if len(container) < 8 {
		return false
	}
	copy(container, []byte(s))
	return true
}

// func (out *OutBuffer) Write(p []byte) (int, error) {
// 	out.Data = append(out.Data, p...)
// 	return len(p), nil
// }

// Append some bytes into buffer.
// func (out *OutBuffer) Append(p ...byte) {
// 	out.Data = append(out.Data, p...)
// }

// // Implement io.Writer interface.
// func (out *OutBuffer) Write(p []byte) (int, error) {
// 	out.Data = append(out.Data, p...)
// 	return len(p), nil
// }

// // Write a byte slice into buffer.
// func (out *OutBuffer) WriteBytes(d []byte) {
// 	out.Append(d...)
// }

// // Write a string into buffer.
// func (out *OutBuffer) WriteString(s string) {
// 	out.Append([]byte(s)...)
// }

// // Write a rune into buffer.
// func (out *OutBuffer) WriteRune(r rune) {
// 	p := []byte{0, 0, 0, 0}
// 	n := utf8.EncodeRune(p, r)
// 	out.Append(p[:n]...)
// }

// // Write a float32 value into buffer using little endian byte order.
// func (out *OutBuffer) WriteFloat32LE(v float32) {
// 	out.WriteUint32LE(math.Float32bits(v))
// }

// // Write a float32 value into buffer using big endian byte order.
// func (out *OutBuffer) WriteFloat32BE(v float32) {
// 	out.WriteUint32BE(math.Float32bits(v))
// }

// // Write a float64 value into buffer using little endian byte order.
// func (out *OutBuffer) WriteFloat64LE(v float64) {
// 	out.WriteUint64LE(math.Float64bits(v))
// }

// // Write a float64 value into buffer using big endian byte order.
// func (out *OutBuffer) WriteFloat64BE(v float64) {
// 	out.WriteUint64BE(math.Float64bits(v))
// }

// // Write a uint64 value into buffer.
// func (out *OutBuffer) WriteUvarint(v uint64) {
// 	for v >= 0x80 {
// 		out.Append(byte(v) | 0x80)
// 		v >>= 7
// 	}
// 	out.Append(byte(v))
// }

// // Write a int64 value into buffer.
// func (out *OutBuffer) WriteVarint(v int64) {
// 	ux := uint64(v) << 1
// 	if v < 0 {
// 		ux = ^ux
// 	}
// 	out.WriteUvarint(ux)
// }
