package service

import (
	"errors"
	"testing"

	coreerrors "github.com/yogayulanda/go-core/errors"
)

type stubErrorDefinitionResolver struct {
	lookup map[string]resolvedErrorDefinition
}

func (s stubErrorDefinitionResolver) Resolve(code string) (string, []coreerrors.Detail, bool) {
	if s.lookup == nil {
		return "", nil, false
	}
	v, ok := s.lookup[code]
	if !ok {
		return "", nil, false
	}
	details := make([]coreerrors.Detail, len(v.details))
	copy(details, v.details)
	return v.userMessage, details, true
}

func TestNewInvalidInputError_UsesTRHValidationTaxonomy(t *testing.T) {
	appErr := NewInvalidInputError("user_id is required", nil)
	if appErr.Code != coreerrors.CodeInvalidRequest {
		t.Fatalf("expected code %s, got %s", coreerrors.CodeInvalidRequest, appErr.Code)
	}
	if appErr.FormatCode() != "TRH-VAL-001" {
		t.Fatalf("expected format code TRH-VAL-001, got %s", appErr.FormatCode())
	}

	grpcErr := coreerrors.ToGRPC(appErr)
	code, domain, category, number, userMessage, isCore := coreerrors.ErrorInfoFromGRPC(grpcErr)
	if !isCore {
		t.Fatal("expected go-core error info")
	}
	if code != coreerrors.CodeInvalidRequest {
		t.Fatalf("expected grpc code INVALID_REQUEST, got %s", code)
	}
	if domain != "TRH" || category != "VAL" || number != "001" {
		t.Fatalf("unexpected taxonomy metadata: domain=%s category=%s number=%s", domain, category, number)
	}
	if userMessage != "Permintaan tidak valid, silakan periksa kembali data yang dimasukkan." {
		t.Fatalf("unexpected user message: %q", userMessage)
	}
}

func TestNewDuplicateReferenceIDError_UsesTRHValidationTaxonomy(t *testing.T) {
	appErr := NewDuplicateReferenceIDError(nil)
	if appErr.Code != coreerrors.CodeInvalidRequest {
		t.Fatalf("expected code %s, got %s", coreerrors.CodeInvalidRequest, appErr.Code)
	}
	if appErr.FormatCode() != "TRH-VAL-002" {
		t.Fatalf("expected format code TRH-VAL-002, got %s", appErr.FormatCode())
	}
}

func TestNewNotFoundError_UsesTRHDBTaxonomy(t *testing.T) {
	appErr := NewNotFoundError("transaction history not found", nil)
	if appErr.Code != coreerrors.CodeNotFound {
		t.Fatalf("expected code %s, got %s", coreerrors.CodeNotFound, appErr.Code)
	}
	if appErr.FormatCode() != "TRH-DB-001" {
		t.Fatalf("expected format code TRH-DB-001, got %s", appErr.FormatCode())
	}
}

func TestNewInternalError_UsesTRHRECTaxonomyAndWraps(t *testing.T) {
	root := errors.New("sql timeout")
	appErr := NewInternalError("failed to get user history", root, nil)

	if appErr.Code != coreerrors.CodeInternal {
		t.Fatalf("expected code %s, got %s", coreerrors.CodeInternal, appErr.Code)
	}
	if appErr.FormatCode() != "TRH-REC-001" {
		t.Fatalf("expected format code TRH-REC-001, got %s", appErr.FormatCode())
	}
	if !errors.Is(appErr, root) {
		t.Fatal("expected wrapped root error")
	}
}

func TestNewInvalidInputError_OverridesUserMessageAndDetailsFromResolver(t *testing.T) {
	resolver := stubErrorDefinitionResolver{lookup: map[string]resolvedErrorDefinition{
		"TRH-VAL-001": {
			userMessage: "permintaan tidak valid",
			details:     []coreerrors.Detail{{Field: "user_id", Reason: "harus diisi"}},
		},
	}}

	appErr := NewInvalidInputError("user_id is required", resolver)
	if appErr.UserMessage != "permintaan tidak valid" {
		t.Fatalf("expected dynamic user message, got %q", appErr.UserMessage)
	}
	if len(appErr.Details) != 1 || appErr.Details[0].Field != "user_id" {
		t.Fatalf("unexpected dynamic details: %+v", appErr.Details)
	}
}
