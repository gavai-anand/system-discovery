package handlers

import (
	"log"
	"net/http"
	"system-discovery/internal/app/services/interfaces"
	"system-discovery/internal/dto"
	"system-discovery/internal/utils"
)

// DiscoveryHandler manages peer discovery and synchronization between nodes.
type DiscoveryHandler struct {
	BaseHandler
	ds interfaces.IDiscoveryService
}

func InitDiscoveryHandler(ds interfaces.IDiscoveryService) *DiscoveryHandler {
	return &DiscoveryHandler{
		ds: ds,
	}
}

// GetAllPeers returns the list of known peers in the cluster.
func (dh *DiscoveryHandler) GetAllPeers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	result := dh.ds.GetAllPeers(ctx)
	dh.ResponseOK(w, result)
}

// SyncPeers updates the local peer list using data received from another node.
func (dh *DiscoveryHandler) SyncPeers(w http.ResponseWriter, r *http.Request) {
	var req dto.SyncPeersRequest

	// Decode request body
	if err := utils.DecodeJSON(r, &req); err != nil {
		log.Printf("[ERROR] Invalid sync peers request: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Basic validation
	if len(req.Peers) == 0 {
		http.Error(w, "peers list cannot be empty", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	dh.ds.SyncAndStorePeers(ctx, req)

	dh.ResponseOK(w, nil)
}
