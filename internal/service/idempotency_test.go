package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yogayulanda/transaction-history-service/internal/domain"
	"gorm.io/gorm"
)

func TestCreateTransactionHistoryIdempotent_EquivalentDuplicateNoOp(t *testing.T) {
	now := time.Date(2026, 5, 25, 1, 2, 3, 0, time.UTC)
	svc := NewTransactionService(fakeRepository{
		createFn: func(context.Context, domain.CreateTransactionHistoryInput) (string, error) {
			return "", gorm.ErrDuplicatedKey
		},
		findByReferenceIDFn: func(_ context.Context, referenceID string) (*domain.TransactionHistoryDetail, error) {
			if referenceID != "ref-1" {
				t.Fatalf("unexpected reference id %q", referenceID)
			}
			return &domain.TransactionHistoryDetail{
				TransactionHistory: domain.TransactionHistory{
					ID:               "tx-existing",
					UserID:           "user-1",
					ReferenceID:      "ref-1",
					ExternalRefID:    "ext-1",
					ProductGroup:     "transfer",
					ProductType:      "transfer_internal",
					TransactionRoute: "internal",
					Channel:          "mobile_app",
					Direction:        "debit",
					Amount:           100,
					Fee:              1,
					TotalAmount:      101,
					Currency:         "IDR",
					StatusCode:       "SUCCESS",
					ErrorCode:        "",
					ErrorMessage:     "",
					SourceService:    "trxFinance",
					TransactionTime:  now,
				},
				MetadataJSON: `{"customer_number":"123"}`,
			}, nil
		},
	}, nil, nil, nil, nil)

	id, noOp, err := svc.CreateTransactionHistoryIdempotent(context.Background(), domain.CreateTransactionHistoryInput{
		UserID:           " user-1 ",
		ReferenceID:      " ref-1 ",
		ExternalRefID:    "ext-1",
		ProductGroup:     "transfer",
		ProductType:      "transfer_internal",
		TransactionRoute: "internal",
		Channel:          "mobile_app",
		Direction:        "debit",
		Amount:           100,
		Fee:              1,
		TotalAmount:      101,
		Currency:         "idr",
		StatusCode:       "success",
		SourceService:    "trxFinance",
		TransactionTime:  now,
		MetadataJSON:     `{"customer_number":"123"}`,
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if id != "tx-existing" || !noOp {
		t.Fatalf("expected existing no-op, got id=%q noOp=%v", id, noOp)
	}
}

func TestCreateTransactionHistoryIdempotent_ConflictingDuplicate(t *testing.T) {
	svc := NewTransactionService(fakeRepository{
		createFn: func(context.Context, domain.CreateTransactionHistoryInput) (string, error) {
			return "", gorm.ErrDuplicatedKey
		},
		findByReferenceIDFn: func(context.Context, string) (*domain.TransactionHistoryDetail, error) {
			return &domain.TransactionHistoryDetail{
				TransactionHistory: domain.TransactionHistory{
					ID:               "tx-existing",
					UserID:           "user-1",
					ReferenceID:      "ref-1",
					ProductGroup:     "transfer",
					ProductType:      "transfer_internal",
					TransactionRoute: "internal",
					Channel:          "mobile_app",
					Direction:        "debit",
					Amount:           999,
					Fee:              0,
					TotalAmount:      999,
					Currency:         "IDR",
					StatusCode:       "SUCCESS",
					SourceService:    "trxFinance",
				},
				MetadataJSON: `{}`,
			}, nil
		},
	}, nil, nil, nil, nil)

	_, _, err := svc.CreateTransactionHistoryIdempotent(context.Background(), domain.CreateTransactionHistoryInput{
		UserID:           "user-1",
		ReferenceID:      "ref-1",
		ProductGroup:     "transfer",
		ProductType:      "transfer_internal",
		TransactionRoute: "internal",
		Channel:          "mobile_app",
		Direction:        "debit",
		Amount:           100,
		Fee:              0,
		TotalAmount:      100,
		Currency:         "IDR",
		StatusCode:       "SUCCESS",
		SourceService:    "trxFinance",
		MetadataJSON:     `{}`,
	})
	if !errors.Is(err, ErrConflictingDuplicateReference) {
		t.Fatalf("expected conflicting duplicate error, got %v", err)
	}
}
