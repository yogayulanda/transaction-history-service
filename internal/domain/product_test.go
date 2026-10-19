package domain

import (
	"errors"
	"strings"
	"testing"
)

func TestNormalizeFiltersAndMasksFields(t *testing.T) {
	version, fields, err := NormalizeDetail("cash_in", map[string]any{
		"code":       "CASHCODE123",
		"warkat_no":  "WR-123",
		"user_id":    "injected-user",
		"unexpected": "discarded",
		"amount":     "1000000000.00",
	})
	if err != nil {
		t.Fatal(err)
	}
	if version != 1 || len(fields) != 2 {
		t.Fatalf("unexpected normalized detail: version=%d fields=%v", version, fields)
	}
	if fields["code"] != "CASHCODE123" || fields["warkat_no"] != "WR-123" {
		t.Fatalf("unexpected normalized fields: %v", fields)
	}
}

func TestNormalizeValidatesTypesAndBounds(t *testing.T) {
	for _, test := range []struct {
		productType string
		fields      map[string]any
	}{
		{"transfer_internal", map[string]any{"bank_code": "12A"}},
		{"pulsa", map[string]any{"serial_number": string(make([]byte, 65))}},
		{"bpjs_kesehatan", map[string]any{"period_months": float64(1.5)}},
		{"bpjs_kesehatan", map[string]any{"participant_count": 51}},
	} {
		if _, _, err := NormalizeDetail(test.productType, test.fields); !errors.Is(err, ErrInvalidDetail) {
			t.Errorf("Normalize(%q) error = %v, want ErrInvalidDetail", test.productType, err)
		}
	}
}

func TestNormalizeDropsUnknownAndReturnsEmpty(t *testing.T) {
	version, fields, err := NormalizeDetail("pulsa", map[string]any{"other": "value", "user_id": "owner"})
	if err != nil || version != 1 || len(fields) != 0 {
		t.Fatalf("Normalize() = %d, %v, %v", version, fields, err)
	}
}

func TestNormalizeAcceptsEveryProductSchema(t *testing.T) {
	tests := []struct {
		productType string
		fields      map[string]any
	}{
		{"transfer_bifast", map[string]any{"bank_code": "014", "note": "sample"}},
		{"transfer_rtol", map[string]any{"bank_code": "008"}},
		{"transfer_internal", map[string]any{"note": "sample"}},
		{"pulsa", map[string]any{"serial_number": "SAMPLE-1"}},
		{"data_package", map[string]any{"product_name": "Sample data", "serial_number": "SAMPLE-2"}},
		{"pln_prepaid", map[string]any{"token": "123", "kwh": "2.5", "tariff_power": "R1-1300"}},
		{"pln_postpaid", map[string]any{"period": "2026-10", "stand_meter": "10-20"}},
		{"bpjs_kesehatan", map[string]any{"period_months": 1, "participant_count": 2, "branch": "Sample City"}},
		{"cash_in", map[string]any{"code": "SAMPLE-CASH-1", "warkat_no": "WR-1"}},
		{"cash_out", map[string]any{"code": "SAMPLE-CASH-2", "warkat_no": "WR-2"}},
	}
	for _, test := range tests {
		t.Run(test.productType, func(t *testing.T) {
			version, fields, err := NormalizeDetail(test.productType, test.fields)
			if err != nil || version != 1 || len(fields) == 0 {
				t.Fatal("Normalize failed for a valid product detail case")
			}
		})
	}
}

func TestNormalizeRejectsEveryProductSchemaBoundary(t *testing.T) {
	tests := []struct {
		productType string
		fields      map[string]any
	}{
		{"transfer_internal", map[string]any{"bank_code": "01A"}},
		{"pulsa", map[string]any{"serial_number": strings.Repeat("s", 65)}},
		{"data_package", map[string]any{"product_name": strings.Repeat("p", 101)}},
		{"pln_prepaid", map[string]any{"token": strings.Repeat("t", 33)}},
		{"pln_prepaid", map[string]any{"kwh": strings.Repeat("k", 17)}},
		{"pln_prepaid", map[string]any{"tariff_power": strings.Repeat("p", 33)}},
		{"pln_postpaid", map[string]any{"period": strings.Repeat("p", 33)}},
		{"pln_postpaid", map[string]any{"stand_meter": strings.Repeat("s", 65)}},
		{"bpjs_kesehatan", map[string]any{"period_months": 25}},
		{"bpjs_kesehatan", map[string]any{"participant_count": 51}},
		{"bpjs_kesehatan", map[string]any{"branch": strings.Repeat("b", 65)}},
		{"cash_in", map[string]any{"warkat_no": strings.Repeat("w", 33)}},
		{"cash_out", map[string]any{"code": 123}},
	}
	for _, test := range tests {
		t.Run(test.productType, func(t *testing.T) {
			if _, _, err := NormalizeDetail(test.productType, test.fields); !errors.Is(err, ErrInvalidDetail) {
				t.Fatalf("Normalize() error = %v, want ErrInvalidDetail", err)
			}
		})
	}
}
