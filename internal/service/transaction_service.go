package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"
	"unicode"

	"github.com/yogayulanda/go-core/cache"
	coreerrors "github.com/yogayulanda/go-core/errors"
	"github.com/yogayulanda/go-core/logger"
	"github.com/yogayulanda/go-core/messaging"
	"github.com/yogayulanda/transaction-history-service/internal/domain"
	"gorm.io/gorm"
)

type TransactionService struct {
	repo      domain.TransactionRepository
	cache     cache.Cache
	publisher messaging.Publisher
	log       logger.Logger
}

func NewTransactionService(
	repo domain.TransactionRepository,
	cache cache.Cache,
	publisher messaging.Publisher,
	log logger.Logger,
) *TransactionService {
	return &TransactionService{
		repo:      repo,
		cache:     cache,
		publisher: publisher,
		log:       log,
	}
}

func (s *TransactionService) CreateTransactionHistory(
	ctx context.Context,
	in domain.CreateTransactionHistoryInput,
) (string, error) {
	startedAt := time.Now()

	normalized, err := sanitizeCreateInput(in)
	if err != nil {
		s.emitServiceLog(ctx, "create_transaction", "failed", startedAt, "validation_failed")
		return "", err
	}

	id, err := s.repo.Create(ctx, normalized)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			s.emitServiceLog(ctx, "create_transaction", "failed", startedAt, "duplicate_reference")
			s.emitTransactionLog(ctx, "create_transaction", "", normalized.UserID, "failed", startedAt, "DUPLICATE_REFERENCE")
			return "", NewDuplicateReferenceIDError()
		}
		s.emitServiceLog(ctx, "create_transaction", "failed", startedAt, "repository_error")
		s.emitTransactionLog(ctx, "create_transaction", "", normalized.UserID, "failed", startedAt, "INTERNAL_ERROR")
		return "", NewInternalError("failed to create transaction history", err)
	}

	s.emitServiceLog(ctx, "create_transaction", "success", startedAt, "")
	s.emitTransactionLog(ctx, "create_transaction", id, normalized.UserID, "success", startedAt, "")

	return id, nil
}

func (s *TransactionService) GetTransactionHistoryDetail(
	ctx context.Context,
	id string,
) (*domain.TransactionHistoryDetail, error) {
	startedAt := time.Now()

	result, err := s.repo.FindDetailByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrTransactionNotFound) {
			s.emitServiceLog(ctx, "get_transaction_detail", "failed", startedAt, "not_found")
			return nil, NewNotFoundError("transaction history not found")
		}
		s.emitServiceLog(ctx, "get_transaction_detail", "failed", startedAt, "repository_error")
		return nil, NewInternalError("failed to get transaction history detail", err)
	}

	s.emitServiceLog(ctx, "get_transaction_detail", "success", startedAt, "")
	return result, nil
}

func (s *TransactionService) GetUserHistory(
	ctx context.Context,
	filter domain.ListUserHistoryFilter,
) ([]domain.TransactionHistory, bool, error) {
	startedAt := time.Now()

	items, hasMore, err := s.repo.ListByUser(ctx, filter)
	if err != nil {
		s.emitServiceLog(ctx, "get_user_history", "failed", startedAt, "repository_error")
		return nil, false, NewInternalError("failed to get user history", err)
	}

	s.emitServiceLog(ctx, "get_user_history", "success", startedAt, "")
	return items, hasMore, nil
}

func (s *TransactionService) emitServiceLog(ctx context.Context, operation, status string, startedAt time.Time, errorCode string) {
	if s.log == nil {
		return
	}
	s.log.LogService(ctx, logger.ServiceLog{
		Operation:  operation,
		Status:     status,
		DurationMs: time.Since(startedAt).Milliseconds(),
		ErrorCode:  errorCode,
	})
}

func (s *TransactionService) emitTransactionLog(ctx context.Context, operation, txID, userID, status string, startedAt time.Time, errorCode string) {
	if s.log == nil {
		return
	}
	s.log.LogTransaction(ctx, logger.TransactionLog{
		Operation:     operation,
		TransactionID: txID,
		UserID:        userID,
		Status:        status,
		DurationMs:    time.Since(startedAt).Milliseconds(),
		ErrorCode:     errorCode,
	})
}

func sanitizeCreateInput(in domain.CreateTransactionHistoryInput) (domain.CreateTransactionHistoryInput, error) {
	in.UserID = strings.TrimSpace(in.UserID)
	in.ReferenceID = strings.TrimSpace(in.ReferenceID)
	in.ExternalRefID = strings.TrimSpace(in.ExternalRefID)
	in.ProductGroup = strings.TrimSpace(in.ProductGroup)
	in.ProductType = strings.TrimSpace(in.ProductType)
	in.TransactionRoute = strings.TrimSpace(in.TransactionRoute)
	in.Channel = strings.TrimSpace(in.Channel)
	in.Direction = strings.TrimSpace(in.Direction)
	in.Currency = strings.ToUpper(strings.TrimSpace(in.Currency))
	in.StatusCode = strings.ToUpper(strings.TrimSpace(in.StatusCode))
	in.ErrorCode = strings.TrimSpace(in.ErrorCode)
	in.ErrorMessage = strings.TrimSpace(in.ErrorMessage)
	in.SourceService = strings.TrimSpace(in.SourceService)

	if in.UserID == "" {
		return in, invalidInput("user_id is required")
	}
	if in.ReferenceID == "" {
		return in, invalidInput("reference_id is required")
	}
	if in.Channel == "" {
		return in, invalidInput("channel is required")
	}
	if in.ProductGroup == "" {
		return in, invalidInput("product_group is required")
	}
	if in.ProductType == "" {
		return in, invalidInput("product_type is required")
	}
	if in.StatusCode == "" {
		return in, invalidInput("status_code is required")
	}
	if in.SourceService == "" {
		return in, invalidInput("source_service is required")
	}
	if in.Currency == "" {
		return in, invalidInput("currency is required")
	}
	if !isAlphaString(in.Currency, 3) {
		return in, invalidInput("currency must be a 3-letter code")
	}
	if in.Amount < 0 {
		return in, invalidInput("amount must be non-negative")
	}
	if in.Fee < 0 {
		return in, invalidInput("fee must be non-negative")
	}
	if in.TotalAmount < 0 {
		return in, invalidInput("total_amount must be non-negative")
	}

	if strings.TrimSpace(in.MetadataJSON) == "" {
		in.MetadataJSON = "{}"
	} else {
		in.MetadataJSON = strings.TrimSpace(in.MetadataJSON)
		if err := validateJSONObject(in.MetadataJSON); err != nil {
			return in, invalidInput(err.Error())
		}
	}

	return in, nil
}

func validateJSONObject(raw string) error {
	var decoded any
	if err := json.Unmarshal([]byte(raw), &decoded); err != nil {
		return errors.New("metadata_json must be valid JSON object")
	}

	if _, ok := decoded.(map[string]any); !ok {
		return errors.New("metadata_json must be valid JSON object")
	}

	return nil
}

func invalidInput(msg string) *coreerrors.AppError {
	return NewInvalidInputError(msg)
}

func isAlphaString(raw string, expectedLen int) bool {
	if len(raw) != expectedLen {
		return false
	}

	for _, r := range raw {
		if !unicode.IsLetter(r) {
			return false
		}
	}

	return true
}
