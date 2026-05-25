package app

import "testing"

func TestLoadKafkaInboundConfig_DisabledByDefault(t *testing.T) {
	t.Setenv("KAFKA_INBOUND_ENABLED", "")

	cfg := loadKafkaInboundConfig()
	if cfg.Enabled {
		t.Fatal("expected kafka inbound disabled by default")
	}
}

func TestLoadKafkaInboundConfig_EnabledExplicitly(t *testing.T) {
	t.Setenv("KAFKA_INBOUND_ENABLED", "true")

	cfg := loadKafkaInboundConfig()
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
