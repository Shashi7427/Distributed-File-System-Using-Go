package p2p

import (
	"fmt"
	"net"
	"reflect"
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
	OnPeer        func(Peer) error
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
	var err error
	defer func() {
		fmt.Printf("Dropping peer connection %s\n", err)
		conn.Close()
	}()

	peer := NewTCPPeer(conn, true)
	fmt.Printf("new incoming connection %v\n", peer)

	if err := t.ShakeHands(peer); err != nil {
		fmt.Printf("error in handshake %s\n", err)
		return
	}

	if err := t.OnPeer(peer); err != nil {
		fmt.Printf("error in peer %s\n", err)
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

		// panic((err))
		if err := t.Decoder.Decode(conn, &rcp); err != nil {
			fmt.Print(reflect.TypeOf(err))
			fmt.Printf("TCP error : %s/n", err)
			return
		}
		rcp.From = conn.RemoteAddr()
		t.rcpch <- rcp
		fmt.Printf("RCP message : %+v\n", &rcp)
	}
}
