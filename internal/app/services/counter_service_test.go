package services_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"system-discovery/internal/app/services"
	"system-discovery/internal/dto"
)

func TestCounterService_Increment(t *testing.T) {
	cs := services.InitCounterService(nil, nil, "node1")

	res := cs.Increment(context.Background())

	assert.Contains(t, res, "New incremented value")
	assert.Equal(t, 1, cs.GetValue()["value"])
}

func TestCounterService_Apply_Idempotent(t *testing.T) {
	cs := services.InitCounterService(nil, nil, "node1")

	req := dto.IncrementRequest{
		ID:    123,
		Value: 1,
	}

	cs.Replicate(context.Background(), req)
	cs.Replicate(context.Background(), req)

	assert.Equal(t, 1, cs.GetValue()["value"])
}

func TestCounterService_GetOperations(t *testing.T) {
	cs := services.InitCounterService(nil, nil, "node1")

	cs.Increment(context.Background())

	ops := cs.GetOperations("0")

	assert.NotEmpty(t, ops)
}
