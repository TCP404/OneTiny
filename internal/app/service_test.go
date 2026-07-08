package app

import (
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/tcp404/OneTiny/internal/config"
	"github.com/tcp404/OneTiny/internal/runtime"
	"github.com/tcp404/OneTiny/internal/server"
)

func freeTCPPort(t *testing.T) int {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen tcp: %v", err)
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port
}

func newTestService(t *testing.T) (*Service, *config.Store, *runtime.Runtime) {
	t.Helper()
	root := t.TempDir()
	path := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(path, []byte(`
server:
  road: `+root+`
  port: 0
  allow_upload: false
  max_level: 1
account:
  secure: false
`), 0o600); err != nil {
		t.Fatalf("WriteFile config: %v", err)
	}
	store := config.NewStore(path)
	cfg, err := store.Load()
	if err != nil {
		t.Fatalf("Load config: %v", err)
	}
	process := runtime.Process{IP: "127.0.0.1", Pwd: root, SessionVal: "session"}
	rt := runtime.New(runtime.SnapshotFromConfig(runtimeConfigFromConfig(cfg), process))
	svc := NewService(Dependencies{
		ConfigStore: store,
		Runtime:     rt,
		Manager:     server.NewManager(rt),
	})
	return svc, store, rt
}

func TestNewServiceReportsInitialStatus(t *testing.T) {
	svc, store, _ := newTestService(t)
	status, err := svc.GetStatus()
	if err != nil {
		t.Fatalf("GetStatus() error = %v", err)
	}
	if status.Running {
		t.Fatal("new service should not report running")
	}
	if status.Config.Port != store.Current().Port {
		t.Fatalf("status port = %d, want %d", status.Config.Port, store.Current().Port)
	}
	if status.ConfigPath != store.Path() {
		t.Fatalf("config path = %q, want %q", status.ConfigPath, store.Path())
	}
}

func TestUpdateConfigPersistsAndUpdatesRuntime(t *testing.T) {
	svc, store, rt := newTestService(t)
	nextPort := 12345
	status, err := svc.UpdateConfig(ConfigPatchDTO{Port: &nextPort})
	if err != nil {
		t.Fatalf("UpdateConfig() error = %v", err)
	}
	if store.Current().Port != nextPort {
		t.Fatalf("stored port = %d, want %d", store.Current().Port, nextPort)
	}
	if rt.Snapshot().Port != nextPort {
		t.Fatalf("runtime port = %d, want %d", rt.Snapshot().Port, nextPort)
	}
	if status.Config.Port != nextPort {
		t.Fatalf("status port = %d, want %d", status.Config.Port, nextPort)
	}
}

func TestUpdateConfigPersistsScratchLimitsAndUpdatesRuntime(t *testing.T) {
	svc, store, rt := newTestService(t)
	maxItems := 25
	maxSize := "2MB"

	status, err := svc.UpdateConfig(ConfigPatchDTO{
		ScratchMaxItems:    &maxItems,
		ScratchMaxItemSize: &maxSize,
	})
	if err != nil {
		t.Fatalf("UpdateConfig returned error: %v", err)
	}

	if store.Current().ScratchMaxItems != 25 || store.Current().ScratchMaxItemSize != "2MB" {
		t.Fatalf("stored scratch config = %d %q", store.Current().ScratchMaxItems, store.Current().ScratchMaxItemSize)
	}
	snapshot := rt.Snapshot()
	if snapshot.ScratchMaxItems != 25 || snapshot.ScratchMaxItemSize != "2MB" || snapshot.ScratchMaxItemBytes != 2*1024*1024 {
		t.Fatalf("runtime scratch = %+v", snapshot)
	}
	if status.Config.ScratchMaxItems != 25 || status.Config.ScratchMaxItemSize != "2MB" {
		t.Fatalf("status scratch = %+v", status.Config)
	}
}

func TestUpdateConfigRestartPortSyncsScratchLimits(t *testing.T) {
	svc, store, rt := newTestService(t)
	startPort := freeTCPPort(t)
	nextPort := freeTCPPort(t)
	if _, err := svc.UpdateConfig(ConfigPatchDTO{Port: &startPort}); err != nil {
		t.Fatalf("UpdateConfig start port: %v", err)
	}
	if _, err := svc.StartSharing(); err != nil {
		t.Fatalf("StartSharing: %v", err)
	}
	defer func() {
		if _, err := svc.StopSharing(); err != nil {
			t.Fatalf("StopSharing: %v", err)
		}
	}()

	maxItems := 25
	maxSize := "2MB"
	status, err := svc.UpdateConfig(ConfigPatchDTO{
		Port:               &nextPort,
		RestartPort:        true,
		ScratchMaxItems:    &maxItems,
		ScratchMaxItemSize: &maxSize,
	})
	if err != nil {
		t.Fatalf("UpdateConfig restart port: %v", err)
	}

	if store.Current().Port != nextPort || store.Current().ScratchMaxItems != 25 || store.Current().ScratchMaxItemSize != "2MB" {
		t.Fatalf("stored config = %+v", store.Current())
	}
	snapshot := rt.Snapshot()
	if snapshot.Port != nextPort || snapshot.ScratchMaxItems != 25 || snapshot.ScratchMaxItemSize != "2MB" || snapshot.ScratchMaxItemBytes != 2*1024*1024 {
		t.Fatalf("runtime snapshot = %+v", snapshot)
	}
	if !status.Running {
		t.Fatal("status should report running after confirmed restart")
	}
	if status.Config.Port != nextPort || status.Config.ScratchMaxItems != 25 || status.Config.ScratchMaxItemSize != "2MB" {
		t.Fatalf("status config = %+v", status.Config)
	}
}
