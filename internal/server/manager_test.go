package server

import (
	"errors"
	"testing"

	"github.com/tcp404/OneTiny/internal/runtime"
)

func newTestRuntime(t *testing.T) *runtime.Runtime {
	t.Helper()
	return runtime.New(runtime.Snapshot{
		RootPath:   t.TempDir(),
		Port:       0,
		MaxLevel:   1,
		SessionVal: "session",
	})
}

func TestManagerStartStop(t *testing.T) {
	manager := NewManager(newTestRuntime(t))
	if err := manager.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if !manager.Running() {
		t.Fatal("manager should report running after Start")
	}
	if err := manager.Stop(); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	if manager.Running() {
		t.Fatal("manager should not report running after Stop")
	}
}

func TestManagerApplyRuntime(t *testing.T) {
	rt := newTestRuntime(t)
	manager := NewManager(rt)
	port := 12345
	if err := manager.ApplyRuntime(runtime.Patch{Port: &port}); err != nil {
		t.Fatalf("ApplyRuntime() error = %v", err)
	}
	if rt.Snapshot().Port != port {
		t.Fatalf("runtime port = %d, want %d", rt.Snapshot().Port, port)
	}
}

func TestManagerStopWhenNotRunning(t *testing.T) {
	manager := NewManager(newTestRuntime(t))
	if err := manager.Stop(); !errors.Is(err, ErrServerNotRunning) {
		t.Fatalf("Stop() error = %v, want %v", err, ErrServerNotRunning)
	}
}
