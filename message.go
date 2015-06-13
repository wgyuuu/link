package link

import (
// "encoding/gob"
// "encoding/json"
// "encoding/xml"
)

type Message interface {
	Size() int
	MarshalTo(dest []byte) (n int, err error)
}

// if you use protobuf , I suggest using  "github.com/gogo/protobuf/proto"
// and it generate  a Size() and MarshalTo(data []byte) (n int, err error)
// you can use it directly

// do not use this , this is not effective ,(if you use Bytes()  you golang Data-------->[]byte-------->dataSize+[]byte)
// if you write your Message then your directly                      golangData---------->dataSize+[]byte
// Convert to bytes message.
func Bytes(v []byte) (m Message) {
	return BytesMessage(v)
}

type BytesMessage []byte

func (message BytesMessage) Size() int {
	return len(message)
}
func (message BytesMessage) MarshalTo(buffer []byte) (n int, err error) {
	if len(buffer) < len(message) {
		return 0, BufferSizeNotEnough
	}
	copy(buffer, message)

	n = len(message)
	return

}
func String(str string) (m Message) {
	return Bytes([]byte(str))
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
