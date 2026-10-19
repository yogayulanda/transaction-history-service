package app

import (
	"strings"
	"testing"

	"github.com/yogayulanda/go-core/config"
)

func setJWTEnvironment(t *testing.T, publicKey, jwksURL, issuer, audience string) {
	t.Helper()
	t.Setenv("HISTORY_JWT_PUBLIC_KEY", publicKey)
	t.Setenv("HISTORY_JWT_JWKS_URL", jwksURL)
	t.Setenv("HISTORY_JWT_ISSUER", issuer)
	t.Setenv("HISTORY_JWT_AUDIENCE", audience)
}

func TestConfigureInternalJWTDisabledWithoutKeySource(t *testing.T) {
	setJWTEnvironment(t, "", "", "", "")
	cfg := &config.Config{}
	if err := ConfigureInternalJWT(cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Auth.InternalJWT.Enabled {
		t.Fatal("InternalJWT should be disabled without a key source")
	}
}

func TestConfigureInternalJWTWithPublicKey(t *testing.T) {
	setJWTEnvironment(t, "test-public-key-pem", "", "sample-issuer", "sample-audience")
	cfg := &config.Config{}
	if err := ConfigureInternalJWT(cfg); err != nil {
		t.Fatal(err)
	}
	if !cfg.Auth.InternalJWT.Enabled || cfg.Auth.InternalJWT.PublicKey != "test-public-key-pem" ||
		cfg.Auth.InternalJWT.Issuer != "sample-issuer" || cfg.Auth.InternalJWT.Audience != "sample-audience" {
		t.Fatalf("unexpected JWT config: %+v", cfg.Auth.InternalJWT)
	}
}

func TestConfigureInternalJWTWithJWKS(t *testing.T) {
	setJWTEnvironment(t, "", "https://keys.example.test/jwks", "sample-issuer", "sample-audience")
	cfg := &config.Config{}
	if err := ConfigureInternalJWT(cfg); err != nil {
		t.Fatal(err)
	}
	if !cfg.Auth.InternalJWT.Enabled || cfg.Auth.InternalJWT.JWKSEndpoint != "https://keys.example.test/jwks" ||
		cfg.Auth.InternalJWT.Issuer != "sample-issuer" || cfg.Auth.InternalJWT.Audience != "sample-audience" {
		t.Fatalf("unexpected JWT config: %+v", cfg.Auth.InternalJWT)
	}
}

func TestConfigureInternalJWTRejectsBothKeySources(t *testing.T) {
	setJWTEnvironment(t, "test-public-key-pem", "https://keys.example.test/jwks", "sample-issuer", "sample-audience")
	err := ConfigureInternalJWT(&config.Config{})
	if err == nil || !strings.Contains(err.Error(), "HISTORY_JWT_PUBLIC_KEY") || !strings.Contains(err.Error(), "HISTORY_JWT_JWKS_URL") {
		t.Fatalf("ConfigureInternalJWT() error = %v, want both variable names", err)
	}
}

func TestConfigureInternalJWTRequiresIssuerAndAudience(t *testing.T) {
	tests := []struct {
		name     string
		issuer   string
		audience string
		missing  string
	}{
		{name: "issuer", audience: "sample-audience", missing: "HISTORY_JWT_ISSUER"},
		{name: "audience", issuer: "sample-issuer", missing: "HISTORY_JWT_AUDIENCE"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			setJWTEnvironment(t, "test-public-key-pem", "", test.issuer, test.audience)
			err := ConfigureInternalJWT(&config.Config{})
			if err == nil || !strings.Contains(err.Error(), test.missing) {
				t.Fatalf("ConfigureInternalJWT() error = %v, want %s", err, test.missing)
			}
		})
	}
}
