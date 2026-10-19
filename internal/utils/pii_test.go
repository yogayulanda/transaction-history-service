package utils

import (
	"fmt"
	"strings"
	"testing"
)

func TestNormalizePhone(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: "+6281234567890", want: "081234567890"},
		{input: "6281234567890", want: "081234567890"},
		{input: "0812-3456-7890", want: "081234567890"},
		{input: "0812 3456.7890", want: "081234567890"},
	}
	for index, test := range tests {
		t.Run(fmt.Sprintf("case_%d", index+1), func(t *testing.T) {
			got, err := NormalizePhone(test.input)
			if err != nil {
				t.Fatalf("NormalizePhone returned an error for valid case: %v", err)
			}
			if got != test.want {
				t.Fatal("NormalizePhone returned an unexpected normalized value")
			}
		})
	}
}

func TestNormalizePhoneRejectsInvalidValues(t *testing.T) {
	for index, input := range []string{"", "081234", "07123456789", "08abc3456789"} {
		if _, err := NormalizePhone(input); err == nil {
			t.Errorf("NormalizePhone accepted invalid case %d", index+1)
		}
	}
}

func TestNormalizeDigits(t *testing.T) {
	for index, input := range []string{"1234", "1234567890123456789012345678901234"} {
		if got, err := NormalizeDigits(input); err != nil || got != input {
			t.Errorf("NormalizeDigits failed for valid case %d", index+1)
		}
	}
	for index, input := range []string{"123", strings.Repeat("1", 35), "12-34", "12a4"} {
		if _, err := NormalizeDigits(input); err == nil {
			t.Errorf("NormalizeDigits accepted invalid case %d", index+1)
		}
	}
}

func TestMaskNumber(t *testing.T) {
	tests := []struct {
		numberType string
		value      string
		want       string
	}{
		{numberType: "account", value: "1234567890", want: "****7890"},
		{numberType: "meter", value: "1234567890", want: "****7890"},
		{numberType: "va", value: "1234567890", want: "****7890"},
		{numberType: "phone", value: "081234567890", want: "0812****7890"},
		{numberType: "cico", value: "ABC1234", want: "****234"},
	}
	for index, test := range tests {
		if got := MaskNumber(test.numberType, test.value); got != test.want {
			t.Errorf("MaskNumber mismatch for case %d", index+1)
		}
	}
}

func TestHasherPrefixesNumberType(t *testing.T) {
	hasher, err := NewHasher([]byte(strings.Repeat("k", 32)))
	if err != nil {
		t.Fatal(err)
	}
	accountHash := hasher.Hash("account", "12345678")
	if accountHash == hasher.Hash("va", "12345678") {
		t.Fatal("equal values of different types must have different hashes")
	}
	if accountHash != hasher.Hash("account", "12345678") {
		t.Fatal("same type and value must have a stable hash")
	}
}
