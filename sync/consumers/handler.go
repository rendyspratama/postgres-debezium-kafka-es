package consumers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Shopify/sarama"
	"github.com/rendyspratama/digital-discovery/sync/models"
	"github.com/rendyspratama/digital-discovery/sync/services"
	"github.com/rendyspratama/digital-discovery/sync/utils"
	"github.com/rendyspratama/digital-discovery/sync/utils/logger"
)

type ConsumerHandler struct {
	syncService *services.SyncService
	logger      logger.Logger
	ready       chan bool
}

type DebeziumEvent struct {
	Payload struct {
		Before json.RawMessage `json:"before"`
		After  json.RawMessage `json:"after"`
		Source struct {
			Version   string `json:"version"`
			Connector string `json:"connector"`
			Database  string `json:"database"`
			Schema    string `json:"schema"`
			Table     string `json:"table"`
			TxId      string `json:"txId"`
			Lsn       string `json:"lsn"`
			Timestamp int64  `json:"ts_ms"`
		} `json:"source"`
		Op string `json:"op"`
	} `json:"payload"`
}

func (h *ConsumerHandler) Setup(sarama.ConsumerGroupSession) error {
	close(h.ready)
	return nil
}

func (h *ConsumerHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *ConsumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				return nil
			}

			ctx := context.WithValue(session.Context(), "requestID", session.GenerationID())

			h.logger.Info(ctx, "Processing message", map[string]interface{}{
				"topic":     message.Topic,
				"partition": message.Partition,
				"offset":    message.Offset,
			})

			if err := h.processMessage(ctx, message); err != nil {
				h.logger.WithError(ctx, err, "Failed to process message", map[string]interface{}{
					"topic":     message.Topic,
					"partition": message.Partition,
					"offset":    message.Offset,
				})
				continue
			}

			session.MarkMessage(message, "")

		case <-session.Context().Done():
			return nil
		}
	}
}

func (h *ConsumerHandler) processMessage(ctx context.Context, message *sarama.ConsumerMessage) error {
	var event DebeziumEvent
	if err := json.Unmarshal(message.Value, &event); err != nil {
		return utils.NewSyncError(
			utils.ErrCodeKafkaDeserialize,
			"Invalid message format",
			err,
			"DESERIALIZE",
			"message",
		)
	}

	if err := h.validateMessage(&event); err != nil {
		return err
	}

	operation := h.mapOperation(event.Payload.Op)
	var category models.Category

	switch operation {
	case models.OperationCreate, models.OperationUpdate:
		if err := json.Unmarshal(event.Payload.After, &category); err != nil {
			return utils.NewSyncError(
				utils.ErrCodeDataTransform,
				"Failed to unmarshal category",
				err,
				operation,
				"category",
			)
		}
	case models.OperationDelete:
		if err := json.Unmarshal(event.Payload.Before, &category); err != nil {
			return utils.NewSyncError(
				utils.ErrCodeDataTransform,
				"Failed to unmarshal category",
				err,
				operation,
				"category",
			)
		}
	default:
		return utils.NewSyncError(
			utils.ErrCodeInvalidPayload,
			fmt.Sprintf("Unknown operation: %s", operation),
			nil,
			operation,
			"category",
		)
	}

	categoryOp := &models.CategoryOperation{
		Operation: operation,
		Payload:   category,
		Timestamp: time.Unix(0, event.Payload.Source.Timestamp*int64(time.Millisecond)),
	}

	err := h.syncService.ProcessCategoryOperation(ctx, categoryOp)
	if err != nil {
		// If the error is retryable, attempt retry
		if utils.IsRetryableError(err) {
			return h.syncService.RetryOperation(ctx, categoryOp)
		}
		return err
	}

	return nil
}

func (h *ConsumerHandler) validateMessage(event *DebeziumEvent) error {
	if event.Payload.Source.Timestamp == 0 {
		return utils.NewSyncError(
			utils.ErrCodeInvalidPayload,
			"Missing timestamp in event",
			nil,
			"VALIDATE",
			"message",
		)
	}

	if event.Payload.Op == "" {
		return utils.NewSyncError(
			utils.ErrCodeInvalidPayload,
			"Missing operation in event",
			nil,
			"VALIDATE",
			"message",
		)
	}

	return nil
}

func (h *ConsumerHandler) mapOperation(op string) string {
	switch op {
	case "c":
		return "CREATE"
	case "u":
		return "UPDATE"
	case "d":
		return "DELETE"
	default:
		return "UNKNOWN"
	}
}

func NewConsumerHandler(syncService *services.SyncService, logger logger.Logger) *ConsumerHandler {
	return &ConsumerHandler{
		syncService: syncService,
		logger:      logger,
		ready:       make(chan bool),
	}
}
