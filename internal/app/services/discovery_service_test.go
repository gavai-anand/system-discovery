package services

import (
	"context"
	"github.com/stretchr/testify/assert"
	"system-discovery/internal/dto"
	"testing"
)

func TestDiscoveryService_GetAllPeers(t *testing.T) {
	store := InitPeerStore()
	store.peers["node1"] = true

	ds := InitDiscoveryService(store, nil, "self", nil)

	peers := ds.GetAllPeers(context.Background())

	assert.Contains(t, peers, "node1")
}

func TestDiscoveryService_SyncAndStorePeers(t *testing.T) {
	store := InitPeerStore()
	ds := InitDiscoveryService(store, nil, "self", nil)

	req := dto.SyncPeersRequest{
		Peers: []string{"node1", "node2"},
	}

	result := ds.SyncAndStorePeers(context.Background(), req)

	assert.Len(t, result, 2)
}
