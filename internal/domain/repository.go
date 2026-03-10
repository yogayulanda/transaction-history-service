package domain

import "context"

type TransactionRepository interface {
	Create(ctx context.Context, in CreateTransactionHistoryInput) (string, error)
	FindDetailByID(ctx context.Context, id string) (*TransactionHistoryDetail, error)
	ListByUser(ctx context.Context, filter ListUserHistoryFilter) ([]TransactionHistory, bool, error)
}
