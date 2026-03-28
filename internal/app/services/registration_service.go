package services

import (
	"context"
	"log"
	"system-discovery/internal/dto"
)

// RegistrationService manages adding/removing peers in the cluster.
type RegistrationService struct {
	store *PeerStore
}

func InitRegistrationService(store *PeerStore) *RegistrationService {
	return &RegistrationService{
		store: store,
	}
}

// Register adds a new node to the peer store.
func (rs *RegistrationService) Register(ctx context.Context, request dto.RegisterRequest) {
	rs.store.mu.Lock()
	defer rs.store.mu.Unlock()

	rs.store.peers[request.NodeID] = true
}

// DeRegister removes a peer from the store if it exists.
func (rs *RegistrationService) DeRegister(ctx context.Context, peer string) {
	rs.store.mu.Lock()
	defer rs.store.mu.Unlock()

	if _, exists := rs.store.peers[peer]; !exists {
		log.Printf("[DEREGISTER] Peer already removed: %s", peer)
		return
	}

	delete(rs.store.peers, peer)
	log.Printf("[DEREGISTER] Peer removed: %s", peer)
}
