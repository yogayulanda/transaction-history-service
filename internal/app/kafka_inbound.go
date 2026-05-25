package app

import (
	"context"
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

func loadKafkaInboundConfig() kafkaInboundConfig {
	return kafkaInboundConfig{
		Enabled:     getEnvBool("KAFKA_INBOUND_ENABLED", false),
		Topic:       handlerkafka.TransactionCreatedTopic,
		GroupID:     handlerkafka.TransactionCreatedGroupID,
		Concurrency: handlerkafka.TransactionCreatedConcurrency,
		RetryMax:    handlerkafka.TransactionCreatedRetryMax,
		RetryDelay:  handlerkafka.TransactionCreatedRetryDelay,
	}
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

func getEnvBool(key string, fallback bool) bool {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(raw)
	if err != nil {
		return fallback
	}
	return parsed
}
