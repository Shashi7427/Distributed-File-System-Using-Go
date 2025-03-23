package p2p

// message represents any arbitary data 
// that is being sent over the each trasport
// between two nodes in the network 
type Message struct{
	Payload [] byte 
}