package service

import (
	coreerrors "github.com/yogayulanda/go-core/errors"
)

// NewInvalidInputError creates a validation error with field-level details.
func NewInvalidInputError(msg string, details ...coreerrors.Detail) *coreerrors.AppError {
	return coreerrors.Validation(msg, details...)
}

// NewDuplicateReferenceIDError creates a duplicate reference error.
func NewDuplicateReferenceIDError() *coreerrors.AppError {
	return coreerrors.New(coreerrors.CodeInvalidRequest, "reference_id already exists")
}

// NewNotFoundError creates a not-found error.
func NewNotFoundError(msg string) *coreerrors.AppError {
	return coreerrors.New(coreerrors.CodeNotFound, msg)
}

// NewInternalError wraps an internal error with go-core error contract.
func NewInternalError(msg string, err error) *coreerrors.AppError {
	return coreerrors.Wrap(coreerrors.CodeInternal, msg, err)
}
