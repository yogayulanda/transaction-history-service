package service

import (
	"context"
	"encoding/json"
	"strings"
	"time"
	"unicode/utf16"

	"github.com/google/uuid"
	"github.com/yogayulanda/go-core/logger"
	"github.com/yogayulanda/transaction-history-service/internal/domain"
	"github.com/yogayulanda/transaction-history-service/internal/utils"
)

type statusEventRepository interface {
	ApplyStatusEvent(ctx context.Context, transaction *domain.Transaction, detail *domain.Detail, ev domain.StatusEvent) (domain.ApplyResult, error)
}

type StatusEventService struct {
	repo    statusEventRepository
	catalog domain.Catalog
	hasher  *utils.Hasher
	log     logger.Logger
}

func NewStatusEventService(repo statusEventRepository, validationCatalog domain.Catalog, hasher *utils.Hasher, log logger.Logger) *StatusEventService {
	return &StatusEventService{repo: repo, catalog: validationCatalog, hasher: hasher, log: log}
}

func (s *StatusEventService) ApplyStatusEvent(
	ctx context.Context,
	input domain.StatusEventInput,
) (domain.ApplyResult, error) {
	startedAt := time.Now()
	transaction, detail, event, err := s.PrepareStatusEvent(input)
	if err != nil {
		s.emit(ctx, "invalid", "validation_failed", startedAt)
		return "", err
	}
	result, err := s.repo.ApplyStatusEvent(ctx, transaction, detail, event)
	if err != nil {
		s.emit(ctx, "failed", "apply_failed", startedAt)
		return "", err
	}
	s.emit(ctx, string(result), "", startedAt)
	return result, nil
}

func (s *StatusEventService) emit(ctx context.Context, status, errorCode string, startedAt time.Time) {
	if s.log == nil {
		return
	}
	s.log.LogService(ctx, logger.ServiceLog{
		Operation:  "status_event_apply",
		Status:     status,
		DurationMs: time.Since(startedAt).Milliseconds(),
		ErrorCode:  errorCode,
	})
}

