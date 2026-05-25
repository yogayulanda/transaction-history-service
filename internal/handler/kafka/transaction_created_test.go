package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/yogayulanda/go-core/messaging"
	"github.com/yogayulanda/transaction-history-service/internal/domain"
	"github.com/yogayulanda/transaction-history-service/internal/service"
)

type stubCreator struct {
	calls int
	input domain.CreateTransactionHistoryInput
	id    string
	noOp  bool
	err   error
}

func (s *stubCreator) CreateTransactionHistoryIdempotent(
	_ context.Context,
	in domain.CreateTransactionHistoryInput,
) (string, bool, error) {
	s.calls++
	s.input = in
	if s.id == "" {
		s.id = "tx-1"
	}
	return s.id, s.noOp, s.err
}

type stubPublisher struct {
	calls int
	msg   messaging.Message
	err   error
}

func (p *stubPublisher) Publish(_ context.Context, msg messaging.Message) error {
	p.calls++
	p.msg = msg
	return p.err
}

func (p *stubPublisher) Close() error { return nil }

func TestHandle_ValidEventCallsService(t *testing.T) {
	creator := &stubCreator{}
	dlq := &stubPublisher{}
	handler := NewTransactionCreatedHandler(creator, dlq, nil)

	err := handler.Handle(context.Background(), messaging.Message{
		Topic:   TransactionCreatedTopic,
		Key:     []byte("ref-1"),
		Payload: validEventPayload(t),
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if creator.calls != 1 {
		t.Fatalf("expected service call once, got %d", creator.calls)
	}
	if creator.input.ReferenceID != "ref-1" || creator.input.StatusCode != "SUCCESS" {
		t.Fatalf("unexpected mapped input: %+v", creator.input)
	}
	if creator.input.MetadataJSON != `{"customer_number":"123"}` {
		t.Fatalf("unexpected metadata json %q", creator.input.MetadataJSON)
	}
	if dlq.calls != 0 {
		t.Fatalf("expected no dlq publish, got %d", dlq.calls)
	}
}

func TestHandle_EquivalentDuplicateCommitsNoOp(t *testing.T) {
	creator := &stubCreator{noOp: true}
	handler := NewTransactionCreatedHandler(creator, &stubPublisher{}, nil)

	err := handler.Handle(context.Background(), messaging.Message{
		Topic:   TransactionCreatedTopic,
		Key:     []byte("ref-1"),
		Payload: validEventPayload(t),
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if creator.calls != 1 {
		t.Fatalf("expected service call once, got %d", creator.calls)
	}
}

func TestHandle_NonRetryableFailuresPublishDLQ(t *testing.T) {
	tests := []struct {
		name    string
		payload []byte
		err     error
	}{
		{name: "malformed json", payload: []byte(`{`)},
		{name: "invalid version", payload: eventPayloadWith(t, map[string]any{"event_version": 2})},
		{name: "unsupported event type", payload: eventPayloadWith(t, map[string]any{"event_type": "transaction.updated"})},
		{name: "producer not allowed", payload: eventPayloadWith(t, map[string]any{"producer": "unknown-service"})},
		{name: "missing required field", payload: eventPayloadWithPayload(t, map[string]any{"reference_id": ""})},
		{name: "conflicting duplicate", payload: validEventPayload(t), err: service.ErrConflictingDuplicateReference},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			creator := &stubCreator{err: tt.err}
			dlq := &stubPublisher{}
			handler := NewTransactionCreatedHandler(creator, dlq, nil)

			err := handler.Handle(context.Background(), messaging.Message{
				Topic:   TransactionCreatedTopic,
				Key:     []byte("ref-1"),
				Payload: tt.payload,
				Headers: map[string]string{"correlation_id": "corr-1"},
			})
			if err != nil {
				t.Fatalf("expected nil error after dlq success, got %v", err)
			}
			if dlq.calls != 1 {
				t.Fatalf("expected one dlq publish, got %d", dlq.calls)
			}
			if dlq.msg.Topic != TransactionCreatedDLQTopic {
				t.Fatalf("expected dlq topic %q, got %q", TransactionCreatedDLQTopic, dlq.msg.Topic)
			}
			if string(dlq.msg.Payload) != string(tt.payload) {
				t.Fatal("expected original payload in dlq message")
			}
			if dlq.msg.Headers["x-original-topic"] != TransactionCreatedTopic {
				t.Fatalf("expected original topic header, got %+v", dlq.msg.Headers)
			}
		})
	}
}

func TestHandle_DLQPublishFailureReturnsRetryableError(t *testing.T) {
	handler := NewTransactionCreatedHandler(&stubCreator{}, &stubPublisher{err: errors.New("publish failed")}, nil)

	err := handler.Handle(context.Background(), messaging.Message{
		Topic:   TransactionCreatedTopic,
		Key:     []byte("ref-1"),
		Payload: []byte(`{`),
	})
	if err == nil {
		t.Fatal("expected error when dlq publish fails")
	}
	if IsNonRetryable(err) {
		t.Fatalf("dlq publish failure must be retryable, got %v", err)
	}
}

func TestHandle_RetryableFailureReturnsErrorWithoutDLQ(t *testing.T) {
	creator := &stubCreator{err: errors.New("database timeout")}
	dlq := &stubPublisher{}
	handler := NewTransactionCreatedHandler(creator, dlq, nil)

	err := handler.Handle(context.Background(), messaging.Message{
		Topic:   TransactionCreatedTopic,
		Key:     []byte("ref-1"),
		Payload: validEventPayload(t),
	})
	if err == nil {
		t.Fatal("expected retryable error")
	}
	if IsNonRetryable(err) {
		t.Fatalf("expected retryable classification, got %v", err)
	}
	if dlq.calls != 0 {
		t.Fatalf("expected no manual dlq for retryable error, got %d", dlq.calls)
	}
}

func validEventPayload(t *testing.T) []byte {
	t.Helper()
	return eventPayloadWith(t, nil)
}

func eventPayloadWith(t *testing.T, overrides map[string]any) []byte {
	t.Helper()
	evt := map[string]any{
		"event_id":      "evt-1",
		"event_type":    TransactionCreatedEventType,
		"event_version": TransactionCreatedVersion,
		"producer":      "trxFinance",
		"payload":       validPayloadMap(),
	}
	for key, value := range overrides {
		evt[key] = value
	}
	out, err := json.Marshal(evt)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	return out
}

func eventPayloadWithPayload(t *testing.T, overrides map[string]any) []byte {
	t.Helper()
	payload := validPayloadMap()
	for key, value := range overrides {
		payload[key] = value
	}
	return eventPayloadWith(t, map[string]any{"payload": payload})
}

func validPayloadMap() map[string]any {
	return map[string]any{
		"user_id":           "user-1",
		"reference_id":      "ref-1",
		"external_ref_id":   "ext-1",
		"product_group":     "transfer",
		"product_type":      "transfer_internal",
		"transaction_route": "internal",
		"channel":           "mobile_app",
		"direction":         "debit",
		"amount":            int64(100),
		"fee":               int64(1),
		"total_amount":      int64(101),
		"currency":          "IDR",
		"status_code":       "SUCCESS",
		"source_service":    "trxFinance",
		"transaction_time":  time.Date(2026, 5, 25, 1, 2, 3, 0, time.UTC).Format(time.RFC3339),
		"metadata_json": map[string]any{
			"customer_number": "123",
		},
	}
}
