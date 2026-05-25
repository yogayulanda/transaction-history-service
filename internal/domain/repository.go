package domain

import "context"

type TransactionRepository interface {
	Create(ctx context.Context, in CreateTransactionHistoryInput) (string, error)
	FindByReferenceID(ctx context.Context, referenceID string) (*TransactionHistoryDetail, error)
	FindDetailByID(ctx context.Context, id string) (*TransactionHistoryDetail, error)
	ListByUser(ctx context.Context, filter ListUserHistoryFilter) ([]TransactionHistory, bool, error)
}

type ErrorDefinitionRepository interface {
	ListActiveErrorDefinitions(ctx context.Context) ([]ErrorDefinition, error)
}
