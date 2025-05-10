package p2p

import (
	"errors"
	"fmt"
	"net"
	"reflect"
	"sync"
)

// remote node over the tcp instablished connection
type TCPPeer struct {
	// underlying connection of with the peer
	net.Conn

	//if we make connection and get conn -> outbound = true
	// if we accept and retrieve a conn -> outbound = false\
	outbound bool

	Wg *sync.WaitGroup
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		Conn:     conn,
		outbound: outbound,
		Wg:       &sync.WaitGroup{},
	}
}

func (p *TCPPeer) Send(msg []byte) error {
	_, err := p.Conn.Write(msg)
	if err != nil {
		fmt.Printf("error to send message %s\n", err)
		return err
	}
	return nil
}

// // close implements the peer interface
// func (p *TCPPeer) Close() error {
// 	return p.Conn.Close()
// }

//	func (p *TCPPeer) RemoteAddr() net.Addr {
//		return p.Conn.RemoteAddr()
//	}
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

func (t *TCPTransport) Close() error {
	fmt.Println("closing the listener")
	return t.listener.Close()
}

func (t *TCPTransport) ListenAndAccept() error {
	var err error
	t.listener, err = net.Listen("tcp", t.ListenAddress)
	if err != nil {
		fmt.Println("error to listen")
		return err
	}

	go t.startAcceptLoop()
	fmt.Println("going to listen on address", t.ListenAddress)
	return nil
}

type Temp struct{}

func (t *TCPTransport) Dial(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Println("error to dial")
		return err
	}
	t.handleConnection(conn, true)
	return nil
}

func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if errors.Is(err, net.ErrClosed) {
			return
		}
		if err != nil {
			fmt.Printf("TCP accept error : %s\n", err)
		}

		go t.handleConnection(conn, false)
	}
}

func (t *TCPTransport) handleConnection(conn net.Conn, outbound bool) {
	var err error
	defer func() {
		fmt.Printf("Dropping peer connection %s\n", err)
		conn.Close()
	}()
	peer := NewTCPPeer(conn, outbound)
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
	// buf := make([]byte, 2000)
	for {
		// n, err := conn.Read(buf)

		// if err != nil {
		// 	fmt.Printf("TCP error : %s\n", err)
		// 	continue
		// }
		rcp := RCP{}
		// panic((err))
		if err := t.Decoder.Decode(conn, &rcp); err != nil {
			fmt.Print(reflect.TypeOf(err))
			fmt.Printf("TCP error : %s/n", err)
			return
		}
		rcp.From = conn.RemoteAddr()

		if rcp.Stream {
			peer.Wg.Add(1)
			fmt.Printf("[%s] incoming stream, waiting for stream to finish\n", conn.RemoteAddr())
			peer.Wg.Wait()
			fmt.Printf("[%s] stream closed, resuming read loop\n", conn.RemoteAddr())
			continue
		}
		// peer.Wg.Add(1)
		// fmt.Printf("waiting till stream is done\n")
		t.rcpch <- rcp
		// peer.Wg.Wait()
		// fmt.Printf("stream done continueing the normal stream flow")
		// fmt.Printf("RCP message : %+v\n", &rcp)
	}
}
