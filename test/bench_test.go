package test

import (
	"github.com/0studio/link"
	"testing"
)

func BenchmarkAuthDecode(b *testing.B) {
	buf := []byte{0, 0, 0, 10, 0, 0, 0, 1, 0, 0, 0, 1, 211, 198, 47, 13}
	protocol := link.AuthPacketN(4, "1111222233334444", link.BigEndian, 13175046)
	for i := 0; i < b.N; i++ {
		protocol.DecodeAuth(buf)
	}
}

func BenchmarkDecode(b *testing.B) {
	buf := []byte{0, 0, 0, 10}
	protocol := link.PacketN(4, link.BigEndian, 13175046)
	for i := 0; i < b.N; i++ {
		protocol.DecodeAuth(buf)
	}
}

func BenchmarkAuthEncode(b *testing.B) {
	bytes := []byte("adasdadadaasd")
	protocol := link.AuthPacketN(4, "1111222233334444", link.BigEndian, 13175046)
	for i := 0; i < b.N; i++ {
		buffer := createBuffer()
		protocol.EncodeAuth(buffer, link.BytesMessage(bytes))
	}
}

func BenchmarkEncode(b *testing.B) {
	bytes := []byte("adasdadadaasd")
	protocol := link.PacketN(4, link.BigEndian, 13175046)
	for i := 0; i < b.N; i++ {
		buffer := createBuffer()
		protocol.EncodeAuth(buffer, link.BytesMessage(bytes))
	}
}

func createBuffer() *link.OutBuffer {
	return &link.OutBuffer{
		Data: make([]byte, 64),
	}
}
