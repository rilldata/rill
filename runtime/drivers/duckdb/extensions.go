package duckdb

import (
	"compress/gzip"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Since DuckDB is called from multiple packages, the extensions are installed in the init function
func init() {
	err := installExtensions()
	if err != nil {
		// If extensions cannot be installed, log the error and continue as the extensions can be downloaded
		// Should it be fatal in order to notice the issue prior to a release?
		log.Printf("Error preparing DuckDB extensions: %v", err)
	}
}

//go:embed embed/extensions/**/*
var embeddedFiles embed.FS

// installExtensions copies embedded DuckDB extensions to DuckDB's extensions directory to avoid downloading them
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
	embedPath := fmt.Sprintf("embed/extensions/%s/%s", duckdbVersion, platformName)
	duckdbExtensionsPath := filepath.Join(os.Getenv("HOME"), ".duckdb", "extensions", duckdbVersion, platformName)

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

	for {
		_, err = io.CopyN(output, gzipReader, 1024) // CopyN is used to prevent a warning (G110: Potential DoS vulnerability...)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}
	}

	return nil
}
