package p2p

import "net"

// message represents any arbitary data
// that is being sent over the each trasport
// between two nodes in the network
type RCP struct {
	From    net.Addr
	Payload []byte
}

