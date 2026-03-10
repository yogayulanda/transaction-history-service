package domain

import "time"

type TransactionHistory struct {
	ID               string
	UserID           string
	ReferenceID      string
	ExternalRefID    string
	ProductGroup     string
	ProductType      string
	TransactionRoute string
	Channel          string
	Direction        string
	Amount           int64
	Fee              int64
	TotalAmount      int64
	Currency         string
	StatusCode       string
	ErrorCode        string
	ErrorMessage     string
	SourceService    string
	TransactionTime  time.Time
}

type TransactionHistoryDetail struct {
	TransactionHistory
	MetadataJSON string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type CreateTransactionHistoryInput struct {
	UserID           string
	ReferenceID      string
	ExternalRefID    string
	ProductGroup     string
	ProductType      string
	TransactionRoute string
	Channel          string
	Direction        string
	Amount           int64
	Fee              int64
	TotalAmount      int64
	Currency         string
	StatusCode       string
	ErrorCode        string
	ErrorMessage     string
	SourceService    string
	TransactionTime  time.Time
	MetadataJSON     string
}

type ListUserHistoryFilter struct {
	UserID           string
	StartDate        *time.Time
	EndDate          *time.Time
	ProductGroup     string
	ProductType      string
	TransactionRoute string
	StatusCode       string
	PageSize         int
	Offset           int
}
