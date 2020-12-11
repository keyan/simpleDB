// package rpc includes the shared message types between client and server.
// Both client and server use rpc to communicate via the Msg{} type. Convinence
// methods are provided that wrap encoding/gob to hide the details of communication
// serialization from client/server. The intention is to have the Msg{} type sent
// as raw bytes in a HTTP POST body.
package rpc

import (
	"bytes"
	"encoding/gob"
	"io"
	"log"
)

// ValueType is the data type stored in the database for each key-value mapping.
type ValueType []byte

// Operation indicates what requested action the caller is specifying.
type Operation int

const (
	Get Operation = iota
	Set
	Delete
)

// Msg is the shared type for communication between client/server. This simplifies
// client calls by having a single call into the server to provide all the required
// actions. The Value field allows for sending arbitrarily complex data as raw bytes.
type Msg struct {
	Op    Operation
	Key   string
	Value ValueType
}

// DecodeMsg converts raw bytes (assumed to be from a HTTP request) to a Msg.
func DecodeMsg(r io.ReadCloser) *Msg {
	dec := gob.NewDecoder(r)
	var msg Msg
	err := dec.Decode(&msg)
	if err != nil {
		log.Fatal("Could not decode message")
	}
	return &msg
}

// NewGetMsg returns a byte buffer resulting for serialization of a Get Msg.
func NewGetMsg(key string) *bytes.Buffer {
	var msgBytes bytes.Buffer
	enc := gob.NewEncoder(&msgBytes)
	msg := Msg{
		Op:    Get,
		Key:   key,
		Value: nil,
	}
	enc.Encode(msg)
	return &msgBytes
}

// NewSetMsg returns a byte buffer resulting for serialization of a Set Msg.
func NewSetMsg(key string, val ValueType) *bytes.Buffer {
	var msgBytes bytes.Buffer
	enc := gob.NewEncoder(&msgBytes)
	msg := Msg{
		Op:    Set,
		Key:   key,
		Value: val,
	}
	enc.Encode(msg)
	return &msgBytes
}

// NewDeleteMsg returns a byte buffer resulting for serialization of a Delete Msg.
func NewDeleteMsg(key string) *bytes.Buffer {
	var msgBytes bytes.Buffer
	enc := gob.NewEncoder(&msgBytes)
	msg := Msg{
		Op:  Delete,
		Key: key,
	}
	enc.Encode(msg)
	return &msgBytes
}
