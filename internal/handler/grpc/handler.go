package grpc

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	coreerrors "github.com/yogayulanda/go-core/errors"
	historyv1 "github.com/yogayulanda/transaction-history-service/gen/go/history/v1"
	"github.com/yogayulanda/transaction-history-service/internal/domain"
	"github.com/yogayulanda/transaction-history-service/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Handler struct {
	historyv1.UnimplementedHistoryServiceServer
	service *service.TransactionService
}

func NewHistoryHandler(
	s *service.TransactionService,
) historyv1.HistoryServiceServer {
	return &Handler{
		service: s,
	}
}

func (h *Handler) CreateTransactionHistory(
	ctx context.Context,
	req *historyv1.CreateTransactionHistoryRequest,
) (*historyv1.CreateTransactionHistoryResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	statusCode := fromStatusCode(req.StatusCode)
	if statusCode == "" {
		return nil, status.Error(codes.InvalidArgument, "status_code is invalid")
	}

	txTime := time.Time{}
	if req.TransactionTime != nil {
		txTime = req.TransactionTime.AsTime()
	}

	id, err := h.service.CreateTransactionHistory(ctx, domain.CreateTransactionHistoryInput{
		UserID:           strings.TrimSpace(req.UserId),
		ReferenceID:      strings.TrimSpace(req.ReferenceId),
		ExternalRefID:    strings.TrimSpace(req.ExternalRefId),
		ProductGroup:     strings.TrimSpace(req.ProductGroup),
		ProductType:      strings.TrimSpace(req.ProductType),
		TransactionRoute: strings.TrimSpace(req.TransactionRoute),
		Channel:          strings.TrimSpace(req.Channel),
		Direction:        strings.TrimSpace(req.Direction),
		Amount:           req.Amount,
		Fee:              req.Fee,
		TotalAmount:      req.TotalAmount,
		Currency:         strings.TrimSpace(req.Currency),
		StatusCode:       statusCode,
		ErrorCode:        strings.TrimSpace(req.ErrorCode),
		ErrorMessage:     strings.TrimSpace(req.ErrorMessage),
		SourceService:    strings.TrimSpace(req.SourceService),
		TransactionTime:  txTime,
		MetadataJSON:     strings.TrimSpace(req.MetadataJson),
	})
	if err != nil {
		return nil, coreerrors.ToGRPC(err)
	}

	return &historyv1.CreateTransactionHistoryResponse{
		Id: id,
	}, nil
}

func (h *Handler) GetUserHistory(
	ctx context.Context,
	req *historyv1.GetUserHistoryRequest,
) (*historyv1.GetUserHistoryResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	userID := strings.TrimSpace(req.UserId)
	if userID == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	startDate, err := parseDate(req.StartDate)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "start_date must be RFC3339")
	}
	endDate, err := parseDate(req.EndDate)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "end_date must be RFC3339")
	}
	if startDate != nil && endDate != nil && startDate.After(*endDate) {
		return nil, status.Error(codes.InvalidArgument, "start_date must be before or equal to end_date")
	}

	pageSize := int(req.PageSize)
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := 0
	if strings.TrimSpace(req.Cursor) != "" {
		parsed, err := strconv.Atoi(strings.TrimSpace(req.Cursor))
		if err != nil || parsed < 0 {
			return nil, status.Error(codes.InvalidArgument, "cursor must be non-negative integer")
		}
		offset = parsed
	}

	statusCode := fromStatusCode(req.StatusCode)
	if req.StatusCode != historyv1.TransactionStatusCode_TRANSACTION_STATUS_CODE_UNSPECIFIED && statusCode == "" {
		return nil, status.Error(codes.InvalidArgument, "status_code is invalid")
	}

	items, hasMore, err := h.service.GetUserHistory(ctx, domain.ListUserHistoryFilter{
		UserID:           userID,
		StartDate:        startDate,
		EndDate:          endDate,
		ProductGroup:     strings.TrimSpace(req.ProductGroup),
		ProductType:      strings.TrimSpace(req.ProductType),
		TransactionRoute: strings.TrimSpace(req.TransactionRoute),
		StatusCode:       statusCode,
		PageSize:         pageSize,
		Offset:           offset,
	})
	if err != nil {
		return nil, coreerrors.ToGRPC(err)
	}

	respItems := make([]*historyv1.TransactionHistory, 0, len(items))
	for _, item := range items {
		respItems = append(respItems, toProtoHistory(item))
	}

	nextCursor := ""
	if hasMore {
		nextCursor = fmt.Sprintf("%d", offset+pageSize)
	}

	return &historyv1.GetUserHistoryResponse{
		Items:      respItems,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

func (h *Handler) GetTransactionHistoryDetail(
	ctx context.Context,
	req *historyv1.GetTransactionHistoryDetailRequest,
) (*historyv1.GetTransactionHistoryDetailResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	if strings.TrimSpace(req.Id) == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	tx, err := h.service.GetTransactionHistoryDetail(ctx, req.Id)
	if err != nil {
		return nil, coreerrors.ToGRPC(err)
	}

	return &historyv1.GetTransactionHistoryDetailResponse{
		Data: toProtoHistoryDetail(*tx),
	}, nil
}

func toStatusCode(status string) historyv1.TransactionStatusCode {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "CREATED":
		return historyv1.TransactionStatusCode_TRANSACTION_STATUS_CODE_CREATED
	case "PENDING":
		return historyv1.TransactionStatusCode_TRANSACTION_STATUS_CODE_PENDING
	case "PROCESSING":
		return historyv1.TransactionStatusCode_TRANSACTION_STATUS_CODE_PROCESSING
	case "SUCCESS":
		return historyv1.TransactionStatusCode_TRANSACTION_STATUS_CODE_SUCCESS
	case "FAILED":
		return historyv1.TransactionStatusCode_TRANSACTION_STATUS_CODE_FAILED
	case "REVERSED":
		return historyv1.TransactionStatusCode_TRANSACTION_STATUS_CODE_REVERSED
	case "EXPIRED":
		return historyv1.TransactionStatusCode_TRANSACTION_STATUS_CODE_EXPIRED
	default:
		return historyv1.TransactionStatusCode_TRANSACTION_STATUS_CODE_UNSPECIFIED
	}
}

