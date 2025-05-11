package main

import (
	"bytes"
	"fmt"
	"log"
	"time"

	"github.com/Shashi7427/Distributed-File-System-Using-Go/p2p"
)

func makeServer(listenAddr string, nodes ...string) *FileServer {
	tcpOpts := p2p.TCPTransportOpts{
		ListenAddress: listenAddr,
		ShakeHands:    p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
		OnPeer: func(peer p2p.Peer) error {
			fmt.Printf("peer %v connected\n", peer)
			return nil
		},
		// this Decoder is an interface and it should be implemented in p2p.GOBDecoder
	}
	transport := p2p.NewTCPTransport(tcpOpts)
	FileServerOpts := FileServerOpts{
		EncKey:            newEncryptionKey(),
		StorageRoot:       listenAddr,
		PathTransformFunc: CASPathTransformFunc,
		Transport:         transport,
		BootStrapNodes:    nodes,
	}
	s := NewFileServer(FileServerOpts)
	transport.OnPeer = s.OnPeer
	return s
}

func main() {

	// fmt.Println("hello world")
	// tr := p2p.NewTCPTransport(tcpOpts)

	// go func() {
	// 	for {
	// 		msg := <-tr.Consume()
	// 		fmt.Printf("+%v\n", msg)
	// 	}
	// }()

	// tr.ListenAndAccept()

	s1 := makeServer("localhost:3000")
	s2 := makeServer("localhost:4000", "localhost:3000")

	go func() {
		// time.Sleep(time.Second * 3)
		s1.Start()
		// s2.Stop()
	}()
	time.Sleep(time.Second * 2)
	go s2.Start()
	time.Sleep(time.Second * 2)
	// r, err := s2.Get("niyati_0")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// b, err := io.ReadAll(r)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Printf("data from the file found -> %s\n", string(b))

	for i := range 4 {
		data := bytes.NewReader([]byte("some jpg bytes / big data data"))
		key := fmt.Sprintf("niyati_%d", i)
		s2.StoreData(key, data)

		if err := s2.store.Delete(key); err != nil {
			log.Fatal(err)
		}
		_, err := s2.Get(key)
		if err != nil {
			log.Fatal(err)
		}
		// _, err := io.ReadAll(r)
		// if err != nil {
		// 	log.Fatal(err)
		// }
		// fmt.Printf("data from the file found -> %s\n", string(b))
	}
	// for i := 0; i < 10; i++ {
	// 	fmt.Printf("looping %d\n", i)
	// 	s2.StoreData(fmt.Sprintf("niyati_%d", i), data)
	// 	time.Sleep(5 * time.Millisecond)
	// }
	// data := bytes.NewReader([]byte("some jpg bytes / big data data"))
	// s2.StoreData(fmt.Sprintf("niyati_%d", 0), data)
	// select {} // to block here and keep the program running
}
