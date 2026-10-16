package domain

const (
	StatusPending    = "PENDING"
	StatusUnknown    = "UNKNOWN"
	StatusSuccess    = "SUCCESS"
	StatusFailed     = "FAILED"
	StatusNeedReview = "NEED_REVIEW"
	StatusRefunded   = "REFUNDED"

	ChangedByUser     = "user"
	ChangedBySystem   = "system"
	ChangedByOperator = "operator"
)

var Statuses = [...]string{
	StatusPending,
	StatusUnknown,
	StatusSuccess,
	StatusFailed,
	StatusNeedReview,
	StatusRefunded,
}

func IsValidStatus(status string) bool {
	for _, candidate := range Statuses {
		if status == candidate {
			return true
		}
	}
	return false
}

func IsFinalStatus(status string) bool {
	return status == StatusSuccess || status == StatusFailed || status == StatusRefunded
}
