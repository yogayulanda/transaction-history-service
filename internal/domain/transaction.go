package domain

import "time"

type TransactionHistory struct {
	ID                  string
	AppName             string
	UserID              string
	ReferenceNo         string
	ProductCategory     string
	ProductType         string
	Provider            string
	Partner             string
	Direction           string
	Denomination        string
	Amount              string
	FeeAmount           string
	TotalAmount         string
	Currency            string
	Status              string
	StatusVersion       int64
	ResponseCode        string
	SourceAccountMasked string
	DestinationName     string
	DestinationNoType   string
	DestinationNoMasked string
	DestinationNoHash   string
	SourceSystem        string
	TransactionAt       time.Time
	CompletedAt         *time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type TransactionSnapshot = TransactionHistory
