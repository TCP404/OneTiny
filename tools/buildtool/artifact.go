package main

import (
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

type Kind string

const (
	KindCLI Kind = "cli"
	KindGUI Kind = "gui"
)

type Config struct {
	AppName string
	CLIName string
	BinDir  string
	DistDir string
}

func DefaultConfig() Config {
	return Config{
		AppName: "OneTiny",
		CLIName: "onetiny-cli",
		BinDir:  "build/bin",
		DistDir: "dist",
	}
}

type ArtifactSpec struct {
	Kind       Kind
	Target     Target
	BaseName   string
	StagingDir string
	ZipPath    string
	BinaryName string
	BinaryPath string
	MacAppPath string
}

func NewArtifactSpec(kind Kind, target Target, cfg Config) (ArtifactSpec, error) {
	if cfg.AppName == "" || cfg.CLIName == "" || cfg.BinDir == "" || cfg.DistDir == "" {
		return ArtifactSpec{}, errors.New("artifact config is incomplete")
	}

	var prefix string
	var binaryName string
	switch kind {
	case KindCLI:
		prefix = cfg.CLIName
		binaryName = cfg.CLIName + target.ExeSuffix()
	case KindGUI:
		prefix = "onetiny-gui"
		binaryName = cfg.AppName + target.ExeSuffix()
	default:
		return ArtifactSpec{}, errors.Errorf("unsupported artifact kind %q", kind)
	}

	baseName := prefix + "-" + target.Label
	stagingDir := filepath.ToSlash(filepath.Join(cfg.DistDir, baseName))

	return ArtifactSpec{
		Kind:       kind,
		Target:     target,
		BaseName:   baseName,
		StagingDir: stagingDir,
		ZipPath:    filepath.ToSlash(filepath.Join(cfg.DistDir, baseName+".zip")),
		BinaryName: binaryName,
		BinaryPath: filepath.ToSlash(filepath.Join(stagingDir, binaryName)),
		MacAppPath: filepath.ToSlash(filepath.Join(cfg.BinDir, cfg.AppName+".app")),
	}, nil
}

func ParseKind(value string) (Kind, error) {
	switch Kind(value) {
	case KindCLI:
		return KindCLI, nil
	case KindGUI:
		return KindGUI, nil
	default:
		return "", errors.Errorf("unsupported kind %q", value)
	}
}

func BuildLdflags(kind Kind, target Target, version string) string {
	parts := []string{"-s", "-w"}
	if kind == KindGUI && target.OS == "windows" {
		parts = append(parts, "-H", "windowsgui")
	}
	parts = append(parts, "-X", "github.com/tcp404/OneTiny/internal/version.Version="+version)
	return strings.Join(parts, " ")
}
