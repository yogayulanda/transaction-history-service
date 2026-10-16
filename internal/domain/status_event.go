package domain

import (
	"time"
)

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
	CursorID        string
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
	CursorID        string
	PageSize        int
}
