package utils

import (
	"fmt"
	"math/big"
	"strings"
)

func AddDecimalStrings(left, right string) (string, error) {
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
	if whole == "" || len(fraction) > 2 || strings.Contains(fraction, ".") {
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
		fraction += strings.Repeat("0", 2-len(fraction))
	} else {
		fraction = "00"
	}
	minor, ok := new(big.Int).SetString(whole+fraction, 10)
	return minor, ok
}
