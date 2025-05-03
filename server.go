package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/Shashi7427/Distributed-File-System-Using-Go/p2p"
)

type FileServerOpts struct {
	StorageRoot       string
	PathTransformFunc PathTransformFunc
	Transport         p2p.Transport
	BootStrapNodes    []string
}

type FileServer struct {
	FileServerOpts
	store *Store

	peerLock sync.Mutex
	peers    map[string]p2p.Peer

	quitch chan struct{}
}

func NewFileServer(opts FileServerOpts) *FileServer {
	StoreOpts := StoreOpts{
		RootPath:          opts.StorageRoot,
		PathTransformFunc: opts.PathTransformFunc,
	}
	return &FileServer{
		FileServerOpts: opts,
		store:          NewStore(StoreOpts),
		quitch:         make(chan struct{}),
		peers:          make(map[string]p2p.Peer),
	}
}

func (fs *FileServer) Stop() {
	fmt.Printf("quiting ")
	close(fs.quitch)
}

func (fs *FileServer) loop() {
	defer func() {
		log.Println("file server stopped due to user quit action")
		fs.Transport.Close()
	}()

	for {
		fmt.Printf("looping loop\n")
		select {
		case msg := <-fs.Transport.Consume():
			var p Payload
			fmt.Println("msg received")
			// fmt.Println(msg)
			if err := gob.NewDecoder(bytes.NewReader(msg.Payload)).Decode(&p); err != nil {
				log.Fatal(err)
				log.Println("error decoding message: ", err)
				continue
			}
			fmt.Printf("%v\n", string(p.Data))
			fmt.Println(p)
		case <-fs.quitch:
			return
		}
	}
}

func (s *FileServer) OnPeer(p p2p.Peer) error {
	s.peerLock.Lock()
	defer s.peerLock.Unlock()

	s.peers[p.RemoteAddr().String()] = p

	log.Printf("connected with remote %s", p.RemoteAddr())

	return nil
}

func (fs *FileServer) bootStrapNetwork() error {
	for _, addr := range fs.BootStrapNodes {
		if len(addr) == 0 {
			continue
		}
		log.Println("dial")
		// why are we calling this can't we just write the dial function
		go func(addr string) {
			// fmt.Printf("[%s] attemping to connect with remote %s\n", fs.Transport.ListenAddress(), addr)
			if err := fs.Transport.Dial(addr); err != nil {
				log.Println("dial error: ", err)
			}
		}(addr)
	}
	return nil
}

type Message struct {
	Payload any
	From    string
}

type Payload struct {
	Key  string
	Data []byte
}

func (s *FileServer) broadcast(msg *Message) error {
	peers := []io.Writer{}
	for _, peer := range s.peers {
		peers = append(peers, peer)
	}
	mw := io.MultiWriter(peers...)
	return gob.NewEncoder(mw).Encode(msg)
}

func (s *FileServer) StoreData(key string, r io.Reader) error {
	// store the file to disk

	buf := new(bytes.Buffer)
	// _, err := io.Copy(buf, r)
	tee := io.TeeReader(r, buf)

	if err := s.store.Write(key, tee); err != nil {
		return err
	}

	p := Payload{
		Key:  key,
		Data: buf.Bytes(),
	}
	fmt.Println("bytes sending", buf.Bytes())
	// create a payload

	// broadcast this file to the all peers in the network

	return s.broadcast(&Message{
		Payload: p,
		From:    "todo",
	})
}

func (fs *FileServer) Start() error {
	// Start the server
	fmt.Printf("looping and accept \n")

	if err := fs.Transport.ListenAndAccept(); err != nil {
		return err
	}
	fmt.Printf("looping finally \n")
	fs.bootStrapNetwork()
	// Start the loop
	fs.loop()
	return nil
}
