package gui

import (
	"testing"

	"github.com/tcp404/OneTiny/internal/app"
)

func TestServiceDialogMethodsAreNoopWithoutAdapter(t *testing.T) {
	service := NewService(&app.Service{}, nil)
	if got, err := service.ChooseDirectory("/tmp"); err != nil || got != "" {
		t.Fatalf("ChooseDirectory() = %q, %v; want empty nil", got, err)
	}
	if got, err := service.ExportLogs(app.LogFilterDTO{}); err != nil || got != "" {
		t.Fatalf("ExportLogs() = %q, %v; want empty nil", got, err)
	}
	if err := service.OpenConfigDir(); err != nil {
		t.Fatalf("OpenConfigDir() error = %v", err)
	}
}
