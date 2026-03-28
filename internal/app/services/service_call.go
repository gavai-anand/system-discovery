package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// ServiceCall provides a reusable HTTP client to make external API requests.
// Supports GET requests with optional headers and timeouts.
type ServiceCall struct {
	client *http.Client // HTTP client used to make requests
}

// InitServiceCall initializes a new ServiceCall instance with a default timeout of 30 seconds.
func InitServiceCall() *ServiceCall {
	return &ServiceCall{
		client: &http.Client{
			Timeout: 30 * time.Second, // Default timeout for all requests
		},
	}
}

// Get performs an HTTP GET request to the given endpoint with optional headers and timeout.
// Returns response bytes, HTTP status code, and an error if any.
func (s *ServiceCall) Get(ctx context.Context, endpoint string, headers map[string]string, timeout ...time.Duration) ([]byte, int, error) {
	// Determine request timeout: use custom if provided, else default client timeout
	reqTimeout := s.client.Timeout
	if len(timeout) > 0 {
		reqTimeout = timeout[0]
	}

	// Create a new context with timeout to ensure request does not hang indefinitely
	ctx, cancel := context.WithTimeout(ctx, reqTimeout)
	defer cancel()

	// Create HTTP GET request with context
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		// Return HTTP 500 on request creation failure
		return nil, http.StatusInternalServerError, err
	}

	// Add any provided headers to the request
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// Execute the request and return the result
	return s.executeRequest(req)
}

// Post performs an HTTP POST request
func (s *ServiceCall) Post(ctx context.Context, endpoint string, body interface{}, headers map[string]string, timeout ...time.Duration) ([]byte, int, error) {
	// Set timeout
	reqTimeout := s.client.Timeout
	if len(timeout) > 0 {
		reqTimeout = timeout[0]
	}

	ctx, cancel := context.WithTimeout(ctx, reqTimeout)
	defer cancel()

	// Convert body → JSON
	var reqBody []byte
	var err error

	if body != nil {
		switch v := body.(type) {
		case []byte:
			reqBody = v
		default:
			reqBody, err = json.Marshal(body)
			if err != nil {
				return nil, http.StatusInternalServerError, err
			}
		}
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	// Default header
	req.Header.Set("Content-Type", "application/json")
	// Add custom headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	// Execute request
	return s.executeRequest(req)
}

// executeRequest executes the provided HTTP request and handles the response.
// Returns response body bytes, HTTP status code, and error if any.
func (s *ServiceCall) executeRequest(req *http.Request) ([]byte, int, error) {
	// Perform the HTTP request
	req.Header.Set("X-Referrer", os.Getenv("SELF"))
	resp, err := s.client.Do(req)
	if err != nil {
		// Check if timeout occurred
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, http.StatusGatewayTimeout, fmt.Errorf("request timed out")
		}
		// Return 500 if request fails (network error, timeout, etc.)
		return nil, http.StatusInternalServerError, err
	}
	defer resp.Body.Close() // Ensure response body is closed to free resources

	// Read the full response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		// Return the actual HTTP status code along with read error
		return nil, resp.StatusCode, err
	}

	// Check for non-2xx HTTP responses
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Treat non-success status codes as errors
		return body, resp.StatusCode, fmt.Errorf("external API error: %s", string(body))
	}
	// Return successful response
	return body, resp.StatusCode, nil
}
