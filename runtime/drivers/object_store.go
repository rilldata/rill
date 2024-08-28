package drivers

import (
	"context"
	"time"
)

type ObjectStore interface {
	// ListObjects returns the paths that match the given properties.
	// It resolves globs with support for all patterns supported by the doublestar package (notably "**").
	ListObjects(ctx context.Context, props map[string]any) ([]ObjectStoreEntry, error)
	// DownloadFiles provides an iterator for downloading and consuming files
	DownloadFiles(ctx context.Context, props map[string]any) (FileIterator, error)
}

// ObjectStoreEntry represents a file listing in an object store.
type ObjectStoreEntry struct {
	Path      string
	IsDir     bool
	UpdatedOn time.Time
}

// FileIterator provides ways to iteratively download files from external sources
// Clients should call close once they are done with iterator to release any resources
type FileIterator interface {
	// Close do cleanup and release resources
	Close() error
	// Next returns a list of file downloaded from external sources
	// and cleanups file created in previous batch
	Next() ([]string, error)
	// Format returns general file format (json, csv, parquet, etc)
	// Returns an empty string if there is no general format
	Format() string
}

type ObjectStoreModelOutputProperties struct {
	Path   string     `mapstructure:"path"`
	Format FileFormat `mapstructure:"format"`
}

type ObjectStoreModelResultProperties struct {
	Path   string `mapstructure:"path"`
	Format string `mapstructure:"format"`
}