func fromStatusCode(status historyv1.TransactionStatusCode) string {
	switch status {
	case historyv1.TransactionStatusCode_TRANSACTION_STATUS_CODE_CREATED:
		return "CREATED"
	case historyv1.TransactionStatusCode_TRANSACTION_STATUS_CODE_PENDING:
		return "PENDING"
	case historyv1.TransactionStatusCode_TRANSACTION_STATUS_CODE_PROCESSING:
		return "PROCESSING"
	case historyv1.TransactionStatusCode_TRANSACTION_STATUS_CODE_SUCCESS:
		return "SUCCESS"
	case historyv1.TransactionStatusCode_TRANSACTION_STATUS_CODE_FAILED:
		return "FAILED"
	case historyv1.TransactionStatusCode_TRANSACTION_STATUS_CODE_REVERSED:
		return "REVERSED"
	case historyv1.TransactionStatusCode_TRANSACTION_STATUS_CODE_EXPIRED:
		return "EXPIRED"
	default:
		return ""
	}
}

func toProtoHistory(in domain.TransactionHistory) *historyv1.TransactionHistory {
	return &historyv1.TransactionHistory{
		Id:               in.ID,
		UserId:           in.UserID,
		ReferenceId:      in.ReferenceID,
		ExternalRefId:    in.ExternalRefID,
		ProductGroup:     in.ProductGroup,
		ProductType:      in.ProductType,
		TransactionRoute: in.TransactionRoute,
		Channel:          in.Channel,
		Direction:        in.Direction,
		Amount:           in.Amount,
		Fee:              in.Fee,
		TotalAmount:      in.TotalAmount,
		Currency:         in.Currency,
		StatusCode:       toStatusCode(in.StatusCode),
		ErrorCode:        in.ErrorCode,
		ErrorMessage:     in.ErrorMessage,
		SourceService:    in.SourceService,
		TransactionTime:  timestamppb.New(in.TransactionTime),
	}
}

func toProtoHistoryDetail(in domain.TransactionHistoryDetail) *historyv1.TransactionHistoryDetail {
	return &historyv1.TransactionHistoryDetail{
		Id:               in.ID,
		UserId:           in.UserID,
		ReferenceId:      in.ReferenceID,
		ExternalRefId:    in.ExternalRefID,
		ProductGroup:     in.ProductGroup,
		ProductType:      in.ProductType,
		TransactionRoute: in.TransactionRoute,
		Channel:          in.Channel,
		Direction:        in.Direction,
		Amount:           in.Amount,
		Fee:              in.Fee,
		TotalAmount:      in.TotalAmount,
		Currency:         in.Currency,
		StatusCode:       toStatusCode(in.StatusCode),
		ErrorCode:        in.ErrorCode,
		ErrorMessage:     in.ErrorMessage,
		SourceService:    in.SourceService,
		TransactionTime:  timestamppb.New(in.TransactionTime),
		MetadataJson:     in.MetadataJSON,
		CreatedAt:        timestamppb.New(in.CreatedAt),
		UpdatedAt:        timestamppb.New(in.UpdatedAt),
	}
}

func parseDate(raw string) (*time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
