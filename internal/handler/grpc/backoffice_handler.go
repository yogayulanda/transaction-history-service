package grpc

import (
	"context"
	"strings"

	coreerrors "github.com/yogayulanda/go-core/errors"
	"github.com/yogayulanda/go-core/logger"
	"github.com/yogayulanda/go-core/security"
	historyv1 "github.com/yogayulanda/transaction-history-service/gen/go/history/v1"
	"github.com/yogayulanda/transaction-history-service/internal/catalog"
	"github.com/yogayulanda/transaction-history-service/internal/domain"
	"github.com/yogayulanda/transaction-history-service/internal/pii"
	"github.com/yogayulanda/transaction-history-service/internal/service"
)

type BackofficeHandler struct {
	historyv1.UnimplementedHistoryBackofficeServiceServer
	service  *service.HistoryService
	catalog  catalog.Catalog
	messages messageResolver
	hasher   *pii.Hasher
	role     string
	log      logger.Logger
}

func NewBackofficeHandler(
	historyService *service.HistoryService,
	validationCatalog catalog.Catalog,
	messages messageResolver,
	hasher *pii.Hasher,
	role string,
	log logger.Logger,
) historyv1.HistoryBackofficeServiceServer {
	if strings.TrimSpace(role) == "" {
		role = "backoffice"
	}
	return &BackofficeHandler{
		service:  historyService,
		catalog:  validationCatalog,
		messages: messages,
		hasher:   hasher,
		role:     role,
		log:      log,
	}
}

func (h *BackofficeHandler) SearchTransactions(
	ctx context.Context,
	req *historyv1.SearchTransactionsRequest,
) (response *historyv1.SearchTransactionsResponse, retErr error) {
	outcome := "success"
	resultCount := 0
	criteriaNames := backofficeCriteriaNames(req)
	destinationHashPrefix := ""
	defer func() {
		h.audit(ctx, outcome, criteriaNames, destinationHashPrefix, resultCount)
	}()
	if req == nil {
		outcome = "invalid_request"
		return nil, invalidRequest("request", "required")
	}
	claims, ok := security.FromContext(ctx)
	if !ok || claims == nil || claims.Role != h.role {
		outcome = "forbidden"
		return nil, coreerrors.ToGRPC(coreerrors.New("CORE-1003", "backoffice role required"))
	}

	transactionID := strings.TrimSpace(req.TransactionId)
	referenceNo := strings.TrimSpace(req.ReferenceNo)
	destinationNo := strings.TrimSpace(req.DestinationNo)
	destinationType := strings.TrimSpace(req.DestinationNoType)
	appName := strings.TrimSpace(req.AppName)
	userID := strings.TrimSpace(req.UserId)
	if (destinationNo == "") != (destinationType == "") || (appName == "") != (userID == "") {
		outcome = "invalid_request"
		return nil, invalidRequest("criteria", "incomplete pair")
	}
	if transactionID == "" && referenceNo == "" && destinationNo == "" && appName == "" {
		outcome = "invalid_request"
		return nil, invalidRequest("criteria", "at least one search criterion is required")
	}
	if len(transactionID) > 64 || len(referenceNo) > 128 || len(appName) > 64 || len(userID) > 64 {
		outcome = "invalid_request"
		return nil, invalidRequest("criteria", "value is too long")
	}
	if appName != "" && !h.catalog.ValidApp(appName) {
		outcome = "invalid_request"
		return nil, invalidRequest("app_name", "invalid")
	}
	var destinationHash string
	if destinationNo != "" {
		normalized, err := normalizeSearchDestination(destinationType, destinationNo)
		if err != nil {
			outcome = "invalid_request"
			return nil, invalidRequest("destination_no", "invalid")
		}
		destinationHash = h.hasher.Hash(destinationType, normalized)
		destinationHashPrefix = destinationHash[:8]
	}
	startDate, endDate, err := parseOptionalDateRange(req.StartDate, req.EndDate)
	if err != nil {
		outcome = "invalid_request"
		return nil, invalidRequest("date_range", err.Error())
	}
	cursorTime, cursorID, err := decodeCursor(req.Cursor)
	if err != nil {
		outcome = "invalid_request"
		return nil, invalidRequest("cursor", "invalid")
	}
	criteria := domain.SearchCriteria{
		TransactionID:   transactionID,
		ReferenceNo:     referenceNo,
		DestinationHash: destinationHash,
		DestinationType: destinationType,
		AppName:         appName,
		UserID:          userID,
		StartDate:       startDate,
		EndDate:         endDate,
		CursorTime:      cursorTime,
		CursorID:        cursorID,
		PageSize:        backofficePageSize(req.PageSize),
	}
	items, hasMore, err := h.service.SearchTransactions(ctx, criteria)
	if err != nil {
		outcome = "failed"
		return nil, coreerrors.ToGRPC(err)
	}
	resultCount = len(items)
	response = &historyv1.SearchTransactionsResponse{Items: make([]*historyv1.BackofficeTransactionSummary, 0, len(items))}
	for _, item := range items {
		response.Items = append(response.Items, &historyv1.BackofficeTransactionSummary{
			Transaction:  toProtoSummary(ctx, item, h.catalog, h.messages),
			AppName:      item.AppName,
			UserId:       item.UserID,
			Partner:      item.Partner,
			ReferenceNo:  item.ReferenceNo,
			ResponseCode: item.ResponseCode,
		})
	}
	if hasMore && len(items) > 0 {
		response.NextCursor = encodeCursor(items[len(items)-1])
	}
	return response, nil
}

func backofficePageSize(requested int32) int {
	if requested <= 0 {
		return 20
	}
	if requested > 100 {
		return 100
	}
	return int(requested)
}

func normalizeSearchDestination(numberType, value string) (string, error) {
	switch numberType {
	case "phone":
		return pii.NormalizePhone(value)
	case "account", "meter", "va":
		return pii.NormalizeDigits(value)
	default:
		return "", pii.ErrInvalidNumber
	}
}

func backofficeCriteriaNames(req *historyv1.SearchTransactionsRequest) []string {
	if req == nil {
		return nil
	}
	criteria := make([]string, 0, 5)
	if req.TransactionId != "" {
		criteria = append(criteria, "transaction_id")
	}
	if req.ReferenceNo != "" {
		criteria = append(criteria, "reference_no")
	}
	if req.DestinationNo != "" {
		criteria = append(criteria, "destination_no")
	}
	if req.AppName != "" || req.UserId != "" {
		criteria = append(criteria, "owner")
	}
	if req.StartDate != "" || req.EndDate != "" {
		criteria = append(criteria, "date_range")
	}
	return criteria
}

func (h *BackofficeHandler) audit(ctx context.Context, outcome string, criteria []string, hashPrefix string, count int) {
	if h.log == nil {
		return
	}
	metadata := map[string]interface{}{
		"criteria":     criteria,
		"result_count": count,
	}
	if hashPrefix != "" {
		metadata["destination_hash_prefix"] = hashPrefix
	}
	h.log.LogEvent(ctx, logger.EventLog{Category: "audit", Action: "transaction_search", Outcome: outcome, Metadata: metadata})
}
