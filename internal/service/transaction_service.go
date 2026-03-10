package service

import (
	"context"

	"github.com/yogayulanda/go-core/cache"
	"github.com/yogayulanda/go-core/logger"
	"github.com/yogayulanda/go-core/messaging"
	"github.com/yogayulanda/transaction-history-service/internal/domain"
)

type TransactionService struct {
	repo      domain.TransactionRepository
	cache     cache.Cache
	publisher messaging.Publisher
	log       logger.Logger
}

func NewTransactionService(
	repo domain.TransactionRepository,
	cache cache.Cache,
	publisher messaging.Publisher,
	log logger.Logger,
) *TransactionService {
	return &TransactionService{
		repo:      repo,
		cache:     cache,
		publisher: publisher,
		log:       log,
	}
}

func (s *TransactionService) CreateTransactionHistory(
	ctx context.Context,
	in domain.CreateTransactionHistoryInput,
) (string, error) {
	return s.repo.Create(ctx, in)
}

func (s *TransactionService) GetTransactionHistoryDetail(
	ctx context.Context,
	id string,
) (*domain.TransactionHistoryDetail, error) {
	return s.repo.FindDetailByID(ctx, id)
}

func (s *TransactionService) GetUserHistory(
	ctx context.Context,
	filter domain.ListUserHistoryFilter,
) ([]domain.TransactionHistory, bool, error) {
	return s.repo.ListByUser(ctx, filter)
}
