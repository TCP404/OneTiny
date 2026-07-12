package main

import (
	"archive/zip"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

func ArchiveZip(inputPath string, outputPath string, entryName string) error {
	info, err := os.Stat(inputPath)
	if err != nil {
		return err
	}
	if entryName == "" {
		entryName = filepath.Base(inputPath)
	}
	entryName = filepath.ToSlash(entryName)

	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return err
	}

	output, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer output.Close()

	writer := zip.NewWriter(output)
	defer writer.Close()

	if !info.IsDir() {
		return addFileToZip(writer, inputPath, entryName, info)
	}

	return filepath.WalkDir(inputPath, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		fileInfo, err := entry.Info()
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(inputPath, path)
		if err != nil {
			return err
		}
		zipName := filepath.ToSlash(filepath.Join(entryName, rel))
		return addFileToZip(writer, path, zipName, fileInfo)
	})
}

func addFileToZip(writer *zip.Writer, path string, name string, info os.FileInfo) error {
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Name = filepath.ToSlash(name)
	header.Method = zip.Deflate

	entry, err := writer.CreateHeader(header)
	if err != nil {
		return err
	}

	input, err := os.Open(path)
	if err != nil {
		return err
	}
	defer input.Close()

	_, err = io.Copy(entry, input)
	return err
}
