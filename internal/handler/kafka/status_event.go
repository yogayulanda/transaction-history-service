package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/yogayulanda/go-core/logger"
	"github.com/yogayulanda/go-core/messaging"
	"github.com/yogayulanda/transaction-history-service/internal/catalog"
	"github.com/yogayulanda/transaction-history-service/internal/detail"
	"github.com/yogayulanda/transaction-history-service/internal/domain"
	"github.com/yogayulanda/transaction-history-service/internal/pii"
)

var decimalPattern = regexp.MustCompile(`^\d{1,16}(\.\d{1,2})?$`)

type statusEventApplier interface {
	ApplyStatusEvent(ctx context.Context, header *domain.TransactionSnapshot, detail *domain.Detail, ev domain.StatusEvent) (domain.ApplyResult, error)
}

type StatusEventHandler struct {
	service statusEventApplier
	catalog catalog.Catalog
	hasher  *pii.Hasher
	dlq     messaging.Publisher
	log     logger.Logger
}

type statusEventPayload struct {
	TransactionID     string         `json:"transaction_id"`
	AppName           string         `json:"app_name"`
	UserID            string         `json:"user_id"`
	ProductCategory   string         `json:"product_category"`
	ProductType       string         `json:"product_type"`
	Provider          string         `json:"provider"`
	Partner           string         `json:"partner"`
	Direction         string         `json:"direction"`
	Denomination      string         `json:"denomination"`
	Amount            string         `json:"amount"`
	FeeAmount         string         `json:"fee_amount"`
	TotalAmount       string         `json:"total_amount"`
	Currency          string         `json:"currency"`
	Status            string         `json:"status"`
	FromStatus        string         `json:"from_status"`
	ResponseCode      string         `json:"response_code"`
	SourceAccountNo   string         `json:"source_account_no"`
	DestinationName   string         `json:"destination_name"`
	DestinationNo     string         `json:"destination_no"`
	DestinationNoType string         `json:"destination_no_type"`
	ReferenceNo       string         `json:"reference_no"`
	SourceSystem      string         `json:"source_system"`
	TransactionAt     string         `json:"transaction_at"`
	ChangedByType     string         `json:"changed_by_type"`
	Detail            map[string]any `json:"detail"`
}

func NewStatusEventHandler(
	svc statusEventApplier,
	validationCatalog catalog.Catalog,
	hasher *pii.Hasher,
	dlq messaging.Publisher,
	log logger.Logger,
) *StatusEventHandler {
	return &StatusEventHandler{service: svc, catalog: validationCatalog, hasher: hasher, dlq: dlq, log: log}
}

func (h *StatusEventHandler) Handle(ctx context.Context, msg messaging.Message) error {
	attrs, ok := messaging.CloudEventsFromContext(ctx)
	if !ok {
		return h.reject(ctx, msg, "missing_cloudevents")
	}
	return h.handleWithEvent(ctx, msg, attrs)
}

func (h *StatusEventHandler) PrepareStatusEvent(
	attrs messaging.CloudEventAttributes,
	payload []byte,
) (*domain.TransactionSnapshot, *domain.Detail, domain.StatusEvent, error) {
	header, detailValue, event, reason := h.decodeStatusEvent(attrs, payload)
	if reason != "" {
		return nil, nil, domain.StatusEvent{}, errors.New(reason)
	}
	return header, detailValue, event, nil
}

func (h *StatusEventHandler) handleWithEvent(ctx context.Context, msg messaging.Message, attrs messaging.CloudEventAttributes) error {
	header, detailValue, event, reason := h.decodeStatusEvent(attrs, msg.Payload)
	if reason != "" {
		return h.reject(ctx, msg, reason)
	}
	result, err := h.service.ApplyStatusEvent(ctx, header, detailValue, event)
	if err != nil {
		h.emit(ctx, "retryable_failed", "database_error", result)
		return err
	}
	h.emit(ctx, "success", "", result)
	return nil
}

