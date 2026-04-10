package grpc

import (
	"context"
	"testing"

	coreerrors "github.com/yogayulanda/go-core/errors"
	historyv1 "github.com/yogayulanda/transaction-history-service/gen/go/history/v1"
	"github.com/yogayulanda/transaction-history-service/internal/domain"
	"github.com/yogayulanda/transaction-history-service/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type stubRepository struct {
	createFn func(ctx context.Context, in domain.CreateTransactionHistoryInput) (string, error)
	detailFn func(ctx context.Context, id string) (*domain.TransactionHistoryDetail, error)
	listFn   func(ctx context.Context, filter domain.ListUserHistoryFilter) ([]domain.TransactionHistory, bool, error)
}

func (s stubRepository) Create(ctx context.Context, in domain.CreateTransactionHistoryInput) (string, error) {
	if s.createFn != nil {
		return s.createFn(ctx, in)
	}
	return "tx-1", nil
}

func (s stubRepository) FindDetailByID(ctx context.Context, id string) (*domain.TransactionHistoryDetail, error) {
	if s.detailFn != nil {
		return s.detailFn(ctx, id)
	}
	return nil, nil
}

func (s stubRepository) ListByUser(ctx context.Context, filter domain.ListUserHistoryFilter) ([]domain.TransactionHistory, bool, error) {
	if s.listFn != nil {
		return s.listFn(ctx, filter)
	}
	return nil, false, nil
}

func TestCreateTransactionHistory_ReturnsInvalidArgumentForBusinessValidation(t *testing.T) {
	handler := &Handler{
		service: service.NewTransactionService(stubRepository{}, nil, nil, nil),
	}

	_, err := handler.CreateTransactionHistory(context.Background(), &historyv1.CreateTransactionHistoryRequest{
		UserId:        "user-1",
		ReferenceId:   "ref-1",
		ProductGroup:  "transfer",
		ProductType:   "transfer_internal",
		StatusCode:    historyv1.TransactionStatusCode_TRANSACTION_STATUS_CODE_SUCCESS,
		SourceService: "trxFinance",
		Currency:      "IDR",
		MetadataJson:  `{}`,
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected invalid argument, got %v", status.Code(err))
	}

	// Verify ErrorInfo contains go-core reason code.
	appCode := coreerrors.CodeFromGRPC(err)
	if appCode != coreerrors.CodeInvalidRequest {
		t.Fatalf("expected INVALID_REQUEST reason, got %s", appCode)
	}
}

func TestCreateTransactionHistory_MapsDuplicateReferenceID(t *testing.T) {
	handler := &Handler{
		service: service.NewTransactionService(stubRepository{
			createFn: func(context.Context, domain.CreateTransactionHistoryInput) (string, error) {
				return "", gorm.ErrDuplicatedKey
			},
		}, nil, nil, nil),
	}

	_, err := handler.CreateTransactionHistory(context.Background(), &historyv1.CreateTransactionHistoryRequest{
		UserId:        "user-1",
		ReferenceId:   "ref-1",
		ProductGroup:  "transfer",
		ProductType:   "transfer_internal",
		Channel:       "app-one",
		StatusCode:    historyv1.TransactionStatusCode_TRANSACTION_STATUS_CODE_SUCCESS,
		SourceService: "trxFinance",
		Currency:      "IDR",
		MetadataJson:  `{}`,
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected invalid argument, got %v", status.Code(err))
	}
}

func TestGetUserHistory_RejectsInvalidDateRange(t *testing.T) {
	handler := &Handler{
		service: service.NewTransactionService(stubRepository{}, nil, nil, nil),
	}

	_, err := handler.GetUserHistory(context.Background(), &historyv1.GetUserHistoryRequest{
		UserId:    "user-1",
		StartDate: "2026-04-10T00:00:00Z",
		EndDate:   "2026-04-09T00:00:00Z",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected invalid argument, got %v", status.Code(err))
	}
}

func TestGetTransactionHistoryDetail_RequiresID(t *testing.T) {
	handler := &Handler{
		service: service.NewTransactionService(stubRepository{}, nil, nil, nil),
	}

	_, err := handler.GetTransactionHistoryDetail(context.Background(), &historyv1.GetTransactionHistoryDetailRequest{})
	if err == nil {
		t.Fatal("expected error")
	}
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected invalid argument, got %v", status.Code(err))
	}
}

func TestGetTransactionHistoryDetail_MapsNotFound(t *testing.T) {
	handler := &Handler{
		service: service.NewTransactionService(stubRepository{
			detailFn: func(_ context.Context, _ string) (*domain.TransactionHistoryDetail, error) {
				return nil, domain.ErrTransactionNotFound
			},
		}, nil, nil, nil),
	}

	_, err := handler.GetTransactionHistoryDetail(context.Background(), &historyv1.GetTransactionHistoryDetailRequest{
		Id: "missing-id",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if status.Code(err) != codes.NotFound {
		t.Fatalf("expected not found, got %v", status.Code(err))
	}
}
