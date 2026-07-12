package main

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pkg/errors"
)

func WriteChecksums(distDir string, outputPath string) error {
	entries, err := os.ReadDir(distDir)
	if err != nil {
		return err
	}

	var zipNames []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(entry.Name(), ".zip") {
			zipNames = append(zipNames, entry.Name())
		}
	}
	sort.Strings(zipNames)

	var builder strings.Builder
	for _, name := range zipNames {
		content, err := os.ReadFile(filepath.Join(distDir, name))
		if err != nil {
			return err
		}
		sum := sha256.Sum256(content)
		builder.WriteString(hex.EncodeToString(sum[:]))
		builder.WriteString("  ")
		builder.WriteString(name)
		builder.WriteByte('\n')
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return err
	}
	if len(zipNames) == 0 {
		return errors.Errorf("no zip files found in %s", distDir)
	}
	return os.WriteFile(outputPath, []byte(builder.String()), 0o644)
}
