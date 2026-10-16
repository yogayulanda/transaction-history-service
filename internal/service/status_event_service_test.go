package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/yogayulanda/transaction-history-service/internal/domain"
	"github.com/yogayulanda/transaction-history-service/internal/repository"
	"github.com/yogayulanda/transaction-history-service/internal/utils"
)

type statusEventRepositoryStub struct {
	result domain.ApplyResult
	err    error
}

func (s statusEventRepositoryStub) ApplyStatusEvent(
	context.Context,
	*domain.Transaction,
	*domain.Detail,
	domain.StatusEvent,
) (domain.ApplyResult, error) {
	return s.result, s.err
}

func testStatusEventInput() domain.StatusEventInput {
	return domain.StatusEventInput{
		EventID: "6ba7b810-9dad-11d1-80b4-00c04fd430c8", EventType: "transaction.status.v1",
		Subject: "tx-1", EventTime: time.Now().UTC(), StatusVersion: 1,
		TransactionID: "tx-1", AppName: "app_a", UserID: "user-1", ProductCategory: "transfer",
		ProductType: "transfer_internal", Provider: "bca", Direction: "debit", Amount: "10.00",
		FeeAmount: "0.00", TotalAmount: "10.00", Status: domain.StatusPending,
		SourceSystem: "sample-producer", TransactionAt: "2026-10-16T00:00:00Z", ChangedByType: domain.ChangedBySystem,
	}
}

func testStatusEventService(repo statusEventRepository) *StatusEventService {
	validationCatalog := repository.NewStaticCatalog([]string{"app_a"}, []string{"sample-producer"})
	hasher, _ := utils.NewHasher([]byte("12345678901234567890123456789012"))
	return NewStatusEventService(repo, validationCatalog, hasher, nil)
}

func TestStatusEventServiceApplyStatusEvent(t *testing.T) {
	service := testStatusEventService(statusEventRepositoryStub{result: domain.Updated})
	result, err := service.ApplyStatusEvent(context.Background(), testStatusEventInput())
	if err != nil || result != domain.Updated {
		t.Fatalf("ApplyStatusEvent() = %q, %v", result, err)
	}
}

func TestStatusEventServiceReturnsRepositoryError(t *testing.T) {
	wantErr := errors.New("database unavailable")
	service := testStatusEventService(statusEventRepositoryStub{err: wantErr})
	_, err := service.ApplyStatusEvent(context.Background(), testStatusEventInput())
	if !errors.Is(err, wantErr) {
		t.Fatalf("ApplyStatusEvent() error = %v, want %v", err, wantErr)
	}
}

func TestPrepareStatusEventMasksAndHashesSensitiveValues(t *testing.T) {
	service := testStatusEventService(statusEventRepositoryStub{})
	input := testStatusEventInput()
	input.SourceAccountNo = "1234567890"
	input.DestinationNoType = "phone"
	input.DestinationNo = "+6281234567890"
	input.ProductCategory = domain.CategoryCICO
	input.ProductType = domain.ProductCashIn
	input.Detail = map[string]any{"code": "SAMPLE-CASH-1234", "warkat_no": "WR-1"}

	transaction, detail, event, err := service.PrepareStatusEvent(input)
	if err != nil {
		t.Fatal(err)
	}
	if transaction.SourceAccountMasked != "****7890" || transaction.DestinationNoMasked != "0812****7890" {
		t.Fatalf("unexpected masked values: source=%q destination=%q", transaction.SourceAccountMasked, transaction.DestinationNoMasked)
	}
	if transaction.DestinationNoHash != service.hasher.Hash("phone", "081234567890") || transaction.TransactionID != "tx-1" || transaction.ID != 0 {
		t.Fatal("prepared transaction did not retain the external ID and normalized destination hash")
	}
	if detail.Fields["code_masked"] != "****234" || detail.Fields["code"] != nil || event.StatusVersion != 1 {
		t.Fatalf("unexpected prepared detail/event: %+v / %+v", detail, event)
	}
}

func TestValidateDetailSizeUsesNVarcharCodeUnits(t *testing.T) {
	if err := validateDetailSize(map[string]any{"x": strings.Repeat("a", 3992)}); err != nil {
		t.Fatalf("4000-character JSON was rejected: %v", err)
	}
	if err := validateDetailSize(map[string]any{"x": strings.Repeat("a", 3993)}); err == nil {
		t.Fatal("JSON larger than NVARCHAR(4000) was accepted")
	}
	if err := validateDetailSize(map[string]any{"x": strings.Repeat("😀", 1997)}); err == nil {
		t.Fatal("JSON larger than 4000 UTF-16 code units was accepted")
	}
}

func TestPrepareStatusEventRejectsNonDecimalDenomination(t *testing.T) {
	service := testStatusEventService(statusEventRepositoryStub{})
	input := testStatusEventInput()
	input.Denomination = "-1.00"
	if _, _, _, err := service.PrepareStatusEvent(input); !errors.Is(err, domain.ErrInvalidEvent) {
		t.Fatalf("PrepareStatusEvent() error = %v, want invalid event", err)
	}
}
