package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/yogayulanda/go-core/dbtx"
	"github.com/yogayulanda/transaction-history-service/internal/domain"
	"gorm.io/gorm"
)

type transactionHistoryModel struct {
	ID               string    `gorm:"column:id;primaryKey"`
	UserID           string    `gorm:"column:user_id"`
	ReferenceID      string    `gorm:"column:reference_id"`
	ExternalRefID    string    `gorm:"column:external_ref_id"`
	ProductGroup     string    `gorm:"column:product_group"`
	ProductType      string    `gorm:"column:product_type"`
	TransactionRoute string    `gorm:"column:transaction_route"`
	Channel          string    `gorm:"column:channel"`
	Direction        string    `gorm:"column:direction"`
	Amount           int64     `gorm:"column:amount"`
	Fee              int64     `gorm:"column:fee"`
	TotalAmount      int64     `gorm:"column:total_amount"`
	Currency         string    `gorm:"column:currency"`
	StatusCode       string    `gorm:"column:status_code"`
	ErrorCode        string    `gorm:"column:error_code"`
	ErrorMessage     string    `gorm:"column:error_message"`
	SourceService    string    `gorm:"column:source_service"`
	TransactionTime  time.Time `gorm:"column:transaction_time"`
	CreatedAt        time.Time `gorm:"column:created_at"`
	UpdatedAt        time.Time `gorm:"column:updated_at"`
}

func (transactionHistoryModel) TableName() string {
	return "dbo.transaction_histories"
}

