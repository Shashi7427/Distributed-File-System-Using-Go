package p2p

// intergace representing an remove node
type Peer interface{}

// transport is anything that handles
// the communication between the nodes in the network. 
// this can be of the form ( TCP, UDP, websockets)
type Transport interface{
	ListenAndAccept() error
}