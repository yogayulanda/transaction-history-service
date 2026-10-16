package repository

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/microsoft/go-mssqldb"
	"github.com/yogayulanda/transaction-history-service/internal/domain"
)

func newHistoryRepositoryMock(t *testing.T) (*historyRepository, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	return &historyRepository{db: db}, mock, func() { _ = db.Close() }
}

func anyArgs(count int) []driver.Value {
	args := make([]driver.Value, count)
	for i := range args {
		args[i] = sqlmock.AnyArg()
	}
	return args
}

func TestApplyStatusEventOrderingAndDeduplication(t *testing.T) {
	tests := []struct {
		name            string
		currentStatus   string
		currentVersion  int64
		eventVersion    int64
		duplicateID     bool
		insertEventErr  string
		wantResult      domain.ApplyResult
		wantHeaderWrite bool
	}{
		{name: "first event version three", eventVersion: 3, wantResult: domain.Created, wantHeaderWrite: true},
		{name: "sequence 1 to 2", currentStatus: "PENDING", currentVersion: 1, eventVersion: 2, wantResult: domain.Updated, wantHeaderWrite: true},
		{name: "sequence 2 to 1 is stale", currentStatus: "PENDING", currentVersion: 2, eventVersion: 1, wantResult: domain.Stale},
		{name: "sequence 1 then duplicate 1", duplicateID: true, eventVersion: 1, wantResult: domain.Duplicate},
		{name: "final to non-final regression is applied", currentStatus: "SUCCESS", currentVersion: 1, eventVersion: 2, wantResult: domain.Updated, wantHeaderWrite: true},
		{name: "version conflict is acknowledged", eventVersion: 1, insertEventErr: "UX_thse_status_version", wantResult: domain.Duplicate, wantHeaderWrite: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo, mock, cleanup := newHistoryRepositoryMock(t)
			defer cleanup()
			mock.ExpectBegin()
			if test.duplicateID {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT 1 FROM dbo.transaction_history_status_events WHERE event_id = @p1`)).
					WithArgs("event-1").WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
				mock.ExpectCommit()
			} else {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT 1 FROM dbo.transaction_history_status_events WHERE event_id = @p1`)).
					WithArgs("event-1").WillReturnRows(sqlmock.NewRows([]string{"1"}))
				if test.currentVersion == 0 {
					mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, status, status_version FROM dbo.transaction_histories WITH (UPDLOCK, HOLDLOCK) WHERE transaction_id = @p1`)).
						WithArgs("tx-1").WillReturnRows(sqlmock.NewRows([]string{"id", "status", "status_version"}))
				} else {
					mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, status, status_version FROM dbo.transaction_histories WITH (UPDLOCK, HOLDLOCK) WHERE transaction_id = @p1`)).
						WithArgs("tx-1").WillReturnRows(sqlmock.NewRows([]string{"id", "status", "status_version"}).AddRow(41, test.currentStatus, test.currentVersion))
				}
				if test.wantHeaderWrite {
					if test.currentVersion == 0 {
						mock.ExpectQuery(`INSERT INTO dbo\.transaction_histories .*OUTPUT INSERTED\.id VALUES`).WithArgs(anyArgs(25)...).
							WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(41))
					} else {
						mock.ExpectExec(`UPDATE dbo\.transaction_histories SET`).WithArgs(anyArgs(25)...).WillReturnResult(sqlmock.NewResult(0, 1))
						mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM dbo.transaction_history_details WHERE transaction_history_id = @p1`)).
							WithArgs(int64(41)).WillReturnResult(sqlmock.NewResult(0, 1))
					}
				}
				insertEvent := mock.ExpectExec(`INSERT INTO dbo\.transaction_history_status_events`)
				insertEvent.WithArgs(anyArgs(9)...)
				if test.insertEventErr != "" {
					insertEvent.WillReturnError(&mssql.Error{Number: 2627, Message: "duplicate key UX_thse_status_version"})
					mock.ExpectRollback()
				} else {
					insertEvent.WillReturnResult(sqlmock.NewResult(1, 1))
					mock.ExpectCommit()
				}
			}

			header := &domain.Transaction{
				TransactionID: "tx-1", AppName: "app_a", UserID: "user-1", ProductCategory: "transfer",
				ProductType: "transfer_internal", Provider: "bca", Direction: "debit",
				Amount: "10.00", FeeAmount: "0.00", TotalAmount: "10.00", Currency: "IDR",
				Status: "PENDING", StatusVersion: test.eventVersion, SourceSystem: "sample-producer",
				TransactionAt: time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC),
			}
			got, err := repo.ApplyStatusEvent(context.Background(), header, nil, domain.StatusEvent{
				EventID: "event-1", EventType: "transaction.status.v1", StatusVersion: test.eventVersion,
				ToStatus: header.Status, ChangedByType: "system", ChangedAt: header.TransactionAt,
			})
			if err != nil || got != test.wantResult {
				t.Fatalf("ApplyStatusEvent() = %q, %v; want %q", got, err, test.wantResult)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("unmet SQL expectations: %v", err)
			}
		})
	}
}

func TestUniqueIndexErrorRequiresSQLServerUniqueNumberAndIndexName(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		indexName string
		want      bool
	}{
		{name: "event ID duplicate value", err: mssql.Error{Number: 2601, Message: "Cannot insert duplicate key row with unique index 'UX_thse_event_id'."}, indexName: "UX_thse_event_id", want: true},
		{name: "wrapped status version duplicate value", err: fmt.Errorf("insert event: %w", mssql.Error{Number: 2627, Message: "Cannot insert duplicate key row with unique index 'UX_thse_status_version'."}), indexName: "UX_thse_status_version", want: true},
		{name: "generic text is not enough", err: errors.New("duplicate UX_thse_event_id"), indexName: "UX_thse_event_id"},
		{name: "wrong SQL error number", err: mssql.Error{Number: 50000, Message: "duplicate UX_thse_event_id"}, indexName: "UX_thse_event_id"},
		{name: "wrong index name", err: mssql.Error{Number: 2627, Message: "duplicate UX_thse_other"}, indexName: "UX_thse_event_id"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := isUniqueIndexError(test.err, test.indexName); got != test.want {
				t.Fatalf("isUniqueIndexError() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestInsertSnapshotUsesDecimalStrings(t *testing.T) {
	if !strings.Contains(snapshotColumns, "amount, fee_amount, total_amount") {
		t.Fatal("snapshot SQL columns do not include all money fields")
	}
	args := snapshotArgs(&domain.Transaction{ID: 42, TransactionID: "tx-1", Amount: "1.25", FeeAmount: "0.10", TotalAmount: "1.35"})
	if got := len(args); got != 25 {
		t.Fatalf("snapshotArgs() = %d values, want 25", got)
	}
	if args[0] != "tx-1" {
		t.Fatalf("first insert argument = %v, want external transaction ID", args[0])
	}
}
