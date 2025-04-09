package middleware

import (
	"log"
	"net/http"
	"time"
)

// LogRequest logs HTTP requests
func LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a custom response writer to capture status code
		lrw := &loggingResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK, // Default to 200 OK
		}

		// Call the next handler
		next.ServeHTTP(lrw, r)

		// Log the request details
		duration := time.Since(start)
		log.Printf(
			"%s %s %s - %d %s - %v",
			r.RemoteAddr,
			r.Method,
			r.URL.Path,
			lrw.statusCode,
			http.StatusText(lrw.statusCode),
			duration,
		)
	})
}

// loggingResponseWriter is a custom ResponseWriter that captures the status code
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code
func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
