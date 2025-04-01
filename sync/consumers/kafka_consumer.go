package consumers

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	"github.com/rendyspratama/digital-discovery/sync/config"
	"github.com/rendyspratama/digital-discovery/sync/services"
	"github.com/rendyspratama/digital-discovery/sync/utils/logger"
)

type KafkaConsumer struct {
	consumer    sarama.ConsumerGroup
	syncService *services.SyncService
	logger      logger.Logger
	topics      []string
	status      string
	statusMu    sync.RWMutex
}

func NewKafkaConsumer(cfg *config.Config, syncService *services.SyncService, logger logger.Logger) (*KafkaConsumer, error) {
	config := sarama.NewConfig()

	// Version must be greater than 0.10.2.0
	config.Version = sarama.V2_8_0_0

	// Consumer group settings
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.AutoCommit.Enable = true
	config.Consumer.Offsets.AutoCommit.Interval = 1 * time.Second

	if cfg.Kafka.SecurityEnabled {
		config.Net.SASL.Enable = true
		config.Net.SASL.User = cfg.Kafka.SASL.Username
		config.Net.SASL.Password = cfg.Kafka.SASL.Password
		config.Net.SASL.Mechanism = sarama.SASLTypePlaintext
	}

	// Add additional consumer configurations
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.AutoCommit.Enable = true
	config.Consumer.Offsets.AutoCommit.Interval = 1 * time.Second

	// Create consumer group
	group, err := sarama.NewConsumerGroup(cfg.Kafka.Brokers, cfg.Kafka.GroupID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer group: %w", err)
	}

	return &KafkaConsumer{
		consumer:    group,
		syncService: syncService,
		logger:      logger,
		topics:      []string{fmt.Sprintf("%s.categories", cfg.Kafka.TopicPrefix)},
		status:      "initialized",
	}, nil
}

func (c *KafkaConsumer) Start(ctx context.Context) error {
	c.setStatus("starting")

	// Handle errors
	go func() {
		for err := range c.consumer.Errors() {
			c.logger.WithError(ctx, err, "Error from consumer", nil)
			c.setStatus("error")
		}
	}()

	c.setStatus("running")

	// Consume messages
	for {
		handler := NewConsumerHandler(c.syncService, c.logger)

		err := c.consumer.Consume(ctx, c.topics, handler)
		if err != nil {
			if err == sarama.ErrClosedConsumerGroup {
				c.setStatus("closed")
				return nil
			}
			c.setStatus("error")
			return fmt.Errorf("error from consumer: %w", err)
		}

		// Check if context was cancelled
		if ctx.Err() != nil {
			c.setStatus("stopped")
			return ctx.Err()
		}
	}
}

func (c *KafkaConsumer) Close() error {
	c.setStatus("closing")
	err := c.consumer.Close()
	if err != nil {
		c.setStatus("error")
		return err
	}
	c.setStatus("closed")
	return nil
}

func (c *KafkaConsumer) HealthCheck() error {
	if c.consumer == nil {
		return fmt.Errorf("consumer is not initialized")
	}

	status := c.getStatus()
	if status == "error" || status == "closed" {
		return fmt.Errorf("consumer is in %s state", status)
	}

	return nil
}

func (c *KafkaConsumer) setStatus(status string) {
	c.statusMu.Lock()
	defer c.statusMu.Unlock()
	c.status = status
}

func (c *KafkaConsumer) getStatus() string {
	c.statusMu.RLock()
	defer c.statusMu.RUnlock()
	return c.status
}
