package link

import (
	"bufio"
	"fmt"
	"net"
	"sync"
)

// Buffered connection.
type bufferConn struct {
	net.Conn
	reader *bufio.Reader
}

func newBufferConn(conn net.Conn, readBufferSize int) *bufferConn {
	return &bufferConn{
		conn,
		bufio.NewReaderSize(conn, readBufferSize),
	}
}

func (conn *bufferConn) Read(d []byte) (int, error) {
	return conn.reader.Read(d)
}

var bufferConnPool sync.Pool

func getBufferConnFromPool(conn net.Conn, readBufferSize int) *bufferConn {
	obj := bufferConnPool.Get()
	if obj == nil {
		fmt.Println("getBufferConnFromPool_miss", conn.RemoteAddr().String())
		return newBufferConn(conn, readBufferSize)
	}
	fmt.Println("getBufferConnFromPool_hit", conn.RemoteAddr().String())
	return obj.(*bufferConn)
}

func putBufferConnToPool(conn net.Conn) {
	if s, ok := conn.(*bufferConn); ok {
		fmt.Println("debug,putBufferConnToPool", conn.RemoteAddr().String())
		bufferConnPool.Put(s)
	}
}
