package p2p

import (
	"net"
	"sync"
)

type TCPTransport struct{
	listenAddress	string
	listener net.Listener

	mu sync.RWMutex
	peer map[net.Addr]Peer
}