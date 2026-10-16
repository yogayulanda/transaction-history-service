package repository

import (
	"context"
	"database/sql"

	"github.com/yogayulanda/transaction-history-service/internal/domain"
)

type messageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) domain.MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) ListActiveErrorDefinitions(ctx context.Context) ([]domain.ErrorDefinition, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT error_code, user_message, details_json, is_active, updated_at
FROM dbo.transaction_error_definitions WHERE is_active = @p1`, true,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	definitions := make([]domain.ErrorDefinition, 0)
	for rows.Next() {
		var definition domain.ErrorDefinition
		var details sql.NullString
		if err := rows.Scan(
			&definition.ErrorCode, &definition.UserMessage, &details, &definition.IsActive, &definition.UpdatedAt,
		); err != nil {
			return nil, err
		}
		definition.DetailsJSON = details.String
		definitions = append(definitions, definition)
	}
	return definitions, rows.Err()
}
