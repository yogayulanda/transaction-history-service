package service

import (
	"context"
	"errors"
	"testing"

	"github.com/yogayulanda/transaction-history-service/internal/domain"
)

type statusEventRepositoryStub struct {
	result domain.ApplyResult
	err    error
}

func (s statusEventRepositoryStub) ApplyStatusEvent(
	context.Context,
	*domain.TransactionSnapshot,
	*domain.Detail,
	domain.StatusEvent,
) (domain.ApplyResult, error) {
	return s.result, s.err
}

func TestStatusEventServiceApplyStatusEvent(t *testing.T) {
	service := NewStatusEventService(statusEventRepositoryStub{result: domain.Updated}, nil)
	result, err := service.ApplyStatusEvent(context.Background(), &domain.TransactionSnapshot{}, nil, domain.StatusEvent{})
	if err != nil || result != domain.Updated {
		t.Fatalf("ApplyStatusEvent() = %q, %v", result, err)
	}
}

func TestStatusEventServiceReturnsRepositoryError(t *testing.T) {
	wantErr := errors.New("database unavailable")
	service := NewStatusEventService(statusEventRepositoryStub{err: wantErr}, nil)
	_, err := service.ApplyStatusEvent(context.Background(), &domain.TransactionSnapshot{}, nil, domain.StatusEvent{})
	if !errors.Is(err, wantErr) {
		t.Fatalf("ApplyStatusEvent() error = %v, want %v", err, wantErr)
	}
}
