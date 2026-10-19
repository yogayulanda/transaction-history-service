package domain

import "context"

type TransactionRepository interface {
	ApplyStatusEvent(ctx context.Context, transaction *Transaction, detail *Detail, ev StatusEvent) (ApplyResult, error)
	FindByOwner(ctx context.Context, owner Owner, transactionID string) (*Transaction, *Detail, error)
	ListByOwner(ctx context.Context, owner Owner, filter TransactionFilter) ([]Transaction, bool, error)
	Search(ctx context.Context, criteria SearchCriteria) ([]Transaction, bool, error)
	ListStatusHistory(ctx context.Context, transactionID string) ([]StatusEvent, error)
	InsertBackfill(ctx context.Context, transaction *Transaction, detail *Detail, ev StatusEvent) (bool, error)
}

type MessageRepository interface {
	ListActiveErrorDefinitions(ctx context.Context) ([]ErrorDefinition, error)
}
