package server

import (
	"errors"
	"net"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/tcp404/OneTiny/internal/runtimeconf"
)

func TestServiceManagerStartStopRunningStatus(t *testing.T) {
	cfg := runtimeconf.NewRuntimeConfig(runtimeconf.ConfigSnapshot{
		RootPath:      t.TempDir(),
		Port:          0,
		MaxLevel:      3,
		IsAllowUpload: true,
		IsSecure:      false,
		IP:            "127.0.0.1",
	})
	manager := NewServiceManager(cfg)

	if manager.Running() {
		t.Fatal("Running() = true before Start")
	}

	if err := manager.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	t.Cleanup(func() {
		if manager.Running() {
			_ = manager.Stop()
		}
	})

	if !manager.Running() {
		t.Fatal("Running() = false after Start")
	}
	if err := manager.Start(); !errors.Is(err, ErrServerAlreadyRunning) {
		t.Fatalf("second Start() error = %v, want ErrServerAlreadyRunning", err)
	}

	status := manager.Status()
	if status.Port != 0 {
		t.Fatalf("Status().Port = %d, want 0", status.Port)
	}
	if status.MaxLevel != 3 {
		t.Fatalf("Status().MaxLevel = %d, want 3", status.MaxLevel)
	}
	if !status.IsAllowUpload {
		t.Fatal("Status().IsAllowUpload = false, want true")
	}

	if err := manager.Stop(); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	if manager.Running() {
		t.Fatal("Running() = true after Stop")
	}
	if err := manager.Stop(); !errors.Is(err, ErrServerNotRunning) {
		t.Fatalf("second Stop() error = %v, want ErrServerNotRunning", err)
	}
}

func TestServiceManagerRestartStartsWhenStopped(t *testing.T) {
	cfg := runtimeconf.NewRuntimeConfig(runtimeconf.ConfigSnapshot{
		RootPath: t.TempDir(),
		Port:     0,
	})
	manager := NewServiceManager(cfg)

	if err := manager.Restart(); err != nil {
		t.Fatalf("Restart() error = %v", err)
	}
	t.Cleanup(func() {
		if manager.Running() {
			_ = manager.Stop()
		}
	})

	if !manager.Running() {
		t.Fatal("Running() = false after Restart")
	}
}

func TestServiceManagerStartInstallsRuntimeConfig(t *testing.T) {
	t.Cleanup(func() {
		runtimeconf.SetCurrent(nil)
	})
	runtimeconf.SetCurrent(nil)

	cfg := runtimeconf.NewRuntimeConfig(runtimeconf.ConfigSnapshot{
		RootPath: t.TempDir(),
		Port:     0,
	})
	manager := NewServiceManager(cfg)
	if err := manager.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	t.Cleanup(func() {
		if manager.Running() {
			_ = manager.Stop()
		}
	})

	if got := runtimeconf.Current(); got != cfg {
		t.Fatalf("runtimeconf.Current() = %p, want %p", got, cfg)
	}
}

func TestServiceManagerApplyRuntimeConfig(t *testing.T) {
	cfg := runtimeconf.NewRuntimeConfig(runtimeconf.ConfigSnapshot{
		RootPath: t.TempDir(),
		Port:     0,
		MaxLevel: 0,
	})
	manager := NewServiceManager(cfg)

	nextRoot := t.TempDir()
	upload := true
	level := uint8(2)
	if err := manager.ApplyRuntimeConfig(runtimeconf.ConfigPatch{
		RootPath:      &nextRoot,
		IsAllowUpload: &upload,
		MaxLevel:      &level,
	}); err != nil {
		t.Fatalf("ApplyRuntimeConfig() error = %v", err)
	}

	if got := manager.Config(); got != cfg {
		t.Fatalf("Config() = %p, want %p", got, cfg)
	}
	got := manager.Status()
	if got.RootPath != nextRoot || !got.IsAllowUpload || got.MaxLevel != 2 {
		t.Fatalf("status = %+v, want updated runtime config", got)
	}
}

