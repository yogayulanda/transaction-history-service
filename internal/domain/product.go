package domain

import (
	"fmt"
	"regexp"
	"unicode"
)

var (
	bankCodePattern = regexp.MustCompile(`^\d*$`)
	amountPattern   = regexp.MustCompile(`^\d{1,16}(\.\d{1,2})?$`)
)

var headerFields = map[string]struct{}{
	"id": {}, "app_name": {}, "user_id": {}, "reference_no": {}, "product_category": {},
	"product_type": {}, "provider": {}, "partner": {}, "direction": {}, "denomination": {},
	"amount": {}, "fee_amount": {}, "total_amount": {}, "currency": {}, "status": {},
	"status_version": {}, "response_code": {}, "source_account_masked": {}, "destination_name": {},
	"destination_no_type": {}, "destination_no_masked": {}, "destination_no_hash": {},
	"source_system": {}, "transaction_at": {}, "completed_at": {}, "created_at": {}, "updated_at": {},
}

type fieldRule struct {
	maxLength int
	integer   bool
	digits    bool
	min       int64
	max       int64
}

var ProductCategories = map[string]string{
	ProductTransferBiFast:   CategoryTransfer,
	ProductTransferRTOL:     CategoryTransfer,
	ProductTransferInternal: CategoryTransfer,
	ProductPulsa:            CategoryPPOB,
	ProductDataPackage:      CategoryPPOB,
	ProductPLNPrepaid:       CategoryPPOB,
	ProductPLNPostpaid:      CategoryPPOB,
	ProductBPJSKesehatan:    CategoryPPOB,
	ProductCashIn:           CategoryCICO,
	ProductCashOut:          CategoryCICO,
}

var detailSchemas = map[string]map[string]fieldRule{
	ProductTransferBiFast:   {"bank_code": {maxLength: 8, digits: true}, "note": {maxLength: 100}},
	ProductTransferRTOL:     {"bank_code": {maxLength: 8, digits: true}, "note": {maxLength: 100}},
	ProductTransferInternal: {"bank_code": {maxLength: 8, digits: true}, "note": {maxLength: 100}},
	ProductPulsa:            {"serial_number": {maxLength: 64}},
	ProductDataPackage:      {"product_name": {maxLength: 100}, "serial_number": {maxLength: 64}},
	ProductPLNPrepaid:       {"token": {maxLength: 32}, "kwh": {maxLength: 16}, "tariff_power": {maxLength: 32}},
	ProductPLNPostpaid:      {"period": {maxLength: 32}, "stand_meter": {maxLength: 64}},
	ProductBPJSKesehatan: {
		"period_months":     {integer: true, min: 1, max: 24},
		"participant_count": {integer: true, min: 1, max: 50},
		"branch":            {maxLength: 64},
	},
	ProductCashIn:  {"code": {maxLength: 0}, "warkat_no": {maxLength: 32}},
	ProductCashOut: {"code": {maxLength: 0}, "warkat_no": {maxLength: 32}},
}

func NormalizeDetail(productType string, fields map[string]any) (int, map[string]any, error) {
	schema, ok := detailSchemas[productType]
	if !ok {
		return 0, nil, fmt.Errorf("%w: unsupported product type", ErrInvalidDetail)
	}

	clean := make(map[string]any, len(fields))
	for key, value := range fields {
		if _, isHeader := headerFields[key]; isHeader {
			continue
		}
		rule, known := schema[key]
		if !known {
			continue
		}
		if productType == ProductCashIn || productType == ProductCashOut {
			if key == "code" {
				code, ok := value.(string)
				if !ok {
					return 0, nil, fmt.Errorf("%w: code must be a string", ErrInvalidDetail)
				}
				clean["code"] = code
				continue
			}
		}
		validated, err := validateField(key, value, rule)
		if err != nil {
			return 0, nil, err
		}
		clean[key] = validated
	}
	return 1, clean, nil
}

func ValidAmount(value string) bool {
	return amountPattern.MatchString(value)
}

func ValidCurrency(value string) bool {
	if len(value) != 3 {
		return false
	}
	for _, char := range value {
		if !unicode.IsLetter(char) || !unicode.IsUpper(char) {
			return false
		}
	}
	return true
}

func ProductCategory(productType string) (string, bool) {
	category, ok := ProductCategories[productType]
	return category, ok
}

func validateField(key string, value any, rule fieldRule) (any, error) {
	if rule.integer {
		integer, ok := integerValue(value)
		if !ok || integer < rule.min || integer > rule.max {
			return nil, fmt.Errorf("%w: %s must be an integer in range", ErrInvalidDetail, key)
		}
		return integer, nil
	}
	text, ok := value.(string)
	if !ok || len(text) > rule.maxLength || (rule.digits && !bankCodePattern.MatchString(text)) {
		return nil, fmt.Errorf("%w: %s has an invalid value", ErrInvalidDetail, key)
	}
	return text, nil
}

func integerValue(value any) (int64, bool) {
	switch number := value.(type) {
	case int:
		return int64(number), true
	case int32:
		return int64(number), true
	case int64:
		return number, true
	case float64:
		if number == float64(int64(number)) {
			return int64(number), true
		}
	}
	return 0, false
}
