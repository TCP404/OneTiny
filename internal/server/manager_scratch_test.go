package server

import (
	"testing"

	"github.com/tcp404/OneTiny/internal/runtime"
	"github.com/tcp404/OneTiny/internal/scratch"
)

func TestNewManagerCreatesScratchStoreFromRuntime(t *testing.T) {
	rt := runtime.New(runtime.Snapshot{ScratchMaxItems: 9, ScratchMaxItemBytes: 1024})
	manager := NewManager(rt)
	if manager.Scratch() == nil {
		t.Fatal("Scratch store is nil")
	}
	if got := manager.Scratch().Limits().MaxItems; got != 9 {
		t.Fatalf("MaxItems = %d, want 9", got)
	}
	if got := manager.Scratch().Limits().MaxItemBytes; got != 1024 {
		t.Fatalf("MaxItemBytes = %d, want 1024", got)
	}
}

func TestNewManagerUsesInjectedScratchStore(t *testing.T) {
	store, err := scratch.NewStore(scratch.Limits{MaxItems: 3, MaxItemBytes: 128})
	if err != nil {
		t.Fatal(err)
	}
	rt := runtime.New(runtime.Snapshot{ScratchMaxItems: 9, ScratchMaxItemBytes: 1024})
	manager := NewManagerWithDependencies(Dependencies{Runtime: rt, Scratch: store})
	if manager.Scratch() != store {
		t.Fatal("manager did not use injected scratch store")
	}
}

func TestApplyRuntimeUpdatesScratchStoreLimits(t *testing.T) {
	rt := runtime.New(runtime.Snapshot{ScratchMaxItems: 3, ScratchMaxItemBytes: 128})
	manager := NewManager(rt)

	initial := manager.Scratch().Limits()
	if initial.MaxItems != 3 || initial.MaxItemBytes != 128 {
		t.Fatalf("initial scratch limits = %+v, want MaxItems 3 MaxItemBytes 128", initial)
	}

	maxItems := 7
	maxItemBytes := int64(2048)
	if err := manager.ApplyRuntime(runtime.Patch{
		ScratchMaxItems:     &maxItems,
		ScratchMaxItemBytes: &maxItemBytes,
	}); err != nil {
		t.Fatalf("ApplyRuntime() error = %v", err)
	}

	got := manager.Scratch().Limits()
	if got.MaxItems != 7 {
		t.Fatalf("MaxItems = %d, want 7", got.MaxItems)
	}
	if got.MaxItemBytes != 2048 {
		t.Fatalf("MaxItemBytes = %d, want 2048", got.MaxItemBytes)
	}
}

func TestRestartWithSnapshotAppliesScratchRuntimeFields(t *testing.T) {
	rt := runtime.New(runtime.Snapshot{
		RootPath:            t.TempDir(),
		Port:                0,
		MaxLevel:            1,
		SessionVal:          "session",
		ScratchMaxItems:     2,
		ScratchMaxItemSize:  "2KB",
		ScratchMaxItemBytes: 2048,
	})
	manager := NewManager(rt)

	next := rt.Snapshot()
	next.ScratchMaxItems = 5
	next.ScratchMaxItemSize = "4KB"
	next.ScratchMaxItemBytes = 4096
	if err := manager.RestartWithSnapshot(next, nil); err != nil {
		t.Fatalf("RestartWithSnapshot() error = %v", err)
	}
	defer func() {
		if err := manager.Stop(); err != nil {
			t.Fatalf("Stop() error = %v", err)
		}
	}()

	gotSnapshot := rt.Snapshot()
	if gotSnapshot.ScratchMaxItems != 5 {
		t.Fatalf("runtime ScratchMaxItems = %d, want 5", gotSnapshot.ScratchMaxItems)
	}
	if gotSnapshot.ScratchMaxItemSize != "4KB" {
		t.Fatalf("runtime ScratchMaxItemSize = %q, want %q", gotSnapshot.ScratchMaxItemSize, "4KB")
	}
	if gotSnapshot.ScratchMaxItemBytes != 4096 {
		t.Fatalf("runtime ScratchMaxItemBytes = %d, want 4096", gotSnapshot.ScratchMaxItemBytes)
	}

	gotLimits := manager.Scratch().Limits()
	if gotLimits.MaxItems != 5 {
		t.Fatalf("MaxItems = %d, want 5", gotLimits.MaxItems)
	}
	if gotLimits.MaxItemBytes != 4096 {
		t.Fatalf("MaxItemBytes = %d, want 4096", gotLimits.MaxItemBytes)
	}
}
