package gui

import (
	"bytes"
	"image/png"
	"testing"
	"testing/fstest"

	"github.com/TCP404/OneTiny-cli/internal/control"
)

func TestAppIconIsEmbeddedPNG(t *testing.T) {
	if len(appIcon) == 0 {
		t.Fatal("appIcon is empty")
	}
	if _, err := png.DecodeConfig(bytes.NewReader(appIcon)); err != nil {
		t.Fatalf("appIcon is not a valid PNG: %v", err)
	}
}

func TestApplicationOptionsUseAppIcon(t *testing.T) {
	service := NewService(control.NewController(), nil)
	options := newApplicationOptions(service, fstest.MapFS{}, nil, nil)

	if !bytes.Equal(options.Icon, appIcon) {
		t.Fatal("application icon does not use appIcon")
	}
}

func TestMainWindowOptionsUseAppIconOnLinux(t *testing.T) {
	options := newMainWindowOptions()

	if !bytes.Equal(options.Linux.Icon, appIcon) {
		t.Fatal("linux window icon does not use appIcon")
	}
}
