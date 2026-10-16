package detail

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/yogayulanda/transaction-history-service/internal/pii"
)

var ErrInvalidDetail = errors.New("invalid transaction detail")

var bankCodePattern = regexp.MustCompile(`^\d*$`)

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

var schemas = map[string]map[string]fieldRule{
	"transfer_bifast":   {"bank_code": {maxLength: 8, digits: true}, "note": {maxLength: 100}},
	"transfer_rtol":     {"bank_code": {maxLength: 8, digits: true}, "note": {maxLength: 100}},
	"transfer_internal": {"bank_code": {maxLength: 8, digits: true}, "note": {maxLength: 100}},
	"pulsa":             {"serial_number": {maxLength: 64}},
	"data_package":      {"product_name": {maxLength: 100}, "serial_number": {maxLength: 64}},
	"pln_prepaid":       {"token": {maxLength: 32}, "kwh": {maxLength: 16}, "tariff_power": {maxLength: 32}},
	"pln_postpaid":      {"period": {maxLength: 32}, "stand_meter": {maxLength: 64}},
	"bpjs_kesehatan": {
		"period_months":     {integer: true, min: 1, max: 24},
		"participant_count": {integer: true, min: 1, max: 50},
		"branch":            {maxLength: 64},
	},
	"cash_in":  {"code": {maxLength: 0}, "warkat_no": {maxLength: 32}},
	"cash_out": {"code": {maxLength: 0}, "warkat_no": {maxLength: 32}},
}

func Normalize(productType string, fields map[string]any) (int, map[string]any, error) {
	schema, ok := schemas[productType]
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
		if productType == "cash_in" || productType == "cash_out" {
			if key == "code" {
				code, ok := value.(string)
				if !ok {
					return 0, nil, fmt.Errorf("%w: code must be a string", ErrInvalidDetail)
				}
				clean["code_masked"] = pii.MaskNumber("cico", code)
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
