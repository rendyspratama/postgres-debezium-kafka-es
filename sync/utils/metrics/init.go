package metrics

import (
	"context"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/trace"
)

func InitPrometheus(port int, path string) error {
	http.Handle(path, promhttp.Handler())
	go func() {
		addr := fmt.Sprintf(":%d", port)
		if err := http.ListenAndServe(addr, nil); err != nil {
			panic(fmt.Sprintf("Failed to start metrics server: %v", err))
		}
	}()
	return nil
}

func InitTracing(serviceName, collectorURL string) error {
	exporter, err := otlptracehttp.New(
		context.Background(),
		otlptracehttp.WithEndpoint(collectorURL),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithSampler(trace.AlwaysSample()),
	)
	otel.SetTracerProvider(tp)

	return nil
}
