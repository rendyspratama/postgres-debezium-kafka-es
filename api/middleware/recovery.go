package middleware

import (
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"time"
)

// RecoveryConfig configures the recovery middleware
type RecoveryConfig struct {
	// DisableStackTrace disables stack trace logging
	DisableStackTrace bool

	// DisableResponseWrite disables writing error response to client
	DisableResponseWrite bool

	// ErrorHandler is called when a panic occurs
	ErrorHandler func(interface{}, http.ResponseWriter, *http.Request)

	// LogHandler is called to log the error
	LogHandler func(interface{}, []byte)
}

// DefaultRecoveryConfig returns the default recovery configuration
func DefaultRecoveryConfig() *RecoveryConfig {
	return &RecoveryConfig{
		DisableStackTrace:    false,
		DisableResponseWrite: false,
		ErrorHandler:         defaultErrorHandler,
		LogHandler:           defaultLogHandler,
	}
}

// Recovery returns a middleware that recovers from panics
func Recovery(config *RecoveryConfig) func(http.Handler) http.Handler {
	if config == nil {
		config = DefaultRecoveryConfig()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Get stack trace
					var stack []byte
					if !config.DisableStackTrace {
						stack = debug.Stack()
					}

					// Log the error
					if config.LogHandler != nil {
						config.LogHandler(err, stack)
					}

					// Handle the error
					if config.ErrorHandler != nil {
						config.ErrorHandler(err, w, r)
					}
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// defaultErrorHandler is the default error handler
func defaultErrorHandler(err interface{}, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, `{"error": "Internal Server Error", "message": "%v"}`, err)
}

// defaultLogHandler is the default log handler
func defaultLogHandler(err interface{}, stack []byte) {
	log.Printf("[PANIC RECOVER] %v\n%s", err, stack)
}

// CircuitBreaker represents a simple circuit breaker
type CircuitBreaker struct {
	failures  int
	threshold int
	timeout   time.Duration
	lastError time.Time
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(threshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		threshold: threshold,
		timeout:   timeout,
	}
}

// WithCircuitBreaker adds circuit breaker functionality to a handler
func WithCircuitBreaker(cb *CircuitBreaker, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if circuit is open
		if cb.failures >= cb.threshold {
			if time.Since(cb.lastError) > cb.timeout {
				// Reset circuit breaker
				cb.failures = 0
			} else {
				http.Error(w, "Service temporarily unavailable", http.StatusServiceUnavailable)
				return
			}
		}

		// Create response writer wrapper to capture status
		rw := NewResponseWriter(w)

		next.ServeHTTP(rw, r)

		// Update circuit breaker state
		if rw.status >= 500 {
			cb.failures++
			cb.lastError = time.Now()
		} else {
			// Reset on successful response
			cb.failures = 0
		}
	})
}

// Retry represents retry configuration
type RetryConfig struct {
	MaxAttempts int
	Delay       time.Duration
	ShouldRetry func(r *http.Request, status int) bool
}

// WithRetry adds retry functionality to a handler
func WithRetry(config RetryConfig, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var lastStatus int
		var lastBody []byte

		for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
			// Create response writer wrapper to capture status and body
			rw := NewResponseWriter(w)

			next.ServeHTTP(rw, r)
			lastStatus = rw.status
			lastBody = rw.body

			// Check if should retry
			if !config.ShouldRetry(r, lastStatus) {
				// Write the successful response
				w.WriteHeader(lastStatus)
				w.Write(lastBody)
				return
			}

			// Don't retry on last attempt
			if attempt == config.MaxAttempts {
				break
			}

			// Wait before retrying
			time.Sleep(config.Delay)
		}

		// If all retries failed, return last response
		w.WriteHeader(lastStatus)
		w.Write(lastBody)
	})
}
