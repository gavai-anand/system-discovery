package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"system-discovery/internal/app/services/interfaces"
	"system-discovery/internal/dto"
	"time"
)

// SyncService periodically pulls missing operations from peers.
// This ensures eventual consistency (anti-entropy mechanism).
type SyncService struct {
	sc       interfaces.IServiceCall
	cs       interfaces.ICounterService
	ds       interfaces.IDiscoveryService
	lastSeen int64      // latest processed operation ID
	mu       sync.Mutex // protects lastSeen
}

func InitSyncService(ds interfaces.IDiscoveryService, sc interfaces.IServiceCall, cs interfaces.ICounterService) *SyncService {
	return &SyncService{
		ds: ds,
		sc: sc,
		cs: cs,
	}
}

// StartSync runs periodic sync with peers.
func (ss *SyncService) StartSync(ctx context.Context) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			peers := ss.ds.GetAllPeers(ctx) // fetch latest peers each cycle
			ss.sync(peers)
		}
	}
}

// sync pulls operations from peers since lastSeen and applies them locally.
func (ss *SyncService) sync(peers []string) {
	ss.mu.Lock()
	last := ss.lastSeen
	ss.mu.Unlock()

	maxTimestamp := last

	for _, peer := range peers {

		endpoint := fmt.Sprintf("%s/operations?since=%d", peer, last)

		resp, _, err := ss.sc.Get(context.Background(), endpoint, nil, 2*time.Second)
		if err != nil {
			// skip unreachable peers
			continue
		}

		var operationRes dto.OperationsResponse
		if err := json.Unmarshal(resp, &operationRes); err != nil {
			log.Printf("[ERROR] Failed to parse operations from peer=%s err=%v", peer, err)
			continue
		}

		for _, op := range operationRes.Data {
			ss.cs.Replicate(context.Background(), op)

			if op.ID > maxTimestamp {
				maxTimestamp = op.ID
			}
		}
	}

	// Update last seen timestamp after processing all peers
	ss.mu.Lock()
	ss.lastSeen = maxTimestamp
	ss.mu.Unlock()
}
