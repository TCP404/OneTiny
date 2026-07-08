package structure_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestCmdRootContainsOnlyCommandDirectories(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	cmdDir := filepath.Join(repoRoot, "cmd")

	entries, err := os.ReadDir(cmdDir)
	if err != nil {
		t.Fatalf("ReadDir(%s): %v", cmdDir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			switch entry.Name() {
			case "cli", "gui":
			default:
				t.Errorf("unexpected cmd subdirectory %q", entry.Name())
			}
			continue
		}
		if filepath.Ext(entry.Name()) == ".go" {
			t.Errorf("cmd root contains Go file %q; move command-specific code under cmd/cli or cmd/gui", entry.Name())
		}
	}
}
