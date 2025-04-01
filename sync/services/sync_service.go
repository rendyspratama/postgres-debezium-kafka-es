package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/rendyspratama/digital-discovery/sync/config"
	"github.com/rendyspratama/digital-discovery/sync/models"
	"github.com/rendyspratama/digital-discovery/sync/repositories/elasticsearch"
	"github.com/rendyspratama/digital-discovery/sync/utils"
	"github.com/rendyspratama/digital-discovery/sync/utils/logger"
	"github.com/rendyspratama/digital-discovery/sync/utils/metrics"
)

type SyncService struct {
	esClient    elasticsearch.Repository
	indexPrefix string
	config      *config.Config
	logger      logger.Logger
	metrics     *metrics.MetricsCollector
	mu          sync.RWMutex
	bulkBuffer  []models.CategoryOperation
}

func NewSyncService(esClient elasticsearch.Repository, cfg *config.Config, logger logger.Logger) *SyncService {
	return &SyncService{
		esClient:    esClient,
		indexPrefix: cfg.ES.IndexPrefix,
		config:      cfg,
		logger:      logger,
		metrics:     metrics.NewMetricsCollector(),
		bulkBuffer:  make([]models.CategoryOperation, 0, cfg.Sync.Custom.BatchSize),
	}
}

func (s *SyncService) ProcessCategoryOperation(ctx context.Context, operation *models.CategoryOperation) error {
	if operation == nil {
		return utils.NewSyncError(
			utils.ErrCodeInvalidPayload,
			"Operation cannot be nil",
			nil,
			"VALIDATE",
			"category",
		)
	}

	// Add operation validation with detailed error
	if err := s.validateOperation(operation); err != nil {
		s.logger.WithError(ctx, err, "Operation validation failed", map[string]interface{}{
			"operation": operation.Operation,
			"id":        operation.Payload.ID,
			"payload":   operation.Payload,
		})
		return err
	}

	// Add context timeout for operation
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	opMetrics := &metrics.OperationMetrics{
		StartTime:   time.Now(),
		Operation:   operation.Operation,
		Entity:      "category",
		EntityID:    operation.Payload.ID,
		Status:      "IN_PROGRESS",
		PayloadSize: 0,
		ErrorCount:  0,
	}

	defer func() {
		opMetrics.EndTime = time.Now()
		opMetrics.Duration = opMetrics.EndTime.Sub(opMetrics.StartTime)
		s.logOperationMetrics(ctx, opMetrics)
		s.recordOperationResult(ctx, operation, opMetrics)
		s.metrics.RecordOperation(opMetrics)
	}()

	s.logger.Info(ctx, "Starting category operation", map[string]interface{}{
		"operation":   operation.Operation,
		"category_id": operation.Payload.ID,
		"timestamp":   operation.Timestamp,
	})

	indexName := s.getCurrentIndexName("categories")
	opMetrics.IndexName = indexName

	// Safe JSON marshaling
	if payloadJSON, err := json.Marshal(operation.Payload); err == nil {
		opMetrics.PayloadSize = len(payloadJSON)
	} else {
		s.logger.WithError(ctx, err, "Failed to marshal payload for metrics", nil)
	}

	var err error
	switch operation.Operation {
	case models.OperationCreate, models.OperationUpdate, models.OperationDelete:
		err = s.processOperation(ctx, indexName, operation)
	default:
		opMetrics.Status = "FAILED"
		opMetrics.ErrorCount++
		return utils.NewSyncError(
			utils.ErrCodeInvalidPayload,
			fmt.Sprintf("Unknown operation: %s", operation.Operation),
			nil,
			operation.Operation,
			"category",
		)
	}

	if err != nil {
		opMetrics.Status = "FAILED"
		opMetrics.ErrorCount++
		s.logger.WithError(ctx, err, "Operation failed", map[string]interface{}{
			"operation":   operation.Operation,
			"category_id": operation.Payload.ID,
			"index":       indexName,
			"duration":    opMetrics.Duration.String(),
		})
		return err
	}

	opMetrics.Status = "SUCCESS"
	s.logger.Info(ctx, "Operation completed successfully", map[string]interface{}{
		"operation":   operation.Operation,
		"category_id": operation.Payload.ID,
		"index":       indexName,
		"duration":    opMetrics.Duration.String(),
	})

	return nil
}

