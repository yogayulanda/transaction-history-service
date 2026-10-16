package service

import "github.com/yogayulanda/go-core/errors"

// NewInvalidInputError creates a validation error with field-level details.
func NewInvalidInputError(msg string, details ...errors.Detail) *errors.AppError {
	return errors.Validation(msg, details...)
}

func NewNotFoundError(msg string) *errors.AppError {
	return errors.New("CORE-1004", msg)
}

func NewInternalError(msg string, err error) *errors.AppError {
	return errors.Wrap("CORE-9000", msg, err)
}
