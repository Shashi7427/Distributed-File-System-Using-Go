package p2p

import (
	"fmt"
	"net"
	"sync"
)

// remote node over the tcp instablished connection
type TCPPeer struct {
	// underlying connection of with the peer
	conn net.Conn

	//if we make connection and get conn -> outbound = true
	// if we accept and retrieve a conn -> outbound = false\
	outbound bool
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		conn:     conn,
		outbound: outbound,
	}
}

type TCPTransport struct {
	listenAddress string
	listener      net.Listener

	mu   sync.RWMutex
	peer map[net.Addr]Peer
}

func NewTCPTransport(listenAddr string) *TCPTransport {
	return &TCPTransport{
		listenAddress: listenAddr,
	}
}

func (t *TCPTransport) ListenAndAccept() error {
	fmt.Println("going to listen")
	var err error
	t.listener, err = net.Listen("tcp", t.listenAddress)
	if err != nil {
		fmt.Println("error to listen")
		return err
	}

	t.startAcceptLoop()
	return nil
}

func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			fmt.Printf("TCP accept error : %s\n", err)
		}
		t.handleConnection(conn)
	}
}

func (t *TCPTransport) handleConnection(conn net.Conn) {
	peer := NewTCPPeer(conn, true)
	fmt.Printf("new incoming connection %v\n", peer)
}