func (s *SyncService) processOperation(ctx context.Context, indexName string, operation *models.CategoryOperation) error {
	switch operation.Operation {
	case models.OperationCreate:
		return s.createCategory(ctx, indexName, operation.Payload)
	case models.OperationUpdate:
		return s.updateCategory(ctx, indexName, operation.Payload)
	case models.OperationDelete:
		return s.deleteCategory(ctx, indexName, operation.Payload.ID)
	default:
		return utils.NewSyncError(
			utils.ErrCodeInvalidPayload,
			"Invalid operation",
			nil,
			operation.Operation,
			"category",
		)
	}
}

func (s *SyncService) validateOperation(operation *models.CategoryOperation) error {
	if operation.Payload.ID == "" {
		return utils.NewSyncError(
			utils.ErrCodeInvalidPayload,
			"Missing category ID",
			nil,
			operation.Operation,
			"category",
		)
	}

	if operation.Operation == models.OperationCreate || operation.Operation == models.OperationUpdate {
		if err := s.validateCategoryFields(operation.Payload); err != nil {
			return err
		}
	}

	return nil
}

func (s *SyncService) validateCategoryFields(category models.Category) error {
	if category.Name == "" {
		return utils.NewSyncError(
			utils.ErrCodeInvalidPayload,
			"Missing category name",
			nil,
			"VALIDATE",
			"category",
		)
	}

	if category.Description == "" {
		return utils.NewSyncError(
			utils.ErrCodeInvalidPayload,
			"Missing category description",
			nil,
			"VALIDATE",
			"category",
		)
	}

	return nil
}

func (s *SyncService) createCategory(ctx context.Context, indexName string, category models.Category) error {
	category.SyncStatus = models.SyncStatusSuccess
	category.LastSync = time.Now()

	body := strings.NewReader(mustJSON(category))
	err := s.esClient.Index(ctx, indexName, category.ID, body)
	if err != nil {
		return utils.NewESIndexError("Failed to index category", err)
	}
	return nil
}

func (s *SyncService) updateCategory(ctx context.Context, indexName string, category models.Category) error {
	category.SyncStatus = models.SyncStatusSuccess
	category.LastSync = time.Now()

	updateBody := map[string]interface{}{
		"doc":           category,
		"doc_as_upsert": true,
	}

	body := strings.NewReader(mustJSON(updateBody))
	err := s.esClient.Update(ctx, indexName, category.ID, body)
	if err != nil {
		return utils.NewESIndexError("Failed to update category", err)
	}
	return nil
}

func (s *SyncService) deleteCategory(ctx context.Context, indexName string, id string) error {
	err := s.esClient.Delete(ctx, indexName, id)
	if err != nil {
		return utils.NewESIndexError("Failed to delete category", err)
	}
	return nil
}

func (s *SyncService) getCurrentIndexName(entity string) string {
	return fmt.Sprintf("%s-%s-%s-%s",
		s.config.App.Environment,
		"digital-discovery",
		entity,
		time.Now().Format("2006-01"))
}

func mustJSON(v interface{}) string {
	defer func() {
		if r := recover(); r != nil {
			panic(fmt.Sprintf("Failed to marshal JSON: %v", r))
		}
	}()

	b, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("Failed to marshal JSON: %v", err))
	}
	return string(b)
}

func (s *SyncService) logOperationMetrics(ctx context.Context, metrics *metrics.OperationMetrics) {
	s.logger.Info(ctx, "Operation metrics", map[string]interface{}{
		"operation":    metrics.Operation,
		"entity":       metrics.Entity,
		"entity_id":    metrics.EntityID,
		"status":       metrics.Status,
		"index":        metrics.IndexName,
		"duration_ms":  metrics.Duration.Milliseconds(),
		"start_time":   metrics.StartTime,
		"end_time":     metrics.EndTime,
		"payload_size": metrics.PayloadSize,
		"error_count":  metrics.ErrorCount,
	})
}

