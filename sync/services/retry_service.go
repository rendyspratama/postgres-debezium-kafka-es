package services

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/rendyspratama/digital-discovery/sync/config"
	"github.com/rendyspratama/digital-discovery/sync/models"
	"github.com/rendyspratama/digital-discovery/sync/utils"
	"github.com/rendyspratama/digital-discovery/sync/utils/logger"
)

type RetryService struct {
	syncService *SyncService
	config      *config.Config
	logger      logger.Logger
}

type RetryAttempt struct {
	Attempt   int
	Error     error
	Timestamp time.Time
	NextRetry time.Time
	Duration  time.Duration
}

type RetryHistory struct {
	OperationID string
	Entity      string
	Operation   string
	Attempts    []RetryAttempt
	Status      string
}

func init() {
	// Initialize random seed for jitter
	rand.Seed(time.Now().UnixNano())
}

func NewRetryService(syncService *SyncService, config *config.Config, logger logger.Logger) *RetryService {
	return &RetryService{
		syncService: syncService,
		config:      config,
		logger:      logger,
	}
}

func (rs *RetryService) calculateNextDelay(attempt int, baseDelay time.Duration) time.Duration {
	// Calculate exponential delay
	delay := float64(baseDelay) * math.Pow(rs.config.Sync.Custom.BackoffFactor, float64(attempt))

	// Add jitter (±20%)
	jitter := rand.Float64()*0.4 - 0.2
	delay = delay * (1 + jitter)

	// Ensure delay doesn't exceed max
	if delay > float64(rs.config.Sync.Custom.MaxRetryDelay) {
		delay = float64(rs.config.Sync.Custom.MaxRetryDelay)
	}

	return time.Duration(delay)
}

func (rs *RetryService) RetryWithBackoff(ctx context.Context, operation *models.CategoryOperation) error {
	history := &RetryHistory{
		OperationID: operation.Payload.ID,
		Entity:      "category",
		Operation:   operation.Operation,
		Attempts:    make([]RetryAttempt, 0),
		Status:      "IN_PROGRESS",
	}

	defer rs.cleanup(ctx, history)

	var lastErr error
	attempt := 0
	baseDelay := rs.config.Sync.Custom.RetryDelay

	rs.logger.Info(ctx, "Starting retry sequence", map[string]interface{}{
		"operation_id": operation.Payload.ID,
		"operation":    operation.Operation,
		"max_retries":  rs.config.Sync.Custom.MaxRetries,
	})

	for attempt < rs.config.Sync.Custom.MaxRetries {
		delay := rs.calculateNextDelay(attempt, baseDelay)
		attemptStart := time.Now()
		err := rs.syncService.ProcessCategoryOperation(ctx, operation)

		retryAttempt := RetryAttempt{
			Attempt:   attempt + 1,
			Error:     err,
			Timestamp: attemptStart,
			Duration:  time.Since(attemptStart),
		}

		if err == nil {
			// Success
			history.Status = "SUCCESS"
			history.Attempts = append(history.Attempts, retryAttempt)
			rs.logRetryHistory(ctx, history)
			return nil
		}

		// Handle failure
		lastErr = err
		attempt++
		nextRetry := time.Now().Add(delay)
		retryAttempt.NextRetry = nextRetry
		history.Attempts = append(history.Attempts, retryAttempt)

		rs.logger.WithError(ctx, err, "Retry attempt failed", map[string]interface{}{
			"operation_id": operation.Payload.ID,
			"attempt":      attempt,
			"next_retry":   nextRetry,
			"delay":        delay.String(),
		})

		select {
		case <-ctx.Done():
			history.Status = "CANCELLED"
			return ctx.Err()
		case <-time.After(delay):
			continue
		}
	}

	// All retries failed
	history.Status = "FAILED"
	return utils.NewSyncError(
		utils.ErrCodeRetryExhausted,
		fmt.Sprintf("Max retries (%d) reached", rs.config.Sync.Custom.MaxRetries),
		lastErr,
		operation.Operation,
		"category",
	)
}

func (rs *RetryService) logRetryHistory(ctx context.Context, history *RetryHistory) {
	rs.logger.Info(ctx, "Retry sequence completed", map[string]interface{}{
		"operation_id":   history.OperationID,
		"entity":         history.Entity,
		"operation":      history.Operation,
		"status":         history.Status,
		"total_attempts": len(history.Attempts),
		"attempts":       history.Attempts,
	})
}

func (rs *RetryService) recordFailedAttempt(ctx context.Context, history *RetryHistory) {
	lastAttempt := history.Attempts[len(history.Attempts)-1]

	record := &models.SyncRecord{
		ID:           history.OperationID,
		EntityType:   history.Entity,
		EntityID:     history.OperationID,
		Operation:    history.Operation,
		Status:       models.SyncStatusFailed,
		ErrorMessage: lastAttempt.Error.Error(),
		RetryCount:   len(history.Attempts),
		LastRetry:    &lastAttempt.Timestamp,
		NextRetry:    &lastAttempt.NextRetry,
		CreatedAt:    history.Attempts[0].Timestamp,
		UpdatedAt:    lastAttempt.Timestamp,
	}

	rs.logger.Info(ctx, "Recording failed retry attempt", map[string]interface{}{
		"sync_record": record,
		"history":     history,
	})
}

func (rs *RetryService) cleanup(ctx context.Context, history *RetryHistory) {
	// Clean up any resources if needed
	if history.Status == "FAILED" {
		rs.recordFailedAttempt(ctx, history)
	}
}
