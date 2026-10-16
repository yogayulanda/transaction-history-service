package grpc

import (
	"context"
	"testing"

	"github.com/yogayulanda/go-core/security"
	"github.com/yogayulanda/transaction-history-service/internal/repository"
)

func TestClaimsResolver(t *testing.T) {
	resolver := NewClaimsResolver(repository.DefaultStaticCatalog("", ""))
	tests := []struct {
		name     string
		claims   *security.Claims
		wantApp  string
		wantUser string
		wantErr  bool
	}{
		{name: "missing claims", wantErr: true},
		{name: "missing subject", claims: &security.Claims{Attributes: map[string]string{"app_name": "app_a"}}, wantErr: true},
		{name: "missing app", claims: &security.Claims{Subject: "user-1"}, wantErr: true},
		{name: "invalid app", claims: &security.Claims{Subject: "user-1", Attributes: map[string]string{"app_name": "other"}}, wantErr: true},
		{name: "valid", claims: &security.Claims{Subject: "user-1", Attributes: map[string]string{"app_name": "app_a"}}, wantApp: "app_a", wantUser: "user-1"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			if test.claims != nil {
				ctx = security.Inject(ctx, test.claims)
			}
			owner, err := resolver.Resolve(ctx)
			if (err != nil) != test.wantErr {
				t.Fatalf("Resolve() error = %v, wantErr=%v", err, test.wantErr)
			}
			if err == nil && (owner.AppName != test.wantApp || owner.UserID != test.wantUser) {
				t.Fatalf("Resolve() = %+v", owner)
			}
		})
	}
}
