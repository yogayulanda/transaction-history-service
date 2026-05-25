package kafka

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	coreerrors "github.com/yogayulanda/go-core/errors"
	"github.com/yogayulanda/go-core/logger"
	"github.com/yogayulanda/go-core/messaging"
	"github.com/yogayulanda/transaction-history-service/internal/domain"
	"github.com/yogayulanda/transaction-history-service/internal/service"
)

const (
	TransactionCreatedTopic      = "transaction-history.transaction.created"
	TransactionCreatedDLQTopic   = "transaction-history.transaction.created.dlq"
	TransactionCreatedGroupID    = "transaction-history-service"
	TransactionCreatedEventType  = "transaction.created"
	TransactionCreatedVersion    = 1
	TransactionCreatedRetryMax   = 5
	TransactionCreatedRetryDelay = 5 * time.Second
	TransactionCreatedConcurrency = 3
)

var defaultProducerAllowlist = map[string]struct{}{
	"ms-agent-payment-purchase": {},
	"trxFinance":                 {},
	"ms-liquiditas":              {},
}

type TransactionCreator interface {
	CreateTransactionHistoryIdempotent(ctx context.Context, in domain.CreateTransactionHistoryInput) (string, bool, error)
}

type TransactionCreatedHandler struct {
	service   TransactionCreator
	dlq       messaging.Publisher
	log       logger.Logger
	allowlist map[string]struct{}
}

func NewTransactionCreatedHandler(
	svc TransactionCreator,
	dlq messaging.Publisher,
	log logger.Logger,
) *TransactionCreatedHandler {
	return &TransactionCreatedHandler{
		service:   svc,
		dlq:       dlq,
		log:       log,
		allowlist: defaultProducerAllowlist,
	}
}

func (h *TransactionCreatedHandler) Handle(ctx context.Context, msg messaging.Message) error {
	evt, input, err := h.decode(msg.Payload)
	if err != nil {
		return h.handleNonRetryable(ctx, msg, evt, "", classifyError(err))
	}

	id, noOp, err := h.service.CreateTransactionHistoryIdempotent(ctx, input)
	if err == nil {
		status := "success"
		if noOp {
			status = "duplicate_noop"
		}
		h.emitLog(ctx, evt, input.ReferenceID, status, "", "not_required")
		_ = id
		return nil
	}

	if IsNonRetryable(err) {
		return h.handleNonRetryable(ctx, msg, evt, input.ReferenceID, classifyError(err))
	}

	h.emitLog(ctx, evt, input.ReferenceID, "retryable_failed", classifyError(err), "not_attempted")
	return err
}

func (h *TransactionCreatedHandler) decode(payload []byte) (transactionCreatedEvent, domain.CreateTransactionHistoryInput, error) {
	var evt transactionCreatedEvent
	if err := json.Unmarshal(payload, &evt); err != nil {
		return evt, domain.CreateTransactionHistoryInput{}, nonRetryableError("malformed_json")
	}

	evt.EventID = strings.TrimSpace(evt.EventID)
	evt.EventType = strings.TrimSpace(evt.EventType)
	evt.Producer = strings.TrimSpace(evt.Producer)
	if evt.EventType != TransactionCreatedEventType {
		return evt, domain.CreateTransactionHistoryInput{}, nonRetryableError("unsupported_event_type")
	}
	if evt.EventVersion != TransactionCreatedVersion {
		return evt, domain.CreateTransactionHistoryInput{}, nonRetryableError("invalid_schema_version")
	}
	if _, ok := h.allowlist[evt.Producer]; !ok {
		return evt, domain.CreateTransactionHistoryInput{}, nonRetryableError("producer_not_allowed")
	}
	if len(evt.Payload) == 0 {
		return evt, domain.CreateTransactionHistoryInput{}, nonRetryableError("payload_required")
	}

	var p transactionCreatedPayload
	if err := json.Unmarshal(evt.Payload, &p); err != nil {
		return evt, domain.CreateTransactionHistoryInput{}, nonRetryableError("malformed_payload")
	}

	txTime, err := parseOptionalTime(p.TransactionTime)
	if err != nil {
		return evt, domain.CreateTransactionHistoryInput{}, nonRetryableError("invalid_transaction_time")
	}
	metadata, err := normalizeMetadataJSON(p.MetadataJSON)
	if err != nil {
		return evt, domain.CreateTransactionHistoryInput{}, nonRetryableError("invalid_metadata_json")
	}

	input := domain.CreateTransactionHistoryInput{
		UserID:           strings.TrimSpace(p.UserID),
		ReferenceID:      strings.TrimSpace(p.ReferenceID),
		ExternalRefID:    strings.TrimSpace(p.ExternalRefID),
		ProductGroup:     strings.TrimSpace(p.ProductGroup),
		ProductType:      strings.TrimSpace(p.ProductType),
		TransactionRoute: strings.TrimSpace(p.TransactionRoute),
		Channel:          strings.TrimSpace(p.Channel),
		Direction:        strings.TrimSpace(p.Direction),
		Amount:           p.Amount,
		Fee:              p.Fee,
		TotalAmount:      p.TotalAmount,
		Currency:         strings.TrimSpace(p.Currency),
		StatusCode:       strings.TrimSpace(p.StatusCode),
		ErrorCode:        strings.TrimSpace(p.ErrorCode),
		ErrorMessage:     strings.TrimSpace(p.ErrorMessage),
		SourceService:    strings.TrimSpace(p.SourceService),
		TransactionTime:  txTime,
		MetadataJSON:     metadata,
	}
	if err := validateRequiredInput(input); err != nil {
		return evt, input, err
	}
	return evt, input, nil
}

