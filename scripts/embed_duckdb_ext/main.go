package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	_ "github.com/duckdb/duckdb-go/v2"
)

// DuckDB extensions Rill depends on
var extensions = []string{"json", "icu", "parquet", "httpfs", "sqlite_scanner", "motherduck"}

// DuckDB platforms to download extensions for
var platforms = []string{"linux_amd64", "linux_arm64", "osx_amd64", "osx_arm64"}

// Embed directory to store the extensions
var embedDirRoot = "runtime/drivers/duckdb/extensions/embed/"

// main downloads and embeds the necessary DuckDB extensions for the current DuckDB version and pre-defined platforms
// https://duckdb.org/docs/extensions/working_with_extensions.html
func main() {
	// Connect to DuckDB and get the version
	db, err := sql.Open("duckdb", "")
	if err != nil {
		log.Fatalf("Failed to connect to DuckDB: %v", err)
	}

	var duckdbVersion string
	err = db.QueryRow("SELECT version();").Scan(&duckdbVersion)
	if err != nil {
		db.Close()
		log.Fatalf("Failed to get DuckDB version: %v", err)
	}
	log.Printf("DuckDB version: %s", duckdbVersion)

	err = db.Close()
	if err != nil {
		log.Fatalf("Failed to close DuckDB connection: %v", err)
	}

	// Download each extension for each platform
	// This might be done in parallel but download time will be almost the same
	for _, platform := range platforms {
		destDir := fmt.Sprintf(embedDirRoot+"%s/%s", platform, duckdbVersion)
		err := os.MkdirAll(destDir, os.ModePerm)
		if err != nil && !os.IsExist(err) {
			log.Fatalf("Failed to create destination directory %s: %v", destDir, err)
		}

		for _, extension := range extensions {
			// Make sure an embedded extension is compressed (gzipped) to save space
			url := fmt.Sprintf("http://extensions.duckdb.org/%s/%s/%s.duckdb_extension.gz", duckdbVersion, platform, extension)
			destPath := filepath.Join(destDir, fmt.Sprintf("%s.duckdb_extension.gz", extension))

			// Check if the extension already exists and skip downloading
			if _, err := os.Stat(destPath); err == nil {
				log.Printf("Extension %s already exists at %s", extension, destPath)
				continue
			}

			// Download the extension
			err = downloadFile(url, destPath)
			if err != nil {
				log.Fatalf("Failed to download %s: %v", url, err)
			}

			log.Printf("Downloaded %s to %s", url, destPath)
		}
	}

	log.Println("All necessary DuckDB extensions have been processed.")
}

// downloadFile downloads a file from a URL and saves it to a destination path
func downloadFile(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to perform HTTP GET request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected HTTP status: %s", resp.Status)
	}

	out, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to copy response body to file: %w", err)
	}

	return nil
}
