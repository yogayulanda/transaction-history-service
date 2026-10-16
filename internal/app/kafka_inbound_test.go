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

	cfg, err := loadKafkaInboundConfig()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.Enabled {
		t.Fatal("expected kafka inbound enabled")
	}
	if cfg.Topic != "transaction-history.transaction.created" {
		t.Fatalf("unexpected topic %q", cfg.Topic)
	}
	if cfg.GroupID != "transaction-history-service" {
		t.Fatalf("unexpected group id %q", cfg.GroupID)
	}
	if cfg.Concurrency != 3 || cfg.RetryMax != 5 || cfg.RetryDelay.String() != "5s" {
		t.Fatalf("unexpected runtime config: %+v", cfg)
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
