package interfaces

import (
	"context"
	"system-discovery/internal/dto"
)

type IDiscoveryService interface {
	GetAllPeers(ctx context.Context) []string
	SyncAndStorePeers(ctx context.Context, request dto.SyncPeersRequest) []string
}
