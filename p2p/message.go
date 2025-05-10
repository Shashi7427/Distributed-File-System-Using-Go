package p2p

import "net"

const (
	IncomingMessage = 0x01
	IncomingStream  = 0x02
)

// message represents any arbitary data
// that is being sent over the each trasport
// between two nodes in the network
type RCP struct {
	From    net.Addr
	Payload []byte
	Stream  bool
}
