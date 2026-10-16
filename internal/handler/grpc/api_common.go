package grpc

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"
	_ "time/tzdata"

	coreerrors "github.com/yogayulanda/go-core/errors"
	historyv1 "github.com/yogayulanda/transaction-history-service/gen/go/history/v1"
	"github.com/yogayulanda/transaction-history-service/internal/catalog"
	"github.com/yogayulanda/transaction-history-service/internal/domain"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type messageResolver interface {
	Message(ctx context.Context, code string) (string, bool)
}

type pageCursor struct {
	TransactionAt time.Time `json:"transaction_at"`
	ID            string    `json:"id"`
}

var statusFallbacks = map[string]string{
	"PENDING":     "Sedang diproses",
	"UNKNOWN":     "Sedang diproses",
	"SUCCESS":     "Berhasil",
	"FAILED":      "Gagal",
	"NEED_REVIEW": "Sedang diperiksa",
	"REFUNDED":    "Dana dikembalikan",
}

func parseDateRange(startDate, endDate string, now time.Time) (*time.Time, *time.Time, error) {
	location, err := time.LoadLocation("Asia/Jakarta")
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
	location, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return time.UTC
	}
	return location
}

func decodeCursor(raw string) (*time.Time, string, error) {
	if raw == "" {
		return nil, "", nil
	}
	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return nil, "", err
	}
	var cursor pageCursor
	if err := json.Unmarshal(decoded, &cursor); err != nil || cursor.ID == "" || cursor.TransactionAt.IsZero() || len(cursor.ID) > 64 {
		return nil, "", errors.New("invalid cursor")
	}
	transactionAt := cursor.TransactionAt.UTC()
	return &transactionAt, cursor.ID, nil
}

func encodeCursor(item domain.TransactionSnapshot) string {
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

func statusLabel(ctx context.Context, resolver messageResolver, status string) string {
	if resolver != nil {
		if value, ok := resolver.Message(ctx, "STATUS-"+status); ok && value != "" {
			return value
		}
	}
	return statusFallbacks[status]
}

func validTransactionStatus(status string) bool {
	switch status {
	case "PENDING", "UNKNOWN", "SUCCESS", "FAILED", "NEED_REVIEW", "REFUNDED":
		return true
	default:
		return false
	}
}

func toProtoSummary(ctx context.Context, item domain.TransactionSnapshot, providerCatalog catalog.Catalog, resolver messageResolver) *historyv1.TransactionSummary {
	providerName := providerCatalog.ProviderName(item.Provider)
	if providerName == "" {
		providerName = item.Provider
	}
	return &historyv1.TransactionSummary{
		TransactionId:       item.ID,
		ProductCategory:     item.ProductCategory,
		ProductType:         item.ProductType,
		Provider:            item.Provider,
		ProviderName:        providerName,
		Direction:           item.Direction,
		Denomination:        item.Denomination,
		Amount:              item.Amount,
		FeeAmount:           item.FeeAmount,
		TotalAmount:         item.TotalAmount,
		Currency:            item.Currency,
		Status:              item.Status,
		StatusLabel:         statusLabel(ctx, resolver, item.Status),
		DestinationName:     item.DestinationName,
		DestinationNoMasked: item.DestinationNoMasked,
		TransactionAt:       timestamppb.New(item.TransactionAt),
	}
}

func toProtoDetail(
	ctx context.Context,
	item domain.TransactionSnapshot,
	detailValue *domain.Detail,
	events []domain.StatusEvent,
	providerCatalog catalog.Catalog,
	resolver messageResolver,
) *historyv1.TransactionDetail {
	summary := toProtoSummary(ctx, item, providerCatalog, resolver)
	response := &historyv1.TransactionDetail{
		TransactionId:       summary.TransactionId,
		ProductCategory:     summary.ProductCategory,
		ProductType:         summary.ProductType,
		Provider:            summary.Provider,
		ProviderName:        summary.ProviderName,
		Direction:           summary.Direction,
		Denomination:        summary.Denomination,
		Amount:              summary.Amount,
		FeeAmount:           summary.FeeAmount,
		TotalAmount:         summary.TotalAmount,
		Currency:            summary.Currency,
		Status:              summary.Status,
		StatusLabel:         summary.StatusLabel,
		DestinationName:     summary.DestinationName,
		DestinationNoMasked: summary.DestinationNoMasked,
		TransactionAt:       summary.TransactionAt,
		ResponseCode:        item.ResponseCode,
		SourceAccountMasked: item.SourceAccountMasked,
		ReferenceNo:         item.ReferenceNo,
	}
	if item.CompletedAt != nil {
		response.CompletedAt = timestamppb.New(*item.CompletedAt)
	}
	for _, event := range events {
		response.StatusHistory = append(response.StatusHistory, &historyv1.TransactionStatusHistory{
			ToStatus:    event.ToStatus,
			StatusLabel: statusLabel(ctx, resolver, event.ToStatus),
			ChangedAt:   timestamppb.New(event.ChangedAt),
		})
	}
	if detailValue == nil {
		return response
	}
	fields := detailValue.Fields
	switch item.ProductType {
	case "transfer_bifast", "transfer_rtol", "transfer_internal":
		response.Detail = &historyv1.TransactionDetail_Transfer{Transfer: &historyv1.TransferDetail{BankCode: stringValue(fields, "bank_code"), Note: stringValue(fields, "note")}}
	case "pulsa":
		response.Detail = &historyv1.TransactionDetail_Pulsa{Pulsa: &historyv1.PulsaDetail{SerialNumber: stringValue(fields, "serial_number")}}
	case "data_package":
		response.Detail = &historyv1.TransactionDetail_DataPackage{DataPackage: &historyv1.DataPackageDetail{ProductName: stringValue(fields, "product_name"), SerialNumber: stringValue(fields, "serial_number")}}
	case "pln_prepaid":
		response.Detail = &historyv1.TransactionDetail_PlnPrepaid{PlnPrepaid: &historyv1.PlnPrepaidDetail{Token: stringValue(fields, "token"), Kwh: stringValue(fields, "kwh"), TariffPower: stringValue(fields, "tariff_power")}}
	case "pln_postpaid":
		response.Detail = &historyv1.TransactionDetail_PlnPostpaid{PlnPostpaid: &historyv1.PlnPostpaidDetail{Period: stringValue(fields, "period"), StandMeter: stringValue(fields, "stand_meter")}}
	case "bpjs_kesehatan":
		response.Detail = &historyv1.TransactionDetail_BpjsKesehatan{BpjsKesehatan: &historyv1.BpjsKesehatanDetail{PeriodMonths: int32Value(fields, "period_months"), ParticipantCount: int32Value(fields, "participant_count"), Branch: stringValue(fields, "branch")}}
	case "cash_in", "cash_out":
		response.Detail = &historyv1.TransactionDetail_Cico{Cico: &historyv1.CicoDetail{CodeMasked: stringValue(fields, "code_masked"), WarkatNo: stringValue(fields, "warkat_no")}}
	}
	return response
}

func stringValue(fields map[string]any, key string) string {
	value, _ := fields[key].(string)
	return value
}

func int32Value(fields map[string]any, key string) int32 {
	switch value := fields[key].(type) {
	case int:
		return int32(value)
	case int32:
		return value
	case int64:
		return int32(value)
	case float64:
		return int32(value)
	case json.Number:
		parsed, _ := strconv.ParseInt(value.String(), 10, 32)
		return int32(parsed)
	default:
		return 0
	}
}