func (h *StatusEventHandler) decodeStatusEvent(
	attrs messaging.CloudEventAttributes,
	payload []byte,
) (*domain.TransactionSnapshot, *domain.Detail, domain.StatusEvent, string) {
	eventID, err := uuid.Parse(attrs.ID)
	if err != nil || attrs.Type == "" || len(attrs.Type) > 128 || attrs.Subject == "" || attrs.Time.IsZero() {
		return nil, nil, domain.StatusEvent{}, "invalid_cloudevents"
	}
	version, err := strconv.ParseInt(attrs.Sequence, 10, 64)
	if err != nil || version < 1 {
		return nil, nil, domain.StatusEvent{}, "invalid_status_version"
	}

	var input statusEventPayload
	if err := json.Unmarshal(payload, &input); err != nil {
		return nil, nil, domain.StatusEvent{}, "malformed_json"
	}
	input.TransactionID = strings.TrimSpace(input.TransactionID)
	if input.TransactionID == "" || attrs.Subject != input.TransactionID {
		return nil, nil, domain.StatusEvent{}, "invalid_transaction_id"
	}
	if !required(input.AppName, input.UserID, input.ProductCategory, input.ProductType, input.Provider,
		input.Direction, input.Amount, input.FeeAmount, input.TotalAmount, input.Status, input.SourceSystem,
		input.TransactionAt, input.ChangedByType) {
		return nil, nil, domain.StatusEvent{}, "required_field_missing"
	}
	if !h.catalog.ValidApp(input.AppName) {
		return nil, nil, domain.StatusEvent{}, "invalid_app"
	}
	category, ok := h.catalog.ProductCategory(input.ProductType)
	if !ok || category != input.ProductCategory || !h.catalog.ValidProvider(input.ProductType, input.Provider) {
		return nil, nil, domain.StatusEvent{}, "invalid_product"
	}
	if input.Partner != "" && !h.catalog.ValidPartner(input.Partner) {
		return nil, nil, domain.StatusEvent{}, "invalid_partner"
	}
	if !h.catalog.AllowedSource(input.SourceSystem) {
		return nil, nil, domain.StatusEvent{}, "source_not_allowed"
	}
	if input.Direction != "debit" && input.Direction != "credit" {
		return nil, nil, domain.StatusEvent{}, "invalid_direction"
	}
	if !validStatus(input.Status) || (input.FromStatus != "" && !validStatus(input.FromStatus)) {
		return nil, nil, domain.StatusEvent{}, "invalid_status"
	}
	if input.ChangedByType != "user" && input.ChangedByType != "system" && input.ChangedByType != "operator" {
		return nil, nil, domain.StatusEvent{}, "invalid_changed_by_type"
	}
	if !validAmount(input.Amount) || !validAmount(input.FeeAmount) || !validAmount(input.TotalAmount) {
		return nil, nil, domain.StatusEvent{}, "invalid_amount"
	}
	if input.Currency == "" {
		input.Currency = "IDR"
	}
	if !validCurrency(input.Currency) || !withinLengths(input) {
		return nil, nil, domain.StatusEvent{}, "invalid_field_length"
	}
	transactionAt, err := time.Parse(time.RFC3339Nano, input.TransactionAt)
	if err != nil {
		return nil, nil, domain.StatusEvent{}, "invalid_transaction_at"
	}

	destinationMasked := ""
	destinationHash := ""
	if input.DestinationNo != "" {
		if input.DestinationNoType == "" {
			return nil, nil, domain.StatusEvent{}, "destination_no_type_required"
		}
		normalized, err := normalizeDestination(input.DestinationNoType, input.DestinationNo)
		if err != nil {
			return nil, nil, domain.StatusEvent{}, "invalid_destination_no"
		}
		destinationMasked = pii.MaskNumber(input.DestinationNoType, normalized)
		destinationHash = h.hasher.Hash(input.DestinationNoType, normalized)
	} else if input.DestinationNoType != "" {
		return nil, nil, domain.StatusEvent{}, "destination_no_without_number"
	}
	sourceMasked := ""
	if input.SourceAccountNo != "" {
		normalized, err := pii.NormalizeDigits(input.SourceAccountNo)
		if err != nil {
			return nil, nil, domain.StatusEvent{}, "invalid_source_account_no"
		}
		sourceMasked = pii.MaskNumber("account", normalized)
	}

	schemaVersion, detailFields, err := detail.Normalize(input.ProductType, input.Detail)
	if err != nil {
		return nil, nil, domain.StatusEvent{}, "invalid_detail"
	}
	header := &domain.TransactionSnapshot{
		ID:                  input.TransactionID,
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
		StatusVersion:       version,
		ResponseCode:        input.ResponseCode,
		SourceAccountMasked: sourceMasked,
		DestinationName:     input.DestinationName,
		DestinationNoType:   input.DestinationNoType,
		DestinationNoMasked: destinationMasked,
		DestinationNoHash:   destinationHash,
		SourceSystem:        input.SourceSystem,
		TransactionAt:       transactionAt,
	}
	if isFinalStatus(input.Status) {
		completedAt := attrs.Time
		header.CompletedAt = &completedAt
	}
	var detailValue *domain.Detail
	if len(detailFields) > 0 {
		detailValue = &domain.Detail{SchemaVersion: schemaVersion, Fields: detailFields}
	}
	event := domain.StatusEvent{
		EventID:       eventID.String(),
		EventType:     attrs.Type,
		StatusVersion: version,
		FromStatus:    input.FromStatus,
		ToStatus:      input.Status,
		ResponseCode:  input.ResponseCode,
		ChangedByType: input.ChangedByType,
		ChangedAt:     attrs.Time,
	}
	return header, detailValue, event, ""
}

