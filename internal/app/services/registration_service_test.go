package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"system-discovery/internal/dto"
)

func TestRegistrationService_Register(t *testing.T) {
	store := InitPeerStore()
	rs := InitRegistrationService(store)

	req := dto.RegisterRequest{
		NodeID: "node1",
	}

	rs.Register(context.Background(), req)

	// ✅ Verify peer added
	store.mu.RLock()
	defer store.mu.RUnlock()

	_, exists := store.peers["node1"]
	assert.True(t, exists)
}

func TestRegistrationService_Register_Overwrite(t *testing.T) {
	store := InitPeerStore()
	rs := InitRegistrationService(store)

	req := dto.RegisterRequest{NodeID: "node1"}

	rs.Register(context.Background(), req)
	rs.Register(context.Background(), req) // duplicate

	store.mu.RLock()
	defer store.mu.RUnlock()

	// ✅ Should still exist (idempotent behavior)
	_, exists := store.peers["node1"]
	assert.True(t, exists)
}

func TestRegistrationService_DeRegister(t *testing.T) {
	store := InitPeerStore()
	rs := InitRegistrationService(store)

	// Preload peer
	store.peers["node1"] = true

	rs.DeRegister(context.Background(), "node1")

	store.mu.RLock()
	defer store.mu.RUnlock()

	_, exists := store.peers["node1"]
	assert.False(t, exists)
}

func TestRegistrationService_DeRegister_NotExists(t *testing.T) {
	store := InitPeerStore()
	rs := InitRegistrationService(store)

	rs.DeRegister(context.Background(), "node1")

	// Should not panic and still be empty
	store.mu.RLock()
	defer store.mu.RUnlock()

	assert.Empty(t, store.peers)
}
