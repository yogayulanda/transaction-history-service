package backfill

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/yogayulanda/go-core/messaging"
	"github.com/yogayulanda/transaction-history-service/internal/domain"
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

type Repository interface {
	InsertBackfill(ctx context.Context, header *domain.TransactionSnapshot, detail *domain.Detail, event domain.StatusEvent) (bool, error)
}

type Preparer interface {
	PrepareStatusEvent(attrs messaging.CloudEventAttributes, payload []byte) (*domain.TransactionSnapshot, *domain.Detail, domain.StatusEvent, error)
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
	repo    Repository
	prepare Preparer
}

func NewRunner(source Source, repo Repository, prepare Preparer) *Runner {
	return &Runner{source: source, repo: repo, prepare: prepare}
}

func (r *Runner) Run(ctx context.Context, batchSize int) (*Report, error) {
	if r.source == nil || r.repo == nil || r.prepare == nil {
		return nil, errors.New("backfill runner dependencies are required")
	}
	if batchSize <= 0 {
		batchSize = 100
	}
	location, err := time.LoadLocation("Asia/Jakarta")
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
			attrs := messaging.CloudEventAttributes{
				ID:       eventID.String(),
				Type:     "transaction.status.v1",
				Subject:  transactionID,
				Time:     transactionAt,
				Sequence: "1",
			}
			header, detailValue, event, err := r.prepare.PrepareStatusEvent(attrs, record.Payload)
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
			event.ChangedByType = "system"
			event.ChangedAt = header.TransactionAt
			group.TotalAmount, err = addDecimalStrings(group.TotalAmount, header.TotalAmount)
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

func addDecimalStrings(left, right string) (string, error) {
	leftMinor, ok := decimalMinorUnits(left)
	if !ok {
		return "", fmt.Errorf("invalid decimal")
	}
	rightMinor, ok := decimalMinorUnits(right)
	if !ok {
		return "", fmt.Errorf("invalid decimal")
	}
	sum := new(big.Int).Add(leftMinor, rightMinor)
	negative := sum.Sign() < 0
	if negative {
		sum.Abs(sum)
	}
	whole, fraction := new(big.Int), new(big.Int)
	whole.QuoRem(sum, big.NewInt(100), fraction)
	formatted := fmt.Sprintf("%s.%02d", whole.String(), fraction.Int64())
	if negative {
		formatted = "-" + formatted
	}
	return formatted, nil
}

func decimalMinorUnits(value string) (*big.Int, bool) {
	if value == "" || strings.HasPrefix(value, "-") || strings.HasPrefix(value, "+") {
		return nil, false
	}
	whole, fraction, hasFraction := strings.Cut(value, ".")
	if whole == "" || len(fraction) > 2 || (strings.Contains(fraction, ".")) {
		return nil, false
	}
	for _, part := range []string{whole, fraction} {
		for _, digit := range part {
			if digit < '0' || digit > '9' {
				return nil, false
			}
		}
	}
	if hasFraction {
		fraction = fraction + strings.Repeat("0", 2-len(fraction))
	} else {
		fraction = "00"
	}
	minor, ok := new(big.Int).SetString(whole+fraction, 10)
	return minor, ok
}
