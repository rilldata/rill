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
	// ListBuckets lists the available buckets. pageSize limits the maximum results
	// returned in one call, and pageToken is non-empty when more results are available.
	// It returns the bucket names, the next page token, and any error.
	ListBuckets(ctx context.Context, pageSize uint32, pageToken string) ([]string, string, error)
	// ListObjects lists the objects and any directory-like prefixes under the given
	// bucket and path. Directories are included only when a delimiter (e.g., "/")
	// is provided. pageSize limits the maximum results returned in one call, and
	// pageToken is non-empty when more results are available. It returns the entries,
	// the next page token, and any error.
	ListObjects(ctx context.Context, bucket, path, delimiter string, pageSize uint32, pageToken string) ([]ObjectStoreEntry, string, error)
	// ListObjectsForGlob returns all objects in the given bucket whose paths match
	// the specified glob pattern. The pattern supports doublestar syntax, including
	// recursive patterns like "**". It returns the matching entries and any error.
	ListObjectsForGlob(ctx context.Context, bucket, glob string) ([]ObjectStoreEntry, error)
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
	URI    string         `mapstructure:"uri"` // Deprecated: use `path` instead
	Format FileFormat     `mapstructure:"format"`
	DuckDB map[string]any `mapstructure:"duckdb"` // Deprecated: use DuckDB directly
}

func (p *ObjectStoreModelInputProperties) Decode(props map[string]any) error {
	err := mapstructure.WeakDecode(props, p)
	if err != nil {
		return fmt.Errorf("failed to parse input properties: %w", err)
	}
	if p.Path == "" && p.URI == "" {
		return fmt.Errorf("missing property `path`")
	}
	if p.Path != "" && p.URI != "" {
		return fmt.Errorf("cannot specify both `path` and `uri`")
	}
	if p.URI != "" { // Backwards compatibility
		p.Path = p.URI
	}
	if !doublestar.ValidatePattern(p.Path) {
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
