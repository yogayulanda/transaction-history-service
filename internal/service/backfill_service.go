package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/yogayulanda/transaction-history-service/internal/domain"
	"github.com/yogayulanda/transaction-history-service/internal/utils"
)

const backfillEventType = "history.backfill.v1"

type Source interface {
	Name() string
	Next(ctx context.Context, batchSize int) ([]Record, error)
}

type Record struct {
	SourceID string
	Payload  json.RawMessage
}

type backfillRepository interface {
	InsertBackfill(ctx context.Context, transaction *domain.Transaction, detail *domain.Detail, event domain.StatusEvent) (bool, error)
}

type statusEventPreparer interface {
	PrepareStatusEvent(input domain.StatusEventInput) (*domain.Transaction, *domain.Detail, domain.StatusEvent, error)
}

type ReportKey struct {
	Day         string
	ProductType string
}

type DailyProductReport struct {
	SourceCount     int
	Inserted        int
	SkippedExisting int
	Invalid         int
	InvalidReasons  map[string]int
	TotalAmount     string
}

type Report struct {
	Groups map[ReportKey]*DailyProductReport
}

type Runner struct {
	source  Source
	repo    backfillRepository
	prepare statusEventPreparer
}

func NewRunner(source Source, repo backfillRepository, prepare statusEventPreparer) *Runner {
	return &Runner{source: source, repo: repo, prepare: prepare}
}

func (r *Runner) Run(ctx context.Context, batchSize int) (*Report, error) {
	if r.source == nil || r.repo == nil || r.prepare == nil {
		return nil, errors.New("backfill runner dependencies are required")
	}
	if batchSize <= 0 {
		batchSize = 100
	}
	location, err := utils.Jakarta()
	if err != nil {
		return nil, err
	}
	report := &Report{Groups: make(map[ReportKey]*DailyProductReport)}
	for {
		records, err := r.source.Next(ctx, batchSize)
		if err != nil {
			return report, err
		}
		if len(records) == 0 {
			return report, nil
		}
		for _, record := range records {
			day, productType, transactionID, transactionAt, sourceSystem, reason := inspectRecord(record.Payload, location)
			group := getGroup(report, ReportKey{Day: day, ProductType: productType})
			group.SourceCount++
			if record.SourceID == "" {
				reason = "missing_source_id"
			}
			if reason != "" {
				recordInvalid(group, reason)
				continue
			}
			eventID := uuid.NewSHA1(uuid.NameSpaceDNS, []byte(r.source.Name()+":"+record.SourceID))
			var input domain.StatusEventInput
			if err := json.Unmarshal(record.Payload, &input); err != nil {
				recordInvalid(group, "malformed_json")
				continue
			}
			input.EventID = eventID.String()
			input.EventType = "transaction.status.v1"
			input.Subject = transactionID
			input.EventTime = transactionAt
			input.StatusVersion = 1
			header, detailValue, event, err := r.prepare.PrepareStatusEvent(input)
			if err != nil {
				recordInvalid(group, safeReason(err))
				continue
			}
			if sourceSystem == "" {
				recordInvalid(group, "source_system_required")
				continue
			}
			header.SourceSystem = "backfill"
			header.StatusVersion = 0
			event.EventID = eventID.String()
			event.EventType = backfillEventType
			event.StatusVersion = 0
			event.ChangedByType = domain.ChangedBySystem
			event.ChangedAt = header.TransactionAt
			group.TotalAmount, err = utils.AddDecimalStrings(group.TotalAmount, header.TotalAmount)
			if err != nil {
				recordInvalid(group, "invalid_total_amount")
				continue
			}
			inserted, err := r.repo.InsertBackfill(ctx, header, detailValue, event)
			if err != nil {
				return report, err
			}
			if inserted {
				group.Inserted++
			} else {
				group.SkippedExisting++
			}
		}
	}
}

func inspectRecord(payload []byte, location *time.Location) (day, productType, transactionID string, transactionAt time.Time, sourceSystem, reason string) {
	var fields struct {
		TransactionID string `json:"transaction_id"`
		ProductType   string `json:"product_type"`
		TransactionAt string `json:"transaction_at"`
		SourceSystem  string `json:"source_system"`
	}
	if err := json.Unmarshal(payload, &fields); err != nil {
		return "unknown", "unknown", "", time.Time{}, "", "malformed_json"
	}
	productType = fields.ProductType
	if productType == "" {
		productType = "unknown"
	}
	transactionID = fields.TransactionID
	sourceSystem = fields.SourceSystem
	transactionAt, err := time.Parse(time.RFC3339Nano, fields.TransactionAt)
	if err != nil {
		return "unknown", productType, transactionID, time.Time{}, sourceSystem, "invalid_transaction_at"
	}
	return transactionAt.In(location).Format("2006-01-02"), productType, transactionID, transactionAt, sourceSystem, ""
}

func getGroup(report *Report, key ReportKey) *DailyProductReport {
	group, ok := report.Groups[key]
	if !ok {
		group = &DailyProductReport{InvalidReasons: make(map[string]int), TotalAmount: "0.00"}
		report.Groups[key] = group
	}
	return group
}

func recordInvalid(group *DailyProductReport, reason string) {
	group.Invalid++
	group.InvalidReasons[reason]++
}

func safeReason(err error) string {
	reason := strings.TrimSpace(err.Error())
	if reason == "" || len(reason) > 64 {
		return "invalid_record"
	}
	return reason
}
