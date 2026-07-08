package safepath

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveWithinRootAcceptsNormalChild(t *testing.T) {
	root := t.TempDir()

	got, ok := ResolveWithinRoot(root, "child.txt")
	if !ok {
		t.Fatal("ResolveWithinRoot rejected normal child")
	}
	want, err := filepath.Abs(filepath.Join(root, "child.txt"))
	if err != nil {
		t.Fatalf("Abs returned error: %v", err)
	}
	if got != want {
		t.Fatalf("path = %q, want %q", got, want)
	}
}

func TestResolveWithinRootRejectsParentEscape(t *testing.T) {
	root := t.TempDir()

	if got, ok := ResolveWithinRoot(root, "../outside.txt"); ok {
		t.Fatalf("ResolveWithinRoot accepted escaped path %q", got)
	}
}

func TestResolveExistingWithinRootRejectsSymlinkEscape(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, "root")
	outside := filepath.Join(parent, "outside.txt")
	mkdirAll(t, root)
	writeFile(t, outside, "secret")
	symlinkOrSkip(t, outside, filepath.Join(root, "link.txt"))

	if got, ok := ResolveExistingWithinRoot(root, "link.txt"); ok {
		t.Fatalf("ResolveExistingWithinRoot accepted symlink escape %q", got)
	}
}

func TestResolveCreateWithinRootRejectsParentSymlinkEscape(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, "root")
	outside := filepath.Join(parent, "outside")
	mkdirAll(t, root)
	mkdirAll(t, outside)
	symlinkOrSkip(t, outside, filepath.Join(root, "uploads"))

	if got, ok := ResolveCreateWithinRoot(root, "uploads", "file.txt"); ok {
		t.Fatalf("ResolveCreateWithinRoot accepted parent symlink escape %q", got)
	}
}

func TestResolveCreateWithinRootRejectsExistingTargetSymlink(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, "root")
	outside := filepath.Join(parent, "outside.txt")
	mkdirAll(t, root)
	writeFile(t, outside, "secret")
	symlinkOrSkip(t, outside, filepath.Join(root, "target.txt"))

	if got, ok := ResolveCreateWithinRoot(root, "target.txt"); ok {
		t.Fatalf("ResolveCreateWithinRoot accepted existing target symlink %q", got)
	}
}

func TestResolveCreateWithinRootAcceptsNormalCreateTarget(t *testing.T) {
	root := t.TempDir()
	mkdirAll(t, filepath.Join(root, "uploads"))

	got, ok := ResolveCreateWithinRoot(root, "uploads", "file.txt")
	if !ok {
		t.Fatal("ResolveCreateWithinRoot rejected normal create target")
	}
	want, err := filepath.Abs(filepath.Join(root, "uploads", "file.txt"))
	if err != nil {
		t.Fatalf("Abs returned error: %v", err)
	}
	if got != want {
		t.Fatalf("path = %q, want %q", got, want)
	}
}

func mkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("MkdirAll %q: %v", path, err)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile %q: %v", path, err)
	}
}

func symlinkOrSkip(t *testing.T, oldname, newname string) {
	t.Helper()
	if err := os.Symlink(oldname, newname); err != nil {
		t.Skipf("os.Symlink unsupported or not permitted: %v", err)
	}
}
