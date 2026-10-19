package grpc

import (
	"context"

	coreerrors "github.com/yogayulanda/go-core/errors"
	"github.com/yogayulanda/go-core/security"
	"github.com/yogayulanda/transaction-history-service/internal/domain"
)

type Resolver interface {
	Resolve(ctx context.Context) (domain.Owner, error)
}

type ClaimsResolver struct {
	catalog domain.Catalog
}

func NewClaimsResolver(validationCatalog domain.Catalog) *ClaimsResolver {
	return &ClaimsResolver{catalog: validationCatalog}
}

func (r *ClaimsResolver) Resolve(ctx context.Context) (domain.Owner, error) {
	claims, ok := security.FromContext(ctx)
	if !ok || claims == nil || claims.Subject == "" || claims.Attributes["app_name"] == "" {
		return domain.Owner{}, coreerrors.New("CORE-1001", "authentication required")
	}
	appName := claims.Attributes["app_name"]
	if !r.catalog.ValidApp(appName) {
		return domain.Owner{}, coreerrors.New("CORE-1003", "application is not allowed")
	}
	return domain.Owner{AppName: appName, UserID: claims.Subject}, nil
}
