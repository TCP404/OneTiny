package gui

import (
	"bytes"
	"image/png"
	"testing"
	"testing/fstest"

	"github.com/tcp404/OneTiny/internal/app"
	"github.com/wailsapp/wails/v3/pkg/application"
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
	service := NewService(app.NewService(app.Dependencies{}), nil)
	options := newApplicationOptions(service, fstest.MapFS{}, nil, nil, nil)

	if !bytes.Equal(options.Icon, appIcon) {
		t.Fatal("application icon does not use appIcon")
	}
}

func TestApplicationOptionsEnableSingleInstance(t *testing.T) {
	service := NewService(app.NewService(app.Dependencies{}), nil)
	var opened bool
	options := newApplicationOptions(service, fstest.MapFS{}, nil, nil, func() {
		opened = true
	})

	if options.SingleInstance == nil {
		t.Fatal("single instance options are not configured")
	}
	if options.SingleInstance.UniqueID != singleInstanceID {
		t.Fatalf("single instance unique ID = %q, want %q", options.SingleInstance.UniqueID, singleInstanceID)
	}

	options.SingleInstance.OnSecondInstanceLaunch(application.SecondInstanceData{})
	if !opened {
		t.Fatal("second instance launch did not open existing panel")
	}
}

func TestMainWindowOptionsUseAppIconOnLinux(t *testing.T) {
	options := newMainWindowOptions()

	if !bytes.Equal(options.Linux.Icon, appIcon) {
		t.Fatal("linux window icon does not use appIcon")
	}
}
