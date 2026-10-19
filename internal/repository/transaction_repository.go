package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	mssql "github.com/microsoft/go-mssqldb"
	"github.com/yogayulanda/go-core/logger"
	"github.com/yogayulanda/transaction-history-service/internal/domain"
)

const snapshotColumns = `id, transaction_id, app_name, user_id, product_category, product_type, provider, partner,
direction, denomination, amount, fee_amount, total_amount, currency, status, status_version, response_code,
source_account_masked, destination_name, destination_no_type, destination_no_masked, destination_no_hash,
reference_no, source_system, transaction_at, completed_at`

type historyRepository struct {
	db  *sql.DB
	log logger.Logger
}

func NewTransactionRepository(db *sql.DB, log logger.Logger) domain.TransactionRepository {
	return &historyRepository{db: db, log: log}
}

func (r *historyRepository) ApplyStatusEvent(
	ctx context.Context,
	transaction *domain.Transaction,
	detail *domain.Detail,
	event domain.StatusEvent,
) (domain.ApplyResult, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer func() { _ = tx.Rollback() }()

	var exists int
	err = tx.QueryRowContext(ctx,
		`SELECT 1 FROM dbo.transaction_history_status_events WHERE event_id = @p1`, event.EventID,
	).Scan(&exists)
	if err == nil {
		if err := tx.Commit(); err != nil {
			return "", err
		}
		return domain.Duplicate, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}

	var previousStatus string
	var previousVersion int64
	var transactionHistoryID int64
	err = tx.QueryRowContext(ctx,
		`SELECT id, status, status_version FROM dbo.transaction_histories WITH (UPDLOCK, HOLDLOCK) WHERE transaction_id = @p1`,
		transaction.TransactionID,
	).Scan(&transactionHistoryID, &previousStatus, &previousVersion)
	newTransaction := errors.Is(err, sql.ErrNoRows)
	result := domain.Created
	switch {
	case newTransaction:
		insertedID, err := insertSnapshot(ctx, tx, transaction)
		if err != nil {
			return "", err
		}
		transaction.ID = insertedID
		transactionHistoryID = insertedID
	case err != nil:
		return "", err
	default:
		transaction.ID = transactionHistoryID
	}
	if !newTransaction {
		transaction.ID = transactionHistoryID
		switch {
		case event.StatusVersion > previousVersion:
			result = domain.Updated
			if domain.IsFinalStatus(previousStatus) && !domain.IsFinalStatus(transaction.Status) {
				r.emit(ctx, "status_regression")
			}
			if err := updateSnapshot(ctx, tx, transaction); err != nil {
				return "", err
			}
		case event.StatusVersion <= previousVersion:
			result = domain.Stale
		}
	}

	if result == domain.Created || result == domain.Updated {
		if detail == nil || len(detail.Fields) == 0 {
			if result == domain.Updated {
				if _, err := tx.ExecContext(ctx,
					`DELETE FROM dbo.transaction_history_details WHERE transaction_history_id = @p1`, transaction.ID,
				); err != nil {
					return "", err
				}
			}
		} else if err := upsertDetail(ctx, tx, transaction.ID, detail); err != nil {
			return "", err
		}
	}
	if event.FromStatus == "" {
		event.FromStatus = previousStatus
	}
	if err := insertStatusEvent(ctx, tx, transaction.ID, event); err != nil {
		if isUniqueIndexError(err, "UX_thse_event_id") {
			r.emit(ctx, "duplicate_event")
			return domain.Duplicate, nil
		}
		if isUniqueIndexError(err, "UX_thse_status_version") {
			r.emit(ctx, "version_conflict")
			return domain.Duplicate, nil
		}
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", err
	}
	return result, nil
}

