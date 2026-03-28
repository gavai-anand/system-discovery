package services

import (
	"context"
	"fmt"
	"github.com/cenkalti/backoff/v4"
	"log"
	"net/http"
	"system-discovery/internal/app/constants"
	"system-discovery/internal/app/services/interfaces"
	"system-discovery/internal/dto"
	"time"
)

// DiscoveryService manages peer discovery, synchronization, and health checks.
type DiscoveryService struct {
	sc         interfaces.IServiceCall
	store      *PeerStore
	regService interfaces.IRegistrationService
	self       string
}

func InitDiscoveryService(store *PeerStore, regService interfaces.IRegistrationService, self string, sc interfaces.IServiceCall) *DiscoveryService {
	return &DiscoveryService{
		store:      store,
		regService: regService,
		self:       self,
		sc:         sc,
	}
}

// GetAllPeers returns the current list of known peers.
func (ds *DiscoveryService) GetAllPeers(ctx context.Context) []string {
	ds.store.mu.RLock()
	defer ds.store.mu.RUnlock()

	result := make([]string, 0, len(ds.store.peers))
	for p := range ds.store.peers {
		result = append(result, p)
	}
	return result
}

// SyncAndStorePeers merges incoming peer list with local state.
func (ds *DiscoveryService) SyncAndStorePeers(ctx context.Context, request dto.SyncPeersRequest) []string {
	ds.store.mu.Lock()
	defer ds.store.mu.Unlock()

	for _, p := range request.Peers {
		if p == "" || p == ds.self {
			continue
		}

		// Add only new peers
		if !ds.store.peers[p] {
			ds.store.peers[p] = true
			log.Printf("[DISCOVERY] New peer added: %s", p)
		}
	}

	result := make([]string, 0, len(ds.store.peers))
	for p := range ds.store.peers {
		result = append(result, p)
	}

	return result
}

// StartHeartbeat periodically checks health of peers.
func (ds *DiscoveryService) StartHeartbeat(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ds.checkPeers(ctx)
		}
	}
}

// checkPeers runs health checks on all peers (excluding self).
func (ds *DiscoveryService) checkPeers(ctx context.Context) {
	ds.store.mu.RLock()
	peers := make([]string, 0, len(ds.store.peers))

	for p := range ds.store.peers {
		if p == ds.self {
			continue
		}
		peers = append(peers, p)
	}
	ds.store.mu.RUnlock()

	// Limit concurrency to avoid overload
	sem := make(chan struct{}, 20)

	for _, peer := range peers {
		sem <- struct{}{}

		go func(peer string) {
			defer func() { <-sem }()
			ds.checkPeer(ctx, peer)
		}(peer)
	}
}

// checkPeer verifies if a peer is alive; removes it if not.
func (ds *DiscoveryService) checkPeer(ctx context.Context, peer string) {
	if !ds.isAlive(ctx, peer) {
		log.Printf("[HEARTBEAT] Peer down: %s", peer)
		ds.regService.DeRegister(ctx, peer)
	}
}

// isAlive checks peer health with retries using exponential backoff.
func (ds *DiscoveryService) isAlive(ctx context.Context, peer string) bool {
	endpoint := fmt.Sprintf(constants.HEARTBEAT_API, peer)

	var alive bool

	operation := func() error {
		_, statusCode, err := ds.sc.Get(ctx, endpoint, nil, 2*time.Second)
		if err == nil && statusCode == http.StatusOK {
			alive = true
			return nil
		}
		return fmt.Errorf("peer not alive")
	}

	expBackoff := backoff.NewExponentialBackOff()
	expBackoff.InitialInterval = 2 * time.Second
	expBackoff.MaxInterval = 2 * time.Second
	expBackoff.Multiplier = 1

	b := backoff.WithMaxRetries(expBackoff, 3)

	if err := backoff.Retry(operation, backoff.WithContext(b, ctx)); err != nil {
		log.Printf("[ERROR] Heartbeat failed for peer=%s err=%v", peer, err)
		return false
	}

	return alive
}