func (s *SyncService) recordOperationResult(ctx context.Context, operation *models.CategoryOperation, metrics *metrics.OperationMetrics) {
	if operation == nil || metrics == nil {
		s.logger.Error(ctx, "Invalid operation or metrics", nil)
		return
	}

	record := &models.SyncRecord{
		ID:           operation.Payload.ID,
		EntityType:   "category",
		EntityID:     operation.Payload.ID,
		Operation:    operation.Operation,
		Status:       models.SyncStatus(metrics.Status),
		ErrorMessage: "",
		RetryCount:   metrics.ErrorCount,
		CreatedAt:    metrics.StartTime,
		UpdatedAt:    metrics.EndTime,
	}

	if metrics.Status == "FAILED" {
		record.MarkAsFailed(
			fmt.Errorf("operation failed with %d errors", metrics.ErrorCount),
			s.config.Sync.Custom.RetryDelay,
		)
		s.metrics.RecordError(operation.Operation, "category", metrics.ErrorCount)
	} else {
		record.MarkAsSuccess()
	}

	s.logger.Info(ctx, "Recording operation result", map[string]interface{}{
		"sync_record": record,
		"metrics":     metrics,
	})
}

func (s *SyncService) processBulkOperations(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.bulkBuffer) == 0 {
		return nil
	}

	bufferSize := len(s.bulkBuffer)
	var buf strings.Builder

	for _, op := range s.bulkBuffer {
		// Add action line
		var action string
		switch op.Operation {
		case models.OperationCreate:
			action = "index"
		case models.OperationUpdate:
			action = "update"
		case models.OperationDelete:
			action = "delete"
		default:
			continue
		}

		actionLine := map[string]interface{}{
			action: map[string]interface{}{
				"_index": s.getCurrentIndexName("categories"),
				"_id":    op.Payload.ID,
			},
		}
		if err := json.NewEncoder(&buf).Encode(actionLine); err != nil {
			s.metrics.RecordBulkOperation("category", bufferSize, true)
			return fmt.Errorf("failed to encode action line: %w", err)
		}

		// Add payload line for non-delete operations
		if op.Operation != models.OperationDelete {
			var payload interface{}
			if op.Operation == models.OperationUpdate {
				payload = map[string]interface{}{
					"doc":           op.Payload,
					"doc_as_upsert": true,
				}
			} else {
				payload = op.Payload
			}

			if err := json.NewEncoder(&buf).Encode(payload); err != nil {
				s.metrics.RecordBulkOperation("category", bufferSize, true)
				return fmt.Errorf("failed to encode payload: %w", err)
			}
		}
	}

	err := s.esClient.Bulk(ctx, strings.NewReader(buf.String()))
	if err != nil {
		s.metrics.RecordBulkOperation("category", bufferSize, true)
		return utils.NewESIndexError("Bulk operation failed", err)
	}

	s.metrics.RecordBulkOperation("category", bufferSize, false)
	s.bulkBuffer = s.bulkBuffer[:0]
	return nil
}

// Add method to check if operation can be bulked
func (s *SyncService) canBulkOperation(operation *models.CategoryOperation) bool {
	return models.IsValidOperation(operation.Operation)
}

// Add context to FlushBulkBuffer
func (s *SyncService) FlushBulkBuffer(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := s.processBulkOperations(ctx); err != nil {
		s.logger.WithError(ctx, err, "Failed to flush bulk buffer", map[string]interface{}{
			"buffer_size": len(s.bulkBuffer),
		})
		return err
	}

	return nil
}

// Update RetryOperation method to pass the logger interface directly
func (s *SyncService) RetryOperation(ctx context.Context, operation *models.CategoryOperation) error {
	retryService := NewRetryService(s, s.config, s.logger)
	return retryService.RetryWithBackoff(ctx, operation)
}