type transactionHistoryDetailModel struct {
	TransactionID string    `gorm:"column:transaction_id;primaryKey"`
	MetadataJSON  string    `gorm:"column:metadata_json"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
}

func (transactionHistoryDetailModel) TableName() string {
	return "dbo.transaction_history_details"
}

type transactionHistoryStatusEventModel struct {
	TransactionID  string    `gorm:"column:transaction_id"`
	FromStatusCode *string   `gorm:"column:from_status_code"`
	ToStatusCode   string    `gorm:"column:to_status_code"`
	ReasonCode     string    `gorm:"column:reason_code"`
	ReasonMessage  string    `gorm:"column:reason_message"`
	EventTime      time.Time `gorm:"column:event_time"`
	RawPayloadJSON string    `gorm:"column:raw_payload_json"`
	CreatedAt      time.Time `gorm:"column:created_at"`
}

func (transactionHistoryStatusEventModel) TableName() string {
	return "dbo.transaction_history_status_events"
}

type transactionRepository struct {
	db    *gorm.DB
	sqlDB *sql.DB
}

func NewTransactionRepository(db *gorm.DB, sqlDB *sql.DB) domain.TransactionRepository {
	return &transactionRepository{
		db:    db,
		sqlDB: sqlDB,
	}
}

func (r *transactionRepository) Create(
	ctx context.Context,
	in domain.CreateTransactionHistoryInput,
) (string, error) {
	id := uuid.NewString()
	now := time.Now().UTC()
	txTime := in.TransactionTime
	if txTime.IsZero() {
		txTime = now
	}
	metadata := in.MetadataJSON
	if metadata == "" {
		metadata = "{}"
	}

	err := dbtx.WithTx(ctx, r.sqlDB, func(txCtx context.Context) error {
		txDB, err := r.dbFromContext(txCtx)
		if err != nil {
			return err
		}

		h := transactionHistoryModel{
			ID:               id,
			UserID:           in.UserID,
			ReferenceID:      in.ReferenceID,
			ExternalRefID:    in.ExternalRefID,
			ProductGroup:     in.ProductGroup,
			ProductType:      in.ProductType,
			TransactionRoute: in.TransactionRoute,
			Channel:          in.Channel,
			Direction:        in.Direction,
			Amount:           in.Amount,
			Fee:              in.Fee,
			TotalAmount:      in.TotalAmount,
			Currency:         in.Currency,
			StatusCode:       in.StatusCode,
			ErrorCode:        in.ErrorCode,
			ErrorMessage:     in.ErrorMessage,
			SourceService:    in.SourceService,
			TransactionTime:  txTime,
			CreatedAt:        now,
			UpdatedAt:        now,
		}

		if err := txDB.Create(&h).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return gorm.ErrDuplicatedKey
			}
			return err
		}

		d := transactionHistoryDetailModel{
			TransactionID: id,
			MetadataJSON:  metadata,
			CreatedAt:     now,
			UpdatedAt:     now,
		}
		if err := txDB.Create(&d).Error; err != nil {
			return err
		}

		e := transactionHistoryStatusEventModel{
			TransactionID: id,
			ToStatusCode:  in.StatusCode,
			EventTime:     now,
			CreatedAt:     now,
		}
		if err := txDB.Create(&e).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	return id, nil
}

func (r *transactionRepository) dbFromContext(ctx context.Context) (*gorm.DB, error) {
	tx, ok := dbtx.FromContext(ctx)
	if !ok || tx == nil {
		return r.db.WithContext(ctx), nil
	}

	txDB := r.db.Session(&gorm.Session{
		NewDB:   true,
		Context: ctx,
	})
	txDB.Statement.ConnPool = tx
	return txDB, nil
}

func (r *transactionRepository) FindDetailByID(
	ctx context.Context,
	id string,
) (*domain.TransactionHistoryDetail, error) {
	var h transactionHistoryModel
	if err := r.db.WithContext(ctx).First(&h, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrTransactionNotFound
		}
		return nil, err
	}

	var d transactionHistoryDetailModel
	if err := r.db.WithContext(ctx).First(&d, "transaction_id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			d.MetadataJSON = "{}"
		} else {
			return nil, err
		}
	}

	return &domain.TransactionHistoryDetail{
		TransactionHistory: toDomainHistory(h),
		MetadataJSON:       d.MetadataJSON,
		CreatedAt:          h.CreatedAt,
		UpdatedAt:          h.UpdatedAt,
	}, nil
}

func (r *transactionRepository) ListByUser(
	ctx context.Context,
	filter domain.ListUserHistoryFilter,
) ([]domain.TransactionHistory, bool, error) {
	q := r.db.WithContext(ctx).
		Model(&transactionHistoryModel{}).
		Where("user_id = ?", filter.UserID).
		Order("transaction_time DESC, id DESC")

	if filter.StartDate != nil {
		q = q.Where("transaction_time >= ?", *filter.StartDate)
	}
	if filter.EndDate != nil {
		q = q.Where("transaction_time <= ?", *filter.EndDate)
	}
	if filter.ProductGroup != "" {
		q = q.Where("product_group = ?", filter.ProductGroup)
	}
	if filter.ProductType != "" {
		q = q.Where("product_type = ?", filter.ProductType)
	}
	if filter.TransactionRoute != "" {
		q = q.Where("transaction_route = ?", filter.TransactionRoute)
	}
	if filter.StatusCode != "" {
		q = q.Where("status_code = ?", filter.StatusCode)
	}

	limit := filter.PageSize + 1
	var rows []transactionHistoryModel
	if err := q.Offset(filter.Offset).Limit(limit).Find(&rows).Error; err != nil {
		return nil, false, err
	}

	hasMore := len(rows) > filter.PageSize
	if hasMore {
		rows = rows[:filter.PageSize]
	}

	out := make([]domain.TransactionHistory, 0, len(rows))
	for _, row := range rows {
		out = append(out, toDomainHistory(row))
	}

	return out, hasMore, nil
}

func toDomainHistory(h transactionHistoryModel) domain.TransactionHistory {
	return domain.TransactionHistory{
		ID:               h.ID,
		UserID:           h.UserID,
		ReferenceID:      h.ReferenceID,
		ExternalRefID:    h.ExternalRefID,
		ProductGroup:     h.ProductGroup,
		ProductType:      h.ProductType,
		TransactionRoute: h.TransactionRoute,
		Channel:          h.Channel,
		Direction:        h.Direction,
		Amount:           h.Amount,
		Fee:              h.Fee,
		TotalAmount:      h.TotalAmount,
		Currency:         h.Currency,
		StatusCode:       h.StatusCode,
		ErrorCode:        h.ErrorCode,
		ErrorMessage:     h.ErrorMessage,
		SourceService:    h.SourceService,
		TransactionTime:  h.TransactionTime,
	}
}
