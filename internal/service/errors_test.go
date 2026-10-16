package service

import (
	"context"
	"errors"
	"testing"

	coreerrors "github.com/yogayulanda/go-core/errors"
	"github.com/yogayulanda/transaction-history-service/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func registerServiceErrorCodes(t *testing.T) {
	t.Helper()
	registry := coreerrors.SnapshotRegistry()
	t.Cleanup(func() { coreerrors.RestoreRegistry(registry) })
	if err := RegisterErrorCodes(); err != nil {
		t.Fatalf("register service error codes: %v", err)
	}
}

func TestNewInvalidInputError_UsesCoreValidationContract(t *testing.T) {
	appErr := NewInvalidInputError("user_id is required", coreerrors.Detail{Field: "user_id", Reason: "required"})
	if appErr.Code != "CORE-1000" {
		t.Fatalf("expected CORE-1000, got %s", appErr.Code)
	}
	if len(appErr.Details) != 1 || appErr.Details[0].Field != "user_id" {
		t.Fatalf("unexpected validation details: %+v", appErr.Details)
	}
}

func TestNewDuplicateReferenceIDError_UsesRegisteredCode(t *testing.T) {
	registerServiceErrorCodes(t)
	appErr := NewDuplicateReferenceIDError()
	if appErr.Code != CodeDuplicateReferenceID {
		t.Fatalf("expected %s, got %s", CodeDuplicateReferenceID, appErr.Code)
	}
}

func TestNewNotFoundError_UsesRegisteredCode(t *testing.T) {
	registerServiceErrorCodes(t)
	appErr := NewNotFoundError("transaction history not found")
	if appErr.Code != CodeTransactionNotFound {
		t.Fatalf("expected %s, got %s", CodeTransactionNotFound, appErr.Code)
	}
}

func TestNewInternalError_UsesRegisteredCodeAndWraps(t *testing.T) {
	registerServiceErrorCodes(t)
	root := errors.New("sql timeout")
	appErr := NewInternalError("failed to get user history", root)
	if appErr.Code != CodeInternal {
		t.Fatalf("expected %s, got %s", CodeInternal, appErr.Code)
	}
	if !errors.Is(appErr, root) {
		t.Fatal("expected wrapped root error")
	}
}

func TestErrorCodes_MapToExpectedGRPCStatusesAndMessages(t *testing.T) {
	registerServiceErrorCodes(t)
	provider := NewDBErrorDefinitionResolver(&fakeErrorDefinitionRepo{rows: []domain.ErrorDefinition{
		{ErrorCode: CodeTransactionNotFound, UserMessage: "localized not found"},
		{ErrorCode: "CORE-1000", UserMessage: "localized invalid request"},
	}}, nil)
	if err := provider.Load(context.Background()); err != nil {
		t.Fatalf("load message provider: %v", err)
	}
	coreerrors.SetMessageProvider(provider)

	tests := []struct {
		name        string
		appErr      *coreerrors.AppError
		grpcCode    codes.Code
		errorCode   string
		message     string
		fieldDetail bool
	}{
		{
			name:      "duplicate reference",
			appErr:    NewDuplicateReferenceIDError(),
			grpcCode:  codes.InvalidArgument,
			errorCode: CodeDuplicateReferenceID,
			message:   "Transaksi dengan referensi ini sudah pernah diproses.",
		},
		{
			name:      "not found uses provider",
			appErr:    NewNotFoundError("transaction history not found"),
			grpcCode:  codes.NotFound,
			errorCode: CodeTransactionNotFound,
			message:   "localized not found",
		},
		{
			name:      "internal error",
			appErr:    NewInternalError("internal", nil),
			grpcCode:  codes.Internal,
			errorCode: CodeInternal,
			message:   "Sistem sedang sibuk. Silakan coba beberapa saat lagi.",
		},
		{
			name:        "validation uses provider and field details",
			appErr:      NewInvalidInputError("invalid", coreerrors.Detail{Field: "user_id", Reason: "required"}),
			grpcCode:    codes.InvalidArgument,
			errorCode:   "CORE-1000",
			message:     "localized invalid request",
			fieldDetail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grpcErr := coreerrors.ToGRPC(tt.appErr)
			if status.Code(grpcErr) != tt.grpcCode {
				t.Fatalf("expected gRPC %s, got %s", tt.grpcCode, status.Code(grpcErr))
			}
			if got := status.Convert(grpcErr).Message(); got != tt.message {
				t.Fatalf("expected message %q, got %q", tt.message, got)
			}
			info, ok := coreerrors.InfoFromGRPC(grpcErr)
			if !ok || info.Code != tt.errorCode {
				t.Fatalf("expected reason %s, got %+v (present=%v)", tt.errorCode, info, ok)
			}
			if tt.fieldDetail {
				details := coreerrors.DetailsFromGRPC(grpcErr)
				if len(details) != 1 || details[0] != (coreerrors.Detail{Field: "user_id", Reason: "required"}) {
					t.Fatalf("unexpected field details: %+v", details)
				}
			}
		})
	}
}
