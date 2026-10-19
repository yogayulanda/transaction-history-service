package app

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/yogayulanda/go-core/messaging"
)

type kafkaInboundConfig struct {
	Enabled        bool
	Topic          string
	GroupID        string
	Concurrency    int
	RetryMax       int
	RetryDelay     time.Duration
	AllowedApps    []string
	AllowedSources []string
	HashKey        []byte
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

	cfg := kafkaInboundConfig{
		Enabled:     enabled,
		Concurrency: 3,
		RetryMax:    5,
		RetryDelay:  5 * time.Second,
	}
	if !enabled {
		return cfg, nil
	}
	cfg.Topic = strings.TrimSpace(os.Getenv("HISTORY_STATUS_TOPIC"))
	if cfg.Topic == "" {
		return kafkaInboundConfig{}, fmt.Errorf("HISTORY_STATUS_TOPIC is required when KAFKA_INBOUND_ENABLED=true")
	}
	cfg.GroupID = strings.TrimSpace(os.Getenv("SERVICE_NAME"))
	if cfg.GroupID == "" {
		return kafkaInboundConfig{}, fmt.Errorf("SERVICE_NAME is required when KAFKA_INBOUND_ENABLED=true")
	}
	allowedSources := strings.TrimSpace(os.Getenv("HISTORY_ALLOWED_SOURCES"))
	if allowedSources == "" {
		return kafkaInboundConfig{}, fmt.Errorf("HISTORY_ALLOWED_SOURCES is required when KAFKA_INBOUND_ENABLED=true")
	}
	cfg.AllowedSources = splitCSV(allowedSources)
	cfg.AllowedApps = splitCSV(os.Getenv("HISTORY_ALLOWED_APPS"))
	if len(cfg.AllowedApps) == 0 {
		cfg.AllowedApps = []string{"app_a", "app_b"}
	}
	key, err := hex.DecodeString(strings.TrimSpace(os.Getenv("HISTORY_HASH_KEY")))
	if err != nil || len(key) < 32 {
		return kafkaInboundConfig{}, fmt.Errorf("HISTORY_HASH_KEY must be hex encoding of at least 32 bytes")
	}
	cfg.HashKey = key
	return cfg, nil
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
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
	return "kafka_inbound_status_events"
}

func (r *kafkaInboundRunner) Start() error {
	return r.consumer.Start(r.ctx)
}
