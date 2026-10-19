package grpc

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	coreerrors "github.com/yogayulanda/go-core/errors"
	"github.com/yogayulanda/transaction-history-service/internal/domain"
	"github.com/yogayulanda/transaction-history-service/internal/utils"
)

type pageCursor struct {
	TransactionAt time.Time `json:"transaction_at"`
	ID            int64     `json:"id"`
}

func parseDateRange(startDate, endDate string, now time.Time) (*time.Time, *time.Time, error) {
	location, err := utils.Jakarta()
	if err != nil {
		return nil, nil, err
	}
	today := time.Date(now.In(location).Year(), now.In(location).Month(), now.In(location).Day(), 0, 0, 0, 0, location)
	start := today.AddDate(0, 0, -29)
	end := today
	if startDate != "" {
		start, err = time.ParseInLocation("2006-01-02", startDate, location)
		if err != nil {
			return nil, nil, fmt.Errorf("start_date must use YYYY-MM-DD")
		}
	}
	if endDate != "" {
		end, err = time.ParseInLocation("2006-01-02", endDate, location)
		if err != nil {
			return nil, nil, fmt.Errorf("end_date must use YYYY-MM-DD")
		}
	}
	if start.After(end) {
		return nil, nil, fmt.Errorf("start_date must not be after end_date")
	}
	if end.Sub(start)/(24*time.Hour)+1 > 92 {
		return nil, nil, fmt.Errorf("date range must not exceed 92 days")
	}
	startUTC := start.UTC()
	endExclusiveUTC := end.AddDate(0, 0, 1).UTC()
	return &startUTC, &endExclusiveUTC, nil
}

func parseOptionalDateRange(startDate, endDate string) (*time.Time, *time.Time, error) {
	if startDate == "" && endDate == "" {
		return nil, nil, nil
	}
	now := time.Now()
	if startDate == "" {
		end, err := time.ParseInLocation("2006-01-02", endDate, mustJakarta())
		if err != nil {
			return nil, nil, fmt.Errorf("end_date must use YYYY-MM-DD")
		}
		startDate = end.AddDate(0, 0, -29).Format("2006-01-02")
	}
	if endDate == "" {
		start, err := time.ParseInLocation("2006-01-02", startDate, mustJakarta())
		if err != nil {
			return nil, nil, fmt.Errorf("start_date must use YYYY-MM-DD")
		}
		endDate = start.Format("2006-01-02")
	}
	return parseDateRange(startDate, endDate, now)
}

func mustJakarta() *time.Location {
	location, err := utils.Jakarta()
	if err != nil {
		return time.UTC
	}
	return location
}

func decodeCursor(raw string) (*time.Time, int64, error) {
	if raw == "" {
		return nil, 0, nil
	}
	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return nil, 0, err
	}
	var cursor pageCursor
	if err := json.Unmarshal(decoded, &cursor); err != nil || cursor.ID <= 0 || cursor.TransactionAt.IsZero() {
		return nil, 0, errors.New("invalid cursor")
	}
	transactionAt := cursor.TransactionAt.UTC()
	return &transactionAt, cursor.ID, nil
}

func encodeCursor(item domain.Transaction) string {
	encoded, _ := json.Marshal(pageCursor{TransactionAt: item.TransactionAt.UTC(), ID: item.ID})
	return base64.RawURLEncoding.EncodeToString(encoded)
}

func pageSize(requested int32) int {
	if requested <= 0 {
		return 20
	}
	if requested > 50 {
		return 50
	}
	return int(requested)
}

func invalidRequest(field, reason string) error {
	return coreerrors.ToGRPC(coreerrors.Validation("invalid request", coreerrors.Detail{Field: field, Reason: reason}))
}