func (r *historyRepository) FindByOwner(
	ctx context.Context,
	owner domain.Owner,
	transactionID string,
) (*domain.Transaction, *domain.Detail, error) {
	header, err := scanSnapshot(r.db.QueryRowContext(ctx,
		`SELECT `+snapshotColumns+` FROM dbo.transaction_histories WHERE app_name = @p1 AND user_id = @p2 AND transaction_id = @p3`,
		owner.AppName, owner.UserID, transactionID,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil, domain.ErrTransactionNotFound
	}
	if err != nil {
		return nil, nil, err
	}
	detailValue, err := readDetail(ctx, r.db, header.ID)
	if err != nil {
		return nil, nil, err
	}
	return header, detailValue, nil
}

func (r *historyRepository) ListByOwner(
	ctx context.Context,
	owner domain.Owner,
	filter domain.TransactionFilter,
) ([]domain.Transaction, bool, error) {
	where := []string{"app_name = @p1", "user_id = @p2"}
	args := []any{owner.AppName, owner.UserID}
	appendFilter := func(column string, value any) {
		args = append(args, value)
		where = append(where, fmt.Sprintf("%s = @p%d", column, len(args)))
	}
	if filter.StartDate != nil {
		args = append(args, *filter.StartDate)
		where = append(where, fmt.Sprintf("transaction_at >= @p%d", len(args)))
	}
	if filter.EndDate != nil {
		args = append(args, *filter.EndDate)
		where = append(where, fmt.Sprintf("transaction_at < @p%d", len(args)))
	}
	if filter.ProductCategory != "" {
		appendFilter("product_category", filter.ProductCategory)
	}
	if filter.ProductType != "" {
		appendFilter("product_type", filter.ProductType)
	}
	if filter.Status != "" {
		appendFilter("status", filter.Status)
	}
	if filter.CursorTime != nil {
		args = append(args, *filter.CursorTime)
		timeParam := fmt.Sprintf("@p%d", len(args))
		args = append(args, filter.CursorID)
		idParam := fmt.Sprintf("@p%d", len(args))
		where = append(where, fmt.Sprintf("(transaction_at < %s OR (transaction_at = %s AND id < %s))", timeParam, timeParam, idParam))
	}
	return r.listSnapshots(ctx, where, args, filter.PageSize)
}

func (r *historyRepository) Search(
	ctx context.Context,
	criteria domain.SearchCriteria,
) ([]domain.Transaction, bool, error) {
	where := make([]string, 0, 8)
	args := make([]any, 0, 12)
	appendFilter := func(column string, value any) {
		args = append(args, value)
		where = append(where, fmt.Sprintf("%s = @p%d", column, len(args)))
	}
	if criteria.TransactionID != "" {
		appendFilter("transaction_id", criteria.TransactionID)
	}
	if criteria.ReferenceNo != "" {
		appendFilter("reference_no", criteria.ReferenceNo)
	}
	if criteria.DestinationHash != "" {
		appendFilter("destination_no_hash", criteria.DestinationHash)
		appendFilter("destination_no_type", criteria.DestinationType)
	}
	if criteria.AppName != "" {
		appendFilter("app_name", criteria.AppName)
	}
	if criteria.UserID != "" {
		appendFilter("user_id", criteria.UserID)
	}
	if criteria.StartDate != nil {
		args = append(args, *criteria.StartDate)
		where = append(where, fmt.Sprintf("transaction_at >= @p%d", len(args)))
	}
	if criteria.EndDate != nil {
		args = append(args, *criteria.EndDate)
		where = append(where, fmt.Sprintf("transaction_at < @p%d", len(args)))
	}
	if criteria.CursorTime != nil {
		args = append(args, *criteria.CursorTime)
		timeParam := fmt.Sprintf("@p%d", len(args))
		args = append(args, criteria.CursorID)
		idParam := fmt.Sprintf("@p%d", len(args))
		where = append(where, fmt.Sprintf("(transaction_at < %s OR (transaction_at = %s AND id < %s))", timeParam, timeParam, idParam))
	}
	return r.listSnapshots(ctx, where, args, criteria.PageSize)
}

func (r *historyRepository) ListStatusHistory(ctx context.Context, transactionID string) ([]domain.StatusEvent, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT event_id, event_type, status_version, from_status, to_status, response_code, changed_by_type, changed_at
FROM dbo.transaction_history_status_events WHERE transaction_history_id =
(SELECT id FROM dbo.transaction_histories WHERE transaction_id = @p1) ORDER BY status_version ASC, id ASC`,
		transactionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	events := make([]domain.StatusEvent, 0)
	for rows.Next() {
		var event domain.StatusEvent
		var fromStatus, responseCode sql.NullString
		if err := rows.Scan(
			&event.EventID, &event.EventType, &event.StatusVersion, &fromStatus, &event.ToStatus,
			&responseCode, &event.ChangedByType, &event.ChangedAt,
		); err != nil {
			return nil, err
		}
		event.FromStatus = fromStatus.String
		event.ResponseCode = responseCode.String
		events = append(events, event)
	}
	return events, rows.Err()
}

func (r *historyRepository) InsertBackfill(
	ctx context.Context,
	transaction *domain.Transaction,
	detail *domain.Detail,
	event domain.StatusEvent,
) (bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer func() { _ = tx.Rollback() }()
	var exists int
	err = tx.QueryRowContext(ctx,
		`SELECT 1 FROM dbo.transaction_histories WITH (UPDLOCK, HOLDLOCK) WHERE transaction_id = @p1`, transaction.TransactionID,
	).Scan(&exists)
	if err == nil {
		return false, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return false, err
	}
	transactionHistoryID, err := insertSnapshot(ctx, tx, transaction)
	if err != nil {
		return false, err
	}
	transaction.ID = transactionHistoryID
	if err := upsertDetail(ctx, tx, transaction.ID, detail); err != nil {
		return false, err
	}
	event.StatusVersion = 0
	if err := insertStatusEvent(ctx, tx, transaction.ID, event); err != nil {
		if isUniqueIndexError(err, "UX_thse_event_id") {
			return false, nil
		}
		return false, err
	}
	if err := tx.Commit(); err != nil {
		return false, err
	}
	return true, nil
}

func (r *historyRepository) listSnapshots(
	ctx context.Context,
	where []string,
	args []any,
	pageSize int,
) ([]domain.Transaction, bool, error) {
	if pageSize <= 0 {
		pageSize = 20
	}
	args = append(args, pageSize+1)
	query := `SELECT ` + snapshotColumns + ` FROM dbo.transaction_histories`
	if len(where) > 0 {
		query += ` WHERE ` + strings.Join(where, " AND ")
	}
	query += fmt.Sprintf(` ORDER BY transaction_at DESC, id DESC OFFSET 0 ROWS FETCH NEXT @p%d ROWS ONLY`, len(args))
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()
	items := make([]domain.Transaction, 0, pageSize+1)
	for rows.Next() {
		item, err := scanSnapshot(rows)
		if err != nil {
			return nil, false, err
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, false, err
	}
	hasMore := len(items) > pageSize
	if hasMore {
		items = items[:pageSize]
	}
	return items, hasMore, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanSnapshot(row rowScanner) (*domain.Transaction, error) {
	var item domain.Transaction
	var referenceNo, partner, denomination, responseCode, sourceAccountMasked sql.NullString
	var destinationName, destinationNoType, destinationNoMasked, destinationNoHash sql.NullString
	var completedAt sql.NullTime
	err := row.Scan(
		&item.ID, &item.TransactionID, &item.AppName, &item.UserID, &item.ProductCategory, &item.ProductType,
		&item.Provider, &partner, &item.Direction, &denomination, &item.Amount, &item.FeeAmount,
		&item.TotalAmount, &item.Currency, &item.Status, &item.StatusVersion, &responseCode,
		&sourceAccountMasked, &destinationName, &destinationNoType, &destinationNoMasked, &destinationNoHash,
		&referenceNo, &item.SourceSystem, &item.TransactionAt, &completedAt,
	)
	if err != nil {
		return nil, err
	}
	item.ReferenceNo = referenceNo.String
	item.Partner = partner.String
	item.Denomination = denomination.String
	item.ResponseCode = responseCode.String
	item.SourceAccountMasked = sourceAccountMasked.String
	item.DestinationName = destinationName.String
	item.DestinationNoType = destinationNoType.String
	item.DestinationNoMasked = destinationNoMasked.String
	item.DestinationNoHash = destinationNoHash.String
	if completedAt.Valid {
		item.CompletedAt = &completedAt.Time
	}
	return &item, nil
}

func insertSnapshot(ctx context.Context, tx *sql.Tx, item *domain.Transaction) (int64, error) {
	query := `INSERT INTO dbo.transaction_histories (transaction_id, app_name, user_id, reference_no,
product_category, product_type, provider, partner, direction, denomination, amount, fee_amount, total_amount,
currency, status, status_version, response_code, source_account_masked, destination_name, destination_no_type,
destination_no_masked, destination_no_hash, source_system, transaction_at, completed_at)
OUTPUT INSERTED.id VALUES (` + placeholders(25) + `)`
	var id int64
	err := tx.QueryRowContext(ctx, query, snapshotArgs(item)...).Scan(&id)
	return id, err
}

func updateSnapshot(ctx context.Context, tx *sql.Tx, item *domain.Transaction) error {
	query := `UPDATE dbo.transaction_histories SET app_name = @p1, user_id = @p2, reference_no = @p3,
product_category = @p4, product_type = @p5, provider = @p6, partner = @p7, direction = @p8,
denomination = @p9, amount = @p10, fee_amount = @p11, total_amount = @p12, currency = @p13,
status = @p14, status_version = @p15, response_code = @p16, source_account_masked = @p17,
destination_name = @p18, destination_no_type = @p19, destination_no_masked = @p20,
destination_no_hash = @p21, source_system = @p22, transaction_at = @p23, completed_at = @p24,
updated_at = SYSUTCDATETIME() WHERE transaction_id = @p25`
	args := snapshotArgs(item)
	args = append(args[1:], item.TransactionID)
	_, err := tx.ExecContext(ctx, query, args...)
	return err
}

func snapshotArgs(item *domain.Transaction) []any {
	return []any{
		item.TransactionID, item.AppName, item.UserID, nullable(item.ReferenceNo), item.ProductCategory, item.ProductType,
		item.Provider, nullable(item.Partner), item.Direction, nullable(item.Denomination), item.Amount,
		item.FeeAmount, item.TotalAmount, item.Currency, item.Status, item.StatusVersion,
		nullable(item.ResponseCode), nullable(item.SourceAccountMasked), nullable(item.DestinationName),
		nullable(item.DestinationNoType), nullable(item.DestinationNoMasked), nullable(item.DestinationNoHash),
		item.SourceSystem, item.TransactionAt, item.CompletedAt,
	}
}

func placeholders(count int) string {
	values := make([]string, count)
	for i := range values {
		values[i] = fmt.Sprintf("@p%d", i+1)
	}
	return strings.Join(values, ", ")
}

func upsertDetail(ctx context.Context, tx *sql.Tx, transactionID int64, item *domain.Detail) error {
	if item == nil || len(item.Fields) == 0 {
		return nil
	}
	encoded, err := json.Marshal(item.Fields)
	if err != nil {
		return err
	}
	result, err := tx.ExecContext(ctx,
		`UPDATE dbo.transaction_history_details SET schema_version = @p1, detail_json = @p2 WHERE transaction_history_id = @p3`,
		item.SchemaVersion, string(encoded), transactionID,
	)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		_, err = tx.ExecContext(ctx,
			`INSERT INTO dbo.transaction_history_details (transaction_history_id, schema_version, detail_json) VALUES (@p1, @p2, @p3)`,
			transactionID, item.SchemaVersion, string(encoded),
		)
	}
	return err
}

func insertStatusEvent(ctx context.Context, tx *sql.Tx, transactionID int64, event domain.StatusEvent) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO dbo.transaction_history_status_events
(transaction_history_id, event_id, event_type, status_version, from_status, to_status, response_code, changed_by_type, changed_at)
VALUES (@p1, @p2, @p3, @p4, @p5, @p6, @p7, @p8, @p9)`,
		transactionID, event.EventID, event.EventType, event.StatusVersion, nullable(event.FromStatus),
		event.ToStatus, nullable(event.ResponseCode), event.ChangedByType, event.ChangedAt,
	)
	return err
}

