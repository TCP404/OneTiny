package main

import "testing"

func TestParseTargetSupportsReleaseTargets(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Target
	}{
		{
			name:  "linux amd64",
			input: "linux-amd64",
			want:  Target{OS: "linux", Arch: "amd64", Label: "linux-x64"},
		},
		{
			name:  "windows amd64",
			input: "windows-amd64",
			want:  Target{OS: "windows", Arch: "amd64", Label: "windows-x64"},
		},
		{
			name:  "darwin amd64",
			input: "darwin-amd64",
			want:  Target{OS: "darwin", Arch: "amd64", Label: "darwin-x64"},
		},
		{
			name:  "darwin arm64",
			input: "darwin-arm64",
			want:  Target{OS: "darwin", Arch: "arm64", Label: "darwin-arm64"},
		},
		{
			name:  "accepts x64 alias",
			input: "windows-x64",
			want:  Target{OS: "windows", Arch: "amd64", Label: "windows-x64"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTarget(tt.input)
			if err != nil {
				t.Fatalf("ParseTarget(%q) returned error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Fatalf("ParseTarget(%q) = %#v, want %#v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseTargetRejectsUnsupportedTargets(t *testing.T) {
	tests := []string{
		"",
		"linux-arm64",
		"windows-arm64",
		"darwin",
		"linux/amd64",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			if _, err := ParseTarget(input); err == nil {
				t.Fatalf("ParseTarget(%q) returned nil error", input)
			}
		})
	}
}

func TestTargetExeSuffix(t *testing.T) {
	windows, err := ParseTarget("windows-amd64")
	if err != nil {
		t.Fatal(err)
	}
	if got := windows.ExeSuffix(); got != ".exe" {
		t.Fatalf("windows ExeSuffix() = %q, want .exe", got)
	}

	linux, err := ParseTarget("linux-amd64")
	if err != nil {
		t.Fatal(err)
	}
	if got := linux.ExeSuffix(); got != "" {
		t.Fatalf("linux ExeSuffix() = %q, want empty suffix", got)
	}
}

func TestRequireTargetOS(t *testing.T) {
	target, err := ParseTarget("darwin-arm64")
	if err != nil {
		t.Fatal(err)
	}
	if err := RequireTargetOS(target, "darwin"); err != nil {
		t.Fatalf("RequireTargetOS returned error: %v", err)
	}
	if err := RequireTargetOS(target, "windows"); err == nil {
		t.Fatal("RequireTargetOS returned nil error for mismatched OS")
	}
}
