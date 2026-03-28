package services

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPeerStore_Init(t *testing.T) {
	store := InitPeerStore()

	assert.NotNil(t, store)
	assert.Empty(t, store.peers)
}
