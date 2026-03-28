package interfaces

import (
	"context"
	"time"
)

type IServiceCall interface {
	Get(ctx context.Context, endpoint string, headers map[string]string, timeout ...time.Duration) ([]byte, int, error)
	Post(ctx context.Context, endpoint string, body interface{}, headers map[string]string, timeout ...time.Duration) ([]byte, int, error)
}
