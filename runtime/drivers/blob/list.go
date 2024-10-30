package blob

import (
	"context"
	"errors"
	"io"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
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

// ListObjects lists objects in the bucket that match the given glob pattern.
// The glob pattern should be a valid path *without* scheme or bucket name.
// E.g. to list gs://my-bucket/path/to/files/*, the glob pattern should be "path/to/files/*".
func (b *Bucket) ListObjects(ctx context.Context, glob string) ([]drivers.ObjectStoreEntry, error) {
	// If it's not a glob, we're pulling a single file.
	// TODO: Should we add support for listing out directories without ** at the end?
	if !fileutil.IsGlob(glob) {
		attrs, err := b.bucket.Attributes(ctx, glob)
		if err != nil {
			return nil, err
		}

		return []drivers.ObjectStoreEntry{{
			Path:      glob,
			IsDir:     false,
			UpdatedOn: attrs.ModTime,
		}}, nil
	}

	// Extract the prefix (if any) that we can push down to the storage provider.
	prefix, _ := doublestar.SplitPattern(glob)
	if prefix == "." {
		prefix = ""
	}

	// Build iterator
	it := b.bucket.List(&blob.ListOptions{
		Prefix: prefix,
		BeforeList: func(as func(interface{}) bool) error {
			var q *storage.Query
			if as(&q) {
				// Only fetch the fields we need.
				_ = q.SetAttrSelection([]string{"Name", "Size", "Created", "Updated"})
			}
			return nil
		},
	})

	// Build output
	var entries []drivers.ObjectStoreEntry
	for {
		obj, err := it.Next(ctx)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}

		// Workaround for some object stores not marking IsDir correctly.
		if strings.HasSuffix(obj.Key, "/") {
			obj.IsDir = true
		}

		ok, err := doublestar.Match(glob, obj.Key)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}

		entries = append(entries, drivers.ObjectStoreEntry{
			Path:      obj.Key,
			IsDir:     obj.IsDir,
			UpdatedOn: obj.ModTime,
		})
	}

	return entries, nil
}
