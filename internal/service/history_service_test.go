package service

import (
	"context"
	"errors"
	"testing"

	coreerrors "github.com/yogayulanda/go-core/errors"
	"github.com/yogayulanda/transaction-history-service/internal/domain"
)

type historyRepositoryStub struct {
	item      *domain.Transaction
	detail    *domain.Detail
	events    []domain.StatusEvent
	listItems []domain.Transaction
	listMore  bool
	err       error
}

func (s *historyRepositoryStub) FindByOwner(context.Context, domain.Owner, string) (*domain.Transaction, *domain.Detail, error) {
	return s.item, s.detail, s.err
}

func (s *historyRepositoryStub) ListByOwner(context.Context, domain.Owner, domain.TransactionFilter) ([]domain.Transaction, bool, error) {
	return s.listItems, s.listMore, s.err
}

func (s *historyRepositoryStub) Search(context.Context, domain.SearchCriteria) ([]domain.Transaction, bool, error) {
	return s.listItems, s.listMore, s.err
}

func (s *historyRepositoryStub) ListStatusHistory(context.Context, string) ([]domain.StatusEvent, error) {
	return s.events, nil
}

func TestHistoryServiceGetTransactionNotFound(t *testing.T) {
	repo := &historyRepositoryStub{err: domain.ErrTransactionNotFound}
	service := NewHistoryService(repo, repo)
	_, _, _, err := service.GetTransaction(context.Background(), domain.Owner{AppName: "app_a", UserID: "user-1"}, "tx-1")
	var appErr *coreerrors.AppError
	if !errors.As(err, &appErr) || appErr.Code != "CORE-1004" {
		t.Fatalf("GetTransaction() error = %v, want CORE-1004", err)
	}
}

func TestHistoryServiceListTransactions(t *testing.T) {
	item := domain.Transaction{ID: 1, TransactionID: "tx-1"}
	repository := &historyRepositoryStub{listItems: []domain.Transaction{item}, listMore: true}
	service := NewHistoryService(repository, repository)
	items, hasMore, err := service.ListTransactions(context.Background(), domain.Owner{AppName: "app_a", UserID: "user-1"}, domain.TransactionFilter{PageSize: 1})
	if err != nil || !hasMore || len(items) != 1 || items[0].ID != item.ID {
		t.Fatalf("ListTransactions() = %+v, %v, %v", items, hasMore, err)
	}
}
