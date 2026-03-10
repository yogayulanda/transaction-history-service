package domain

import "errors"

var (
	ErrTransactionNotFound = errors.New("transaction not found")
	ErrInvalidStatus       = errors.New("invalid status transition")
)
