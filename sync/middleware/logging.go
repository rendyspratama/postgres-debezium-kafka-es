package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create response writer wrapper to capture status code
		rw := &responseWriter{w, http.StatusOK}

		// Process request
		next.ServeHTTP(rw, r)

		// Calculate duration
		duration := time.Since(start)

		// Log request details
		logEntry := map[string]interface{}{
			"request_id": uuid.New().String(),
			"timestamp":  time.Now().Format("2006-01-02 15:04:05.999"),
			"method":     r.Method,
			"path":       r.URL.Path,
			"status":     rw.statusCode,
			"duration":   duration.String(),
			"ip":         r.RemoteAddr,
			"user_agent": r.UserAgent(),
		}

		prettyJSON, _ := json.MarshalIndent(logEntry, "", "  ")
		fmt.Printf("\n%s\n\n", string(prettyJSON))
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