func (h *TransactionCreatedHandler) handleNonRetryable(
	ctx context.Context,
	msg messaging.Message,
	evt transactionCreatedEvent,
	referenceID string,
	errorCode string,
) error {
	if h.dlq == nil {
		h.emitLog(ctx, evt, referenceID, "dlq_failed", errorCode, "missing_publisher")
		return retryableError("dlq_publish_failed")
	}

	dlqMsg := messaging.Message{
		Topic:   TransactionCreatedDLQTopic,
		Key:     msg.Key,
		Payload: msg.Payload,
		Headers: map[string]string{},
	}
	dlqMsg.Headers["x-original-topic"] = msg.Topic
	dlqMsg.Headers["x-error-code"] = errorCode

	if err := h.dlq.Publish(ctx, dlqMsg); err != nil {
		h.emitLog(ctx, evt, referenceID, "dlq_failed", errorCode, "failed")
		return retryableError("dlq_publish_failed")
	}

	h.emitLog(ctx, evt, referenceID, "dlq_success", errorCode, "success")
	return nil
}

func (h *TransactionCreatedHandler) emitLog(
	ctx context.Context,
	evt transactionCreatedEvent,
	referenceID string,
	status string,
	errorCode string,
	dlqStatus string,
) {
	if h.log == nil {
		return
	}
	h.log.LogService(ctx, logger.ServiceLog{
		Operation: "kafka_transaction_created",
		Status:    status,
		ErrorCode: errorCode,
		Metadata: map[string]interface{}{
			"reference_id":  referenceID,
			"event_id":      evt.EventID,
			"event_type":    evt.EventType,
			"event_version": evt.EventVersion,
			"status":        status,
			"retry_count":   0,
			"dlq_status":    dlqStatus,
		},
	})
}

type transactionCreatedEvent struct {
	EventID      string          `json:"event_id"`
	EventType    string          `json:"event_type"`
	EventVersion int             `json:"event_version"`
	Producer     string          `json:"producer"`
	Payload      json.RawMessage `json:"payload"`
}

type transactionCreatedPayload struct {
	UserID           string          `json:"user_id"`
	ReferenceID      string          `json:"reference_id"`
	ExternalRefID    string          `json:"external_ref_id"`
	ProductGroup     string          `json:"product_group"`
	ProductType      string          `json:"product_type"`
	TransactionRoute string          `json:"transaction_route"`
	Channel          string          `json:"channel"`
	Direction        string          `json:"direction"`
	Amount           int64           `json:"amount"`
	Fee              int64           `json:"fee"`
	TotalAmount      int64           `json:"total_amount"`
	Currency         string          `json:"currency"`
	StatusCode       string          `json:"status_code"`
	ErrorCode        string          `json:"error_code"`
	ErrorMessage     string          `json:"error_message"`
	SourceService    string          `json:"source_service"`
	TransactionTime  string          `json:"transaction_time"`
	MetadataJSON     json.RawMessage `json:"metadata_json"`
}

type nonRetryableError string

func (e nonRetryableError) Error() string { return string(e) }

type retryableError string

func (e retryableError) Error() string { return string(e) }

func IsNonRetryable(err error) bool {
	var nr nonRetryableError
	if errors.As(err, &nr) {
		return true
	}
	if errors.Is(err, service.ErrConflictingDuplicateReference) {
		return true
	}

	var appErr *coreerrors.AppError
	if errors.As(err, &appErr) {
		return appErr.Code == coreerrors.CodeInvalidRequest
	}
	return false
}

func classifyError(err error) string {
	if err == nil {
		return ""
	}
	var nr nonRetryableError
	if errors.As(err, &nr) {
		return nr.Error()
	}
	if errors.Is(err, service.ErrConflictingDuplicateReference) {
		return "conflicting_duplicate"
	}
	var appErr *coreerrors.AppError
	if errors.As(err, &appErr) {
		return appErr.FormatCode()
	}
	return "retryable_runtime_error"
}

func validateRequiredInput(in domain.CreateTransactionHistoryInput) error {
	switch {
	case in.UserID == "":
		return nonRetryableError("invalid_required_fields")
	case in.ReferenceID == "":
		return nonRetryableError("invalid_required_fields")
	case in.Channel == "":
		return nonRetryableError("invalid_required_fields")
	case in.ProductGroup == "":
		return nonRetryableError("invalid_required_fields")
	case in.ProductType == "":
		return nonRetryableError("invalid_required_fields")
	case in.StatusCode == "":
		return nonRetryableError("invalid_required_fields")
	case in.SourceService == "":
		return nonRetryableError("invalid_required_fields")
	case in.Currency == "":
		return nonRetryableError("invalid_required_fields")
	default:
		return nil
	}
}

func parseOptionalTime(raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, nil
	}
	return time.Parse(time.RFC3339, raw)
}

func normalizeMetadataJSON(raw json.RawMessage) (string, error) {
	if len(raw) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return "{}", nil
	}

	var asString string
	if err := json.Unmarshal(raw, &asString); err == nil {
		asString = strings.TrimSpace(asString)
		if asString == "" {
			return "{}", nil
		}
		return validateJSONObjectString(asString)
	}

	return validateJSONObjectString(string(raw))
}

func validateJSONObjectString(raw string) (string, error) {
	var decoded any
	if err := json.Unmarshal([]byte(raw), &decoded); err != nil {
		return "", err
	}
	if _, ok := decoded.(map[string]any); !ok {
		return "", fmt.Errorf("metadata_json must be object")
	}
	compact := bytes.Buffer{}
	if err := json.Compact(&compact, []byte(raw)); err != nil {
		return "", err
	}
	return compact.String(), nil
}
