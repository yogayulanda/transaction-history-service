package grpc

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"

	coreerrors "github.com/yogayulanda/go-core/errors"
	"github.com/yogayulanda/go-core/logger"
	"github.com/yogayulanda/go-core/security"
	historyv1 "github.com/yogayulanda/transaction-history-service/gen/go/history/v1"
	"github.com/yogayulanda/transaction-history-service/internal/catalog"
	"github.com/yogayulanda/transaction-history-service/internal/domain"
	"github.com/yogayulanda/transaction-history-service/internal/identity"
	"github.com/yogayulanda/transaction-history-service/internal/pii"
	"github.com/yogayulanda/transaction-history-service/internal/service"
	"google.golang.org/grpc/status"
)

type handlerRepository struct {
	owner          domain.Owner
	item           domain.TransactionSnapshot
	searchCriteria domain.SearchCriteria
	searchItems    []domain.TransactionSnapshot
	err            error
}

func (r *handlerRepository) FindByOwner(_ context.Context, owner domain.Owner, _ string) (*domain.TransactionSnapshot, *domain.Detail, error) {
	r.owner = owner
	if r.err != nil {
		return nil, nil, r.err
	}
	return &r.item, nil, nil
}

func (r *handlerRepository) ListByOwner(_ context.Context, owner domain.Owner, _ domain.TransactionFilter) ([]domain.TransactionSnapshot, bool, error) {
	r.owner = owner
	if r.err != nil {
		return nil, false, r.err
	}
	return []domain.TransactionSnapshot{r.item}, false, nil
}

func (r *handlerRepository) Search(_ context.Context, criteria domain.SearchCriteria) ([]domain.TransactionSnapshot, bool, error) {
	r.searchCriteria = criteria
	return r.searchItems, false, nil
}

func (r *handlerRepository) ListStatusHistory(context.Context, string) ([]domain.StatusEvent, error) {
	return []domain.StatusEvent{{ToStatus: "SUCCESS", ChangedAt: time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC)}}, nil
}

func newTestHistoryHandler(repo *handlerRepository) *HistoryHandler {
	validationCatalog := catalog.DefaultStaticCatalog("", "")
	return &HistoryHandler{
		service:  service.NewHistoryService(repo),
		identity: identity.NewClaimsResolver(validationCatalog),
		catalog:  validationCatalog,
	}
}

func claimsContext() context.Context {
	return security.Inject(context.Background(), &security.Claims{Subject: "user-1", Attributes: map[string]string{"app_name": "app_a"}})
}

