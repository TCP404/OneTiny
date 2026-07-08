package config

import (
	"errors"
	"strconv"
	"strings"
	"unicode"
)

var ErrInvalidByteSize = errors.New("临时内容大小上限无效")

func ParseByteSize(value string) (int64, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return 0, ErrInvalidByteSize
	}

	split := len(trimmed)
	for i, r := range trimmed {
		if !unicode.IsDigit(r) {
			split = i
			break
		}
	}
	if split == 0 {
		return 0, ErrInvalidByteSize
	}

	number, err := strconv.ParseInt(trimmed[:split], 10, 64)
	if err != nil || number < 1 {
		return 0, ErrInvalidByteSize
	}

	unit := strings.ToUpper(strings.TrimSpace(trimmed[split:]))
	switch unit {
	case "", "B":
		return number, nil
	case "KB", "K":
		return number * 1024, nil
	case "MB", "M":
		return number * 1024 * 1024, nil
	case "GB", "G":
		return number * 1024 * 1024 * 1024, nil
	default:
		return 0, ErrInvalidByteSize
	}
}
