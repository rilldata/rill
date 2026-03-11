package drivers

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/pkg/pagination"
)

// ObjectStore is an interface for object storage systems.
type ObjectStore interface {
	// ListBuckets lists the available buckets. pageSize limits the maximum results
	// returned in one call, and pageToken is non-empty when more results are available.
	// It returns the bucket names, the next page token, and any error.
	ListBuckets(ctx context.Context, pageSize uint32, pageToken string) ([]string, string, error)
	// ListObjects lists objects and directory-like prefixes under the given bucket and path,
	// using the provided delimiter (defaults to "/"). This is a non-recursive listing.
	// pageSize limits results, and pageToken is non-empty when more results are available.
	// It returns the entries, the next page token, and any error.
	ListObjects(ctx context.Context, bucket, path, delimiter string, pageSize uint32, pageToken string) ([]ObjectStoreEntry, string, error)
	// ListObjectsForGlob returns all objects in the given bucket whose paths match
	// the specified glob pattern. The pattern supports doublestar syntax, including
	// recursive patterns like "**". It returns the matching entries and any error.
	ListObjectsForGlob(ctx context.Context, bucket, glob string, pageSize uint32, pageToken string) ([]ObjectStoreEntry, string, error)
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

// ListBucketsFromPathPrefixes returns the list of buckets allowed by PathPrefixes,
// with deterministic sorting and pagination.
func ListBucketsFromPathPrefixes(pathPrefixes []string, pageSize uint32, pageToken string) ([]string, string, error) {
	uniqueBuckets := make(map[string]bool)
	for _, p := range pathPrefixes {
		u, err := url.Parse(p)
		if err != nil {
			return nil, "", fmt.Errorf("invalid path prefix %q: %w", p, err)
		}
		bucket := u.Hostname()
		if bucket == "" {
			return nil, "", fmt.Errorf("can't parse bucket in path prefix %q; proper format is <schema>://<bucket>/path", p)
		}
		uniqueBuckets[bucket] = true
	}

	buckets := make([]string, 0, len(uniqueBuckets))
	for b := range uniqueBuckets {
		buckets = append(buckets, b)
	}
	sort.Strings(buckets)

	validPageSize := pagination.ValidPageSize(pageSize, DefaultPageSize)
	startIndex := 0
	if pageToken != "" {
		if err := pagination.UnmarshalPageToken(pageToken, &startIndex); err != nil {
			return nil, "", fmt.Errorf("invalid page token: %w", err)
		}
	}

	endIndex := startIndex + validPageSize
	if endIndex > len(buckets) {
		endIndex = len(buckets)
	}

	next := ""
	if endIndex < len(buckets) {
		next = pagination.MarshalPageToken(endIndex)
	}
	return buckets[startIndex:endIndex], next, nil
}

type BlobListfn func(ctx context.Context, path, delimiter string, pageSize uint32, pageToken string) ([]ObjectStoreEntry, string, error)

// ListObjects restricts listing to allowed path prefixes. If the requested path
// is within an allowed prefix, a normal blob listing is performed. If the path
// is a parent of allowed prefixes, a synthetic directory listing is returned.
// Otherwise, access is denied.
func ListObjects(ctx context.Context, pathPrefixes []string, blobListfn BlobListfn, bucket, path, delimiter string, pageSize uint32, pageToken string) ([]ObjectStoreEntry, string, error) {
	if delimiter == "" {
		delimiter = "/"
	}

	if path == "." || path == "/" {
		path = ""
	}

	if len(pathPrefixes) == 0 {
		return blobListfn(ctx, path, delimiter, pageSize, pageToken)
	}

	// Extract allowed prefixes for this bucket, reduced to non-nested forms
	allowedPaths, err := buildAllowedPrefixesForBucket(pathPrefixes, bucket)
	if err != nil {
		return nil, "", err
	}

	pathInAllowedPrefix := false // allowed prefix contains requested path → real listing
	matchedChild := []string{}   // allowed prefix is deeper → synthetic listing

	for _, ap := range allowedPaths {
		switch {
		case strings.HasPrefix(path, ap):
			pathInAllowedPrefix = true

		case strings.HasPrefix(ap, path):
			matchedChild = append(matchedChild, ap)
		}
	}

	switch {
	case pathInAllowedPrefix:
		// User is within allowed scope → list real objects
		return blobListfn(ctx, path, delimiter, pageSize, pageToken)

	case len(matchedChild) > 0:
		// User is above allowed scope → list child prefixes instead of objects
		return listObjectsFromPathPrefixes(ctx, matchedChild, blobListfn, path, delimiter, pageSize, pageToken)
	}

	return nil, "", fmt.Errorf("path %q not allowed by path_prefixes", path)
}

// listObjectsFromPathPrefixes returns directory entries for allowed prefixes
// that are deeper than the requested path. If the remaining part of the prefix
// has no delimiter, real blobs under that prefix are listed. Otherwise, only
// the first child segment is returned as a synthetic directory. Results are
// paginated.
func listObjectsFromPathPrefixes(ctx context.Context, matchingPathPrefixes []string, blobListfn BlobListfn, path, delimiter string, pageSize uint32, pageToken string) ([]ObjectStoreEntry, string, error) {
	result := make([]ObjectStoreEntry, 0)
	for _, ap := range matchingPathPrefixes {
		rest := strings.TrimPrefix(ap, path)
		// If 'rest' has no delimiter (e.g. "y="), list real blobs under that prefix.
		// Otherwise, return only the first child segment as a synthetic directory.
		if !strings.Contains(rest, delimiter) {
			// we are list all prefix without pagination it will become too complex if we add pagination here.
			objs, err := pagination.CollectAll(ctx,
				func(ctx context.Context, pz uint32, tk string) ([]ObjectStoreEntry, string, error) {
					return blobListfn(ctx, ap, delimiter, pz, tk)
				},
				DefaultPageSize,
			)
			if err != nil {
				return nil, "", err
			}
			result = append(result, objs...)
		} else {
			result = append(result, ObjectStoreEntry{
				Path:      path + strings.SplitN(rest, delimiter, 2)[0] + delimiter,
				IsDir:     true,
				Size:      0,
				UpdatedOn: time.Time{},
			})
		}
	}
	// Pagination
	validPageSize := pagination.ValidPageSize(pageSize, DefaultPageSize)
	startIndex := 0
	if pageToken != "" {
		if err := pagination.UnmarshalPageToken(pageToken, &startIndex); err != nil {
			return nil, "", fmt.Errorf("invalid page token: %w", err)
		}
	}
	endIndex := startIndex + validPageSize
	if endIndex > len(result) {
		endIndex = len(result)
	}

	next := ""
	if endIndex < len(result) {
		next = pagination.MarshalPageToken(endIndex)
	}

	return result[startIndex:endIndex], next, nil
}

// buildAllowedPrefixesForBucket returns allowed prefixes for the given bucket,
// removing ones nested under others. Errors if none match or format is invalid.
func buildAllowedPrefixesForBucket(pathPrefixes []string, bucket string) ([]string, error) {
	var paths []string
	for _, p := range pathPrefixes {
		u, err := url.Parse(p)
		if err != nil {
			return nil, fmt.Errorf("invalid path prefix %q: %w", p, err)
		}
		p := strings.TrimPrefix(u.Path, "/")
		if u.Hostname() == bucket {
			paths = append(paths, p)
		}
	}

	if len(paths) == 0 {
		return nil, fmt.Errorf("bucket %q not allowed by path_prefixes", bucket)
	}

	// Remove nested paths → only root access points remain
	sort.Strings(paths)
	reduced := make([]string, 0, len(paths))
	for _, p := range paths {
		include := true
		for _, r := range reduced {
			if strings.HasPrefix(p, r) {
				include = false
				break
			}
		}
		if include {
			reduced = append(reduced, p)
		}
	}

	return reduced, nil
}
