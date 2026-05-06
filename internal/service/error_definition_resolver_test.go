package service

import (
	"context"
	"errors"
	"testing"
	"time"

	coreerrors "github.com/yogayulanda/go-core/errors"
	"github.com/yogayulanda/transaction-history-service/internal/domain"
)

type fakeErrorDefinitionRepo struct {
	rows []domain.ErrorDefinition
	err  error
}

func (f *fakeErrorDefinitionRepo) ListActiveErrorDefinitions(context.Context) ([]domain.ErrorDefinition, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.rows, nil
}

func TestDBErrorDefinitionResolver_LoadAndResolveSuccess(t *testing.T) {
	repo := &fakeErrorDefinitionRepo{rows: []domain.ErrorDefinition{
		{
			ErrorCode:   "TRH-VAL-001",
			UserMessage: "permintaan tidak valid",
			DetailsJSON: `[{"field":"user_id","reason":"required"}]`,
			IsActive:    true,
			UpdatedAt:   time.Now().UTC(),
		},
	}}

	resolver := NewDBErrorDefinitionResolver(repo, nil)
	if err := resolver.Load(context.Background()); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	userMessage, details, ok := resolver.Resolve("TRH-VAL-001")
	if !ok {
		t.Fatal("expected definition to exist")
	}
	if userMessage != "permintaan tidak valid" {
		t.Fatalf("unexpected user message: %q", userMessage)
	}
	if len(details) != 1 || details[0].Field != "user_id" {
		t.Fatalf("unexpected details: %+v", details)
	}
}

func TestDBErrorDefinitionResolver_LoadSkipsMalformedDetails(t *testing.T) {
	repo := &fakeErrorDefinitionRepo{rows: []domain.ErrorDefinition{
		{
			ErrorCode:   "TRH-REC-001",
			UserMessage: "internal server error",
			DetailsJSON: `{not-json}`,
			IsActive:    true,
			UpdatedAt:   time.Now().UTC(),
		},
	}}

	resolver := NewDBErrorDefinitionResolver(repo, nil)
	if err := resolver.Load(context.Background()); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	_, details, ok := resolver.Resolve("TRH-REC-001")
	if !ok {
		t.Fatal("expected definition to exist")
	}
	if len(details) != 0 {
		t.Fatalf("expected malformed details to be ignored, got %+v", details)
	}
}

func TestDBErrorDefinitionResolver_RefreshFailureKeepsPreviousCache(t *testing.T) {
	repo := &fakeErrorDefinitionRepo{rows: []domain.ErrorDefinition{
		{
			ErrorCode:   "TRH-VAL-002",
			UserMessage: "reference duplicate",
			DetailsJSON: `[]`,
			IsActive:    true,
			UpdatedAt:   time.Now().UTC(),
		},
	}}

	resolver := NewDBErrorDefinitionResolver(repo, nil)
	if err := resolver.Load(context.Background()); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	repo.err = errors.New("db unavailable")
	if err := resolver.loadWithOperation(context.Background(), "error_definition_refresh"); err == nil {
		t.Fatal("expected refresh error")
	}

	msg, _, ok := resolver.Resolve("TRH-VAL-002")
	if !ok {
		t.Fatal("expected previous cache to remain available")
	}
	if msg != "reference duplicate" {
		t.Fatalf("unexpected cached message: %q", msg)
	}
}

func TestParseDefinitionDetails_EmptyAndInvalid(t *testing.T) {
	details, err := parseDefinitionDetails("")
	if err != nil {
		t.Fatalf("expected nil error for empty input, got %v", err)
	}
	if len(details) != 0 {
		t.Fatalf("expected no details, got %+v", details)
	}

	_, err = parseDefinitionDetails("{bad}")
	if err == nil {
		t.Fatal("expected parse error")
	}

	details, err = parseDefinitionDetails(`[{"field":"amount","reason":"must be >= 0"}]`)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(details) != 1 || details[0] != (coreerrors.Detail{Field: "amount", Reason: "must be >= 0"}) {
		t.Fatalf("unexpected parsed details: %+v", details)
	}
}
