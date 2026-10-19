package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/yogayulanda/go-core/messaging"
	"github.com/yogayulanda/transaction-history-service/internal/domain"
)

type statusApplierStub struct {
	input  domain.StatusEventInput
	result domain.ApplyResult
	err    error
}

func (s *statusApplierStub) ApplyStatusEvent(_ context.Context, input domain.StatusEventInput) (domain.ApplyResult, error) {
	s.input = input
	return s.result, s.err
}

type publisherStub struct {
	message messaging.Message
	err     error
}

func (s *publisherStub) Publish(_ context.Context, message messaging.Message) error {
	s.message = message
	return s.err
}

func (s *publisherStub) Close() error { return nil }

func statusEventAttrs() messaging.CloudEventAttributes {
	return messaging.CloudEventAttributes{
		ID: "6ba7b810-9dad-11d1-80b4-00c04fd430c8", Type: "transaction.status.v1",
		Subject: "tx-1", Time: time.Date(2026, 10, 16, 1, 0, 0, 0, time.UTC), Sequence: "3",
	}
}

func validStatusPayload() []byte {
	payload, _ := json.Marshal(map[string]any{
		"transaction_id": "tx-1", "app_name": "app_a", "user_id": "user-1",
		"product_category": "transfer", "product_type": "transfer_internal", "provider": "bca",
		"direction": "debit", "amount": "100.00", "fee_amount": "0", "total_amount": "100.00",
		"status": "SUCCESS", "source_system": "sample-producer", "transaction_at": "2026-10-16T00:00:00Z",
		"changed_by_type": "system", "source_account_no": "1234567890",
		"destination_no": "+6281234567890", "destination_no_type": "phone",
		"detail": map[string]any{"note": "private note"},
	})
	return payload
}

func TestStatusEventHandlerDecodesAndDelegates(t *testing.T) {
	applier := &statusApplierStub{result: domain.Created}
	handler := NewStatusEventHandler(applier, nil, nil)
	msg := messaging.Message{Topic: "status", Payload: validStatusPayload()}
	if err := handler.handleWithEvent(context.Background(), msg, statusEventAttrs()); err != nil {
		t.Fatal(err)
	}
	if applier.input.TransactionID != "tx-1" || applier.input.StatusVersion != 3 || applier.input.EventID != statusEventAttrs().ID {
		t.Fatalf("handler passed incomplete decoded input: %+v", applier.input)
	}
	if applier.input.SourceAccountNo != "1234567890" || applier.input.DestinationNo != "+6281234567890" {
		t.Fatal("handler should pass decoded values to the service without business normalization")
	}
}

func TestStatusEventHandlerRejectsMissingCloudEventsToDLQ(t *testing.T) {
	publisher := &publisherStub{}
	handler := NewStatusEventHandler(&statusApplierStub{}, publisher, nil)
	err := handler.Handle(context.Background(), messaging.Message{Topic: "status", Payload: []byte(`{}`)})
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if publisher.message.Topic != "status.dlq" || publisher.message.Headers["x-dlq-reason"] != "missing_cloudevents" {
		t.Fatal("DLQ message did not contain the expected topic and reason")
	}
}

func TestStatusEventHandlerRoutesServiceValidationErrorsToDLQ(t *testing.T) {
	publisher := &publisherStub{}
	applier := &statusApplierStub{err: &domain.EventValidationError{Reason: "invalid_detail"}}
	handler := NewStatusEventHandler(applier, publisher, nil)
	if err := handler.handleWithEvent(context.Background(), messaging.Message{Topic: "status", Payload: validStatusPayload()}, statusEventAttrs()); err != nil {
		t.Fatal(err)
	}
	if publisher.message.Headers["x-dlq-reason"] != "invalid_detail" {
		t.Fatalf("DLQ reason = %q, want invalid_detail", publisher.message.Headers["x-dlq-reason"])
	}
}

func TestStatusEventHandlerReturnsDatabaseErrorsForRetry(t *testing.T) {
	dbErr := errors.New("database unavailable")
	applier := &statusApplierStub{err: dbErr}
	handler := NewStatusEventHandler(applier, nil, nil)
	err := handler.handleWithEvent(context.Background(), messaging.Message{Topic: "status", Payload: validStatusPayload()}, statusEventAttrs())
	if !errors.Is(err, dbErr) {
		t.Fatalf("handleWithEvent() error = %v, want %v", err, dbErr)
	}
}
