package services

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"system-discovery/internal/dto"
	"system-discovery/internal/mocks"
)

func TestSyncService_Sync_Success(t *testing.T) {
	mockSC := new(mocks.IServiceCall)
	mockCS := new(mocks.ICounterService)
	mockDS := new(mocks.IDiscoveryService)

	ss := InitSyncService(mockDS, mockSC, mockCS)

	peers := []string{"peer1"}

	response := dto.OperationsResponse{
		Data: []dto.IncrementRequest{
			{ID: 1, Value: 1},
			{ID: 2, Value: 1},
		},
	}

	bytes, _ := json.Marshal(response)

	mockSC.On("Get",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(bytes, 200, nil)

	mockCS.On("Replicate", mock.Anything, mock.Anything).Return("ok")

	ss.sync(peers)

	mockCS.AssertNumberOfCalls(t, "Replicate", 2)
}

func TestSyncService_Sync_HTTPError(t *testing.T) {
	mockSC := new(mocks.IServiceCall)
	mockCS := new(mocks.ICounterService)
	mockDS := new(mocks.IDiscoveryService)

	ss := InitSyncService(mockDS, mockSC, mockCS)

	mockSC.On("Get",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(nil, 500, errors.New("network error"))

	// Should skip without panic
	ss.sync([]string{"peer1"})
}

func TestSyncService_Sync_InvalidJSON(t *testing.T) {
	mockSC := new(mocks.IServiceCall)
	mockCS := new(mocks.ICounterService)
	mockDS := new(mocks.IDiscoveryService)

	ss := InitSyncService(mockDS, mockSC, mockCS)

	mockSC.On("Get",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return([]byte("invalid-json"), 200, nil)

	ss.sync([]string{"peer1"})
}

func TestSyncService_Sync_EmptyOperations(t *testing.T) {
	mockSC := new(mocks.IServiceCall)
	mockCS := new(mocks.ICounterService)
	mockDS := new(mocks.IDiscoveryService)

	ss := InitSyncService(mockDS, mockSC, mockCS)

	response := dto.OperationsResponse{
		Data: []dto.IncrementRequest{},
	}

	bytes, _ := json.Marshal(response)

	mockSC.On("Get",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(bytes, 200, nil)

	ss.sync([]string{"peer1"})

	// No replicate calls expected
	mockCS.AssertNotCalled(t, "Replicate", mock.Anything, mock.Anything)
}

func TestSyncService_Sync_LastSeenUpdate(t *testing.T) {
	mockSC := new(mocks.IServiceCall)
	mockCS := new(mocks.ICounterService)
	mockDS := new(mocks.IDiscoveryService)

	ss := InitSyncService(mockDS, mockSC, mockCS)

	response := dto.OperationsResponse{
		Data: []dto.IncrementRequest{
			{ID: 5, Value: 1},
			{ID: 10, Value: 1},
		},
	}

	bytes, _ := json.Marshal(response)

	mockSC.On("Get",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(bytes, 200, nil)

	mockCS.On("Replicate", mock.Anything, mock.Anything).Return("ok")

	ss.sync([]string{"peer1"})

	// Validate lastSeen updated to max ID
	assert.Equal(t, int64(10), ss.lastSeen)
}

func TestSyncService_Sync_MultiplePeers(t *testing.T) {
	mockSC := new(mocks.IServiceCall)
	mockCS := new(mocks.ICounterService)
	mockDS := new(mocks.IDiscoveryService)

	ss := InitSyncService(mockDS, mockSC, mockCS)

	response := dto.OperationsResponse{
		Data: []dto.IncrementRequest{
			{ID: 1, Value: 1},
		},
	}

	bytes, _ := json.Marshal(response)

	mockSC.On("Get",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(bytes, 200, nil)

	mockCS.On("Replicate", mock.Anything, mock.Anything).Return("ok")

	ss.sync([]string{"peer1", "peer2"})

	// Called twice (once per peer)
	mockCS.AssertNumberOfCalls(t, "Replicate", 2)
}

func TestStartSync_ContextCancel(t *testing.T) {
	mockDS := new(mocks.IDiscoveryService)
	mockSC := new(mocks.IServiceCall)
	mockCS := new(mocks.ICounterService)

	ss := InitSyncService(mockDS, mockSC, mockCS)

	ctx, cancel := context.WithCancel(context.Background())

	// cancel immediately
	cancel()

	// should exit instantly
	ss.StartSync(ctx)
}
