package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

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

type Message struct {
	Payload any
	// From    string
}

type MessageStoreFile struct {
	Key  string
	Size int64
}

func (fs *FileServer) loop() {
	defer func() {
		log.Println("file server stopped due to user quit action")
		fs.Transport.Close()
	}()

	for {
		fmt.Printf("looping loop\n")
		select {
		case rcp := <-fs.Transport.Consume():
			var msg Message
			fmt.Printf("msg received by server %s\n", fs.store.RootPath)
			// fmt.Println(msg)
			if err := gob.NewDecoder(bytes.NewReader(rcp.Payload)).Decode(&msg); err != nil {
				log.Fatal(err)
				log.Println("error decoding message: ", err)
				continue
			}
			if err := fs.handleMessage(rcp.From.String(), &msg); err != nil {
				log.Println("error handling message: ", err)
			}

			// fmt.Printf("msg received %s\n", msg.Payload.([]byte))
			// fmt.Printf("msg received %s\n", msg.Payload)

			// peer, ok := fs.peers[rcp.From.String()]
			// if !ok {
			// 	log.Panic("peer not found")
			// }
			// b := make([]byte, 1024)
			// if _, err := peer.Read(b); err != nil {
			// 	log.Panic(err)
			// }

			// println("peer found")
			// println(peer)
			// // this is not a good practice but we are doing it for now
			// peer.(*p2p.TCPPeer).Wg.Wait()

			// fmt.Printf("%s\n", b)

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

func (s *FileServer) handleMessage(from string, msg *Message) error {
	switch msg.Payload.(type) {
	case MessageStoreFile:
		return s.handleMessageStoreFile(from, msg.Payload.(MessageStoreFile))
	}
	return nil
}

func (s *FileServer) handleMessageStoreFile(from string, msg MessageStoreFile) error {

	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("peer (%s) not found", from)
	}
	fmt.Printf("peer found %s msg of size %d\n", peer.RemoteAddr(), msg.Size)
	n, err := s.store.Write(msg.Key, io.LimitReader(peer, msg.Size))
	if err != nil {
		return err
	}
	// this is not a good practice but we are doing it for now
	fmt.Printf("received private data and stored of size : %v\n", n)
	peer.(*p2p.TCPPeer).Wg.Done()
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

// type Payload struct {
// 	Key  string
// 	Data []byte
// }

func (s *FileServer) broadcast(msg *Message) error {
	peers := []io.Writer{}
	for _, peer := range s.peers {
		peers = append(peers, peer)
	}
	mw := io.MultiWriter(peers...)
	return gob.NewEncoder(mw).Encode(msg)
}

func (s *FileServer) StoreData(key string, r io.Reader) error {

	// storing the data to disk
	buf := new(bytes.Buffer)
	tee := io.TeeReader(r, buf)

	size, err := s.store.Write(key, tee)
	if err != nil {
		return err
	}

	// sending the data to all the peers in the network
	msgbuf := new(bytes.Buffer)
	msg := &Message{
		Payload: MessageStoreFile{
			Key:  key,
			Size: size,
		},
	}

	if err := gob.NewEncoder(msgbuf).Encode(msg); err != nil {
		log.Fatal(err)
		return err
	}
	// first informing that we are going to send the data
	for _, peer := range s.peers {
		if err := peer.Send(msgbuf.Bytes()); err != nil {
			log.Println("error sending message to peer: ", err)
		}
	}

	time.Sleep(time.Second * 3)
	// sending th large data now :
	fmt.Println("sending the large data now")
	for _, peer := range s.peers {
		// if err := peer.Send(payload); err != nil {
		// 	log.Println("error sending message to peer: ", err)
		// }
		n, err := io.Copy(peer, buf)
		if err != nil {
			return err
		}
		fmt.Println("received and written bytes to disk", n)
	}
	return nil

	// // store the file to disk

	// buf := new(bytes.Buffer)
	// // _, err := io.Copy(buf, r)
	// tee := io.TeeReader(r, buf)

	// if err := s.store.Write(key, tee); err != nil {
	// 	return err
	// }

	// p := Payload{
	// 	Key:  key,
	// 	Data: buf.Bytes(),
	// }
	// fmt.Println("bytes sending", buf.Bytes())
	// // create a payload

	// // broadcast this file to the all peers in the network

	// return s.broadcast(&Message{
	// 	Payload: p,
	// 	From:    "todo",
	// })
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

// registering the gob type for interface main.MessageStorefile
func init() {
	gob.Register(MessageStoreFile{})
}
