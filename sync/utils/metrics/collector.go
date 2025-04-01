package metrics

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type OperationMetrics struct {
	StartTime   time.Time
	EndTime     time.Time
	Duration    time.Duration
	Operation   string
	Entity      string
	EntityID    string
	Status      string
	IndexName   string
	PayloadSize int
	ErrorCount  int
}

type MetricsCollector struct {
	mu sync.RWMutex

	// Operation metrics
	operationDuration *prometheus.HistogramVec
	operationTotal    *prometheus.CounterVec
	operationErrors   *prometheus.CounterVec
	payloadSize       *prometheus.HistogramVec

	// Bulk operation metrics
	bulkOperations *prometheus.HistogramVec
}

func NewMetricsCollector() *MetricsCollector {
	mc := &MetricsCollector{}
	mc.initMetrics()
	return mc
}

func (mc *MetricsCollector) initMetrics() {
	mc.operationDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "sync",
			Name:      "operation_duration_seconds",
			Help:      "Duration of sync operations",
		},
		[]string{"operation", "entity", "status"},
	)
	prometheus.MustRegister(mc.operationDuration)

	mc.operationTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "sync",
			Name:      "operations_total",
			Help:      "Total number of sync operations",
		},
		[]string{"operation", "entity", "status"},
	)
	prometheus.MustRegister(mc.operationTotal)

	mc.operationErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "sync",
			Name:      "operation_errors_total",
			Help:      "Total number of sync operation errors",
		},
		[]string{"operation", "entity"},
	)
	prometheus.MustRegister(mc.operationErrors)

	mc.payloadSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "sync",
			Name:      "payload_size_bytes",
			Help:      "Size of sync operation payloads",
			Buckets:   prometheus.ExponentialBuckets(100, 2, 10),
		},
		[]string{"operation", "entity"},
	)
	prometheus.MustRegister(mc.payloadSize)

	mc.bulkOperations = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "sync",
			Name:      "bulk_operations_total",
			Help:      "Number of operations in bulk requests",
		},
		[]string{"entity", "status"},
	)
	prometheus.MustRegister(mc.bulkOperations)
}

func (mc *MetricsCollector) RecordOperation(metrics *OperationMetrics) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	mc.operationDuration.WithLabelValues(
		metrics.Operation,
		metrics.Entity,
		metrics.Status,
	).Observe(metrics.Duration.Seconds())

	mc.operationTotal.WithLabelValues(
		metrics.Operation,
		metrics.Entity,
		metrics.Status,
	).Inc()

	mc.payloadSize.WithLabelValues(
		metrics.Operation,
		metrics.Entity,
	).Observe(float64(metrics.PayloadSize))
}

func (mc *MetricsCollector) RecordError(operation, entity string, count int) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	mc.operationErrors.WithLabelValues(operation, entity).Add(float64(count))
}

func (mc *MetricsCollector) RecordBulkOperation(entity string, size int, hasError bool) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	status := "success"
	if hasError {
		status = "error"
	}
	mc.bulkOperations.WithLabelValues(entity, status).Observe(float64(size))
}

func (mc *MetricsCollector) Cleanup() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	// Unregister all metrics
	prometheus.Unregister(mc.operationDuration)
	prometheus.Unregister(mc.operationTotal)
	prometheus.Unregister(mc.operationErrors)
	prometheus.Unregister(mc.payloadSize)
	prometheus.Unregister(mc.bulkOperations)
}
