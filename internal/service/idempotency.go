package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/yogayulanda/transaction-history-service/internal/domain"
	"gorm.io/gorm"
)

var ErrConflictingDuplicateReference = errors.New("conflicting duplicate reference_id")

func (s *TransactionService) CreateTransactionHistoryIdempotent(
	ctx context.Context,
	in domain.CreateTransactionHistoryInput,
) (string, bool, error) {
	normalized, err := s.sanitizeCreateInput(in)
	if err != nil {
		return "", false, err
	}

	id, err := s.repo.Create(ctx, normalized)
	if err == nil {
		return id, false, nil
	}
	if !errors.Is(err, gorm.ErrDuplicatedKey) {
		return "", false, NewInternalError("failed to create transaction history", err, s.errorResolver)
	}

	existing, lookupErr := s.repo.FindByReferenceID(ctx, normalized.ReferenceID)
	if lookupErr != nil {
		return "", false, NewInternalError("failed to resolve duplicate transaction history", lookupErr, s.errorResolver)
	}
	if existing == nil {
		return "", false, NewInternalError("failed to resolve duplicate transaction history", domain.ErrTransactionNotFound, s.errorResolver)
	}
	if !equivalentTransactionHistory(normalized, *existing) {
		return "", false, ErrConflictingDuplicateReference
	}

	return existing.ID, true, nil
}

func equivalentTransactionHistory(in domain.CreateTransactionHistoryInput, existing domain.TransactionHistoryDetail) bool {
	return in.UserID == existing.UserID &&
		in.ReferenceID == existing.ReferenceID &&
		in.ExternalRefID == existing.ExternalRefID &&
		in.ProductGroup == existing.ProductGroup &&
		in.ProductType == existing.ProductType &&
		in.TransactionRoute == existing.TransactionRoute &&
		in.Channel == existing.Channel &&
		in.Direction == existing.Direction &&
		in.Amount == existing.Amount &&
		in.Fee == existing.Fee &&
		in.TotalAmount == existing.TotalAmount &&
		in.Currency == existing.Currency &&
		in.StatusCode == existing.StatusCode &&
		in.ErrorCode == existing.ErrorCode &&
		in.ErrorMessage == existing.ErrorMessage &&
		in.SourceService == existing.SourceService &&
		equivalentTransactionTime(in.TransactionTime, existing.TransactionTime) &&
		equivalentJSON(in.MetadataJSON, existing.MetadataJSON)
}

func equivalentTransactionTime(in, existing time.Time) bool {
	if in.IsZero() {
		return true
	}
	return in.UTC().Equal(existing.UTC())
}

func equivalentJSON(left, right string) bool {
	left = strings.TrimSpace(left)
	right = strings.TrimSpace(right)
	if left == "" {
		left = "{}"
	}
	if right == "" {
		right = "{}"
	}

	var leftDecoded any
	var rightDecoded any
	if err := json.Unmarshal([]byte(left), &leftDecoded); err != nil {
		return left == right
	}
	if err := json.Unmarshal([]byte(right), &rightDecoded); err != nil {
		return left == right
	}
	return jsonEqual(leftDecoded, rightDecoded)
}

func jsonEqual(left, right any) bool {
	leftBytes, leftErr := json.Marshal(left)
	rightBytes, rightErr := json.Marshal(right)
	if leftErr != nil || rightErr != nil {
		return false
	}
	return string(leftBytes) == string(rightBytes)
}
