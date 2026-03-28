package services

import (
	"context"
	"fmt"
	"github.com/cenkalti/backoff/v4"
	"log"
	"net/http"
	"strconv"
	"sync"
	"system-discovery/internal/app/constants"
	"system-discovery/internal/app/services/interfaces"
	"system-discovery/internal/dto"
	"time"
)

// CounterService is a distributed counter with:
// - idempotent operations (no duplicate increments)
// - eventual consistency via peer replication
type CounterService struct {
	mu      sync.Mutex
	value   int
	seenOps map[int64]bool         // prevents duplicate processing
	ops     []dto.IncrementRequest // operation history (for sync)

	ds   interfaces.IDiscoveryService
	sc   interfaces.IServiceCall
	self string
}

func InitCounterService(ds interfaces.IDiscoveryService, sc interfaces.IServiceCall, self string) *CounterService {
	return &CounterService{
		value:   0,
		seenOps: make(map[int64]bool),
		ops:     []dto.IncrementRequest{},
		ds:      ds,
		sc:      sc,
		self:    self,
	}
}

// Increment creates a new operation, applies it locally,
// and asynchronously propagates it to peers.
func (cs *CounterService) Increment(ctx context.Context) string {
	req := dto.IncrementRequest{
		ID:     time.Now().UnixNano(),
		Source: cs.self,
		Value:  1,
	}

	log.Printf("[INCREMENT] ID=%d Source=%s", req.ID, req.Source)

	response := cs.apply(req)

	if cs.ds != nil && cs.sc != nil {
		go cs.propagate(context.Background(), req)
	}

	return response
}

// GetValue returns the current counter value.
func (cs *CounterService) GetValue() map[string]int {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	return map[string]int{"value": cs.value}
}

// Replicate applies an operation received from another node.
func (cs *CounterService) Replicate(ctx context.Context, req dto.IncrementRequest) string {
	return cs.apply(req)
}

// apply safely applies an operation once (idempotent).
func (cs *CounterService) apply(req dto.IncrementRequest) string {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if cs.seenOps[req.ID] {
		log.Printf("[DUPLICATE] ID=%d ignored", req.ID)
		return fmt.Sprintf("already incremented: %d", req.Value)
	}

	cs.value += req.Value
	cs.seenOps[req.ID] = true
	cs.ops = append(cs.ops, req)

	log.Printf("[APPLY] ID=%d NewValue=%d", req.ID, cs.value)

	return fmt.Sprintf("New incremented value is: %d", cs.value)
}

// propagate sends the operation to all peers (fire-and-forget).
func (cs *CounterService) propagate(ctx context.Context, req dto.IncrementRequest) {
	peers := cs.ds.GetAllPeers(ctx)

	log.Printf("[PROPAGATE] ID=%d Peers=%d", req.ID, len(peers))

	for _, peer := range peers {
		if peer == cs.self {
			continue
		}

		go func(p string) {
			if err := cs.sendWithRetry(ctx, p, req); err != nil {
				log.Printf("[ERROR] send failed ID=%d peer=%s err=%v", req.ID, p, err)
			}
		}(peer)
	}
}

// sendWithRetry retries failed requests using exponential backoff.
func (cs *CounterService) sendWithRetry(ctx context.Context, peer string, req dto.IncrementRequest) error {
	operation := func() error {
		return cs.send(ctx, peer, req)
	}

	expBackoff := backoff.NewExponentialBackOff()
	expBackoff.InitialInterval = 500 * time.Millisecond
	expBackoff.MaxInterval = 2 * time.Second
	expBackoff.Multiplier = 2
	expBackoff.MaxElapsedTime = 0

	b := backoff.WithMaxRetries(expBackoff, 2)

	return backoff.Retry(operation, backoff.WithContext(b, ctx))
}

// send performs HTTP replication to a peer node.
func (cs *CounterService) send(ctx context.Context, peer string, req dto.IncrementRequest) error {
	endpoint := fmt.Sprintf(constants.REPLICATE_API, peer)

	_, status, err := cs.sc.Post(ctx, endpoint, req, nil, 2*time.Second)
	if err != nil {
		return err
	}

	if status != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", status)
	}

	return nil
}

// GetOperations returns all operations or those after a given timestamp.
func (cs *CounterService) GetOperations(sinceStr string) []dto.IncrementRequest {
	if len(sinceStr) > 0 {
		since, _ := strconv.ParseInt(sinceStr, 10, 64)

		var filtered []dto.IncrementRequest
		for _, op := range cs.ops {
			if op.ID > since {
				filtered = append(filtered, op)
			}
		}
		return filtered
	}
	return cs.ops
}