func (s *StatusEventService) PrepareStatusEvent(input domain.StatusEventInput) (*domain.Transaction, *domain.Detail, domain.StatusEvent, error) {
	eventID, err := uuid.Parse(input.EventID)
	if err != nil || input.EventType == "" || len(input.EventType) > 128 || input.Subject == "" || input.EventTime.IsZero() {
		return nil, nil, domain.StatusEvent{}, invalidEvent("invalid_cloudevents")
	}
	if input.StatusVersion < 1 {
		return nil, nil, domain.StatusEvent{}, invalidEvent("invalid_status_version")
	}
	input.TransactionID = strings.TrimSpace(input.TransactionID)
	if input.TransactionID == "" || input.Subject != input.TransactionID {
		return nil, nil, domain.StatusEvent{}, invalidEvent("invalid_transaction_id")
	}
	if !required(input.AppName, input.UserID, input.ProductCategory, input.ProductType, input.Provider,
		input.Direction, input.Amount, input.FeeAmount, input.TotalAmount, input.Status, input.SourceSystem,
		input.TransactionAt, input.ChangedByType) {
		return nil, nil, domain.StatusEvent{}, invalidEvent("required_field_missing")
	}
	if !s.catalog.ValidApp(input.AppName) {
		return nil, nil, domain.StatusEvent{}, invalidEvent("invalid_app")
	}
	category, ok := s.catalog.ProductCategory(input.ProductType)
	if !ok || category != input.ProductCategory || !s.catalog.ValidProvider(input.ProductType, input.Provider) {
		return nil, nil, domain.StatusEvent{}, invalidEvent("invalid_product")
	}
	if input.Partner != "" && !s.catalog.ValidPartner(input.Partner) {
		return nil, nil, domain.StatusEvent{}, invalidEvent("invalid_partner")
	}
	if !s.catalog.AllowedSource(input.SourceSystem) {
		return nil, nil, domain.StatusEvent{}, invalidEvent("source_not_allowed")
	}
	if input.Direction != "debit" && input.Direction != "credit" {
		return nil, nil, domain.StatusEvent{}, invalidEvent("invalid_direction")
	}
	if !domain.IsValidStatus(input.Status) || (input.FromStatus != "" && !domain.IsValidStatus(input.FromStatus)) {
		return nil, nil, domain.StatusEvent{}, invalidEvent("invalid_status")
	}
	if input.ChangedByType != domain.ChangedByUser && input.ChangedByType != domain.ChangedBySystem && input.ChangedByType != domain.ChangedByOperator {
		return nil, nil, domain.StatusEvent{}, invalidEvent("invalid_changed_by_type")
	}
	if !domain.ValidAmount(input.Amount) || !domain.ValidAmount(input.FeeAmount) || !domain.ValidAmount(input.TotalAmount) {
		return nil, nil, domain.StatusEvent{}, invalidEvent("invalid_amount")
	}
	if input.Denomination != "" && !domain.ValidAmount(input.Denomination) {
		return nil, nil, domain.StatusEvent{}, invalidEvent("invalid_amount")
	}
	if input.Currency == "" {
		input.Currency = "IDR"
	}
	if !domain.ValidCurrency(input.Currency) || !input.WithinLimits() {
		return nil, nil, domain.StatusEvent{}, invalidEvent("invalid_field_length")
	}
	transactionAt, err := time.Parse(time.RFC3339Nano, input.TransactionAt)
	if err != nil {
		return nil, nil, domain.StatusEvent{}, invalidEvent("invalid_transaction_at")
	}

	destinationMasked, destinationHash := "", ""
	if input.DestinationNo != "" {
		if input.DestinationNoType == "" {
			return nil, nil, domain.StatusEvent{}, invalidEvent("destination_no_type_required")
		}
		normalized, err := normalizeDestination(input.DestinationNoType, input.DestinationNo)
		if err != nil {
			return nil, nil, domain.StatusEvent{}, invalidEvent("invalid_destination_no")
		}
		destinationMasked = utils.MaskNumber(input.DestinationNoType, normalized)
		destinationHash = s.hasher.Hash(input.DestinationNoType, normalized)
	} else if input.DestinationNoType != "" {
		return nil, nil, domain.StatusEvent{}, invalidEvent("destination_no_without_number")
	}
	sourceMasked := ""
	if input.SourceAccountNo != "" {
		normalized, err := utils.NormalizeDigits(input.SourceAccountNo)
		if err != nil {
			return nil, nil, domain.StatusEvent{}, invalidEvent("invalid_source_account_no")
		}
		sourceMasked = utils.MaskNumber("account", normalized)
	}

	schemaVersion, detailFields, err := domain.NormalizeDetail(input.ProductType, input.Detail)
	if err != nil {
		return nil, nil, domain.StatusEvent{}, invalidEvent("invalid_detail")
	}
	if code, ok := detailFields["code"].(string); ok {
		detailFields["code_masked"] = utils.MaskNumber("cico", code)
		delete(detailFields, "code")
	}
	if err := validateDetailSize(detailFields); err != nil {
		return nil, nil, domain.StatusEvent{}, err
	}

	transaction := &domain.Transaction{
		TransactionID:       input.TransactionID,
		AppName:             input.AppName,
		UserID:              input.UserID,
		ReferenceNo:         input.ReferenceNo,
		ProductCategory:     input.ProductCategory,
		ProductType:         input.ProductType,
		Provider:            strings.ToLower(input.Provider),
		Partner:             input.Partner,
		Direction:           input.Direction,
		Denomination:        input.Denomination,
		Amount:              input.Amount,
		FeeAmount:           input.FeeAmount,
		TotalAmount:         input.TotalAmount,
		Currency:            input.Currency,
		Status:              input.Status,
		StatusVersion:       input.StatusVersion,
		ResponseCode:        input.ResponseCode,
		SourceAccountMasked: sourceMasked,
		DestinationName:     input.DestinationName,
		DestinationNoType:   input.DestinationNoType,
		DestinationNoMasked: destinationMasked,
		DestinationNoHash:   destinationHash,
		SourceSystem:        input.SourceSystem,
		TransactionAt:       transactionAt,
	}
	if domain.IsFinalStatus(input.Status) {
		completedAt := input.EventTime
		transaction.CompletedAt = &completedAt
	}
	var detail *domain.Detail
	if len(detailFields) > 0 {
		detail = &domain.Detail{SchemaVersion: schemaVersion, Fields: detailFields}
	}
	event := domain.StatusEvent{
		EventID:       eventID.String(),
		EventType:     input.EventType,
		StatusVersion: input.StatusVersion,
		FromStatus:    input.FromStatus,
		ToStatus:      input.Status,
		ResponseCode:  input.ResponseCode,
		ChangedByType: input.ChangedByType,
		ChangedAt:     input.EventTime,
	}
	return transaction, detail, event, nil
}

func invalidEvent(reason string) error {
	return &domain.EventValidationError{Reason: reason}
}

func validateDetailSize(fields map[string]any) error {
	if len(fields) == 0 {
		return nil
	}
	encoded, err := json.Marshal(fields)
	if err != nil || len(utf16.Encode([]rune(string(encoded)))) > 4000 {
		return invalidEvent("invalid_detail")
	}
	return nil
}

func required(values ...string) bool {
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			return false
		}
	}
	return true
}

func normalizeDestination(numberType, value string) (string, error) {
	switch numberType {
	case "phone":
		return utils.NormalizePhone(value)
	case "account", "meter", "va":
		return utils.NormalizeDigits(value)
	default:
		return "", utils.ErrInvalidNumber
	}
}
