package interfaces

import (
	"context"
)

type ISyncService interface {
	StartSync(ctx context.Context)
}
