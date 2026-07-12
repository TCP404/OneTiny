package main

import "testing"

func TestArtifactSpecForCLI(t *testing.T) {
	target, err := ParseTarget("windows-amd64")
	if err != nil {
		t.Fatal(err)
	}

	spec, err := NewArtifactSpec(KindCLI, target, DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}

	if spec.BaseName != "onetiny-cli-windows-x64" {
		t.Fatalf("BaseName = %q", spec.BaseName)
	}
	if spec.BinaryName != "onetiny-cli.exe" {
		t.Fatalf("BinaryName = %q", spec.BinaryName)
	}
	if spec.BinaryPath != "dist/onetiny-cli-windows-x64/onetiny-cli.exe" {
		t.Fatalf("BinaryPath = %q", spec.BinaryPath)
	}
	if spec.ZipPath != "dist/onetiny-cli-windows-x64.zip" {
		t.Fatalf("ZipPath = %q", spec.ZipPath)
	}
}

func TestArtifactSpecForGUI(t *testing.T) {
	target, err := ParseTarget("darwin-arm64")
	if err != nil {
		t.Fatal(err)
	}

	spec, err := NewArtifactSpec(KindGUI, target, DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}

	if spec.BaseName != "onetiny-gui-darwin-arm64" {
		t.Fatalf("BaseName = %q", spec.BaseName)
	}
	if spec.BinaryName != "OneTiny" {
		t.Fatalf("BinaryName = %q", spec.BinaryName)
	}
	if spec.BinaryPath != "dist/onetiny-gui-darwin-arm64/OneTiny" {
		t.Fatalf("BinaryPath = %q", spec.BinaryPath)
	}
	if spec.MacAppPath != "build/bin/OneTiny.app" {
		t.Fatalf("MacAppPath = %q", spec.MacAppPath)
	}
}

func TestNewArtifactSpecRejectsUnknownKind(t *testing.T) {
	target, err := ParseTarget("linux-amd64")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := NewArtifactSpec("server", target, DefaultConfig()); err == nil {
		t.Fatal("NewArtifactSpec returned nil error for unsupported kind")
	}
}

func TestBuildLdflagsInjectsVersion(t *testing.T) {
	target, err := ParseTarget("linux-amd64")
	if err != nil {
		t.Fatal(err)
	}

	got := BuildLdflags(KindCLI, target, "v0.6.0")
	want := "-s -w -X github.com/tcp404/OneTiny/internal/version.Version=v0.6.0"
	if got != want {
		t.Fatalf("BuildLdflags() = %q, want %q", got, want)
	}
}

func TestBuildLdflagsHidesWindowsGUIConsole(t *testing.T) {
	target, err := ParseTarget("windows-amd64")
	if err != nil {
		t.Fatal(err)
	}

	got := BuildLdflags(KindGUI, target, "v0.6.0")
	want := "-s -w -H windowsgui -X github.com/tcp404/OneTiny/internal/version.Version=v0.6.0"
	if got != want {
		t.Fatalf("BuildLdflags() = %q, want %q", got, want)
	}
}
