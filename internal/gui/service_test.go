package gui

import (
	"net"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/tcp404/OneTiny/internal/app"
	"github.com/tcp404/OneTiny/internal/config"
	"github.com/tcp404/OneTiny/internal/runtime"
	"github.com/tcp404/OneTiny/internal/server"
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
	if err := service.OpenShareAddress(); err != nil {
		t.Fatalf("OpenShareAddress() error = %v", err)
	}
}

func TestOpenShareAddressSkipsEmptyAddress(t *testing.T) {
	service := NewService(newTestAppService(t), &recordingDialogs{})

	if err := service.OpenShareAddress(); err != nil {
		t.Fatalf("OpenShareAddress() error = %v", err)
	}

	dialogs := service.dialogs.(*recordingDialogs)
	if len(dialogs.openedURLs) != 0 {
		t.Fatalf("opened URLs = %v, want none", dialogs.openedURLs)
	}
}

func TestOpenShareAddressOpensRunningAddress(t *testing.T) {
	appService := newTestAppService(t)
	dialogs := &recordingDialogs{}
	service := NewService(appService, dialogs)

	status, err := appService.StartSharing()
	if err != nil {
		t.Fatalf("StartSharing() error = %v", err)
	}
	t.Cleanup(func() {
		_, _ = appService.StopSharing()
	})

	if err := service.OpenShareAddress(); err != nil {
		t.Fatalf("OpenShareAddress() error = %v", err)
	}
	if len(dialogs.openedURLs) != 1 {
		t.Fatalf("opened URLs = %v, want one URL", dialogs.openedURLs)
	}
	if dialogs.openedURLs[0] != status.Address {
		t.Fatalf("opened URL = %q, want %q", dialogs.openedURLs[0], status.Address)
	}
}

type recordingDialogs struct {
	openedURLs []string
}

func (d *recordingDialogs) ChooseDirectory(string) (string, error) {
	return "", nil
}

func (d *recordingDialogs) ChooseExportPath() (string, error) {
	return "", nil
}

func (d *recordingDialogs) OpenConfigDir() error {
	return nil
}

func (d *recordingDialogs) OpenURL(url string) error {
	d.openedURLs = append(d.openedURLs, url)
	return nil
}

func (d *recordingDialogs) ConfirmQuitWhileRunning(func()) error {
	return nil
}

func newTestAppService(t *testing.T) *app.Service {
	t.Helper()
	root := t.TempDir()
	port := freeTCPPort(t)
	path := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(path, []byte(`
server:
  road: `+root+`
  port: `+strconv.Itoa(port)+`
  allow_upload: false
  max_level: 1
account:
  secure: false
scratch:
  max_items: 500
  max_item_size: 10MB
`), 0o600); err != nil {
		t.Fatalf("WriteFile config: %v", err)
	}
	store := config.NewStore(path)
	cfg, err := store.Load()
	if err != nil {
		t.Fatalf("Load config: %v", err)
	}
	rt := runtime.New(runtime.SnapshotFromConfig(runtime.PersistentConfig{
		RootPath:            cfg.RootPath,
		Port:                cfg.Port,
		MaxLevel:            cfg.MaxLevel,
		IsAllowUpload:       cfg.IsAllowUpload,
		IsSecure:            cfg.IsSecure,
		Username:            cfg.Username,
		PasswordHash:        cfg.PasswordHash,
		ScratchMaxItems:     cfg.ScratchMaxItems,
		ScratchMaxItemSize:  cfg.ScratchMaxItemSize,
		ScratchMaxItemBytes: 10 * 1024 * 1024,
	}, runtime.Process{IP: "127.0.0.1", Pwd: root, SessionVal: "session"}))
	return app.NewService(app.Dependencies{
		ConfigStore: store,
		Runtime:     rt,
		Manager:     server.NewManager(rt),
	})
}

func freeTCPPort(t *testing.T) int {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen tcp: %v", err)
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port
}
