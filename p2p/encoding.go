package p2p

import (
	"encoding/gob"
	"io"
)

type Decoder interface {
	Decode(io.Reader, *Message) error
}

// In Go, a type implements an interface by implementing
// all the methods defined in the interface.
type GOBDecoder struct{}

func (dec GOBDecoder) Decode(r io.Reader, msg *Message) error {
	return gob.NewDecoder(r).Decode(msg)
}

type DefaultDecoder struct{}

func (dec DefaultDecoder) Decode(r io.Reader, msg *Message) error {
	buf := make([]byte, 1028)
	println("inside the decoder")
	n, err := r.Read(buf)
	println("bute read")
	if err != nil {
		return err
	}
	msg.Payload = buf[:n]
	return nil
}
