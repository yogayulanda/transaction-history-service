package service

import "github.com/yogayulanda/go-core/errors"

// NewInvalidInputError creates a validation error with field-level details.
func NewInvalidInputError(msg string, details ...errors.Detail) *errors.AppError {
	return errors.Validation(msg, details...)
}

// NewDuplicateReferenceIDError creates a duplicate reference error.
func NewDuplicateReferenceIDError() *errors.AppError {
	return errors.New(CodeDuplicateReferenceID, "reference_id already exists")
}

// NewNotFoundError creates a not-found error.
func NewNotFoundError(msg string) *errors.AppError {
	return errors.New(CodeTransactionNotFound, msg)
}

// NewInternalError wraps an internal error with go-core error contract.
func NewInternalError(msg string, err error) *errors.AppError {
	return errors.Wrap(CodeInternal, msg, err)
}
