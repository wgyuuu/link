package link

import (
	// "encoding/gob"
	// "encoding/json"
	// "encoding/xml"
	"errors"
)

type Message interface {
	OutBufferSize() int
	MarshalTo(buffer *OutBuffer) (n int, err error)
	// WriteOutBuffer(*OutBuffer) error
}

type BytesMessage struct {
	data []byte
}

func (message BytesMessage) OutBufferSize() int {
	return len(message.data)
}
func (message BytesMessage) MarshalTo(buffer *OutBuffer) (n int, err error) {
	container := buffer.GetContainer()
	if len(container) < len(message.data) {
		return 0, errors.New("buffer_size_not_enough")
	}
	copy(container, message.data)
	n = len(message.data)
	return

}

// Convert to bytes message.
func Bytes(v []byte) (m Message) {
	return BytesMessage{v}
}

// // Convert to string message.
// func String(v string) Message {
// 	return MessageFunc(func(buffer *OutBuffer) error {
// 		buffer.WriteString(v)
// 		return nil
// 	})
// }

// // Create a json message.
// func Json(v interface{}) Message {
// 	return MessageFunc(func(buffer *OutBuffer) error {
// 		return json.NewEncoder(buffer).Encode(v)
// 	})
// }

// // Create a gob message.
// func Gob(v interface{}) Message {
// 	return MessageFunc(func(buffer *OutBuffer) error {
// 		return gob.NewEncoder(buffer).Encode(v)
// 	})
// }

// // Create a xml message.
// func Xml(v interface{}) Message {
// 	return MessageFunc(func(buffer *OutBuffer) error {
// 		return xml.NewEncoder(buffer).Encode(v)
// 	})
// }
