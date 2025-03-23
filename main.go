package main

import (
	"fmt"

	"github.com/Shashi7427/Distributed-File-System-Using-Go/p2p"
)

func main() {
	tcpOpts := p2p.TCPTransportOpts{
		ListenAddress: ":3000",
		ShakeHands: p2p.NOPHandshakeFunc,
		Decoder: p2p.DefaultDecoder{}, 
		// this Decoder is an interface and it should be implemented in p2p.GOBDecoder
		
	}


	fmt.Println("hello world")
	tr := p2p.NewTCPTransport(tcpOpts)
	tr.ListenAndAccept()
}
