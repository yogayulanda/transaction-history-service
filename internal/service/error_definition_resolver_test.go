package service

import (
	"context"
	"errors"
	"testing"
	"time"

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

func TestDBErrorDefinitionResolver_LoadAndResolveMessage(t *testing.T) {
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

	userMessage, ok := resolver.Message(context.Background(), " trh-val-001 ")
	if !ok {
		t.Fatal("expected definition to exist")
	}
	if userMessage != "permintaan tidak valid" {
		t.Fatalf("unexpected user message: %q", userMessage)
	}
}

func TestDBErrorDefinitionResolver_EmptyMessageIsNotProvided(t *testing.T) {
	repo := &fakeErrorDefinitionRepo{rows: []domain.ErrorDefinition{
		{
			ErrorCode:   "TRH-REC-001",
			UserMessage: "  ",
			DetailsJSON: `{not-json}`,
			IsActive:    true,
			UpdatedAt:   time.Now().UTC(),
		},
	}}

	resolver := NewDBErrorDefinitionResolver(repo, nil)
	if err := resolver.Load(context.Background()); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if message, ok := resolver.Message(context.Background(), "TRH-REC-001"); ok || message != "" {
		t.Fatalf("expected no message for empty user message, got %q (present=%v)", message, ok)
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

	msg, ok := resolver.Message(context.Background(), "TRH-VAL-002")
	if !ok {
		t.Fatal("expected previous cache to remain available")
	}
	if msg != "reference duplicate" {
		t.Fatalf("unexpected cached message: %q", msg)
	}
}