func readDetail(ctx context.Context, db *sql.DB, transactionID int64) (*domain.Detail, error) {
	var version int
	var encoded string
	err := db.QueryRowContext(ctx,
		`SELECT schema_version, detail_json FROM dbo.transaction_history_details WHERE transaction_history_id = @p1`,
		transactionID,
	).Scan(&version, &encoded)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	fields := make(map[string]any)
	if err := json.Unmarshal([]byte(encoded), &fields); err != nil {
		return nil, err
	}
	return &domain.Detail{SchemaVersion: version, Fields: fields}, nil
}

func nullable(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func isUniqueIndexError(err error, indexName string) bool {
	var numberedError interface{ SQLErrorNumber() int32 }
	if !errors.As(err, &numberedError) {
		return false
	}
	number := numberedError.SQLErrorNumber()
	if number != 2601 && number != 2627 {
		return false
	}
	message := err.Error()
	var sqlError mssql.Error
	if errors.As(err, &sqlError) {
		message = sqlError.Message
	}
	return strings.Contains(message, indexName)
}

func (r *historyRepository) emit(ctx context.Context, reason string) {
	if r.log == nil {
		return
	}
	r.log.LogService(ctx, logger.ServiceLog{
		Operation: "status_event_anomaly",
		Status:    "observed",
		ErrorCode: reason,
	})
}
