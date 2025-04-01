package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rendyspratama/digital-discovery/sync/config"
	"github.com/rendyspratama/digital-discovery/sync/consumers"
	"github.com/rendyspratama/digital-discovery/sync/middleware"
	"github.com/rendyspratama/digital-discovery/sync/models"
	"github.com/rendyspratama/digital-discovery/sync/repositories/elasticsearch"
	"github.com/rendyspratama/digital-discovery/sync/services"
	"github.com/rendyspratama/digital-discovery/sync/utils/logger"
	"github.com/rendyspratama/digital-discovery/sync/utils/metrics"
)

type App struct {
	cfg          *config.Config
	logger       logger.Logger
	esClient     elasticsearch.Repository
	syncService  *services.SyncService
	retryService *services.RetryService
	consumer     *consumers.KafkaConsumer
	httpServer   *http.Server
	metrics      *metrics.MetricsCollector
}

// Add health check handler
func (a *App) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"status":    "UP",
		"timestamp": time.Now().Format(time.RFC3339),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// Add readiness check handler
func (a *App) handleReadinessCheck(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	status := map[string]interface{}{
		"status":        "UP",
		"timestamp":     time.Now().Format(time.RFC3339),
		"elasticsearch": "UP",
		"kafka":         "UP",
	}

	// Check Elasticsearch using repository method
	if err := a.esClient.CheckHealth(ctx); err != nil {
		status["elasticsearch"] = "DOWN"
		status["status"] = "DOWN"
		a.logger.WithError(ctx, err, "Elasticsearch health check failed", map[string]interface{}{
			"component": "elasticsearch",
		})
	}

	// Check Kafka consumer
	if err := a.consumer.HealthCheck(); err != nil {
		status["kafka"] = "DOWN"
		status["status"] = "DOWN"
		a.logger.WithError(ctx, err, "Kafka health check failed", map[string]interface{}{
			"component": "kafka",
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if status["status"] == "DOWN" {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	json.NewEncoder(w).Encode(status)
}

func main() {
	logger := logger.NewPrettyLogger("Digital Discovery Sync")

	// Print startup banner
	logger.Info(context.Background(), "Server starting", map[string]interface{}{
		"port":        8082,
		"time":        time.Now().Format("2006-01-02 15:04:05"),
		"environment": os.Getenv("APP_ENV"),
	})

	app, err := initializeApp(logger)
	if err != nil {
		logger.WithError(context.Background(), err, "Failed to initialize application", nil)
		os.Exit(1)
	}
	defer app.cleanup()

	// Initialize context with cancellation for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start application
	go func() {
		if err := app.Start(ctx); err != nil {
			logger.Error(ctx, "Application failed to start", map[string]interface{}{
				"error": err.Error(),
			})
			os.Exit(1)
		}
	}()

	// Wait for shutdown signal
	sig := <-sigChan
	logger.Info(ctx, "Shutdown initiated", map[string]interface{}{
		"signal": sig.String(),
	})

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Perform graceful shutdown
	if err := app.Stop(shutdownCtx); err != nil {
		logger.Error(ctx, "Shutdown error", map[string]interface{}{
			"error": err.Error(),
		})
	}

	logger.Info(ctx, "Shutdown complete", map[string]interface{}{
		"message": "Application shutdown completed successfully",
	})
}

func initializeApp(appLogger logger.Logger) (*App, error) {
	ctx := context.Background()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize metrics collector
	// metricsCollector := metrics.NewMetricsCollector()

	// Initialize Elasticsearch repository
	esConfig := &elasticsearch.Config{
		Addresses:      cfg.ES.Hosts,
		Username:       cfg.ES.Username,
		Password:       cfg.ES.Password,
		MaxRetries:     cfg.ES.MaxRetries,
		RetryBackoff:   cfg.ES.RetryBackoff,
		EnableRetry:    cfg.ES.EnableRetry,
		MaxConns:       cfg.ES.MaxConns,
		RequestTimeout: cfg.ES.RequestTimeout,
		GzipEnabled:    cfg.ES.GzipEnabled,
	}

	// Use NewRepository instead of NewClient
	esClient, err := elasticsearch.NewRepository(esConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Elasticsearch repository: %w", err)
	}

	// Initialize services with repository
	syncService := services.NewSyncService(esClient, cfg, appLogger)
	retryService := services.NewRetryService(syncService, cfg, appLogger)

	// Initialize Kafka consumer
	consumer, err := consumers.NewKafkaConsumer(cfg, syncService, appLogger)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka consumer: %w", err)
	}

	app := &App{
		cfg:          cfg,
		logger:       appLogger,
		esClient:     esClient,
		syncService:  syncService,
		retryService: retryService,
		consumer:     consumer,
		// metrics:      metricsCollector,
	}

	// Initialize HTTP server for metrics and health checks
	if err := app.initHTTPServer(); err != nil {
		return nil, fmt.Errorf("failed to initialize HTTP server: %w", err)
	}

	app.logger.Info(ctx, "Application initialized successfully", map[string]interface{}{
		"service": cfg.App.ServiceName,
		"env":     cfg.App.Environment,
	})

	return app, nil
}

func (a *App) Start(ctx context.Context) error {
	// Initialize services
	if err := a.initializeServices(ctx); err != nil {
		return fmt.Errorf("failed to initialize services: %w", err)
	}

	// Start API server for both modes
	go func() {
		if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.WithError(ctx, err, "API server failed", map[string]interface{}{
				"port": a.httpServer.Addr,
			})
		}
	}()

	// Start sync based on mode
	switch a.cfg.Sync.Mode {
	case "custom":
		if !a.cfg.Sync.Custom.Enabled {
			return fmt.Errorf("custom sync is not enabled")
		}
		return a.startCustomSync(ctx)
	case "kafka-connect":
		if !a.cfg.Sync.KafkaConnect.Enabled {
			return fmt.Errorf("kafka connect is not enabled")
		}
		return a.startKafkaConnectSync(ctx)
	default:
		return fmt.Errorf("invalid sync mode: %s", a.cfg.Sync.Mode)
	}
}

func (a *App) startCustomSync(ctx context.Context) error {
	a.logger.Info(ctx, "Starting custom sync mode", map[string]interface{}{
		"mode": "custom",
	})
	return a.consumer.Start(ctx)
}

func (a *App) startKafkaConnectSync(ctx context.Context) error {
	a.logger.Info(ctx, "Starting Kafka Connect sync mode", map[string]interface{}{
		"mode": "kafka-connect",
	})
	return a.monitorKafkaConnect(ctx)
}

func (a *App) monitorKafkaConnect(ctx context.Context) error {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			status, err := a.checkConnectorStatus()
			if err != nil {
				a.logger.WithError(ctx, err, "Failed to check connector status", map[string]interface{}{
					"mode": "kafka-connect",
				})
				continue
			}
			a.logger.Info(ctx, "Connector status", map[string]interface{}{
				"status": status,
			})
		}
	}
}

