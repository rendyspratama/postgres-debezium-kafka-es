package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// MetricType represents the type of metric being tracked
type MetricType string

const (
	MetricLatency   MetricType = "latency"
	MetricErrors    MetricType = "errors"
	MetricRequests  MetricType = "requests"
	MetricResponses MetricType = "responses"
)

// MetricValue represents a metric value with timestamp
type MetricValue struct {
	Value     float64
	Timestamp time.Time
}

// MiddlewareMetrics tracks metrics for middleware
type MiddlewareMetrics struct {
	mu      sync.RWMutex
	metrics map[string]map[MetricType][]MetricValue
}

// NewMiddlewareMetrics creates a new middleware metrics tracker
func NewMiddlewareMetrics() *MiddlewareMetrics {
	return &MiddlewareMetrics{
		metrics: make(map[string]map[MetricType][]MetricValue),
	}
}

// Track creates a middleware that tracks metrics
func (mm *MiddlewareMetrics) Track(name string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create response writer wrapper to capture status code
		rw := newResponseWriter(w)

		// Track the request
		mm.recordMetric(name, MetricRequests, 1)

		// Call the next handler
		next.ServeHTTP(rw, r)

		// Record metrics
		duration := time.Since(start)
		mm.recordMetric(name, MetricLatency, float64(duration.Milliseconds()))
		mm.recordMetric(name, MetricResponses, float64(rw.status))

		if rw.status >= 400 {
			mm.recordMetric(name, MetricErrors, 1)
		}
	})
}

// recordMetric records a metric value
func (mm *MiddlewareMetrics) recordMetric(middleware string, metricType MetricType, value float64) {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	if _, exists := mm.metrics[middleware]; !exists {
		mm.metrics[middleware] = make(map[MetricType][]MetricValue)
	}

	mm.metrics[middleware][metricType] = append(
		mm.metrics[middleware][metricType],
		MetricValue{Value: value, Timestamp: time.Now()},
	)

	// Keep only last 1000 values
	if len(mm.metrics[middleware][metricType]) > 1000 {
		mm.metrics[middleware][metricType] = mm.metrics[middleware][metricType][1:]
	}
}

// GetMetrics returns metrics for a middleware
func (mm *MiddlewareMetrics) GetMetrics(middleware string) map[MetricType][]MetricValue {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	if metrics, exists := mm.metrics[middleware]; exists {
		return metrics
	}
	return nil
}

// GetAverageLatency returns the average latency for a middleware
func (mm *MiddlewareMetrics) GetAverageLatency(middleware string) float64 {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	if metrics, exists := mm.metrics[middleware]; exists {
		if latencies, exists := metrics[MetricLatency]; exists {
			var sum float64
			for _, v := range latencies {
				sum += v.Value
			}
			return sum / float64(len(latencies))
		}
	}
	return 0
}

// GetErrorRate returns the error rate for a middleware
func (mm *MiddlewareMetrics) GetErrorRate(middleware string) float64 {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	if metrics, exists := mm.metrics[middleware]; exists {
		if requests, hasReqs := metrics[MetricRequests]; hasReqs {
			if errors, hasErrs := metrics[MetricErrors]; hasErrs {
				totalReqs := 0.0
				totalErrs := 0.0
				for _, v := range requests {
					totalReqs += v.Value
				}
				for _, v := range errors {
					totalErrs += v.Value
				}
				if totalReqs > 0 {
					return (totalErrs / totalReqs) * 100
				}
			}
		}
	}
	return 0
}

// String returns a string representation of middleware metrics
func (mm *MiddlewareMetrics) String() string {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	result := "Middleware Metrics:\n"
	for middleware := range mm.metrics {
		result += fmt.Sprintf("\n%s:\n", middleware)
		result += fmt.Sprintf("  Average Latency: %.2fms\n", mm.GetAverageLatency(middleware))
		result += fmt.Sprintf("  Error Rate: %.2f%%\n", mm.GetErrorRate(middleware))
	}
	return result
}
