package link

import (
	"github.com/funny/unitest"
	"runtime"
	"testing"
)

func TestBufferPrepare(t *testing.T) {
	var buffer = &OutBuffer{Data: make([]byte, 3, 3)}
	buffer.Data[0] = 1
	buffer.Data[1] = 2
	buffer.Data[2] = 3
	buffer.Prepare(1)

	unitest.Pass(t, len(buffer.Data) == 4)
	unitest.Pass(t, cap(buffer.Data) == 4)
	unitest.Pass(t, buffer.Data[0] == 1)
	unitest.Pass(t, buffer.Data[1] == 2)
	unitest.Pass(t, buffer.Data[2] == 3)
	unitest.Pass(t, buffer.Data[3] == 0)
}

func TestBufferPrepare2(t *testing.T) {
	var buffer = &OutBuffer{Data: make([]byte, 1, 3)}
	buffer.pos = 1
	buffer.Data[0] = 1
	buffer.Prepare(3)

	unitest.Pass(t, len(buffer.Data) == 4)
	unitest.Pass(t, cap(buffer.Data) == 4)
	unitest.Pass(t, buffer.Data[0] == 1)
	unitest.Pass(t, buffer.Data[1] == 0)
	unitest.Pass(t, buffer.Data[2] == 0)
	unitest.Pass(t, buffer.Data[3] == 0)
}
func TestBuffer(t *testing.T) {
	var buffer = newOutBuffer()

	PrepareBuffer(buffer)

	VerifyBuffer(t, &InBuffer{Data: buffer.Data})
}

func TestBuffer2(t *testing.T) {
	var buffer = newOutBufferWithDefaultCap(0)

	PrepareBuffer(buffer)

	VerifyBuffer(t, &InBuffer{Data: buffer.Data})
}

func PrepareBuffer(buffer *OutBuffer) {
	buffer.Prepare(1)
	buffer.WriteUint8(123)
	buffer.Prepare(2)
	buffer.WriteUint16LE(0xFFEE)
	buffer.Prepare(2)
	buffer.WriteUint16BE(0xFFEE)
	buffer.Prepare(4)
	buffer.WriteUint32LE(0xFFEEDDCC)
	buffer.Prepare(4)
	buffer.WriteUint32BE(0xFFEEDDCC)
	buffer.Prepare(8)
	buffer.WriteUint64LE(0xFFEEDDCCBBAA9988)
	buffer.Prepare(8)
	buffer.WriteUint64BE(0xFFEEDDCCBBAA9988)
}

func VerifyBuffer(t *testing.T, buffer *InBuffer) {
	unitest.Pass(t, buffer.ReadUint8() == 123)
	unitest.Pass(t, buffer.ReadUint16LE() == 0xFFEE)
	unitest.Pass(t, buffer.ReadUint16BE() == 0xFFEE)
	unitest.Pass(t, buffer.ReadUint32LE() == 0xFFEEDDCC)
	unitest.Pass(t, buffer.ReadUint32BE() == 0xFFEEDDCC)
	unitest.Pass(t, buffer.ReadUint64LE() == 0xFFEEDDCCBBAA9988)
	unitest.Pass(t, buffer.ReadUint64BE() == 0xFFEEDDCCBBAA9988)
}

func Benchmark_SetFinalizer1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var x = &InBuffer{}
		runtime.SetFinalizer(x, func(x *InBuffer) {
		})
	}
}

func Benchmark_SetFinalizer2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var x = &InBuffer{}
		runtime.SetFinalizer(x, nil)
	}
}

func Benchmark_MakeBytes_512(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = make([]byte, 512)
	}
}

func Benchmark_MakeBytes_1024(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = make([]byte, 1024)
	}
}

func Benchmark_MakeBytes_4096(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = make([]byte, 4096)
	}
}

func Benchmark_MakeBytes_8192(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = make([]byte, 8192)
	}
}
