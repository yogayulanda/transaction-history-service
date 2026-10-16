package grpc

import (
	"context"
	"strings"
	"time"

	coreerrors "github.com/yogayulanda/go-core/errors"
	historyv1 "github.com/yogayulanda/transaction-history-service/gen/go/history/v1"
	"github.com/yogayulanda/transaction-history-service/internal/domain"
	"github.com/yogayulanda/transaction-history-service/internal/service"
)

type HistoryHandler struct {
	historyv1.UnimplementedHistoryServiceServer
	service  *service.HistoryService
	identity Resolver
	catalog  domain.Catalog
	messages messageResolver
}

func NewHistoryHandler(
	historyService *service.HistoryService,
	resolver Resolver,
	validationCatalog domain.Catalog,
	messages messageResolver,
) historyv1.HistoryServiceServer {
	return &HistoryHandler{service: historyService, identity: resolver, catalog: validationCatalog, messages: messages}
}

func (h *HistoryHandler) ListTransactions(
	ctx context.Context,
	req *historyv1.ListTransactionsRequest,
) (*historyv1.ListTransactionsResponse, error) {
	if req == nil {
		return nil, invalidRequest("request", "required")
	}
	owner, err := h.identity.Resolve(ctx)
	if err != nil {
		return nil, coreerrors.ToGRPC(err)
	}
	startDate, endDate, err := parseDateRange(req.StartDate, req.EndDate, time.Now())
	if err != nil {
		return nil, invalidRequest("date_range", err.Error())
	}
	statusFilter := strings.TrimSpace(req.Status)
	if statusFilter != "" && !domain.IsValidStatus(statusFilter) {
		return nil, invalidRequest("status", "invalid")
	}
	cursorTime, cursorID, err := decodeCursor(req.Cursor)
	if err != nil {
		return nil, invalidRequest("cursor", "invalid")
	}
	filter := domain.TransactionFilter{
		StartDate:       startDate,
		EndDate:         endDate,
		ProductCategory: strings.TrimSpace(req.ProductCategory),
		ProductType:     strings.TrimSpace(req.ProductType),
		Status:          statusFilter,
		CursorTime:      cursorTime,
		CursorID:        cursorID,
		PageSize:        pageSize(req.PageSize),
	}
	items, hasMore, err := h.service.ListTransactions(ctx, owner, filter)
	if err != nil {
		return nil, coreerrors.ToGRPC(err)
	}
	response := &historyv1.ListTransactionsResponse{Items: make([]*historyv1.TransactionSummary, 0, len(items))}
	for _, item := range items {
		response.Items = append(response.Items, toProtoSummary(ctx, item, h.catalog, h.messages))
	}
	if hasMore && len(items) > 0 {
		response.NextCursor = encodeCursor(items[len(items)-1])
	}
	return response, nil
}

func (h *HistoryHandler) GetTransaction(
	ctx context.Context,
	req *historyv1.GetTransactionRequest,
) (*historyv1.GetTransactionResponse, error) {
	if req == nil || strings.TrimSpace(req.TransactionId) == "" {
		return nil, invalidRequest("transaction_id", "required")
	}
	owner, err := h.identity.Resolve(ctx)
	if err != nil {
		return nil, coreerrors.ToGRPC(err)
	}
	header, detailValue, events, err := h.service.GetTransaction(ctx, owner, strings.TrimSpace(req.TransactionId))
	if err != nil {
		return nil, coreerrors.ToGRPC(err)
	}
	return &historyv1.GetTransactionResponse{Transaction: toProtoDetail(ctx, *header, detailValue, events, h.catalog, h.messages)}, nil
}
