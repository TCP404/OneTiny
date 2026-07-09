package config

import (
	"errors"
	"math"
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
		return multiplyByteSize(number, 1024)
	case "MB", "M":
		return multiplyByteSize(number, 1024*1024)
	case "GB", "G":
		return multiplyByteSize(number, 1024*1024*1024)
	default:
		return 0, ErrInvalidByteSize
	}
}

func multiplyByteSize(number, multiplier int64) (int64, error) {
	if number > math.MaxInt64/multiplier {
		return 0, ErrInvalidByteSize
	}
	return number * multiplier, nil
}
