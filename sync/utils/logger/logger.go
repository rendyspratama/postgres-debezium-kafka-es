package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
)

const (
	// ANSI color codes
	red     = "\033[31m"
	green   = "\033[32m"
	yellow  = "\033[33m"
	blue    = "\033[34m"
	magenta = "\033[35m"
	cyan    = "\033[36m"
	reset   = "\033[0m"
)

type Logger interface {
	Info(ctx context.Context, msg string, fields map[string]interface{})
	Error(ctx context.Context, msg string, fields map[string]interface{})
	WithError(ctx context.Context, err error, msg string, fields map[string]interface{})
}

type logger struct {
	format string
}

func NewLogger(format string) Logger {
	return &logger{
		format: format,
	}
}

func (l *logger) Info(ctx context.Context, msg string, fields map[string]interface{}) {
	l.log(ctx, "INFO", green, msg, fields)
}

func (l *logger) Error(ctx context.Context, msg string, fields map[string]interface{}) {
	l.log(ctx, "ERROR", red, msg, fields)
}

func (l *logger) WithError(ctx context.Context, err error, msg string, fields map[string]interface{}) {
	if fields == nil {
		fields = make(map[string]interface{})
	}
	fields["error"] = err.Error()
	l.log(ctx, "ERROR", red, msg, fields)
}

func (l *logger) log(ctx context.Context, level, colorCode string, msg string, fields map[string]interface{}) {
	if fields == nil {
		fields = make(map[string]interface{})
	}

	// Add standard fields
	fields["timestamp"] = time.Now().Format(time.RFC3339)
	fields["level"] = level
	fields["message"] = msg

	// Get environment from context if available
	if env, ok := ctx.Value("environment").(string); ok {
		fields["environment"] = env
	}

	// Format the log entry
	if l.format == "json" {
		// JSON format
		jsonData, _ := json.Marshal(fields)
		fmt.Fprintf(os.Stdout, "%s%s%s\n", colorCode, string(jsonData), reset)
	} else {
		// Pretty format with colors
		fmt.Printf("%s[%s] %s%s\n", colorCode, level, msg, reset)
		if len(fields) > 0 {
			for k, v := range fields {
				if k != "level" && k != "message" {
					fmt.Printf("%s  %s: %v%s\n", yellow, k, v, reset)
				}
			}
		}
	}
}

type PrettyLogger struct {
	serviceName string
}

func NewPrettyLogger(serviceName string) *PrettyLogger {
	// Print service banner
	fmt.Printf("\n=== %s ===\n\n", serviceName)
	return &PrettyLogger{
		serviceName: serviceName,
	}
}

func (l *PrettyLogger) Info(ctx context.Context, message string, fields map[string]interface{}) {
	logEntry := l.formatLogEntry(ctx, "INFO", message, fields)
	fmt.Printf("▶ %s\n", message)
	if len(fields) > 0 {
		prettyJSON, _ := json.MarshalIndent(logEntry, "", "  ")
		fmt.Printf("\n%s\n\n", string(prettyJSON))
	}
}

func (l *PrettyLogger) Error(ctx context.Context, message string, fields map[string]interface{}) {
	logEntry := l.formatLogEntry(ctx, "ERROR", message, fields)
	fmt.Printf("❌ %s\n", message)
	prettyJSON, _ := json.MarshalIndent(logEntry, "", "  ")
	fmt.Printf("\n%s\n\n", string(prettyJSON))
}

func (l *PrettyLogger) WithError(ctx context.Context, err error, message string, fields map[string]interface{}) {
	if fields == nil {
		fields = make(map[string]interface{})
	}
	fields["error"] = err.Error()
	fields["error_type"] = fmt.Sprintf("%T", err)
	l.Error(ctx, message, fields)
}

func (l *PrettyLogger) formatLogEntry(ctx context.Context, level, message string, fields map[string]interface{}) map[string]interface{} {
	entry := make(map[string]interface{})

	// Add standard fields
	entry["timestamp"] = time.Now().Format("2006-01-02 15:04:05.999")
	entry["level"] = level
	entry["service"] = l.serviceName

	// Add message if present
	if message != "" {
		entry["message"] = message
	}

	// Add request_id if present in context
	if reqID := l.getRequestID(ctx); reqID != "" {
		entry["request_id"] = reqID
	}

	// Add all additional fields
	for k, v := range fields {
		// Don't overwrite standard fields
		if _, exists := entry[k]; !exists {
			entry[k] = v
		}
	}

	return entry
}

func (l *PrettyLogger) getRequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	// You can implement your own request ID retrieval logic here
	// For example, if you're using a request ID middleware:
	if reqID, ok := ctx.Value("request_id").(string); ok {
		return reqID
	}
	return ""
}

// Example usage of request ID middleware
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := uuid.New().String()
		ctx := context.WithValue(r.Context(), "request_id", requestID)
		w.Header().Set("X-Request-ID", requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
