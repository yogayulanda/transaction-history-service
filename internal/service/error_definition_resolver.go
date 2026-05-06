package service

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"

	coreerrors "github.com/yogayulanda/go-core/errors"
	"github.com/yogayulanda/go-core/logger"
	"github.com/yogayulanda/transaction-history-service/internal/domain"
)

const defaultErrorDefinitionRefreshInterval = 60 * time.Second

type ErrorDefinitionResolver interface {
	Resolve(code string) (string, []coreerrors.Detail, bool)
}

type resolvedErrorDefinition struct {
	userMessage string
	details     []coreerrors.Detail
}

type DBErrorDefinitionResolver struct {
	repo domain.ErrorDefinitionRepository
	log  logger.Logger

	mu          sync.RWMutex
	definitions map[string]resolvedErrorDefinition
}

func NewDBErrorDefinitionResolver(
	repo domain.ErrorDefinitionRepository,
	log logger.Logger,
) *DBErrorDefinitionResolver {
	return &DBErrorDefinitionResolver{
		repo:        repo,
		log:         log,
		definitions: make(map[string]resolvedErrorDefinition),
	}
}

func (r *DBErrorDefinitionResolver) Load(ctx context.Context) error {
	return r.loadWithOperation(ctx, "error_definition_load")
}

func (r *DBErrorDefinitionResolver) StartAutoRefresh(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = defaultErrorDefinitionRefreshInterval
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				_ = r.loadWithOperation(ctx, "error_definition_refresh")
			}
		}
	}()
}

func (r *DBErrorDefinitionResolver) Resolve(code string) (string, []coreerrors.Detail, bool) {
	normalizedCode := strings.ToUpper(strings.TrimSpace(code))
	if normalizedCode == "" {
		return "", nil, false
	}

	r.mu.RLock()
	definition, ok := r.definitions[normalizedCode]
	r.mu.RUnlock()
	if !ok {
		r.emitServiceLog(context.Background(), "error_definition_miss", "failed", "not_found", map[string]interface{}{
			"error_code": normalizedCode,
		})
		return "", nil, false
	}

	details := make([]coreerrors.Detail, len(definition.details))
	copy(details, definition.details)
	return definition.userMessage, details, true
}

func (r *DBErrorDefinitionResolver) loadWithOperation(ctx context.Context, operation string) error {
	if r.repo == nil {
		return nil
	}

	rows, err := r.repo.ListActiveErrorDefinitions(ctx)
	if err != nil {
		r.emitServiceLog(ctx, operation, "failed", "query_failed", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	next := make(map[string]resolvedErrorDefinition, len(rows))
	malformedCount := 0
	for _, row := range rows {
		code := strings.ToUpper(strings.TrimSpace(row.ErrorCode))
		if code == "" {
			continue
		}

		details, parseErr := parseDefinitionDetails(row.DetailsJSON)
		if parseErr != nil {
			malformedCount++
			details = nil
		}

		next[code] = resolvedErrorDefinition{
			userMessage: strings.TrimSpace(row.UserMessage),
			details:     details,
		}
	}

	r.mu.Lock()
	r.definitions = next
	r.mu.Unlock()

	r.emitServiceLog(ctx, operation, "success", "", map[string]interface{}{
		"definition_count": len(next),
		"malformed_count":  malformedCount,
	})

	return nil
}

func parseDefinitionDetails(raw string) ([]coreerrors.Detail, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	var parsed []coreerrors.Detail
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return nil, err
	}

	if len(parsed) == 0 {
		return nil, nil
	}

	out := make([]coreerrors.Detail, 0, len(parsed))
	for _, detail := range parsed {
		field := strings.TrimSpace(detail.Field)
		reason := strings.TrimSpace(detail.Reason)
		if field == "" && reason == "" {
			continue
		}
		out = append(out, coreerrors.Detail{Field: field, Reason: reason})
	}
	if len(out) == 0 {
		return nil, nil
	}
	return out, nil
}

func (r *DBErrorDefinitionResolver) emitServiceLog(
	ctx context.Context,
	operation,
	status,
	errorCode string,
	metadata map[string]interface{},
) {
	if r.log == nil {
		return
	}
	r.log.LogService(ctx, logger.ServiceLog{
		Operation: operation,
		Status:    status,
		ErrorCode: errorCode,
		Metadata:  metadata,
	})
}
