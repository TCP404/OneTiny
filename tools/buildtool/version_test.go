package main

import "testing"

func TestValidateVersionAcceptsReleaseTags(t *testing.T) {
	tests := []string{
		"v0.6.0",
		"v1.2.3",
		"v1.2.3-alpha.1",
	}

	for _, version := range tests {
		t.Run(version, func(t *testing.T) {
			if err := ValidateVersion(version); err != nil {
				t.Fatalf("ValidateVersion(%q) returned error: %v", version, err)
			}
		})
	}
}

func TestValidateVersionRejectsAmbiguousVersions(t *testing.T) {
	tests := []string{
		"",
		"0.6.0",
		"v1",
		"latest",
		"refs/tags/v0.6.0",
	}

	for _, version := range tests {
		t.Run(version, func(t *testing.T) {
			if err := ValidateVersion(version); err == nil {
				t.Fatalf("ValidateVersion(%q) returned nil error", version)
			}
		})
	}
}
