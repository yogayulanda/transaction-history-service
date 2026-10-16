package domain

import "errors"

var (
	ErrTransactionNotFound = errors.New("transaction not found")
	ErrInvalidDetail       = errors.New("invalid transaction detail")
	ErrInvalidEvent        = errors.New("invalid status event")
)

type EventValidationError struct {
	Reason string
}

func (e *EventValidationError) Error() string {
	return e.Reason
}

func (e *EventValidationError) Unwrap() error {
	return ErrInvalidEvent
}
