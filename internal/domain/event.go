package domain

import "time"

type Owner struct {
	AppName string
	UserID  string
}

type StatusEvent struct {
	EventID       string
	EventType     string
	StatusVersion int64
	FromStatus    string
	ToStatus      string
	ResponseCode  string
	ChangedByType string
	ChangedAt     time.Time
}

type StatusEventInput struct {
	EventID           string         `json:"-"`
	EventType         string         `json:"-"`
	Subject           string         `json:"-"`
	EventTime         time.Time      `json:"-"`
	StatusVersion     int64          `json:"-"`
	TransactionID     string         `json:"transaction_id"`
	AppName           string         `json:"app_name"`
	UserID            string         `json:"user_id"`
	ProductCategory   string         `json:"product_category"`
	ProductType       string         `json:"product_type"`
	Provider          string         `json:"provider"`
	Partner           string         `json:"partner"`
	Direction         string         `json:"direction"`
	Denomination      string         `json:"denomination"`
	Amount            string         `json:"amount"`
	FeeAmount         string         `json:"fee_amount"`
	TotalAmount       string         `json:"total_amount"`
	Currency          string         `json:"currency"`
	Status            string         `json:"status"`
	FromStatus        string         `json:"from_status"`
	ResponseCode      string         `json:"response_code"`
	SourceAccountNo   string         `json:"source_account_no"`
	DestinationName   string         `json:"destination_name"`
	DestinationNo     string         `json:"destination_no"`
	DestinationNoType string         `json:"destination_no_type"`
	ReferenceNo       string         `json:"reference_no"`
	SourceSystem      string         `json:"source_system"`
	TransactionAt     string         `json:"transaction_at"`
	ChangedByType     string         `json:"changed_by_type"`
	Detail            map[string]any `json:"detail"`
}

type Detail struct {
	SchemaVersion int
	Fields        map[string]any
}

type ApplyResult string

const (
	Created   ApplyResult = "Created"
	Updated   ApplyResult = "Updated"
	Stale     ApplyResult = "Stale"
	Duplicate ApplyResult = "Duplicate"
)

type TransactionFilter struct {
	StartDate       *time.Time
	EndDate         *time.Time
	ProductCategory string
	ProductType     string
	Status          string
	CursorTime      *time.Time
	CursorID        int64
	PageSize        int
}

type SearchCriteria struct {
	TransactionID   string
	ReferenceNo     string
	DestinationHash string
	DestinationType string
	AppName         string
	UserID          string
	StartDate       *time.Time
	EndDate         *time.Time
	CursorTime      *time.Time
	CursorID        int64
	PageSize        int
}
