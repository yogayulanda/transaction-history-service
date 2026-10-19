package domain

// PayloadMaxLengths holds the single source of payload length limits; they match migration 0001 column sizes.
var PayloadMaxLengths = map[string]int{
	"transaction_id":      64,
	"app_name":            16,
	"user_id":             64,
	"product_category":    16,
	"product_type":        32,
	"provider":            32,
	"partner":             32,
	"direction":           6,
	"status":              16,
	"response_code":       16,
	"destination_name":    100,
	"destination_no_type": 8,
	"reference_no":        64,
	"source_system":       32,
	"changed_by_type":     16,
}

func (i StatusEventInput) fieldValues() map[string]string {
	return map[string]string{
		"transaction_id": i.TransactionID, "app_name": i.AppName, "user_id": i.UserID,
		"product_category": i.ProductCategory, "product_type": i.ProductType, "provider": i.Provider,
		"partner": i.Partner, "direction": i.Direction, "status": i.Status, "response_code": i.ResponseCode,
		"destination_name": i.DestinationName, "destination_no_type": i.DestinationNoType,
		"reference_no": i.ReferenceNo, "source_system": i.SourceSystem, "changed_by_type": i.ChangedByType,
	}
}

func (i StatusEventInput) WithinLimits() bool {
	for field, value := range i.fieldValues() {
		if len(value) > PayloadMaxLengths[field] {
			return false
		}
	}
	return true
}
