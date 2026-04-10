package service

import (
	"context"
	"errors"
	"testing"

	"github.com/yogayulanda/transaction-history-service/internal/domain"
	"gorm.io/gorm"
)

type fakeRepository struct {
	createFn func(ctx context.Context, in domain.CreateTransactionHistoryInput) (string, error)
}

func (f fakeRepository) Create(ctx context.Context, in domain.CreateTransactionHistoryInput) (string, error) {
	if f.createFn != nil {
		return f.createFn(ctx, in)
	}
	return "tx-1", nil
}

func (f fakeRepository) FindDetailByID(context.Context, string) (*domain.TransactionHistoryDetail, error) {
	return nil, nil
}

func (f fakeRepository) ListByUser(context.Context, domain.ListUserHistoryFilter) ([]domain.TransactionHistory, bool, error) {
	return nil, false, nil
}

func TestCreateTransactionHistory_ValidatesBusinessFields(t *testing.T) {
	svc := NewTransactionService(fakeRepository{}, nil, nil, nil)

	_, err := svc.CreateTransactionHistory(context.Background(), domain.CreateTransactionHistoryInput{
		UserID:        "user-1",
		ReferenceID:   "ref-1",
		ProductGroup:  "transfer",
		ProductType:   "transfer_internal",
		StatusCode:    "SUCCESS",
		SourceService: "trxFinance",
		Currency:      "IDR",
		MetadataJSON:  `[]`,
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !errors.Is(err, ErrInvalidTransactionHistoryInput) {
		t.Fatalf("expected invalid input error, got %v", err)
	}
}

func TestCreateTransactionHistory_NormalizesAndDefaultsInput(t *testing.T) {
	var saved domain.CreateTransactionHistoryInput
	svc := NewTransactionService(fakeRepository{
		createFn: func(_ context.Context, in domain.CreateTransactionHistoryInput) (string, error) {
			saved = in
			return "tx-123", nil
		},
	}, nil, nil, nil)

	id, err := svc.CreateTransactionHistory(context.Background(), domain.CreateTransactionHistoryInput{
		UserID:        " user-1 ",
		ReferenceID:   " ref-1 ",
		ProductGroup:  " transfer ",
		ProductType:   " transfer_internal ",
		Channel:       " app-one ",
		StatusCode:    " success ",
		SourceService: " trxFinance ",
		Currency:      " idr ",
		MetadataJSON:  "",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if id != "tx-123" {
		t.Fatalf("unexpected id %q", id)
	}
	if saved.ReferenceID != "ref-1" || saved.Channel != "app-one" || saved.Currency != "IDR" || saved.StatusCode != "SUCCESS" {
		t.Fatalf("input was not normalized: %+v", saved)
	}
	if saved.MetadataJSON != "{}" {
		t.Fatalf("expected metadata default, got %q", saved.MetadataJSON)
	}
}

func TestCreateTransactionHistory_MapsDuplicateReferenceID(t *testing.T) {
	svc := NewTransactionService(fakeRepository{
		createFn: func(context.Context, domain.CreateTransactionHistoryInput) (string, error) {
			return "", gorm.ErrDuplicatedKey
		},
	}, nil, nil, nil)

	_, err := svc.CreateTransactionHistory(context.Background(), domain.CreateTransactionHistoryInput{
		UserID:        "user-1",
		ReferenceID:   "ref-1",
		ProductGroup:  "transfer",
		ProductType:   "transfer_internal",
		Channel:       "app-one",
		StatusCode:    "SUCCESS",
		SourceService: "trxFinance",
		Currency:      "IDR",
		MetadataJSON:  `{}`,
	})
	if !errors.Is(err, ErrDuplicateReferenceID) {
		t.Fatalf("expected duplicate reference error, got %v", err)
	}
}
