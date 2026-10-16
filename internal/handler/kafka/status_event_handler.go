package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	"github.com/yogayulanda/go-core/logger"
	"github.com/yogayulanda/go-core/messaging"
	"github.com/yogayulanda/transaction-history-service/internal/domain"
)

type statusEventApplier interface {
	ApplyStatusEvent(ctx context.Context, input domain.StatusEventInput) (domain.ApplyResult, error)
}

type StatusEventHandler struct {
	service statusEventApplier
	dlq     messaging.Publisher
	log     logger.Logger
}

func NewStatusEventHandler(svc statusEventApplier, dlq messaging.Publisher, log logger.Logger) *StatusEventHandler {
	return &StatusEventHandler{service: svc, dlq: dlq, log: log}
}

func (h *StatusEventHandler) Handle(ctx context.Context, msg messaging.Message) error {
	attrs, ok := messaging.CloudEventsFromContext(ctx)
	if !ok {
		return h.reject(ctx, msg, "missing_cloudevents")
	}
	return h.handleWithEvent(ctx, msg, attrs)
}

func (h *StatusEventHandler) handleWithEvent(ctx context.Context, msg messaging.Message, attrs messaging.CloudEventAttributes) error {
	input, reason := decodeStatusEvent(attrs, msg.Payload)
	if reason != "" {
		return h.reject(ctx, msg, reason)
	}
	result, err := h.service.ApplyStatusEvent(ctx, input)
	if errors.Is(err, domain.ErrInvalidEvent) {
		var validationError *domain.EventValidationError
		if errors.As(err, &validationError) {
			return h.reject(ctx, msg, validationError.Reason)
		}
		return h.reject(ctx, msg, "invalid_status_event")
	}
	if err != nil {
		h.emit(ctx, "retryable_failed", "database_error", result)
		return err
	}
	h.emit(ctx, "success", "", result)
	return nil
}

func decodeStatusEvent(attrs messaging.CloudEventAttributes, payload []byte) (domain.StatusEventInput, string) {
	if _, err := uuid.Parse(attrs.ID); err != nil || attrs.Type == "" || len(attrs.Type) > 128 || attrs.Subject == "" || attrs.Time.IsZero() {
		return domain.StatusEventInput{}, "invalid_cloudevents"
	}
	version, err := strconv.ParseInt(attrs.Sequence, 10, 64)
	if err != nil || version < 1 {
		return domain.StatusEventInput{}, "invalid_status_version"
	}
	var input domain.StatusEventInput
	if err := json.Unmarshal(payload, &input); err != nil {
		return domain.StatusEventInput{}, "malformed_json"
	}
	input.EventID = attrs.ID
	input.EventType = attrs.Type
	input.Subject = attrs.Subject
	input.EventTime = attrs.Time
	input.StatusVersion = version
	return input, ""
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
