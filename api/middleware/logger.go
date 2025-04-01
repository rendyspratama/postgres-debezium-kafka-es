package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/rendyspratama/digital-discovery/api/config"
)

type LoggerMiddleware struct {
	config config.MiddlewareConfig
}

func NewLoggerMiddleware(cfg config.MiddlewareConfig) *LoggerMiddleware {
	return &LoggerMiddleware{config: cfg}
}

type LogEntry struct {
	RequestID    string      `json:"request_id"`
	Timestamp    string      `json:"timestamp"`
	Method       string      `json:"method"`
	Path         string      `json:"path"`
	Status       int         `json:"status"`
	Duration     string      `json:"duration"`
	IP           string      `json:"ip"`
	UserAgent    string      `json:"user_agent"`
	QueryParams  string      `json:"query_params,omitempty"`
	RequestBody  interface{} `json:"request_body,omitempty"`
	ResponseBody interface{} `json:"response_body,omitempty"`
}

func (l *LoggerMiddleware) Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestID := uuid.New().String()

		// Create new response writer to capture status and body
		rw := NewResponseWriter(w)

		// Store request ID in context
		ctx := r.Context()
		ctx = context.WithValue(ctx, "requestID", requestID)
		r = r.WithContext(ctx)

		// Process request
		next.ServeHTTP(rw, r)

		// Create log entry
		entry := LogEntry{
			RequestID: requestID,
			Timestamp: time.Now().Format("2006-01-02 15:04:05.000"),
			Method:    r.Method,
			Path:      r.URL.Path,
			Status:    rw.status,
			Duration:  fmt.Sprintf("%.3fms", float64(time.Since(start).Microseconds())/1000),
			IP:        r.RemoteAddr,
			UserAgent: r.UserAgent(),
		}

		if r.URL.RawQuery != "" {
			entry.QueryParams = r.URL.RawQuery
		}

		// Pretty print the log entry
		logJSON, _ := json.MarshalIndent(entry, "", "  ")

		// Color codes
		green := "\033[32m"
		yellow := "\033[33m"
		red := "\033[31m"
		blue := "\033[34m"
		reset := "\033[0m"

		// Choose color based on status code
		var color string
		switch {
		case entry.Status >= 500:
			color = red
		case entry.Status >= 400:
			color = yellow
		case entry.Status >= 300:
			color = blue
		default:
			color = green
		}

		fmt.Printf("\n%s%s%s\n", color, string(logJSON), reset)
	})
}
