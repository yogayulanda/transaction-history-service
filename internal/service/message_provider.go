package service

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/yogayulanda/go-core/logger"
	"github.com/yogayulanda/transaction-history-service/internal/domain"
)

const defaultErrorDefinitionRefreshInterval = 60 * time.Second

type DBMessageProvider struct {
	repo domain.MessageRepository
	log  logger.Logger

	mu          sync.RWMutex
	definitions map[string]string
}

func NewDBMessageProvider(
	repo domain.MessageRepository,
	log logger.Logger,
) *DBMessageProvider {
	return &DBMessageProvider{
		repo:        repo,
		log:         log,
		definitions: make(map[string]string),
	}
}

func (r *DBMessageProvider) Load(ctx context.Context) error {
	return r.loadWithOperation(ctx, "error_definition_load")
}

func (r *DBMessageProvider) StartAutoRefresh(ctx context.Context, interval time.Duration) {
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

func (r *DBMessageProvider) Message(_ context.Context, code string) (string, bool) {
	normalizedCode := strings.ToUpper(strings.TrimSpace(code))
	if normalizedCode == "" {
		return "", false
	}

	r.mu.RLock()
	message, ok := r.definitions[normalizedCode]
	r.mu.RUnlock()
	return message, ok && message != ""
}

func (r *DBMessageProvider) loadWithOperation(ctx context.Context, operation string) error {
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

	next := make(map[string]string, len(rows))
	for _, row := range rows {
		code := strings.ToUpper(strings.TrimSpace(row.ErrorCode))
		if code == "" {
			continue
		}
		next[code] = strings.TrimSpace(row.UserMessage)
	}

	r.mu.Lock()
	r.definitions = next
	r.mu.Unlock()

	r.emitServiceLog(ctx, operation, "success", "", map[string]interface{}{
		"definition_count": len(next),
	})

	return nil
}

func (r *DBMessageProvider) emitServiceLog(
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
