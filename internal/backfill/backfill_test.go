package backfill

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/yogayulanda/transaction-history-service/internal/catalog"
	"github.com/yogayulanda/transaction-history-service/internal/domain"
	"github.com/yogayulanda/transaction-history-service/internal/handler/kafka"
	"github.com/yogayulanda/transaction-history-service/internal/pii"
)

type fakeSource struct {
	name    string
	records []Record
	read    bool
}

func (s *fakeSource) Name() string { return s.name }

func (s *fakeSource) Next(_ context.Context, _ int) ([]Record, error) {
	if s.read {
		return nil, nil
	}
	s.read = true
	return s.records, nil
}

type memoryRepository struct {
	items  map[string]domain.TransactionSnapshot
	events map[string]domain.StatusEvent
}

func newMemoryRepository() *memoryRepository {
	return &memoryRepository{items: make(map[string]domain.TransactionSnapshot), events: make(map[string]domain.StatusEvent)}
}

func (r *memoryRepository) InsertBackfill(_ context.Context, header *domain.TransactionSnapshot, _ *domain.Detail, event domain.StatusEvent) (bool, error) {
	if _, exists := r.items[header.ID]; exists {
		return false, nil
	}
	r.items[header.ID] = *header
	r.events[header.ID] = event
	return true, nil
}

func (r *memoryRepository) ApplyStatusEvent(_ context.Context, header *domain.TransactionSnapshot, _ *domain.Detail, event domain.StatusEvent) (domain.ApplyResult, error) {
	current, exists := r.items[header.ID]
	if !exists {
		r.items[header.ID] = *header
		r.events[header.ID] = event
		return domain.Created, nil
	}
	if event.StatusVersion <= current.StatusVersion {
		return domain.Stale, nil
	}
	r.items[header.ID] = *header
	r.events[header.ID] = event
	return domain.Updated, nil
}

func newRunner(source Source, repository *memoryRepository) *Runner {
	hasher, _ := pii.NewHasher([]byte(strings.Repeat("k", 32)))
	preparer := kafka.NewStatusEventHandler(nil, catalog.DefaultStaticCatalog("", "sample-producer"), hasher, nil, nil)
	return NewRunner(source, repository, preparer)
}

func validPayload(transactionID, status, amount string) json.RawMessage {
	payload, _ := json.Marshal(map[string]string{
		"transaction_id":   transactionID,
		"app_name":         "app_a",
		"user_id":          "user-1",
		"product_category": "transfer",
		"product_type":     "transfer_internal",
		"provider":         "bca",
		"direction":        "debit",
		"amount":           amount,
		"fee_amount":       "0.10",
		"total_amount":     amount,
		"currency":         "IDR",
		"status":           status,
		"source_system":    "sample-producer",
		"transaction_at":   "2026-10-15T17:30:00Z",
		"changed_by_type":  "system",
	})
	return payload
}

func TestRunnerRerunIsInsertOnlyAndUsesDeterministicBackfillEvent(t *testing.T) {
	record := Record{SourceID: "legacy-42", Payload: validPayload("tx-1", "SUCCESS", "12.34")}
	repository := newMemoryRepository()
	first, err := newRunner(&fakeSource{name: "sample-source", records: []Record{record}}, repository).Run(context.Background(), 10)
	if err != nil {
		t.Fatal(err)
	}
	key := ReportKey{Day: "2026-10-16", ProductType: "transfer_internal"}
	if got := first.Groups[key]; got == nil || got.SourceCount != 1 || got.Inserted != 1 || got.TotalAmount != "12.34" {
		t.Fatalf("unexpected first run report: %+v", got)
	}
	stored := repository.items["tx-1"]
	storedEvent := repository.events["tx-1"]
	wantID := uuid.NewSHA1(uuid.NameSpaceDNS, []byte("sample-source:legacy-42")).String()
	if stored.StatusVersion != 0 || stored.SourceSystem != "backfill" || storedEvent.EventID != wantID ||
		storedEvent.EventType != backfillEventType || storedEvent.StatusVersion != 0 ||
		storedEvent.ChangedByType != "system" || !storedEvent.ChangedAt.Equal(stored.TransactionAt) {
		t.Fatalf("unexpected backfill snapshot/event: %+v / %+v", stored, storedEvent)
	}

	second, err := newRunner(&fakeSource{name: "sample-source", records: []Record{record}}, repository).Run(context.Background(), 10)
	if err != nil {
		t.Fatal(err)
	}
	if got := second.Groups[key]; got == nil || got.Inserted != 0 || got.SkippedExisting != 1 {
		t.Fatalf("unexpected rerun report: %+v", got)
	}
	if repository.items["tx-1"].Status != "SUCCESS" {
		t.Fatalf("rerun changed existing row: %+v", repository.items["tx-1"])
	}
}

