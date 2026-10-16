package grpc

import (
	"context"
	"encoding/json"
	"strconv"

	historyv1 "github.com/yogayulanda/transaction-history-service/gen/go/history/v1"
	"github.com/yogayulanda/transaction-history-service/internal/domain"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type messageResolver interface {
	Message(ctx context.Context, code string) (string, bool)
}

var statusFallbacks = map[string]string{
	domain.StatusPending:    "Sedang diproses",
	domain.StatusUnknown:    "Sedang diproses",
	domain.StatusSuccess:    "Berhasil",
	domain.StatusFailed:     "Gagal",
	domain.StatusNeedReview: "Sedang diperiksa",
	domain.StatusRefunded:   "Dana dikembalikan",
}

func statusLabel(ctx context.Context, resolver messageResolver, status string) string {
	if resolver != nil {
		if value, ok := resolver.Message(ctx, "STATUS-"+status); ok && value != "" {
			return value
		}
	}
	return statusFallbacks[status]
}

func toProtoSummary(ctx context.Context, item domain.Transaction, providerCatalog domain.Catalog, resolver messageResolver) *historyv1.TransactionSummary {
	providerName := providerCatalog.ProviderName(item.Provider)
	if providerName == "" {
		providerName = item.Provider
	}
	return &historyv1.TransactionSummary{
		TransactionId:       item.TransactionID,
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
	item domain.Transaction,
	detailValue *domain.Detail,
	events []domain.StatusEvent,
	providerCatalog domain.Catalog,
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
	case domain.ProductTransferBiFast, domain.ProductTransferRTOL, domain.ProductTransferInternal:
		response.Detail = &historyv1.TransactionDetail_Transfer{Transfer: &historyv1.TransferDetail{BankCode: stringValue(fields, "bank_code"), Note: stringValue(fields, "note")}}
	case domain.ProductPulsa:
		response.Detail = &historyv1.TransactionDetail_Pulsa{Pulsa: &historyv1.PulsaDetail{SerialNumber: stringValue(fields, "serial_number")}}
	case domain.ProductDataPackage:
		response.Detail = &historyv1.TransactionDetail_DataPackage{DataPackage: &historyv1.DataPackageDetail{ProductName: stringValue(fields, "product_name"), SerialNumber: stringValue(fields, "serial_number")}}
	case domain.ProductPLNPrepaid:
		response.Detail = &historyv1.TransactionDetail_PlnPrepaid{PlnPrepaid: &historyv1.PlnPrepaidDetail{Token: stringValue(fields, "token"), Kwh: stringValue(fields, "kwh"), TariffPower: stringValue(fields, "tariff_power")}}
	case domain.ProductPLNPostpaid:
		response.Detail = &historyv1.TransactionDetail_PlnPostpaid{PlnPostpaid: &historyv1.PlnPostpaidDetail{Period: stringValue(fields, "period"), StandMeter: stringValue(fields, "stand_meter")}}
	case domain.ProductBPJSKesehatan:
		response.Detail = &historyv1.TransactionDetail_BpjsKesehatan{BpjsKesehatan: &historyv1.BpjsKesehatanDetail{PeriodMonths: int32Value(fields, "period_months"), ParticipantCount: int32Value(fields, "participant_count"), Branch: stringValue(fields, "branch")}}
	case domain.ProductCashIn, domain.ProductCashOut:
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
