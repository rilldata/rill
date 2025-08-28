package drivers

import (
	"context"
	"fmt"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/mitchellh/mapstructure"
)

// ObjectStore is an interface for object storage systems.
type ObjectStore interface {
	// ListObjects returns the paths that match the given properties.
	// It resolves globs with support for all patterns supported by the doublestar package (notably "**").
	ListObjects(ctx context.Context, path string) ([]ObjectStoreEntry, error)
	// DownloadFiles provides an iterator for downloading and consuming files.
	// It resolves globs similar to ListObjects.
	DownloadFiles(ctx context.Context, path string) (FileIterator, error)
}

// ObjectStoreEntry represents a file listing in an object store.
type ObjectStoreEntry struct {
	Path      string
	IsDir     bool
	Size      int64
	UpdatedOn time.Time
}

// FileIterator provides ways to iteratively download files from external sources
// Clients should call close once they are done with iterator to release any resources
type FileIterator interface {
	// Close do cleanup and release resources
	Close() error
	// Format returns general file format (json, csv, parquet, etc)
	// Returns an empty string if there is no general format
	Format() string
	// SetKeepFilesUntilClose configures the iterator to keep all files until Close() is called.
	SetKeepFilesUntilClose()
	// Next returns a list of file downloaded from external sources
	// and cleanups file created in previous batch
	Next(ctx context.Context) ([]string, error)
}

// ObjectStoreModelInputProperties contain common input properties for object store models.
type ObjectStoreModelInputProperties struct {
	Path   string         `mapstructure:"path"`
	SQL    string         `mapstructure:"sql"` // SQL to be executed on the output connector if it supports SQL. Cannot be combined with `path`.
	URI    string         `mapstructure:"uri"` // Deprecated: use `path` instead
	Format FileFormat     `mapstructure:"format"`
	DuckDB map[string]any `mapstructure:"duckdb"` // Deprecated: use DuckDB directly
}

func (p *ObjectStoreModelInputProperties) Decode(props map[string]any) error {
	err := mapstructure.WeakDecode(props, p)
	if err != nil {
		return fmt.Errorf("failed to parse input properties: %w", err)
	}
	if p.URI != "" { // Backwards compatibility
		p.Path = p.URI
	}
	if p.Path == "" && p.SQL == "" {
		return fmt.Errorf("missing property `path` or `sql`")
	}
	if p.Path != "" && p.SQL != "" {
		return fmt.Errorf("cannot specify both `path` and `sql`")
	}
	if p.Path != "" && !doublestar.ValidatePattern(p.Path) {
		return fmt.Errorf("glob pattern %q is invalid", p.Path)
	}
	return nil
}

// ObjectStoreModelOutputProperties contain common output properties for object store models.
type ObjectStoreModelOutputProperties struct {
	Path   string     `mapstructure:"path"`
	Format FileFormat `mapstructure:"format"`
}

// ObjectStoreModelResultProperties contain common result properties for object store models.
type ObjectStoreModelResultProperties struct {
	Path   string `mapstructure:"path"`
	Format string `mapstructure:"format"`
}
