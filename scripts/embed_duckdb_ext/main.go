package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	_ "github.com/marcboeker/go-duckdb"
)

// DuckDB extensions Rill depends on
var extensions = []string{"json", "icu", "parquet", "httpfs", "sqlite_scanner", "motherduck"}

// DuckDB platforms to download extensions for
var platforms = []string{"linux_amd64", "osx_amd64", "osx_arm64"}

// Embed directory to store the extensions
var embedDirRoot = "runtime/drivers/duckdb/extensions/embed/"

// main downloads and embeds the necessary DuckDB extensions for the current DuckDB version and pre-defined platforms
// https://duckdb.org/docs/extensions/working_with_extensions.html
func main() {
	// Parse command-line arguments
	platformFlag := flag.String("platform", "", "Specify the platform to download extensions for (e.g., linux_amd64). If not provided, downloads extensions for all platforms.")
	flag.Parse()

	// Pre-clean: Remove all existing directories from embedDirRoot
	err := removeAllDirectories(embedDirRoot)
	if err != nil {
		log.Fatalf("Failed to remove existing directories in %s: %v", embedDirRoot, err)
	}

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

	// Determine the platforms to process
	var platformsToProcess []string
	if *platformFlag != "" {
		platformsToProcess = []string{*platformFlag}
	} else {
		platformsToProcess = platforms
	}

	// Download each extension for each platform
	for _, platform := range platformsToProcess {
		destDir := fmt.Sprintf(embedDirRoot+"%s/%s", duckdbVersion, platform)
		err := os.MkdirAll(destDir, os.ModePerm)
		if err != nil && !os.IsExist(err) {
			log.Fatalf("Failed to create destination directory %s: %v", destDir, err)
		}

		for _, extension := range extensions {
			// Make sure an embedded extension is compressed (gzipped) to save space
			url := fmt.Sprintf("http://extensions.duckdb.org/%s/%s/%s.duckdb_extension.gz", duckdbVersion, platform, extension)
			destPath := filepath.Join(destDir, fmt.Sprintf("%s.duckdb_extension.gz", extension))

			// Download the extension (silently overwrite if it exists)
			err = downloadFile(url, destPath)
			if err != nil {
				log.Fatalf("Failed to download %s: %v", url, err)
			}

			log.Printf("Downloaded %s to %s", url, destPath)
		}
	}

	log.Println("All necessary DuckDB extensions have been processed.")
}

// removeAllDirectories removes all directories under the specified root directory
func removeAllDirectories(root string) error {
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Directory does not exist, nothing to remove
		}
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			err := os.RemoveAll(filepath.Join(root, entry.Name()))
			if err != nil {
				return err
			}
		}
	}
	return nil
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
