package main

import (
	"fmt"

	"github.com/Shashi7427/Distributed-File-System-Using-Go/p2p"
)

func main() {
	fmt.Println("hello world")
	tr := p2p.NewTCPTransport(":3000")
	tr.ListenAndAccept()
}
