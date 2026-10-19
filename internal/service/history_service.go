package service

import (
	"context"
	"errors"

	coreerrors "github.com/yogayulanda/go-core/errors"
	"github.com/yogayulanda/transaction-history-service/internal/domain"
)

type historyReadRepository interface {
	FindByOwner(ctx context.Context, owner domain.Owner, transactionID string) (*domain.Transaction, *domain.Detail, error)
	ListByOwner(ctx context.Context, owner domain.Owner, filter domain.TransactionFilter) ([]domain.Transaction, bool, error)
	Search(ctx context.Context, criteria domain.SearchCriteria) ([]domain.Transaction, bool, error)
}

type statusHistoryRepository interface {
	ListStatusHistory(ctx context.Context, transactionID string) ([]domain.StatusEvent, error)
}

type HistoryService struct {
	repo          historyReadRepository
	statusHistory statusHistoryRepository
}

func NewHistoryService(repo historyReadRepository, statusHistory statusHistoryRepository) *HistoryService {
	return &HistoryService{repo: repo, statusHistory: statusHistory}
}

func (s *HistoryService) ListTransactions(
	ctx context.Context,
	owner domain.Owner,
	filter domain.TransactionFilter,
) ([]domain.Transaction, bool, error) {
	items, hasMore, err := s.repo.ListByOwner(ctx, owner, filter)
	if err != nil {
		return nil, false, coreerrors.Wrap("CORE-9000", "failed to list transaction history", err)
	}
	return items, hasMore, nil
}

func (s *HistoryService) GetTransaction(
	ctx context.Context,
	owner domain.Owner,
	transactionID string,
) (*domain.Transaction, *domain.Detail, []domain.StatusEvent, error) {
	header, detailValue, err := s.repo.FindByOwner(ctx, owner, transactionID)
	if errors.Is(err, domain.ErrTransactionNotFound) {
		return nil, nil, nil, coreerrors.New("CORE-1004", "transaction history not found")
	}
	if err != nil {
		return nil, nil, nil, coreerrors.Wrap("CORE-9000", "failed to get transaction history", err)
	}
	events, err := s.statusHistory.ListStatusHistory(ctx, transactionID)
	if err != nil {
		return nil, nil, nil, coreerrors.Wrap("CORE-9000", "failed to get transaction status history", err)
	}
	return header, detailValue, events, nil
}

func (s *HistoryService) SearchTransactions(
	ctx context.Context,
	criteria domain.SearchCriteria,
) ([]domain.Transaction, bool, error) {
	items, hasMore, err := s.repo.Search(ctx, criteria)
	if err != nil {
		return nil, false, coreerrors.Wrap("CORE-9000", "failed to search transaction history", err)
	}
	return items, hasMore, nil
}