func (h *StatusEventHandler) reject(ctx context.Context, msg messaging.Message, reason string) error {
	if h.dlq == nil {
		h.emit(ctx, "dlq_failed", reason, "")
		return errors.New("status event DLQ publisher is unavailable")
	}
	message := messaging.Message{
		Topic:   msg.Topic + ".dlq",
		Key:     msg.Key,
		Payload: msg.Payload,
		Headers: map[string]string{
			"x-original-topic": msg.Topic,
			"x-dlq-reason":     reason,
		},
	}
	if err := h.dlq.Publish(ctx, message); err != nil {
		h.emit(ctx, "dlq_failed", "dlq_publish_failed", "")
		return fmt.Errorf("publish invalid status event to DLQ: %w", err)
	}
	h.emit(ctx, "dlq_success", reason, "")
	return nil
}

func (h *StatusEventHandler) emit(ctx context.Context, status, reason string, result domain.ApplyResult) {
	if h.log == nil {
		return
	}
	h.log.LogService(ctx, logger.ServiceLog{
		Operation: "kafka_status_event",
		Status:    status,
		ErrorCode: reason,
		Metadata: map[string]interface{}{
			"result": result,
		},
	})
}

func normalizeDestination(numberType, value string) (string, error) {
	switch numberType {
	case "phone":
		return pii.NormalizePhone(value)
	case "account", "meter", "va":
		return pii.NormalizeDigits(value)
	default:
		return "", pii.ErrInvalidNumber
	}
}

func required(values ...string) bool {
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			return false
		}
	}
	return true
}

func validAmount(value string) bool {
	return decimalPattern.MatchString(value)
}

func validCurrency(value string) bool {
	if len(value) != 3 {
		return false
	}
	for _, char := range value {
		if !unicode.IsLetter(char) || !unicode.IsUpper(char) {
			return false
		}
	}
	return true
}

func validStatus(value string) bool {
	switch value {
	case "PENDING", "UNKNOWN", "SUCCESS", "FAILED", "NEED_REVIEW", "REFUNDED":
		return true
	default:
		return false
	}
}

func isFinalStatus(value string) bool {
	return value == "SUCCESS" || value == "FAILED" || value == "REFUNDED"
}

func withinLengths(input statusEventPayload) bool {
	return len(input.TransactionID) <= 64 && len(input.AppName) <= 64 && len(input.UserID) <= 64 &&
		len(input.ReferenceNo) <= 128 && len(input.ProductCategory) <= 16 && len(input.ProductType) <= 64 &&
		len(input.Provider) <= 64 && len(input.Partner) <= 64 && len(input.Direction) <= 16 &&
		len(input.Denomination) <= 64 && len(input.Currency) == 3 && len(input.Status) <= 32 &&
		len(input.ResponseCode) <= 64 && len(input.DestinationName) <= 128 &&
		len(input.DestinationNoType) <= 16 && len(input.SourceSystem) <= 64 && len(input.ChangedByType) <= 16
}
