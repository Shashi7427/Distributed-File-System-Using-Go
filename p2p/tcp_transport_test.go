package p2p

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTCPTransport(t *testing.T) {
	listenerAddr := ":4000"
	tr := NewTCPTransport(listenerAddr)
	assert.Equal(t, listenerAddr, tr.listenAddress)
	// tr.ListenAndAccept()
	assert.Nil(t, tr.ListenAndAccept())
}
