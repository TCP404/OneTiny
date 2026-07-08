package structure_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestGoModulePathMatchesRepository(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	data, err := os.ReadFile(filepath.Join(repoRoot, "go.mod"))
	if err != nil {
		t.Fatalf("ReadFile(go.mod): %v", err)
	}

	firstLine := strings.SplitN(string(data), "\n", 2)[0]
	const want = "module github.com/tcp404/OneTiny"
	if firstLine != want {
		t.Fatalf("module line = %q, want %q", firstLine, want)
	}
}
