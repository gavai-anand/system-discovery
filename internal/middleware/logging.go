package middleware

import (
	"log"
	"net/http"
	"time"
)

// custom response writer to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Middleware function
func LoggingMiddleware(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		start := time.Now()

		// wrap response writer
		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// call next handler
		next.ServeHTTP(rw, r)

		duration := time.Since(start)

		log.Printf(
			"[REFERRER] %s | [REQUEST] %s %s | Status: %d | Duration: %v",
			r.Header.Get("X-Referrer"),
			r.Method,
			r.URL.Path,
			rw.statusCode,
			duration,
		)
	})
}
