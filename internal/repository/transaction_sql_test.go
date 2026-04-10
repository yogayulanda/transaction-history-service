package repository

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/yogayulanda/transaction-history-service/internal/domain"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

func newMockRepository(t *testing.T) (*transactionRepository, sqlmock.Sqlmock, func()) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}

	gdb, err := gorm.Open(sqlserver.New(sqlserver.Config{Conn: sqlDB}), &gorm.Config{
		TranslateError: true,
	})
	if err != nil {
		t.Fatalf("gorm.Open: %v", err)
	}

	cleanup := func() {
		_ = sqlDB.Close()
	}

	return &transactionRepository{db: gdb, sqlDB: sqlDB}, mock, cleanup
}

func TestCreate_MapsDuplicateReferenceIDToDuplicatedKey(t *testing.T) {
	repo, mock, cleanup := newMockRepository(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "dbo"."transaction_histories"`)).
		WillReturnError(gorm.ErrDuplicatedKey)
	mock.ExpectRollback()

	_, err := repo.Create(context.Background(), domain.CreateTransactionHistoryInput{
		UserID:           "user-1",
		ReferenceID:      "ref-1",
		ProductGroup:     "transfer",
		ProductType:      "transfer_internal",
		TransactionRoute: "internal",
		Channel:          "app-one",
		Direction:        "debit",
		Amount:           100,
		Fee:              0,
		TotalAmount:      100,
		Currency:         "IDR",
		StatusCode:       "SUCCESS",
		SourceService:    "trxFinance",
		MetadataJSON:     `{}`,
	})
	if !errors.Is(err, gorm.ErrDuplicatedKey) {
		t.Fatalf("expected duplicated key error, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestFindDetailByID_ReturnsNotFound(t *testing.T) {
	repo, mock, cleanup := newMockRepository(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "dbo"."transaction_histories" WHERE id = @p1 ORDER BY "transaction_histories"."id" OFFSET 0 ROW FETCH NEXT 1 ROWS ONLY`)).
		WithArgs("missing-id").
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	_, err := repo.FindDetailByID(context.Background(), "missing-id")
	if !errors.Is(err, domain.ErrTransactionNotFound) {
		t.Fatalf("expected not found error, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestListByUser_ReturnsHasMoreAndMapsChannel(t *testing.T) {
	repo, mock, cleanup := newMockRepository(t)
	defer cleanup()

	now := time.Date(2026, 4, 9, 1, 2, 3, 0, time.UTC)
	rows := sqlmock.NewRows([]string{
		"id", "user_id", "reference_id", "external_ref_id", "product_group", "product_type",
		"transaction_route", "channel", "direction", "amount", "fee", "total_amount",
		"currency", "status_code", "error_code", "error_message", "source_service",
		"transaction_time", "created_at", "updated_at",
	}).
		AddRow("tx-3", "user-1", "ref-3", sql.NullString{}, "transfer", "transfer_internal", "internal", "app-one", "debit", 300, 0, 300, "IDR", "SUCCESS", "", "", "trxFinance", now.Add(2*time.Minute), now, now).
		AddRow("tx-2", "user-1", "ref-2", sql.NullString{}, "transfer", "transfer_internal", "internal", "app-two", "debit", 200, 0, 200, "IDR", "SUCCESS", "", "", "trxFinance", now.Add(time.Minute), now, now).
		AddRow("tx-1", "user-1", "ref-1", sql.NullString{}, "transfer", "transfer_internal", "internal", "app-one", "debit", 100, 0, 100, "IDR", "SUCCESS", "", "", "trxFinance", now, now, now)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "dbo"."transaction_histories" WHERE user_id = @p1 ORDER BY transaction_time DESC, id DESC OFFSET 0 ROW FETCH NEXT 3 ROWS ONLY`)).
		WithArgs("user-1").
		WillReturnRows(rows)

	items, hasMore, err := repo.ListByUser(context.Background(), domain.ListUserHistoryFilter{
		UserID:   "user-1",
		PageSize: 2,
		Offset:   0,
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !hasMore {
		t.Fatal("expected hasMore=true")
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[1].Channel != "app-two" {
		t.Fatalf("expected channel app-two, got %q", items[1].Channel)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
