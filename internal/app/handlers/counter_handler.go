package handlers

import (
	"log"
	"net/http"
	"system-discovery/internal/app/services/interfaces"
	"system-discovery/internal/dto"
	"system-discovery/internal/utils"
)

// CounterHandler handles HTTP requests for counter operations.
// It acts as a thin layer between HTTP and business logic.
type CounterHandler struct {
	BaseHandler
	cs interfaces.ICounterService
}

func InitCounterHandler(cs interfaces.ICounterService) *CounterHandler {
	return &CounterHandler{
		cs: cs,
	}
}

// Increment handles client requests to increment the counter.
func (h *CounterHandler) Increment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	log.Println("[HTTP] Increment request received")

	response := h.cs.Increment(ctx)
	h.ResponseOK(w, response)
}

// Replicate handles replication requests from peer nodes.
func (h *CounterHandler) Replicate(w http.ResponseWriter, r *http.Request) {
	var req dto.IncrementRequest

	if err := utils.DecodeJSON(r, &req); err != nil {
		h.ResponseError(w, http.StatusBadRequest, "Invalid replicate request")
		return
	}

	ctx := r.Context()
	response := h.cs.Replicate(ctx, req)
	h.ResponseOK(w, response)
}

// GetCount returns the current counter value.
func (h *CounterHandler) GetCount(w http.ResponseWriter, r *http.Request) {
	response := h.cs.GetValue()
	h.ResponseOK(w, response)
}

// GetOperations returns operation history.
// Optional query param: ?since=<timestamp> for incremental fetch.
func (h *CounterHandler) GetOperations(w http.ResponseWriter, r *http.Request) {
	sinceStr := r.URL.Query().Get("since")

	operations := h.cs.GetOperations(sinceStr)
	h.ResponseOK(w, operations)
}
