package main

import (
	"runtime"
	"strings"

	"github.com/pkg/errors"
)

type Target struct {
	OS    string
	Arch  string
	Label string
}

func ParseTarget(value string) (Target, error) {
	value = strings.TrimSpace(value)
	parts := strings.Split(value, "-")
	if len(parts) != 2 {
		return Target{}, errors.Errorf("target %q must use GOOS-GOARCH format", value)
	}

	goos := parts[0]
	goarch := normalizeArch(parts[1])
	label := platformLabel(goos, goarch)
	if label == "" {
		return Target{}, errors.Errorf("unsupported target %q", value)
	}

	return Target{
		OS:    goos,
		Arch:  goarch,
		Label: label,
	}, nil
}

func CurrentTarget() (Target, error) {
	return ParseTarget(runtime.GOOS + "-" + runtime.GOARCH)
}

func (t Target) Name() string {
	return t.OS + "-" + t.Arch
}

func (t Target) ExeSuffix() string {
	if t.OS == "windows" {
		return ".exe"
	}
	return ""
}

func RequireTargetOS(target Target, wantOS string) error {
	if target.OS != wantOS {
		return errors.Errorf("target %s requires GOOS=%s", target.Name(), wantOS)
	}
	return nil
}

func normalizeArch(value string) string {
	if value == "x64" {
		return "amd64"
	}
	return value
}

func platformLabel(goos string, goarch string) string {
	switch goos + "-" + goarch {
	case "linux-amd64":
		return "linux-x64"
	case "windows-amd64":
		return "windows-x64"
	case "darwin-amd64":
		return "darwin-x64"
	case "darwin-arm64":
		return "darwin-arm64"
	default:
		return ""
	}
}