func TestHistoryHandlerUsesClaimsOwner(t *testing.T) {
	repo := &handlerRepository{item: domain.TransactionSnapshot{ID: "tx-1", TransactionAt: time.Now().UTC()}}
	handler := newTestHistoryHandler(repo)
	_, err := handler.ListTransactions(claimsContext(), &historyv1.ListTransactionsRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if repo.owner != (domain.Owner{AppName: "app_a", UserID: "user-1"}) {
		t.Fatalf("repository owner = %+v", repo.owner)
	}
}

func TestHistoryHandlerMissingClaimsReturnsUnauthenticated(t *testing.T) {
	handler := newTestHistoryHandler(&handlerRepository{})
	_, err := handler.ListTransactions(context.Background(), &historyv1.ListTransactionsRequest{})
	if status.Code(err).String() != "Unauthenticated" {
		t.Fatalf("ListTransactions() status = %v", status.Code(err))
	}
}

func TestHistoryHandlerRejectsUnknownStatusFilter(t *testing.T) {
	handler := newTestHistoryHandler(&handlerRepository{})
	_, err := handler.ListTransactions(claimsContext(), &historyv1.ListTransactionsRequest{Status: "PROCESSING"})
	info, ok := coreerrors.InfoFromGRPC(err)
	if !ok || info.Code != "CORE-1000" {
		t.Fatalf("ListTransactions() error info = %+v, present=%v", info, ok)
	}
}

func TestHistoryHandlerUnknownAndOtherOwnerHaveSameNotFound(t *testing.T) {
	unknown := newTestHistoryHandler(&handlerRepository{err: domain.ErrTransactionNotFound})
	otherOwner := newTestHistoryHandler(&handlerRepository{err: domain.ErrTransactionNotFound})
	_, unknownErr := unknown.GetTransaction(claimsContext(), &historyv1.GetTransactionRequest{TransactionId: "tx-unknown"})
	_, otherErr := otherOwner.GetTransaction(claimsContext(), &historyv1.GetTransactionRequest{TransactionId: "tx-other"})
	if status.Code(unknownErr) != status.Code(otherErr) || status.Code(unknownErr).String() != "NotFound" {
		t.Fatalf("not-found responses differ: %v vs %v", unknownErr, otherErr)
	}
	unknownInfo, _ := coreerrors.InfoFromGRPC(unknownErr)
	otherInfo, _ := coreerrors.InfoFromGRPC(otherErr)
	if unknownInfo.Code != "CORE-1004" || otherInfo.Code != unknownInfo.Code || status.Convert(unknownErr).Message() != status.Convert(otherErr).Message() {
		t.Fatalf("not-found response bodies differ: %+v / %+v", unknownInfo, otherInfo)
	}
}

func TestParseDateRangeUsesJakartaInclusiveEnd(t *testing.T) {
	now := time.Date(2026, 10, 16, 16, 30, 0, 0, time.UTC)
	start, end, err := parseDateRange("2026-10-16", "2026-10-16", now)
	if err != nil {
		t.Fatal(err)
	}
	if !start.Equal(time.Date(2026, 10, 15, 17, 0, 0, 0, time.UTC)) || !end.Equal(time.Date(2026, 10, 16, 17, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected UTC date bounds: %s to %s", start, end)
	}
}

func TestCursorRoundTripAndRejectsInvalid(t *testing.T) {
	item := domain.TransactionSnapshot{ID: "tx-1", TransactionAt: time.Date(2026, 10, 16, 1, 2, 3, 4, time.UTC)}
	encoded := encodeCursor(item)
	when, id, err := decodeCursor(encoded)
	if err != nil || id != item.ID || !when.Equal(item.TransactionAt) {
		t.Fatalf("cursor roundtrip = %s, %q, %v", when, id, err)
	}
	if _, _, err := decodeCursor("not-base64"); err == nil {
		t.Fatal("expected malformed cursor error")
	}
	invalid, _ := json.Marshal(pageCursor{ID: "tx-1"})
	if _, _, err := decodeCursor(base64.RawURLEncoding.EncodeToString(invalid)); err == nil {
		t.Fatal("expected zero timestamp cursor error")
	}
}

func TestDetailOneofMapsEveryProduct(t *testing.T) {
	tests := []struct {
		productType string
		fields      map[string]any
		present     bool
	}{
		{"transfer_internal", map[string]any{"bank_code": "001"}, true},
		{"pulsa", map[string]any{"serial_number": "S1"}, true},
		{"data_package", map[string]any{"product_name": "Data"}, true},
		{"pln_prepaid", map[string]any{"token": "123"}, true},
		{"pln_postpaid", map[string]any{"period": "2026-10"}, true},
		{"bpjs_kesehatan", map[string]any{"period_months": int64(1)}, true},
		{"cash_in", map[string]any{"code_masked": "****123"}, true},
	}
	for _, test := range tests {
		got := toProtoDetail(context.Background(), domain.TransactionSnapshot{ProductType: test.productType}, &domain.Detail{Fields: test.fields}, nil, catalog.DefaultStaticCatalog("", ""), nil)
		if (got.GetDetail() != nil) != test.present {
			t.Errorf("%s detail oneof present=%v", test.productType, got.GetDetail() != nil)
		}
	}
}

func TestCursorEncodingIsStable(t *testing.T) {
	item := domain.TransactionSnapshot{ID: "tx-2", TransactionAt: time.Date(2026, 10, 16, 1, 0, 0, 0, time.UTC)}
	first := encodeCursor(item)
	second := encodeCursor(item)
	if first != second || first == "" {
		t.Fatal("cursor should be stable and non-empty")
	}
}

type auditLogger struct {
	events []logger.EventLog
}

func (*auditLogger) Info(context.Context, string, ...logger.Field)  {}
func (*auditLogger) Error(context.Context, string, ...logger.Field) {}
func (*auditLogger) Debug(context.Context, string, ...logger.Field) {}
func (*auditLogger) Warn(context.Context, string, ...logger.Field)  {}
func (*auditLogger) LogService(context.Context, logger.ServiceLog)  {}
func (*auditLogger) LogDB(context.Context, logger.DBLog)            {}
func (l *auditLogger) LogEvent(_ context.Context, event logger.EventLog) {
	l.events = append(l.events, event)
}
func (*auditLogger) LogTransaction(context.Context, logger.TransactionLog) {}
func (l *auditLogger) WithComponent(string) logger.Logger                  { return l }

func TestBackofficeSearchHashesDestinationAndAuditsWithoutRawNumber(t *testing.T) {
	validationCatalog := catalog.DefaultStaticCatalog("", "")
	hasher, err := pii.NewHasher([]byte(strings.Repeat("k", 32)))
	if err != nil {
		t.Fatal(err)
	}
	repo := &handlerRepository{searchItems: []domain.TransactionSnapshot{{ID: "tx-1", Status: "SUCCESS"}}}
	log := &auditLogger{}
	handler := &BackofficeHandler{
		service: service.NewHistoryService(repo),
		catalog: validationCatalog,
		hasher:  hasher,
		role:    "backoffice",
		log:     log,
	}
	ctx := security.Inject(context.Background(), &security.Claims{Role: "backoffice"})
	response, err := handler.SearchTransactions(ctx, &historyv1.SearchTransactionsRequest{
		DestinationNo:     "+6281234567890",
		DestinationNoType: "phone",
	})
	if err != nil {
		t.Fatal(err)
	}
	wantHash := hasher.Hash("phone", "081234567890")
	if repo.searchCriteria.DestinationHash != wantHash || repo.searchCriteria.DestinationType != "phone" {
		t.Fatalf("search received unexpected criteria: %+v", repo.searchCriteria)
	}
	if len(response.Items) != 1 || len(log.events) != 1 {
		t.Fatalf("unexpected response/audit count: items=%d events=%d", len(response.Items), len(log.events))
	}
	event := log.events[0]
	if event.Category != "audit" || event.Action != "transaction_search" || event.Outcome != "success" {
		t.Fatalf("unexpected audit event: %+v", event)
	}
	encoded, _ := json.Marshal(event.Metadata)
	if strings.Contains(string(encoded), "+6281234567890") || strings.Contains(string(encoded), "081234567890") {
		t.Fatal("audit metadata contained a raw destination number")
	}
	if event.Metadata["destination_hash_prefix"] != wantHash[:8] || event.Metadata["result_count"] != 1 {
		t.Fatalf("unexpected audit metadata: %+v", event.Metadata)
	}
}

func TestBackofficeSearchRequiresRole(t *testing.T) {
	validationCatalog := catalog.DefaultStaticCatalog("", "")
	handler := &BackofficeHandler{service: service.NewHistoryService(&handlerRepository{}), catalog: validationCatalog, role: "backoffice"}
	ctx := security.Inject(context.Background(), &security.Claims{Role: "customer"})
	_, err := handler.SearchTransactions(ctx, &historyv1.SearchTransactionsRequest{TransactionId: "tx-1"})
	if status.Code(err).String() != "PermissionDenied" {
		t.Fatalf("SearchTransactions() status = %v", status.Code(err))
	}
}
