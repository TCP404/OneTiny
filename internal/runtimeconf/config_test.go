package runtimeconf

import (
	"sync"
	"testing"
)

func TestUpdateOnlyChangesPatchedFields(t *testing.T) {
	cfg := NewRuntimeConfig(ConfigSnapshot{
		RootPath:      "/tmp/one",
		Port:          9090,
		MaxLevel:      2,
		IsAllowUpload: false,
		IsSecure:      true,
		IP:            "127.0.0.1",
	})

	rootPath := "/tmp/two"
	maxLevel := uint8(4)
	allowUpload := true

	cfg.Update(ConfigPatch{
		RootPath:      &rootPath,
		MaxLevel:      &maxLevel,
		IsAllowUpload: &allowUpload,
	})

	got := cfg.Snapshot()
	if got.RootPath != rootPath {
		t.Fatalf("RootPath = %q, want %q", got.RootPath, rootPath)
	}
	if got.Port != 9090 {
		t.Fatalf("Port = %d, want unchanged 9090", got.Port)
	}
	if got.MaxLevel != maxLevel {
		t.Fatalf("MaxLevel = %d, want %d", got.MaxLevel, maxLevel)
	}
	if !got.IsAllowUpload {
		t.Fatal("IsAllowUpload = false, want true")
	}
	if !got.IsSecure {
		t.Fatal("IsSecure = false, want unchanged true")
	}
	if got.IP != "127.0.0.1" {
		t.Fatalf("IP = %q, want unchanged 127.0.0.1", got.IP)
	}
}

func TestSnapshotReturnsCopy(t *testing.T) {
	cfg := NewRuntimeConfig(ConfigSnapshot{
		RootPath: "/tmp/source",
		Port:     9090,
		IP:       "127.0.0.1",
	})

	snapshot := cfg.Snapshot()
	snapshot.RootPath = "/tmp/changed"
	snapshot.Port = 8080

	got := cfg.Snapshot()
	if got.RootPath != "/tmp/source" {
		t.Fatalf("RootPath = %q, want original /tmp/source", got.RootPath)
	}
	if got.Port != 9090 {
		t.Fatalf("Port = %d, want original 9090", got.Port)
	}
}

func TestCurrentCanBeSetAndLoaded(t *testing.T) {
	t.Cleanup(func() {
		SetCurrent(nil)
	})

	cfg := NewRuntimeConfig(ConfigSnapshot{RootPath: "/tmp/current", Port: 9090})
	SetCurrent(cfg)

	if got := Current(); got != cfg {
		t.Fatalf("Current() = %p, want %p", got, cfg)
	}
}

func TestConcurrentSnapshotAndUpdate(t *testing.T) {
	cfg := NewRuntimeConfig(ConfigSnapshot{
		RootPath: "/tmp/source",
		Port:     9090,
		MaxLevel: 1,
	})

	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func(port int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				if j%2 == 0 {
					cfg.Update(ConfigPatch{Port: &port})
				} else {
					_ = cfg.Snapshot()
				}
			}
		}(10000 + i)
	}

	wg.Wait()
}
