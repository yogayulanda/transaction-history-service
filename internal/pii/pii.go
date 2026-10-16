package pii

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"regexp"
	"strings"
)

var (
	phonePattern  = regexp.MustCompile(`^08\d{7,13}$`)
	digitsPattern = regexp.MustCompile(`^\d{4,34}$`)
)

var ErrInvalidNumber = errors.New("invalid destination number")

type Hasher struct {
	key []byte
}

func NewHasher(key []byte) (*Hasher, error) {
	if len(key) < 32 {
		return nil, errors.New("hash key must contain at least 32 bytes")
	}
	return &Hasher{key: append([]byte(nil), key...)}, nil
}

func NormalizePhone(value string) (string, error) {
	value = strings.NewReplacer(" ", "", "-", "", ".", "").Replace(value)
	value = strings.TrimPrefix(value, "+")
	if strings.HasPrefix(value, "62") {
		value = "0" + value[2:]
	}
	if !phonePattern.MatchString(value) {
		return "", ErrInvalidNumber
	}
	return value, nil
}

func NormalizeDigits(value string) (string, error) {
	if !digitsPattern.MatchString(value) {
		return "", ErrInvalidNumber
	}
	return value, nil
}

func MaskNumber(numberType, normalized string) string {
	switch numberType {
	case "phone":
		if len(normalized) <= 8 {
			return normalized
		}
		return normalized[:4] + "****" + normalized[len(normalized)-4:]
	case "cico":
		if len(normalized) <= 3 {
			return strings.Repeat("*", len(normalized))
		}
		return "****" + normalized[len(normalized)-3:]
	default:
		if len(normalized) <= 4 {
			return strings.Repeat("*", len(normalized))
		}
		return "****" + normalized[len(normalized)-4:]
	}
}

func (h *Hasher) Hash(numberType, normalized string) string {
	mac := hmac.New(sha256.New, h.key)
	_, _ = mac.Write([]byte(numberType + ":" + normalized))
	return hex.EncodeToString(mac.Sum(nil))
}
