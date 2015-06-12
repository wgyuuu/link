package link

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"sync"
	"sync/atomic"
	"unicode/utf8"
)

var (
	enableBufferPool = true
	globalPool       = newBufferPool()
)

// Turn On/Off buffer pool. Default is enable.
func BufferPoolEnable(enable bool) {
	enableBufferPool = enable
}

type bufferPool struct {
	inPool  sync.Pool
	outPool sync.Pool
}

func newBufferPool() *bufferPool {
	return &bufferPool{
		inPool:  sync.Pool{New: newInBufferObj},
		outPool: sync.Pool{New: newOutBufferObj},
	}
}

func (pool *bufferPool) GetInBuffer() (in *InBuffer) {
	bufferObj := pool.inPool.Get()
	return bufferObj.(*InBuffer)
}

func (pool *bufferPool) GetOutBuffer() (out *OutBuffer) {
	bufferObj := pool.outPool.Get()
	return bufferObj.(*OutBuffer)
}

func (pool *bufferPool) PutInBuffer(in *InBuffer) {
	pool.inPool.Put(in)
}

func (pool *bufferPool) PutOutBuffer(out *OutBuffer) {
	pool.outPool.Put(out)
}

// Incomming message buffer.
type InBuffer struct {
	Data    []byte // Buffer data.
	ReadPos int    // Read position.
	isFreed bool
}

func newInBufferObj() interface{} {
	return &InBuffer{
		Data: make([]byte, 0, 1024),
	}
}

func newInBuffer() *InBuffer {
	if enableBufferPool == true {
		return globalPool.GetInBuffer()
	}
	return &InBuffer{Data: make([]byte, 0, 1024)}
}

func (in *InBuffer) reset() {
	in.Data = in.Data[0:0]
	in.ReadPos = 0
}

// Return the buffer to buffer pool.
func (in *InBuffer) free() {
	if enableBufferPool {
		if in.isFreed {
			panic("link.InBuffer: double free")
		}
		in.reset()
		globalPool.PutInBuffer(in)
	}
}

// Prepare buffer for next message.
// This method is for custom protocol only.
// Dont' use it in application logic.
func (in *InBuffer) Prepare(size int) {
	if cap(in.Data) < size {
		in.Data = make([]byte, size)
	} else {
		in.Data = in.Data[0:size]
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
	x := make([]byte, n)
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
	Data        []byte // Buffer data.
	isFreed     bool
	isBroadcast bool
	refCount    int32
	pos         int
}

func newOutBufferObj() interface{} {
	return &OutBuffer{Data: make([]byte, 0, 1024)}
}

func newOutBuffer() *OutBuffer {
	if enableBufferPool == true {
		return globalPool.GetOutBuffer()
	}
	return &OutBuffer{Data: make([]byte, 0, 1024)}
}

func newOutBufferWithDefaultCap(cap int) *OutBuffer {
	return &OutBuffer{Data: make([]byte, 0, cap)}
}
func (out *OutBuffer) broadcastUse() {
	if enableBufferPool {
		atomic.AddInt32(&out.refCount, 1)
	}
}

func (out *OutBuffer) broadcastFree() {
	if enableBufferPool {
		if out.isBroadcast && atomic.AddInt32(&out.refCount, -1) == 0 {
			out.free()
		}
	}
}

func (out *OutBuffer) reset() {
	out.Data = out.Data[0:0]
	out.pos = 0
}

// Return the buffer to buffer pool.
func (out *OutBuffer) free() {
	if enableBufferPool {
		if out.isFreed {
			panic("link.OutBuffer: double free")
		}
		out.reset()
		globalPool.PutOutBuffer(out)
	}
}

// Prepare for next message.
// This method is for custom protocol only.
// Don't use it in application logic.
func (out *OutBuffer) Prepare(size int) {
	if cap(out.Data)-out.pos < size {
		data := make([]byte, out.pos+size, out.pos+size)
		fmt.Println("outpos", out.pos, len(out.Data))
		if out.pos > 0 && len(out.Data) > 0 {
			copy(data, out.Data[0:out.pos])
		}
		out.Data = data
	} else {
		out.Data = out.Data[0 : out.pos+size]
	}
	fmt.Printf("after_prepare out.pos=%d,len(data)=%d,cap(data)=%d\n", out.pos, len(out.Data), cap(out.Data))
}
func (out *OutBuffer) GetContainer() (data []byte) {
	data = out.Data[out.pos:]
	fmt.Println("container.len", len(data), out.pos, len(out.Data))
	return
}

func (out *OutBuffer) WriteMessage(protocol ProtocolState, message Message) (err error) {
	var n int
	n, err = message.MarshalTo(out)
	out.pos += n
	return
}

// Write a uint8 value into buffer.
func (out *OutBuffer) WriteUint8(v uint8) {
	out.GetContainer()[0] = byte(v)
	out.pos++
}

func (out *OutBuffer) WriteUint16(v uint16, order binary.ByteOrder) {
	order.PutUint16(out.GetContainer(), v)
	out.pos += 2
}

// Write a uint16 value into buffer using little endian byte order.
func (out *OutBuffer) WriteUint16LE(v uint16) {
	binary.LittleEndian.PutUint16(out.GetContainer(), v)
	out.pos += 2
}

// Write a uint16 value into buffer using big endian byte order.
func (out *OutBuffer) WriteUint16BE(v uint16) {
	binary.BigEndian.PutUint16(out.GetContainer(), v)
	out.pos += 2
}
func (out *OutBuffer) WriteUint32(v uint32, order binary.ByteOrder) {
	order.PutUint32(out.GetContainer(), v)
	out.pos += 4
}

// Write a uint32 value into buffer using little endian byte order.
func (out *OutBuffer) WriteUint32LE(v uint32) {
	binary.LittleEndian.PutUint32(out.GetContainer(), v)
	out.pos += 4
}

// Write a uint32 value into buffer using big endian byte order.
func (out *OutBuffer) WriteUint32BE(v uint32) {
	binary.BigEndian.PutUint32(out.GetContainer(), v)
	out.pos += 4
}

func (out *OutBuffer) WriteUint64(v uint64, order binary.ByteOrder) {
	order.PutUint64(out.GetContainer(), v)
	out.pos += 8
}

// Write a uint64 value into buffer using little endian byte order.
func (out *OutBuffer) WriteUint64LE(v uint64) {
	binary.LittleEndian.PutUint64(out.GetContainer(), v)
	out.pos += 8
}

// Write a uint64 value into buffer using big endian byte order.
func (out *OutBuffer) WriteUint64BE(v uint64) {
	binary.BigEndian.PutUint64(out.GetContainer(), v)
	out.pos += 8
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
