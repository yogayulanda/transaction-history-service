package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/yogayulanda/go-core/cache"
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
	normalized, err := sanitizeCreateInput(in)
	if err != nil {
		return "", err
	}

	id, err := s.repo.Create(ctx, normalized)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return "", ErrDuplicateReferenceID
		}
		return "", err
	}

	return id, nil
}

func (s *TransactionService) GetTransactionHistoryDetail(
	ctx context.Context,
	id string,
) (*domain.TransactionHistoryDetail, error) {
	return s.repo.FindDetailByID(ctx, id)
}

func (s *TransactionService) GetUserHistory(
	ctx context.Context,
	filter domain.ListUserHistoryFilter,
) ([]domain.TransactionHistory, bool, error) {
	return s.repo.ListByUser(ctx, filter)
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
		return fmt.Errorf("metadata_json must be valid JSON object")
	}

	if _, ok := decoded.(map[string]any); !ok {
		return fmt.Errorf("metadata_json must be valid JSON object")
	}

	return nil
}

func invalidInput(msg string) error {
	return fmt.Errorf("%w: %s", ErrInvalidTransactionHistoryInput, msg)
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
