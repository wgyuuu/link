package main

import (
	"fmt"
	"github.com/0studio/link"
	"time"
)

// This is an echo client demo work with the echo_server.
// usage:
//     go run echo_client/main.go
func main() {
	link.DefaultProtocol = link.AuthPacketN(4, "1111222233334444", link.BigEndian, 13175046)
	client, err := link.Dial("tcp", "127.0.0.1:10010")
	if err != nil {
		panic(err)
	}
	go client.Process(func(msg *link.InBuffer) error {
		println(string(msg.Data))
		return nil
	})

	for {
		var input string
		if _, err := fmt.Scanf("%s\n", &input); err != nil {
			break
		}
		client.Send(link.Bytes([]byte(input)), time.Now())
	}

	client.Close()

	println("bye")
}