func (a *App) checkConnectorStatus() (string, error) {
	resp, err := http.Get(fmt.Sprintf("%s/connectors/%s/status",
		a.cfg.Sync.KafkaConnect.SinkConnector.URL,
		a.cfg.Sync.KafkaConnect.SinkConnector.Name))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var status struct {
		Connector struct {
			State string `json:"state"`
		} `json:"connector"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return "", err
	}
	return status.Connector.State, nil
}

func (a *App) setupElasticsearch(ctx context.Context) error {
	// Create index template using repository
	if err := a.esClient.CreateTemplate(ctx); err != nil {
		return fmt.Errorf("failed to create index template: %w", err)
	}

	// Create lifecycle policy using repository
	if err := a.esClient.CreateLifecyclePolicy(ctx, "digital-discovery-policy"); err != nil {
		return fmt.Errorf("failed to create lifecycle policy: %w", err)
	}

	// Verify setup using repository
	if err := a.esClient.VerifySetup(ctx); err != nil {
		return fmt.Errorf("failed to verify elasticsearch setup: %w", err)
	}

	a.logger.Info(ctx, "Elasticsearch setup completed", map[string]interface{}{
		"templates": []string{"categories-template"},
		"policies":  []string{"digital-discovery-policy"},
		"status":    "success",
	})

	return nil
}

func (a *App) initMetrics() error {
	// Initialize Prometheus metrics
	if err := metrics.InitPrometheus(a.cfg.Monitoring.MetricsPort, a.cfg.Monitoring.PrometheusPath); err != nil {
		return fmt.Errorf("failed to initialize Prometheus metrics: %w", err)
	}

	// Initialize OpenTelemetry if enabled
	if a.cfg.Monitoring.TracingEnabled {
		if err := metrics.InitTracing(a.cfg.App.ServiceName, a.cfg.Monitoring.OtelCollector); err != nil {
			return fmt.Errorf("failed to initialize tracing: %w", err)
		}
	}

	return nil
}

func (a *App) initHTTPServer() error {
	mux := http.NewServeMux()

	// Wrap all handlers with logging middleware
	handler := middleware.LoggingMiddleware(mux)

	// Add health check endpoint
	mux.HandleFunc("/health", a.handleHealthCheck)

	// Add metrics endpoint
	mux.Handle("/metrics", promhttp.Handler())

	// Add readiness check endpoint
	mux.HandleFunc("/ready", a.handleReadinessCheck)

	// Add API endpoints
	mux.HandleFunc("/api/v1/categories", a.handleCategories)
	mux.HandleFunc("/api/v1/category", a.handleCategory)

	a.httpServer = &http.Server{
		Addr:         ":8082", // API server port
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return nil
}

func (a *App) handleCategories(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	switch r.Method {
	case http.MethodGet:
		categories, err := a.syncService.ListCategories(ctx)
		if err != nil {
			a.respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		a.respondWithJSON(w, http.StatusOK, categories)
	case http.MethodPost:
		var category models.Category
		if err := json.NewDecoder(r.Body).Decode(&category); err != nil {
			a.respondWithError(w, http.StatusBadRequest, "Invalid request format")
			return
		}

		// Set default values if not provided
		if category.Description == "" {
			category.Description = "No description provided"
		}
		if category.Status == 0 {
			category.Status = 1 // Default status
		}

		// Validate category
		if err := category.Validate(); err != nil {
			a.respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		// Set timestamps
		now := time.Now()
		category.CreatedAt = now
		category.UpdatedAt = now

		// Create category
		if err := a.syncService.CreateCategory(ctx, category); err != nil {
			a.respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		a.respondWithJSON(w, http.StatusCreated, category)
	default:
		a.respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (a *App) handleCategory(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		a.respondWithError(w, http.StatusBadRequest, "Category ID is required")
		return
	}

	switch r.Method {
	case http.MethodGet:
		category, err := a.syncService.GetCategory(r.Context(), id)
		if err != nil {
			a.respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		a.respondWithJSON(w, http.StatusOK, category)
	case http.MethodPut:
		var category models.Category
		if err := json.NewDecoder(r.Body).Decode(&category); err != nil {
			a.respondWithError(w, http.StatusBadRequest, "Invalid request body")
			return
		}
		if err := a.syncService.UpdateCategory(r.Context(), category); err != nil {
			a.respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		a.respondWithJSON(w, http.StatusOK, map[string]string{"message": "Category updated successfully"})
	case http.MethodDelete:
		if err := a.syncService.DeleteCategory(r.Context(), id); err != nil {
			a.respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		a.respondWithJSON(w, http.StatusOK, map[string]string{"message": "Category deleted successfully"})
	default:
		a.respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// Helper methods for consistent responses
func (a *App) respondWithError(w http.ResponseWriter, code int, message string) {
	a.respondWithJSON(w, code, map[string]interface{}{
		"status":     "error",
		"message":    message,
		"request_id": uuid.New().String(),
	})
}

func (a *App) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func (a *App) cleanup() {
	startTime := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cleanupInfo := map[string]interface{}{
		"event":     "cleanup_started",
		"timestamp": time.Now().Format(time.RFC3339),
		"service":   a.cfg.App.ServiceName,
		"components": []string{
			"http_server",
			"kafka_consumer",
			"elasticsearch_client",
			"metrics_collector",
		},
	}

	jsonBytes, _ := json.MarshalIndent(cleanupInfo, "", "  ")
	a.logger.Info(ctx, "Starting cleanup", map[string]interface{}{
		"cleanup_info": string(jsonBytes),
	})

	var wg sync.WaitGroup
	errChan := make(chan error, 4) // Buffer for all cleanup operations

	// Cleanup HTTP server
	if a.httpServer != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := a.httpServer.Shutdown(ctx); err != nil {
				errChan <- fmt.Errorf("http server shutdown: %w", err)
			}
		}()
	}

	// Cleanup Kafka consumer
	if a.consumer != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := a.consumer.Close(); err != nil {
				errChan <- fmt.Errorf("kafka consumer cleanup: %w", err)
			}
		}()
	}

	// Cleanup Elasticsearch client
	if a.esClient != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := a.esClient.Close(); err != nil {
				errChan <- fmt.Errorf("elasticsearch cleanup: %w", err)
			}
		}()
	}

	// Cleanup metrics
	if a.metrics != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			a.metrics.Cleanup()
		}()
	}

	// Wait for all cleanup operations
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	// Wait for cleanup or timeout
	select {
	case <-done:
		// Check for any errors
		close(errChan)
		for err := range errChan {
			a.logger.WithError(ctx, err, "Cleanup error", nil)
		}
	case <-ctx.Done():
		a.logger.WithError(ctx, ctx.Err(), "Cleanup timeout", nil)
	}

	cleanupCompleteInfo := map[string]interface{}{
		"event":       "cleanup_completed",
		"timestamp":   time.Now().Format(time.RFC3339),
		"service":     a.cfg.App.ServiceName,
		"duration_ms": time.Since(startTime).Milliseconds(),
	}

	jsonBytes, _ = json.MarshalIndent(cleanupCompleteInfo, "", "  ")
	a.logger.Info(ctx, "Cleanup completed", map[string]interface{}{
		"cleanup_info": string(jsonBytes),
	})
}

func (a *App) initializeServices(ctx context.Context) error {
	// Setup Elasticsearch
	if err := a.setupElasticsearch(ctx); err != nil {
		return fmt.Errorf("failed to setup elasticsearch: %w", err)
	}

	// Initialize metrics
	if err := a.initMetrics(); err != nil {
		return fmt.Errorf("failed to initialize metrics: %w", err)
	}

	return nil
}

func (a *App) Stop(ctx context.Context) error {
	var err error
	// Shutdown HTTP server
	if a.httpServer != nil {
		if err = a.httpServer.Shutdown(ctx); err != nil {
			a.logger.WithError(ctx, err, "Failed to shutdown HTTP server", nil)
		}
	}

	// Close Kafka consumer
	if a.consumer != nil {
		if err = a.consumer.Close(); err != nil {
			a.logger.WithError(ctx, err, "Failed to close Kafka consumer", nil)
		}
	}

	// Close Elasticsearch client
	if a.esClient != nil {
		if err = a.esClient.Close(); err != nil {
			a.logger.WithError(ctx, err, "Failed to close Elasticsearch client", nil)
		}
	}

	return err
}
