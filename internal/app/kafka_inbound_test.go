package app

import (
	"strings"
	"testing"
)

func TestLoadKafkaInboundConfig_DisabledByDefault(t *testing.T) {
	t.Setenv("KAFKA_INBOUND_ENABLED", "")

	cfg, err := loadKafkaInboundConfig()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.Enabled {
		t.Fatal("expected kafka inbound disabled by default")
	}
}

func TestLoadKafkaInboundConfig_EnabledExplicitly(t *testing.T) {
	t.Setenv("KAFKA_INBOUND_ENABLED", "true")
	t.Setenv("SERVICE_NAME", "transaction-history-service")
	t.Setenv("HISTORY_STATUS_TOPIC", "sample.transaction.status")
	t.Setenv("HISTORY_ALLOWED_SOURCES", "sample-producer")
	t.Setenv("HISTORY_HASH_KEY", strings.Repeat("6b", 32))

	cfg, err := loadKafkaInboundConfig()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.Enabled {
		t.Fatal("expected kafka inbound enabled")
	}
	if cfg.Topic != "sample.transaction.status" {
		t.Fatalf("unexpected topic %q", cfg.Topic)
	}
	if cfg.GroupID != "transaction-history-service" || len(cfg.AllowedSources) != 1 || len(cfg.HashKey) != 32 {
		t.Fatalf("unexpected group id %q", cfg.GroupID)
	}
	if cfg.Concurrency != 3 || cfg.RetryMax != 5 || cfg.RetryDelay.String() != "5s" {
		t.Fatalf("unexpected runtime config: %+v", cfg)
	}
}

func TestLoadKafkaInboundConfig_RequiresStatusConfiguration(t *testing.T) {
	t.Setenv("KAFKA_INBOUND_ENABLED", "true")
	t.Setenv("SERVICE_NAME", "transaction-history-service")
	t.Setenv("HISTORY_STATUS_TOPIC", "")
	t.Setenv("HISTORY_ALLOWED_SOURCES", "sample-producer")
	t.Setenv("HISTORY_HASH_KEY", strings.Repeat("6b", 32))
	if _, err := loadKafkaInboundConfig(); err == nil || !strings.Contains(err.Error(), "HISTORY_STATUS_TOPIC") {
		t.Fatalf("expected missing status topic error, got %v", err)
	}
}

func TestLoadKafkaInboundConfig_RejectsMalformedEnabledValue(t *testing.T) {
	t.Setenv("KAFKA_INBOUND_ENABLED", "sometimes")

	_, err := loadKafkaInboundConfig()
	if err == nil {
		t.Fatal("expected malformed KAFKA_INBOUND_ENABLED to fail")
	}
	if got := err.Error(); got == "" || !strings.Contains(got, "KAFKA_INBOUND_ENABLED") {
		t.Fatalf("expected error to name KAFKA_INBOUND_ENABLED, got %q", got)
	}
}
