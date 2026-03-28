package services

import "sync"

// PeerStore is an in-memory thread-safe store for managing peer nodes.
// It uses RWMutex to allow concurrent reads and safe writes.
type PeerStore struct {
	mu    sync.RWMutex
	peers map[string]bool // set of peer addresses
}

// InitPeerStore initializes an empty peer store.
func InitPeerStore() *PeerStore {
	return &PeerStore{
		peers: make(map[string]bool),
	}
}
