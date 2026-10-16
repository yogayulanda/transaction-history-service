package app

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/yogayulanda/go-core/config"
)

func ConfigureInternalJWT(cfg *config.Config) error {
	publicKey := strings.TrimSpace(os.Getenv("HISTORY_JWT_PUBLIC_KEY"))
	jwksURL := strings.TrimSpace(os.Getenv("HISTORY_JWT_JWKS_URL"))
	if publicKey != "" && jwksURL != "" {
		return fmt.Errorf("HISTORY_JWT_PUBLIC_KEY and HISTORY_JWT_JWKS_URL cannot both be set")
	}

	jwtConfig := config.InternalJWTConfig{Leeway: 30 * time.Second}
	if publicKey == "" && jwksURL == "" {
		cfg.Auth.InternalJWT = jwtConfig
		return nil
	}

	issuer := strings.TrimSpace(os.Getenv("HISTORY_JWT_ISSUER"))
	if issuer == "" {
		return fmt.Errorf("HISTORY_JWT_ISSUER is required when HISTORY_JWT_PUBLIC_KEY or HISTORY_JWT_JWKS_URL is set")
	}
	audience := strings.TrimSpace(os.Getenv("HISTORY_JWT_AUDIENCE"))
	if audience == "" {
		return fmt.Errorf("HISTORY_JWT_AUDIENCE is required when HISTORY_JWT_PUBLIC_KEY or HISTORY_JWT_JWKS_URL is set")
	}

	jwtConfig.Enabled = true
	jwtConfig.PublicKey = publicKey
	jwtConfig.JWKSEndpoint = jwksURL
	jwtConfig.Issuer = issuer
	jwtConfig.Audience = audience
	cfg.Auth.InternalJWT = jwtConfig
	return nil
}
