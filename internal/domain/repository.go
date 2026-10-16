package domain

import "context"

type TransactionRepository interface {
	ApplyStatusEvent(ctx context.Context, header *TransactionHistory, detail *Detail, ev StatusEvent) (ApplyResult, error)
	FindByOwner(ctx context.Context, owner Owner, transactionID string) (*TransactionHistory, *Detail, error)
	ListByOwner(ctx context.Context, owner Owner, filter TransactionFilter) ([]TransactionHistory, bool, error)
	Search(ctx context.Context, criteria SearchCriteria) ([]TransactionHistory, bool, error)
	InsertBackfill(ctx context.Context, header *TransactionHistory, detail *Detail, ev StatusEvent) (bool, error)
}

type TransactionHistoryRepository = TransactionRepository

type StatusHistoryRepository interface {
	ListStatusHistory(ctx context.Context, transactionID string) ([]StatusEvent, error)
}

type ErrorDefinitionRepository interface {
	ListActiveErrorDefinitions(ctx context.Context) ([]ErrorDefinition, error)
}
