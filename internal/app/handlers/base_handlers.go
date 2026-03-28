package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// BaseHandler provides common functionality for all HTTP handlers.
// It holds a context and helper methods to send JSON responses.
type BaseHandler struct {
	ctx context.Context
}

// SetContext sets the context for the handler.
// Returns the handler itself to allow method chaining.
func (h *BaseHandler) SetContext(c context.Context) *BaseHandler {
	h.ctx = c
	return h
}

// GetContext retrieves the context associated with the handler.
// Useful when passing context downstream to services or DB calls.
func (h *BaseHandler) GetContext() context.Context {
	return h.ctx
}

// ResponseOK sends a successful JSON response (HTTP 200).
// `data` is any payload you want to include in the response.
func (h *BaseHandler) ResponseOK(w http.ResponseWriter, data interface{}) {
	// Set response content type as JSON
	w.Header().Set("Content-Type", "application/json")
	// Set HTTP status code to 200 OK
	w.WriteHeader(http.StatusOK)

	// Construct a standardized success response
	resp := map[string]interface{}{
		"success": true,
		"message": "Success",
		"data":    data,
	}

	// Encode response to JSON and write it to the client
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		// Log error if encoding fails
		fmt.Println(err)
		return
	}
}

// ResponseError sends an error JSON response with a custom HTTP status code and message.
func (h *BaseHandler) ResponseError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	resp := map[string]interface{}{
		"success": false,
		"message": message,
	}

	// Encode and send the error response
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		fmt.Println(err)
		return
	}
}

// ResponseNotFound sends a 404 Not Found error response with a custom message.
// Internally calls ResponseError with http.StatusNotFound.
func (h *BaseHandler) ResponseNotFound(w http.ResponseWriter, message string) {
	h.ResponseError(w, http.StatusNotFound, message)
}
