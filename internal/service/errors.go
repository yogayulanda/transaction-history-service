package service

import (
	"strings"

	coreerrors "github.com/yogayulanda/go-core/errors"
)

const (
	errorDomainTRH = "TRH"

	errNumberValidationInvalidInput         = "001"
	errNumberValidationDuplicateRefID       = "002"
	errNumberDataTransactionHistoryNotFound = "001"
	errNumberTechnicalInternal              = "001"
)

// NewInvalidInputError creates a validation error with field-level details.
func NewInvalidInputError(
	msg string,
	resolver ErrorDefinitionResolver,
	details ...coreerrors.Detail,
) *coreerrors.AppError {
	builder := coreerrors.Build(errorDomainTRH, coreerrors.CategoryVAL, errNumberValidationInvalidInput).
		Code(coreerrors.CodeInvalidRequest).
		Message(msg).
		UserMessage("Permintaan tidak valid, silakan periksa kembali data yang dimasukkan.").
		Finality(coreerrors.FinalityBusiness).
		Retryable(false)
	if len(details) > 0 {
		builder.Details(details...)
	}
	return withDynamicErrorDefinition(builder.Done(), resolver)
}

// NewDuplicateReferenceIDError creates a duplicate reference error.
func NewDuplicateReferenceIDError(resolver ErrorDefinitionResolver) *coreerrors.AppError {
	appErr := coreerrors.Build(errorDomainTRH, coreerrors.CategoryVAL, errNumberValidationDuplicateRefID).
		Code(coreerrors.CodeInvalidRequest).
		Message("reference_id already exists").
		UserMessage("Transaksi dengan referensi ini sudah pernah diproses.").
		Finality(coreerrors.FinalityBusiness).
		Retryable(false).
		Done()
	return withDynamicErrorDefinition(appErr, resolver)
}

// NewNotFoundError creates a not-found error.
func NewNotFoundError(msg string, resolver ErrorDefinitionResolver) *coreerrors.AppError {
	appErr := coreerrors.Build(errorDomainTRH, coreerrors.CategoryDB, errNumberDataTransactionHistoryNotFound).
		Code(coreerrors.CodeNotFound).
		Message(msg).
		UserMessage("Riwayat transaksi tidak ditemukan.").
		Finality(coreerrors.FinalityBusiness).
		Retryable(false).
		Done()
	return withDynamicErrorDefinition(appErr, resolver)
}

// NewInternalError wraps an internal error with go-core error contract.
func NewInternalError(msg string, err error, resolver ErrorDefinitionResolver) *coreerrors.AppError {
	appErr := coreerrors.Build(errorDomainTRH, coreerrors.CategoryREC, errNumberTechnicalInternal).
		Code(coreerrors.CodeInternal).
		Message(msg).
		UserMessage("Sistem sedang sibuk. Silakan coba beberapa saat lagi.").
		Finality(coreerrors.FinalityTechnicalNonRecoverable).
		Retryable(false).
		Err(err).
		Done()
	return withDynamicErrorDefinition(appErr, resolver)
}

func withDynamicErrorDefinition(
	appErr *coreerrors.AppError,
	resolver ErrorDefinitionResolver,
) *coreerrors.AppError {
	if appErr == nil || resolver == nil {
		return appErr
	}

	userMessage, details, ok := resolver.Resolve(appErr.FormatCode())
	if !ok {
		return appErr
	}

	if strings.TrimSpace(userMessage) != "" {
		appErr.UserMessage = userMessage
	}
	if len(details) > 0 {
		appErr.Details = details
	}

	return appErr
}
