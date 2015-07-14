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

func getBufferConnFromPool(conn net.Conn, readBufferSize int) (bc *bufferConn) {
	obj := bufferConnPool.Get()
	if obj == nil {
		fmt.Println("getBufferConnFromPool_miss", conn.RemoteAddr().String())
		return newBufferConn(conn, readBufferSize)
	}
	fmt.Println("getBufferConnFromPool_hit", conn.RemoteAddr().String())
	bc = obj.(*bufferConn)
	bc.reader.Reset(conn)
	bc.Conn = conn
	return
}

func putBufferConnToPool(session *Session) {
	if s, ok := session.Conn().(*bufferConn); ok {
		fmt.Println("debug,putBufferConnToPool", session.Conn().RemoteAddr().String())
		bufferConnPool.Put(s)
		session.conn = s.Conn
	}
}