func TestRunnerPreservesExistingAndLiveEventUpdatesBackfill(t *testing.T) {
	repository := newMemoryRepository()
	repository.items["tx-existing"] = domain.TransactionSnapshot{ID: "tx-existing", Status: "FAILED", StatusVersion: 2}
	_, err := newRunner(&fakeSource{name: "sample-source", records: []Record{{SourceID: "existing", Payload: validPayload("tx-existing", "SUCCESS", "1")}}}, repository).Run(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if repository.items["tx-existing"].Status != "FAILED" || repository.items["tx-existing"].StatusVersion != 2 {
		t.Fatalf("backfill overwrote existing row: %+v", repository.items["tx-existing"])
	}

	_, err = newRunner(&fakeSource{name: "sample-source", records: []Record{{SourceID: "new", Payload: validPayload("tx-new", "PENDING", "1")}}}, repository).Run(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	liveHeader := repository.items["tx-new"]
	liveHeader.Status = "SUCCESS"
	liveHeader.StatusVersion = 1
	result, err := repository.ApplyStatusEvent(context.Background(), &liveHeader, nil, domain.StatusEvent{EventID: "live-1", StatusVersion: 1, ToStatus: "SUCCESS"})
	if err != nil || result != domain.Updated || repository.items["tx-new"].Status != "SUCCESS" {
		t.Fatalf("live event did not update backfill: result=%s err=%v item=%+v", result, err, repository.items["tx-new"])
	}
}

func TestRunnerSkipsInvalidAndAddsDecimalStringsExactly(t *testing.T) {
	bad := json.RawMessage(`{"transaction_id":"bad","product_type":"transfer_internal","transaction_at":"not-a-time"}`)
	repository := newMemoryRepository()
	runner := newRunner(&fakeSource{name: "sample-source", records: []Record{
		{SourceID: "one", Payload: validPayload("tx-1", "PENDING", "9999999999999999.99")},
		{SourceID: "two", Payload: validPayload("tx-2", "PENDING", "0.01")},
		{SourceID: "bad", Payload: bad},
	}}, repository)
	report, err := runner.Run(context.Background(), 3)
	if err != nil {
		t.Fatal(err)
	}
	group := report.Groups[ReportKey{Day: "2026-10-16", ProductType: "transfer_internal"}]
	if group == nil || group.SourceCount != 2 || group.Inserted != 2 || group.Invalid != 0 || group.TotalAmount != "10000000000000000.00" {
		t.Fatalf("unexpected aggregate report: %+v", group)
	}
	invalidGroup := report.Groups[ReportKey{Day: "unknown", ProductType: "transfer_internal"}]
	if invalidGroup == nil || invalidGroup.SourceCount != 1 || invalidGroup.Invalid != 1 || invalidGroup.InvalidReasons["invalid_transaction_at"] != 1 {
		t.Fatalf("unexpected invalid record report: %+v", invalidGroup)
	}
}

func TestAddDecimalStrings(t *testing.T) {
	got, err := addDecimalStrings("999999999999999999999999.99", "0.01")
	if err != nil || got != "1000000000000000000000000.00" {
		t.Fatalf("addDecimalStrings() = %q, %v", got, err)
	}
	if _, err := addDecimalStrings("1.001", "2"); err == nil {
		t.Fatal("expected malformed decimal to fail")
	}
}
