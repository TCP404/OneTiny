package structure_test

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestRepositoryOwnedPathsUseLowercaseOwner(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	uppercaseOwner := "TCP" + "404"

	var matches []string
	err := filepath.WalkDir(repoRoot, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			switch entry.Name() {
			case ".git", "node_modules":
				return filepath.SkipDir
			}
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if !utf8.Valid(data) || bytes.IndexByte(data, 0) >= 0 {
			return nil
		}
		if strings.Contains(string(data), uppercaseOwner) {
			rel, err := filepath.Rel(repoRoot, path)
			if err != nil {
				rel = path
			}
			matches = append(matches, rel)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("WalkDir(%s): %v", repoRoot, err)
	}
	if len(matches) > 0 {
		t.Fatalf("uppercase owner found outside canonical external dependencies: %s", strings.Join(matches, ", "))
	}
}
