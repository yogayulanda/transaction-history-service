package domain

import "time"

// ErrorDefinition represents service-owned dynamic error presentation config.
type ErrorDefinition struct {
	ErrorCode   string
	UserMessage string
	DetailsJSON string
	IsActive    bool
	UpdatedAt   time.Time
}
