package service

import "errors"

var (
	ErrInvalidTransactionHistoryInput = errors.New("invalid transaction history input")
	ErrDuplicateReferenceID           = errors.New("reference_id already exists")
)
