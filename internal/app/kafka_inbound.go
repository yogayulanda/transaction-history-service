package app

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/yogayulanda/go-core/messaging"
	handlerkafka "github.com/yogayulanda/transaction-history-service/internal/handler/kafka"
)

type kafkaInboundConfig struct {
	Enabled     bool
	Topic       string
	GroupID     string
	Concurrency int
	RetryMax    int
	RetryDelay  time.Duration
}

func loadKafkaInboundConfig() (kafkaInboundConfig, error) {
	var enabled bool
	if raw := strings.TrimSpace(os.Getenv("KAFKA_INBOUND_ENABLED")); raw != "" {
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			return kafkaInboundConfig{}, fmt.Errorf("invalid KAFKA_INBOUND_ENABLED: %w", err)
		}
		enabled = parsed
	}

	return kafkaInboundConfig{
		Enabled:     enabled,
		Topic:       handlerkafka.TransactionCreatedTopic,
		GroupID:     handlerkafka.TransactionCreatedGroupID,
		Concurrency: handlerkafka.TransactionCreatedConcurrency,
		RetryMax:    handlerkafka.TransactionCreatedRetryMax,
		RetryDelay:  handlerkafka.TransactionCreatedRetryDelay,
	}, nil
}

type kafkaInboundRunner struct {
	consumer messaging.Consumer
	ctx      context.Context
}

func newKafkaInboundRunner(ctx context.Context, consumer messaging.Consumer) *kafkaInboundRunner {
	return &kafkaInboundRunner{
		consumer: consumer,
		ctx:      ctx,
	}
}

func (r *kafkaInboundRunner) Name() string {
	return "kafka_inbound_transaction_created"
}

func (r *kafkaInboundRunner) Start() error {
	return r.consumer.Start(r.ctx)
}
