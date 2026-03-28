package interfaces

import (
	"context"
	"system-discovery/internal/dto"
)

type ICounterService interface {
	Increment(ctx context.Context) string
	Replicate(ctx context.Context, req dto.IncrementRequest) string
	GetValue() map[string]int
	GetOperations(sinceStr string) []dto.IncrementRequest
}