func TestServiceManagerRestartWithSnapshotPrepareFailureKeepsOldServer(t *testing.T) {
	cfg := runtimeconf.NewRuntimeConfig(runtimeconf.ConfigSnapshot{
		RootPath: t.TempDir(),
		Port:     freeServerTestPort(t),
		IP:       "127.0.0.1",
	})
	manager := NewServiceManager(cfg)
	if err := manager.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	t.Cleanup(func() {
		if manager.Running() {
			_ = manager.Stop()
		}
	})
	oldPort := cfg.Snapshot().Port
	if !waitServerTCPReachable("127.0.0.1", oldPort, true) {
		t.Fatalf("old port %d did not become reachable", oldPort)
	}

	prepareErr := errors.New("prepare new server failed")
	previousBuilder := buildHTTPServer
	buildHTTPServer = func(listener net.Listener) (*http.Server, error) {
		return nil, prepareErr
	}
	t.Cleanup(func() {
		buildHTTPServer = previousBuilder
	})

	next := cfg.Snapshot()
	next.Port = freeServerTestPort(t)
	err := manager.RestartWithSnapshot(next, nil)
	if !errors.Is(err, prepareErr) {
		t.Fatalf("RestartWithSnapshot error = %v, want %v", err, prepareErr)
	}
	if !manager.Running() {
		t.Fatal("manager stopped old server after prepare failure")
	}
	if got := manager.Status().Port; got != oldPort {
		t.Fatalf("runtime port = %d, want old port %d", got, oldPort)
	}
	if !waitServerTCPReachable("127.0.0.1", oldPort, true) {
		t.Fatalf("old port %d should remain reachable after prepare failure", oldPort)
	}
}

func TestServeDoesNotClearStateWhileStopping(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}
	if err := listener.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	srv := &http.Server{}
	done := make(chan error, 1)
	manager := &ServiceManager{
		srv:      srv,
		listener: listener,
		done:     done,
		stopping: true,
	}

	manager.serve(srv, listener, done)
	if err := <-done; err != nil {
		t.Fatalf("serve() done error = %v", err)
	}

	if manager.srv != srv {
		t.Fatal("serve() cleared srv while Stop owns cleanup")
	}
	if manager.listener != listener {
		t.Fatal("serve() cleared listener while Stop owns cleanup")
	}
	if manager.done != done {
		t.Fatal("serve() cleared done while Stop owns cleanup")
	}
	if !manager.stopping {
		t.Fatal("serve() cleared stopping while Stop owns cleanup")
	}
}

func TestServiceManagerDoneRemainsAvailableAfterServeExits(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}
	if err := listener.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	srv := &http.Server{}
	done := make(chan error, 1)
	manager := &ServiceManager{
		srv:      srv,
		listener: listener,
		done:     done,
	}

	manager.serve(srv, listener, done)
	if err := <-done; err != nil {
		t.Fatalf("serve() done error = %v", err)
	}
	if manager.Done() != done {
		t.Fatal("Done() did not keep the serve channel available after serve exited")
	}
	if manager.Running() {
		t.Fatal("Running() = true after serve exited")
	}
}

func TestRunCoreWithManagerReturnsOnServeError(t *testing.T) {
	serveErr := errors.New("serve failed")
	manager := &fakeCoreManager{done: make(chan error, 1)}
	manager.done <- serveErr
	signals := make(chan os.Signal)

	shouldExit := runCoreWithManager(manager, signals, func() {})

	if shouldExit {
		t.Fatal("runCoreWithManager() exit = true, want false for serve error")
	}
	if manager.started != 1 {
		t.Fatalf("Start() calls = %d, want 1", manager.started)
	}
	if manager.stopped != 0 {
		t.Fatalf("Stop() calls = %d, want 0", manager.stopped)
	}
}

func TestRunCoreWithManagerStopsOnSignal(t *testing.T) {
	manager := &fakeCoreManager{done: make(chan error)}
	signals := make(chan os.Signal, 1)
	signals <- os.Interrupt

	shouldExit := runCoreWithManager(manager, signals, func() {})

	if !shouldExit {
		t.Fatal("runCoreWithManager() exit = false, want true for signal")
	}
	if manager.started != 1 {
		t.Fatalf("Start() calls = %d, want 1", manager.started)
	}
	if manager.stopped != 1 {
		t.Fatalf("Stop() calls = %d, want 1", manager.stopped)
	}
}

type fakeCoreManager struct {
	startErr error
	stopErr  error
	done     chan error
	started  int
	stopped  int
}

func (m *fakeCoreManager) Start() error {
	m.started++
	return m.startErr
}

func (m *fakeCoreManager) Stop() error {
	m.stopped++
	return m.stopErr
}

func freeServerTestPort(t *testing.T) int {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port
}

func waitServerTCPReachable(host string, port int, want bool) bool {
	deadline := time.Now().Add(500 * time.Millisecond)
	address := net.JoinHostPort(host, strconv.Itoa(port))
	for {
		conn, err := net.DialTimeout("tcp", address, 25*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			if want {
				return true
			}
		} else if !want {
			return false
		}
		if time.Now().After(deadline) {
			return false
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func (m *fakeCoreManager) Done() <-chan error {
	return m.done
}
