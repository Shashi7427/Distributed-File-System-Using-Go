package p2p

import "errors"

// error returned if the handshake between local and remote note could not be established
var ErrInvalidHandshake = errors.New("invalid handshake")

// HandshakeFunc
type HandshakeFunc func(Peer) error

func NOPHandshakeFunc(Peer) error {return nil}