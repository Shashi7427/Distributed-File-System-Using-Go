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

// close implements the peer interface
func (p *TCPPeer) Close() error {
	return p.conn.Close()
}

type TCPTransportOpts struct {
	ListenAddress string
	ShakeHands    HandshakeFunc
	Decoder       Decoder
}

type TCPTransport struct {
	TCPTransportOpts
	listener net.Listener
	rcpch    chan RCP

	mu    sync.RWMutex
	peers map[net.Addr]Peer
}

func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: opts,
		rcpch:            make(chan RCP),
	}
}

// consume implements the trasport interface, which will return read-only channle
// for reading the incoming messsages received from  another peer in the network
func (t *TCPTransport) Consume() <-chan RCP {
	return t.rcpch
}

func (t *TCPTransport) ListenAndAccept() error {
	fmt.Println("going to listen")
	var err error
	t.listener, err = net.Listen("tcp", t.ListenAddress)
	if err != nil {
		fmt.Println("error to listen")
		return err
	}

	t.startAcceptLoop()
	return nil
}

type Temp struct{}

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

	if err := t.ShakeHands(peer); err != nil {
		conn.Close()
		fmt.Printf("TCP handshake erorr %s\n", err)
		return
	}

	// Read loop
	rcp := RCP{}
	// buf := make([]byte, 2000)
	for {
		// n, err := conn.Read(buf)

		// if err != nil {
		// 	fmt.Printf("TCP error : %s\n", err)
		// 	continue
		// }
		if err := t.Decoder.Decode(conn, &rcp); err != nil {
			fmt.Printf("TCP error : %s/n", err)
			continue
		}
		rcp.From = conn.RemoteAddr()
		t.rcpch <- rcp
		fmt.Printf("RCP message : %+v\n", &rcp)
	}
}
