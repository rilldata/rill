package blob

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"github.com/rilldata/rill/runtime/pkg/pagination"
	"go.uber.org/zap"
	"gocloud.dev/blob"
)

// Bucket wraps a blob.Bucket with functionality for implementing the drivers.ObjectStore interface.
// NOTE: It currently only supports listing objects, but eventually we should refactor NewIterator to a member function of this struct.
type Bucket struct {
	bucket *blob.Bucket
	logger *zap.Logger
}

// NewBucket wraps a *blob.Bucket.
// It takes ownership of the bucket and will close it when Close is called.
func NewBucket(bucket *blob.Bucket, logger *zap.Logger) (*Bucket, error) {
	return &Bucket{
		bucket: bucket,
		logger: logger,
	}, nil
}

// Close the underlying bucket.
func (b *Bucket) Close() error {
	return b.bucket.Close()
}

// Underlying returns the underlying *blob.Bucket.
func (b *Bucket) Underlying() *blob.Bucket {
	return b.bucket
}

// ListObjects lists objects in the bucket that match the given glob pattern.
// The glob pattern should be a valid path *without* scheme or bucket name.
// E.g. to list gs://my-bucket/path/to/files/*, the glob pattern should be "path/to/files/*".
func (b *Bucket) ListObjects(ctx context.Context, glob, delimiter string, pageSize int, pageToken string) ([]drivers.ObjectStoreEntry, string, error) {
	validPageSize := pagination.ValidPageSize(uint32(pageSize), drivers.DefaultPageSize)
	driverPageToken := blob.FirstPageToken
	var offset string
	if pageToken != "" {
		if err := pagination.UnmarshalPageToken(pageToken, &driverPageToken, &offset); err != nil {
			return nil, "", fmt.Errorf("invalid page token: %w", err)
		}
	}

	isGlob := fileutil.IsGlob(glob)
	prefix := glob
	// Extract the prefix (if any) that we can push down to the storage provider.
	if isGlob {
		prefix, _ = doublestar.SplitPattern(glob)
		if prefix == "." {
			prefix = ""
		}
	}

	entries := make([]drivers.ObjectStoreEntry, 0, pageSize)
	for len(entries) < pageSize && driverPageToken != nil {
		retval, nextDriverPageToken, err := b.bucket.ListPage(ctx, driverPageToken, validPageSize, &blob.ListOptions{
			Prefix:    prefix,
			Delimiter: delimiter,
			BeforeList: func(as func(interface{}) bool) error {
				// For GCS
				var q *storage.Query
				if as(&q) {
					// Only fetch the fields we need.
					_ = q.SetAttrSelection([]string{"Name", "Size", "Created", "Updated"})
				}
				return nil
			},
		})
		if err != nil {
			if errors.Is(err, io.EOF) {
				break // no more backend pages
			}
			return nil, "", err
		}

		for _, obj := range retval {
			// start after offset
			if offset != "" && obj.Key <= offset {
				continue
			}
			offset = obj.Key

			// if glob path then add to result entries only if matches the glob pattern
			if isGlob {
				ok, err := doublestar.Match(glob, obj.Key)
				if err != nil {
					return nil, "", err
				}
				if !ok {
					continue
				}
			}

			entries = append(entries, drivers.ObjectStoreEntry{
				Path:      obj.Key,
				IsDir:     strings.HasSuffix(obj.Key, "/"), // Workaround for some object stores not marking IsDir correctly
				Size:      obj.Size,
				UpdatedOn: obj.ModTime,
			})

			// Pagination cutoff
			if len(entries) == validPageSize {
				break
			}
		}
		driverPageToken = nextDriverPageToken
	}
	nextToken := ""
	if driverPageToken != nil {
		nextToken = pagination.MarshalPageToken(driverPageToken, offset)
	}
	return entries, nextToken, nil
}
