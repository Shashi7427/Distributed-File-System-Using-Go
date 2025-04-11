package p2p

import (
	"encoding/gob"
	"io"
)

type Decoder interface {
	Decode(io.Reader, *RCP) error
}

// In Go, a type implements an interface by implementing
// all the methods defined in the interface.
type GOBDecoder struct{}

func (dec GOBDecoder) Decode(r io.Reader, msg *RCP) error {
	return gob.NewDecoder(r).Decode(msg)
}

type DefaultDecoder struct{}

func (dec DefaultDecoder) Decode(r io.Reader, msg *RCP) error {
	buf := make([]byte, 1028)
	println("inside the decoder")
	n, err := r.Read(buf)
	println("byte read")
	if err != nil {
		return err
	}
	msg.Payload = buf[:n]
	return nil
}
