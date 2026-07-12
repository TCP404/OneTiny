package main

import (
	"regexp"

	"github.com/pkg/errors"
)

var releaseVersionPattern = regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+(?:-[0-9A-Za-z.-]+)?$`)

func ValidateVersion(version string) error {
	if !releaseVersionPattern.MatchString(version) {
		return errors.Errorf("version %q must be a release tag like v0.6.0", version)
	}
	return nil
}
