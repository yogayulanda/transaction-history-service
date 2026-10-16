package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/yogayulanda/go-core/messaging"
	"github.com/yogayulanda/transaction-history-service/internal/catalog"
	"github.com/yogayulanda/transaction-history-service/internal/domain"
	"github.com/yogayulanda/transaction-history-service/internal/pii"
)

type statusApplierStub struct {
	header *domain.TransactionSnapshot
	detail *domain.Detail
	event  domain.StatusEvent
	result domain.ApplyResult
	err    error
}

func (s *statusApplierStub) ApplyStatusEvent(
	_ context.Context,
	header *domain.TransactionSnapshot,
	detail *domain.Detail,
	event domain.StatusEvent,
) (domain.ApplyResult, error) {
	s.header, s.detail, s.event = header, detail, event
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

func newStatusTestHandler(applier statusEventApplier, publisher messaging.Publisher) *StatusEventHandler {
	hasher, _ := pii.NewHasher([]byte(strings.Repeat("k", 32)))
	return NewStatusEventHandler(applier, catalog.DefaultStaticCatalog("", "sample-producer"), hasher, publisher, nil)
}

func TestStatusEventHandlerMasksNumbersBeforeApply(t *testing.T) {
	applier := &statusApplierStub{result: domain.Created}
	handler := newStatusTestHandler(applier, nil)
	payload, _ := json.Marshal(map[string]any{
		"transaction_id": "tx-1", "app_name": "app_a", "user_id": "user-1",
		"product_category": "transfer", "product_type": "transfer_internal", "provider": "bca",
		"direction": "debit", "amount": "100.00", "fee_amount": "0", "total_amount": "100.00",
		"status": "SUCCESS", "source_system": "sample-producer", "transaction_at": "2026-10-16T00:00:00Z",
		"changed_by_type": "system", "source_account_no": "1234567890",
		"destination_no": "+6281234567890", "destination_no_type": "phone",
		"detail": map[string]any{"note": "private note"},
	})
	attrs := messaging.CloudEventAttributes{
		ID: "6ba7b810-9dad-11d1-80b4-00c04fd430c8", Type: "transaction.status.v1",
		Subject: "tx-1", Time: time.Date(2026, 10, 16, 1, 0, 0, 0, time.UTC), Sequence: "3",
	}
	header, detailValue, event, reason := handler.decodeStatusEvent(attrs, payload)
	if reason != "" {
		t.Fatalf("decodeStatusEvent() reason = %q", reason)
	}
	if header.DestinationNoMasked != "0812****7890" || header.DestinationNoHash == "" || header.SourceAccountMasked != "****7890" {
		t.Fatalf("numbers were not masked/hashed: %+v", header)
	}
	if strings.Contains(header.DestinationNoMasked, "+62812") || strings.Contains(header.SourceAccountMasked, "123456") {
		t.Fatal("raw number reached the repository header")
	}
	if detailValue.Fields["note"] != "private note" || event.StatusVersion != 3 || event.EventID != attrs.ID {
		t.Fatal("normalized detail or event metadata did not match expected values")
	}
}

func TestStatusEventHandlerRejectsMissingCloudEventsToDLQ(t *testing.T) {
	publisher := &publisherStub{}
	handler := newStatusTestHandler(&statusApplierStub{}, publisher)
	err := handler.Handle(context.Background(), messaging.Message{Topic: "status", Payload: []byte(`{}`)})
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if publisher.message.Topic != "status.dlq" || publisher.message.Headers["x-dlq-reason"] != "missing_cloudevents" {
		t.Fatal("DLQ message did not contain the expected topic and reason")
	}
}

func TestStatusEventHandlerReturnsDatabaseErrorsForRetry(t *testing.T) {
	dbErr := errors.New("database unavailable")
	applier := &statusApplierStub{err: dbErr}
	handler := newStatusTestHandler(applier, nil)
	attrs := messaging.CloudEventAttributes{
		ID: "6ba7b810-9dad-11d1-80b4-00c04fd430c8", Type: "transaction.status.v1",
		Subject: "tx-1", Time: time.Now().UTC(), Sequence: "1",
	}
	payload, _ := json.Marshal(map[string]string{
		"transaction_id": "tx-1", "app_name": "app_a", "user_id": "user-1",
		"product_category": "transfer", "product_type": "transfer_internal", "provider": "bca",
		"direction": "debit", "amount": "100", "fee_amount": "0", "total_amount": "100",
		"status": "PENDING", "source_system": "sample-producer", "transaction_at": "2026-10-16T00:00:00Z",
		"changed_by_type": "system",
	})
	err := handler.handleWithEvent(context.Background(), messaging.Message{Topic: "status", Payload: payload}, attrs)
	if !errors.Is(err, dbErr) {
		t.Fatalf("handleWithEvent() error = %v, want %v", err, dbErr)
	}
}
