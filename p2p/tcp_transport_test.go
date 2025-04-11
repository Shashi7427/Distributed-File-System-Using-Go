package p2p

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTCPTransport(t *testing.T) {
	tcpOpts := TCPTransportOpts{
		ListenAddress: ":3000",
		ShakeHands:    NOPHandshakeFunc,
		Decoder:       DefaultDecoder{},
	}

	tr := NewTCPTransport(tcpOpts)
	assert.Equal(t, tr.ListenAddress, tcpOpts.ListenAddress)
	assert.Nil(t, tr.ListenAndAccept())
}
