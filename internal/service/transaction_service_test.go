package service

import (
	"context"
	"errors"
	"testing"

	coreerrors "github.com/yogayulanda/go-core/errors"
	"github.com/yogayulanda/go-core/observability"
	"github.com/yogayulanda/transaction-history-service/internal/domain"
	"gorm.io/gorm"
)

type fakeRepository struct {
	createFn            func(ctx context.Context, in domain.CreateTransactionHistoryInput) (string, error)
	findByReferenceIDFn func(ctx context.Context, referenceID string) (*domain.TransactionHistoryDetail, error)
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

func (f fakeRepository) FindByReferenceID(ctx context.Context, referenceID string) (*domain.TransactionHistoryDetail, error) {
	if f.findByReferenceIDFn != nil {
		return f.findByReferenceIDFn(ctx, referenceID)
	}
	return nil, domain.ErrTransactionNotFound
}

func (f fakeRepository) ListByUser(context.Context, domain.ListUserHistoryFilter) ([]domain.TransactionHistory, bool, error) {
	return nil, false, nil
}

func TestCreateTransactionHistory_ValidatesBusinessFields(t *testing.T) {
	svc := NewTransactionService(fakeRepository{}, nil, nil, nil)

	_, err := svc.CreateTransactionHistory(context.Background(), domain.CreateTransactionHistoryInput{
		UserID:        "user-1",
		ReferenceID:   "ref-1",
		Channel:       "app-one",
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
	var appErr *coreerrors.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got %T: %v", err, err)
	}
	if appErr.Code != "CORE-1000" {
		t.Fatalf("expected CORE-1000 code, got %s", appErr.Code)
	}
	if len(appErr.Details) == 0 || appErr.Details[0].Field != "metadata_json" {
		t.Fatalf("expected metadata_json validation detail, got %+v", appErr.Details)
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
	var appErr *coreerrors.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got %T: %v", err, err)
	}
	if appErr.Code != CodeDuplicateReferenceID {
		t.Fatalf("expected %s code, got %s", CodeDuplicateReferenceID, appErr.Code)
	}
}

func TestCreateTransactionHistory_StoresIdentityInObservabilityFieldBag(t *testing.T) {
	ctx := observability.EnsureFieldBag(context.Background())
	svc := NewTransactionService(fakeRepository{}, nil, nil, nil)
	_, err := svc.CreateTransactionHistory(ctx, domain.CreateTransactionHistoryInput{
		UserID:        "user-1",
		ReferenceID:   "ref-1",
		ProductGroup:  "transfer",
		ProductType:   "transfer_internal",
		Channel:       "app-one",
		StatusCode:    "SUCCESS",
		SourceService: "trxFinance",
		Currency:      "IDR",
		MetadataJSON:  "{}",
	})
	if err != nil {
		t.Fatalf("expected successful create, got %v", err)
	}
	if got := observability.GetUserID(ctx); got != "user-1" {
		t.Fatalf("expected user id user-1, got %q", got)
	}
	if got := observability.Labels(ctx)["reference_id"]; got != "ref-1" {
		t.Fatalf("expected reference_id label ref-1, got %q", got)
	}
}
