package main

import (
	"flag"
	"fmt"
	"github.com/0studio/link"
	"time"
)

var (
	benchmark  = flag.Bool("bench", false, "is for benchmark, will disable print")
	buffersize = flag.Int("buffer", 1024, "session read buffer size")
)

func log(v ...interface{}) {
	if !*benchmark {
		fmt.Println(v...)
	}
}

// This is an echo server demo work with the echo_client.
// usage:
//     go run echo_server/main.go
func main() {
	flag.Parse()

	link.DefaultProtocol = link.AuthPacketN(4, "1111222233334444", link.BigEndian, 13175046)
	link.DefaultConnBufferSize = *buffersize

	server, err := link.Listen("tcp", "127.0.0.1:10010")
	if err != nil {
		panic(err)
	}

	println("server start:", server.Listener().Addr().String())

	server.Serve(func(session link.SessionAble) {
		log("client", session.Conn().RemoteAddr().String(), "in")

		session.Process(func(msg *link.InBuffer) error {
			log("client", session.Conn().RemoteAddr().String(), "say:", string(msg.Data))
			return session.Send(link.Bytes(msg.Data), time.Now())
		})

		log("client", session.Conn().RemoteAddr().String(), "close")
	})
}
