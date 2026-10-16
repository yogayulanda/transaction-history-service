package service

import (
	"context"
	"time"

	"github.com/yogayulanda/go-core/logger"
	"github.com/yogayulanda/transaction-history-service/internal/domain"
)

type statusEventRepository interface {
	ApplyStatusEvent(ctx context.Context, header *domain.TransactionSnapshot, detail *domain.Detail, ev domain.StatusEvent) (domain.ApplyResult, error)
}

type StatusEventService struct {
	repo statusEventRepository
	log  logger.Logger
}

func NewStatusEventService(repo statusEventRepository, log logger.Logger) *StatusEventService {
	return &StatusEventService{repo: repo, log: log}
}

func (s *StatusEventService) ApplyStatusEvent(
	ctx context.Context,
	header *domain.TransactionSnapshot,
	detail *domain.Detail,
	ev domain.StatusEvent,
) (domain.ApplyResult, error) {
	startedAt := time.Now()
	result, err := s.repo.ApplyStatusEvent(ctx, header, detail, ev)
	if err != nil {
		s.emit(ctx, "failed", "apply_failed", startedAt)
		return "", err
	}
	s.emit(ctx, string(result), "", startedAt)
	return result, nil
}

func (s *StatusEventService) emit(ctx context.Context, status, errorCode string, startedAt time.Time) {
	if s.log == nil {
		return
	}
	s.log.LogService(ctx, logger.ServiceLog{
		Operation:  "status_event_apply",
		Status:     status,
		DurationMs: time.Since(startedAt).Milliseconds(),
		ErrorCode:  errorCode,
	})
}
