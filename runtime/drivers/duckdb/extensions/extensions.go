package extensions

import (
	"compress/gzip"
	"database/sql"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	// once is used to ensure that DuckDB extensions are installed only once
	once sync.Once
	//go:embed embed/*
	embeddedFiles embed.FS
)

// InstallExtensionsOnce installs the embedded DuckDB extensions once
func InstallExtensionsOnce() error {
	var installErr error
	once.Do(func() {
		// Check if no extensions are embedded (for any versions and platform)
		// This is likely a development run or tests, return no error
		if noExtensions, err := noEmbeddedExtensions(); err != nil {
			installErr = err
			return
		} else if noExtensions {
			return
		}
		// Install embedded extensions
		installErr = installExtensions()
	})
	return installErr
}

// noEmbeddedExtensions checks if extensions are not embedded
func noEmbeddedExtensions() (bool, error) {
	empty := true

	err := fs.WalkDir(embeddedFiles, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && path != "." {
			empty = false
			return fs.SkipDir
		}
		return nil
	})

	return empty, err
}

// installExtensions installs the embedded DuckDB extensions
func installExtensions() error {
	// Connect to DuckDB and get the version
	db, err := sql.Open("duckdb", "")
	if err != nil {
		return err
	}
	defer db.Close()

	var duckdbVersion string
	err = db.QueryRow("SELECT version();").Scan(&duckdbVersion)
	if err != nil {
		return err
	}

	// Get the platform name using PRAGMA
	var platformName string
	err = db.QueryRow("PRAGMA platform;").Scan(&platformName)
	if err != nil {
		return err
	}

	// Define source and destination paths
	embedPath := fmt.Sprintf("embed/%s/%s", duckdbVersion, platformName)
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	duckdbExtensionsPath := filepath.Join(homeDir, ".duckdb", "extensions", duckdbVersion, platformName)

	// Create the destination directory if it doesn't exist
	err = os.MkdirAll(duckdbExtensionsPath, os.ModePerm)
	if err != nil {
		return err
	}

	// Walk through the embedded files and copy them to the destination directory
	err = fs.WalkDir(embeddedFiles, embedPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Determine the relative path and the destination path
		relativePath, err := filepath.Rel(embedPath, path)
		if err != nil {
			return err
		}

		// Remove the .gz extension for the destination file
		destPath := filepath.Join(duckdbExtensionsPath, strings.TrimSuffix(relativePath, ".gz"))

		// Check if the file already exists in the destination
		if _, err := os.Stat(destPath); err == nil {
			return nil
		} else if !os.IsNotExist(err) {
			return err
		}

		// Copy and uncompress the file
		err = uncompressEmbeddedFile(path, destPath)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func uncompressEmbeddedFile(src, dest string) error {
	input, err := embeddedFiles.Open(src)
	if err != nil {
		return err
	}
	defer input.Close()

	gzipReader, err := gzip.NewReader(input)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	output, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer output.Close()

	_, err = io.Copy(output, gzipReader) //nolint:gosec // Source is trusted, no risk of G110: Potential DoS vulnerability
	if err != nil {
		return err
	}

	return nil
}
