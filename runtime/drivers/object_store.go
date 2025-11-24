package drivers

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strconv"
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

func ListObjects(ctx context.Context, pathPrefixes []string, blobListfn BlobListfn, path, delimiter string, pageSize uint32, pageToken string, bucket string) ([]ObjectStoreEntry, string, error) {
	if delimiter == "" {
		delimiter = "/"
	}
	if len(pathPrefixes) == 0 {
		return blobListfn(ctx, path, delimiter, pageSize, pageToken)
	}

	// Build allowed map
	allowed, err := buildAllowedPrefixMap(pathPrefixes)
	if err != nil {
		return nil, "", err
	}

	// Find prefixes allowed for this bucket
	allowedPaths, ok := allowed[bucket]
	if !ok {
		return nil, "", fmt.Errorf("bucket %q not allowed by path_prefixes", bucket)
	}

	// Classify prefix relationships
	matchedParent := []string{} // allowed prefix is parent -> real listing
	matchedChild := []string{}  // allowed prefix is child -> synthetic listing

	for _, ap := range allowedPaths {
		switch {
		case strings.HasPrefix(path, ap):
			// Path is deeper than or equal to allowed → directly allowed
			matchedParent = append(matchedParent, ap)

		case strings.HasPrefix(ap, path):
			// Path is parent of multiple allowed → synthetic listing needed
			matchedChild = append(matchedChild, ap)
		}
	}

	// Determine access behavior
	switch {
	case len(matchedParent) > 0:
		// List directly at path scope
		return blobListfn(ctx, path, delimiter, pageSize, pageToken)

	case len(matchedChild) > 0:
		// Path is parent → synthetic "directories"
		return synthesizeDirectoryListing(path, matchedChild, delimiter, pageSize, pageToken)
	}
	return nil, "", fmt.Errorf("path %q not allowed by path_prefixes", path)
}

// synthesizeDirectoryListing groups deeper allowed prefixes into direct child folders
// under the requested path, applying pagination.
func synthesizeDirectoryListing(path string, matching []string, delimiter string, pageSize uint32, pageToken string) ([]ObjectStoreEntry, string, error) {
	childrenSet := make(map[string]struct{})
	for _, ap := range matching {
		rest := strings.TrimPrefix(ap, path)        // remove requested prefix base
		parts := strings.SplitN(rest, delimiter, 2) // only direct children
		if parts[0] != "" {
			childrenSet[parts[0]] = struct{}{}
		}
	}

	// Convert to sorted list
	children := make([]string, 0, len(childrenSet))
	for c := range childrenSet {
		children = append(children, c)
	}
	sort.Strings(children)

	// Pagination
	start := 0
	if pageToken != "" {
		var s int
		if err := pagination.UnmarshalPageToken(pageToken, &s); err == nil {
			start = s
		}
	}
	end := start + int(pagination.ValidPageSize(pageSize, DefaultPageSize))
	if end > len(children) {
		end = len(children)
	}

	entries := make([]ObjectStoreEntry, 0, end-start)
	for _, c := range children[start:end] {
		entries = append(entries, ObjectStoreEntry{
			Path:      path + c + delimiter,
			IsDir:     true,
			Size:      0,
			UpdatedOn: time.Time{},
		})
	}

	var next string
	if end < len(children) {
		next = pagination.MarshalPageToken(strconv.Itoa(end))
	}

	return entries, next, nil
}

// BuildAllowedPrefixMap constructs a reduced set of allowed path prefixes grouped by bucket
//
// Input example:
//
//	  []string{
//	      "s3://my-bucket/foo/",
//	      "s3://my-bucket/bar/",
//		  "s3://my-bucket/bar/baz",
//	      "s3://other-bucket/alpha",
//	  }
//
// Output example:
//
//	{
//	    "my-bucket":   []string{"s3://my-bucket/foo/", "s3://my-bucket/bar/"},
//	    "other-bucket": []string{"s3://other-bucket/alpha"},
//	}
//
// Returns an error if any prefix is malformed.
func buildAllowedPrefixMap(pathPrefixes []string) (map[string][]string, error) {
	result := make(map[string][]string)
	for _, p := range pathPrefixes {
		// Parse prefix and extract bucket + path
		u, err := url.Parse(p)
		if err != nil {
			return nil, fmt.Errorf("invalid path prefix %q: %w", p, err)
		}
		bucket := u.Hostname()
		if bucket == "" {
			return nil, fmt.Errorf("can't parse bucket in path prefix %q; proper format is <schema>://<bucket>/path", p)
		}
		// Group paths by bucket
		result[bucket] = append(result[bucket], u.Path)
	}
	// Reduce prefixes: remove more specific paths under the same root
	for bucket, prs := range result {
		sort.Strings(prs)
		reduced := make([]string, 0, len(prs))
		for _, p := range prs {
			include := true
			// If prefix starts with an already included shorter prefix, skip it
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
		result[bucket] = reduced
	}

	return result, nil
}
