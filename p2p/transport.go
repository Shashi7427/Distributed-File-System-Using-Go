package p2p

import "net"

// intergace representing an remove node
type Peer interface {
	net.Conn // we can use net.Conn directly
	Send([]byte) error
	CloseStream()
	// RemoteAddr() net.Addr // not needed if we use net.Conn
	// Close() error // not need if we use net.Conn
}

// transport is anything that handles
// the communication between the nodes in the network.
// this can be of the form ( TCP, UDP, websockets)
type Transport interface {
	ListenAndAccept() error
	Consume() <-chan RCP
	Close() error
	Dial(string) error
}
