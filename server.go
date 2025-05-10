package main

import (
	"bytes"
	"encoding/binary"
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
				log.Println("decoding error: ", err)
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
	case MessageGetFile:
		return s.handleMessageGetFile(from, msg.Payload.(MessageGetFile))
	}
	return nil
}

func (s *FileServer) handleMessageGetFile(from string, msg MessageGetFile) error {
	fmt.Println("need to get file from the disk and send it to the peer")
	if !s.store.Has(msg.Key) {
		return fmt.Errorf("file not found in the local store : " + msg.Key)
	}
	fileSize, r, err := s.store.Read(msg.Key)
	if err != nil {
		return err
	}
	// asserting the reader to io.ReadCloser
	if rc, ok := r.(io.ReadCloser); ok {
		fmt.Println("closing the readclose")
		defer rc.Close()
	}

	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("peer (%s) not found", from)
	}
	// sending incoming strea message,
	peer.Send([]byte{p2p.IncomingStream})
	// sending the file size
	binary.Write(peer, binary.LittleEndian, fileSize)
	n, err := io.Copy(peer, r)
	if err != nil {
		return err
	}
	fmt.Printf("written %d bytes to the network to %s\n", n, from)
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

func (s *FileServer) stream(msg *Message) error {
	peers := []io.Writer{}
	for _, peer := range s.peers {
		peers = append(peers, peer)
	}
	mw := io.MultiWriter(peers...)
	return gob.NewEncoder(mw).Encode(msg)
}

func (s *FileServer) broadcast(msg *Message) error {
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		log.Fatal(err)
		return err
	}
	// first informing that we are going to send the data
	for _, peer := range s.peers {
		// we should let the remote peer know that we are going to send the the stream
		peer.Send([]byte{p2p.IncomingMessage})
		if err := peer.Send(buf.Bytes()); err != nil {
			log.Println("error sending message to peer: ", err)
		}
	}
	return nil
}

type MessageGetFile struct {
	Key string
}

func (s *FileServer) Get(key string) (io.Reader, error) {
	if s.store.Has(key) {
		_, f, err := s.store.Read(key)
		return f, err
	}
	// if the file is not present in the local store then we need to get it from the networkd
	println("file not found in the local store searching in the network")
	msg := Message{
		Payload: MessageGetFile{
			Key: key,
		},
	}
	if err := s.broadcast(&msg); err != nil {
		return nil, err
	}
	time.Sleep(time.Millisecond * 500)

	for _, peer := range s.peers {
		fmt.Println("received the message from the peer")
		var fileSize int64
		binary.Read(peer, binary.LittleEndian, &fileSize)
		n, err := s.store.Write(key, io.LimitReader(peer, fileSize))
		if err != nil {
			return nil, err
		}
		fmt.Printf("received %d bytes from the peer from %s \n", n, peer.RemoteAddr())
		peer.Close()

		// filebuffer := new(bytes.Buffer)
		// n, err := io.CopyN(filebuffer, peer, 10)
		// if err != nil {
		// 	return nil, err
		// }
		// fmt.Printf("received %d bytes from the peer\n", n)
		// fmt.Println(filebuffer.String())
	}
	_, f, err := s.store.Read(key)
	return f, err
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
	msg := &Message{
		Payload: MessageStoreFile{
			Key:  key,
			Size: size,
		},
	}
	if err := s.broadcast(msg); err != nil {
		log.Println("error broadcasting message: ", err)
	}

	time.Sleep(time.Millisecond * 5)
	// sending th actual data now :
	fmt.Println("sending the large data now")
	for _, peer := range s.peers {
		// if err := peer.Send(payload); err != nil {
		// 	log.Println("error sending message to peer: ", err)
		// }
		// first letting them know that this is the incoming stream
		peer.Send([]byte{p2p.IncomingStream})
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
	gob.Register(MessageGetFile{})
}