// Update addToBulkBuffer to be exported for use in bulk operations
func (s *SyncService) AddToBulkBuffer(operation models.CategoryOperation) error {
	if !s.canBulkOperation(&operation) {
		return utils.NewSyncError(
			utils.ErrCodeInvalidPayload,
			"Operation not supported for bulk processing",
			nil,
			operation.Operation,
			"category",
		)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.bulkBuffer = append(s.bulkBuffer, operation)

	// Auto-flush if buffer is full
	if len(s.bulkBuffer) >= s.config.Sync.Custom.BatchSize {
		return s.FlushBulkBuffer(context.Background())
	}

	return nil
}

// CreateCategory creates a new category in Elasticsearch
func (s *SyncService) CreateCategory(ctx context.Context, category models.Category) error {
	indexName := s.getCurrentIndexName("categories")
	return s.createCategory(ctx, indexName, category)
}

// UpdateCategory updates an existing category in Elasticsearch
func (s *SyncService) UpdateCategory(ctx context.Context, category models.Category) error {
	indexName := s.getCurrentIndexName("categories")
	return s.updateCategory(ctx, indexName, category)
}

// DeleteCategory deletes a category from Elasticsearch
func (s *SyncService) DeleteCategory(ctx context.Context, id string) error {
	indexName := s.getCurrentIndexName("categories")
	return s.deleteCategory(ctx, indexName, id)
}

// GetCategory retrieves a category from Elasticsearch
func (s *SyncService) GetCategory(ctx context.Context, id string) (*models.Category, error) {
	indexName := s.getCurrentIndexName("categories")

	// Create a search query to find the document
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"term": map[string]interface{}{
				"_id": id,
			},
		},
	}

	// Execute search
	docs, err := s.esClient.Search(ctx, indexName, query)
	if err != nil {
		return nil, utils.NewESIndexError("Failed to search category", err)
	}

	if len(docs) == 0 {
		return nil, utils.NewESIndexError("Category not found", nil)
	}

	// Parse document into Category struct
	var category models.Category
	if err := json.Unmarshal(docs[0], &category); err != nil {
		return nil, utils.NewESIndexError("Failed to parse category", err)
	}

	return &category, nil
}

// ListCategories retrieves all categories from Elasticsearch
func (s *SyncService) ListCategories(ctx context.Context) ([]models.Category, error) {
	indexName := s.getCurrentIndexName("categories")

	// Create a search query to find all documents
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match_all": map[string]interface{}{},
		},
	}

	// Execute search
	docs, err := s.esClient.Search(ctx, indexName, query)
	if err != nil {
		return nil, utils.NewESIndexError("Failed to search categories", err)
	}

	// Parse documents into Category structs
	var categories []models.Category
	for _, doc := range docs {
		var category models.Category
		if err := json.Unmarshal(doc, &category); err != nil {
			return nil, utils.NewESIndexError("Failed to parse category", err)
		}
		categories = append(categories, category)
	}

	return categories, nil
}

func (s *SyncService) GetCurrentIndexName(entity string) string {
	return s.getCurrentIndexName(entity)
}

func (s *SyncService) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check Elasticsearch connection using basic request if Ping is not available
	_, err := s.esClient.Search(ctx, "_all", map[string]interface{}{
		"size": 0,
	})
	if err != nil {
		return fmt.Errorf("elasticsearch health check failed: %w", err)
	}

	// Check current index exists using Search with size 0 if IndexExists is not available
	indexName := s.getCurrentIndexName("categories")
	_, err = s.esClient.Search(ctx, indexName, map[string]interface{}{
		"size": 0,
	})
	if err != nil {
		return fmt.Errorf("failed to check index existence: %w", err)
	}

	// Check bulk buffer status using default size if not configured
	s.mu.RLock()
	bufferSize := len(s.bulkBuffer)
	maxSize := s.config.Sync.Custom.BatchSize
	s.mu.RUnlock()

	if bufferSize >= maxSize {
		return fmt.Errorf("bulk buffer is full: %d items", bufferSize)
	}

	return nil
}
