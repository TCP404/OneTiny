package updater

import (
	"strconv"
	"strings"
)

func parseVersion(version string) ([3]int, bool) {
	var parsed [3]int

	version = strings.TrimSpace(version)
	if version == "" || version == "dev" {
		return parsed, false
	}
	version = strings.TrimPrefix(version, "v")

	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return parsed, false
	}

	for i, part := range parts {
		if part == "" {
			return parsed, false
		}
		for j := 0; j < len(part); j++ {
			if part[j] < '0' || part[j] > '9' {
				return parsed, false
			}
		}

		value, err := strconv.Atoi(part)
		if err != nil {
			return parsed, false
		}
		parsed[i] = value
	}

	return parsed, true
}

func CompareVersions(left, right string) (int, bool) {
	leftVersion, leftOK := parseVersion(left)
	rightVersion, rightOK := parseVersion(right)
	if !leftOK || !rightOK {
		return 0, false
	}

	for i := range leftVersion {
		switch {
		case leftVersion[i] > rightVersion[i]:
			return 1, true
		case leftVersion[i] < rightVersion[i]:
			return -1, true
		}
	}
	return 0, true
}

func IsUpdateAvailable(current, latest string) Availability {
	result := Availability{
		Current: current,
		Latest:  latest,
	}

	if _, ok := parseVersion(current); !ok {
		result.Reason = ErrUnknownVersion.Error()
		return result
	}
	if _, ok := parseVersion(latest); !ok {
		result.Reason = ErrUnknownVersion.Error()
		return result
	}

	comparison, ok := CompareVersions(latest, current)
	if !ok {
		result.Reason = ErrUnknownVersion.Error()
		return result
	}

	result.Known = true
	result.Available = comparison > 0
	return result
}
