package main

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func TestArchiveZipCreatesSingleFileArchive(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "OneTiny.exe")
	output := filepath.Join(dir, "onetiny-gui-windows-x64.zip")
	if err := os.WriteFile(input, []byte("binary"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := ArchiveZip(input, output, "OneTiny.exe"); err != nil {
		t.Fatal(err)
	}

	reader, err := zip.OpenReader(output)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()

	if len(reader.File) != 1 {
		t.Fatalf("archive contains %d files, want 1", len(reader.File))
	}
	if reader.File[0].Name != "OneTiny.exe" {
		t.Fatalf("archive entry = %q", reader.File[0].Name)
	}
}

func TestWriteChecksumsUsesSortedZipNames(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "b.zip"), "second")
	writeTestFile(t, filepath.Join(dir, "a.zip"), "first")
	writeTestFile(t, filepath.Join(dir, "ignore.txt"), "not included")

	output := filepath.Join(dir, "onetiny-checksums.txt")
	if err := WriteChecksums(dir, output); err != nil {
		t.Fatal(err)
	}

	firstHash := sha256.Sum256([]byte("first"))
	secondHash := sha256.Sum256([]byte("second"))
	want := hex.EncodeToString(firstHash[:]) + "  a.zip\n" +
		hex.EncodeToString(secondHash[:]) + "  b.zip\n"

	got, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != want {
		t.Fatalf("checksum file = %q, want %q", string(got), want)
	}
}

func writeTestFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
