package blob

import (
	"context"
	"fmt"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
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

// ListObjectsForGlob lists objects in the bucket that match the given glob pattern.
// The glob pattern should be a valid path *without* scheme or bucket name.
// E.g. to list gs://my-bucket/path/to/files/*, the glob pattern should be "path/to/files/*".
func (b *Bucket) ListObjectsForGlob(ctx context.Context, glob string, pageSize uint32, pageToken string) ([]drivers.ObjectStoreEntry, string, error) {
	validPageSize := pagination.ValidPageSize(pageSize, drivers.DefaultPageSizeForObjects)
	var startAfter string
	driverPageToken := blob.FirstPageToken
	if pageToken != "" {
		if err := pagination.UnmarshalPageToken(pageToken, &driverPageToken, &startAfter); err != nil {
			return nil, "", fmt.Errorf("invalid page token: %w", err)
		}
	}

	// If it's not a glob, we're pulling a single file.
	// TODO: Should we add support for listing out directories without ** at the end?
	if !fileutil.IsGlob(glob) {
		attrs, err := b.bucket.Attributes(ctx, glob)
		if err != nil {
			return nil, "", err
		}

		return []drivers.ObjectStoreEntry{{
			Path:      glob,
			IsDir:     false,
			Size:      attrs.Size,
			UpdatedOn: attrs.ModTime,
		}}, "", nil
	}

	prefix := fileutil.GlobPrefix(glob)

	delimiter := byte('/')
	globLevel := fileutil.PathLevel(glob, delimiter)

	hasDoubleStar := fileutil.IsDoubleStarGlob(glob)

	var entries []drivers.ObjectStoreEntry
	var currentDir *drivers.ObjectStoreEntry // Track current directory being accumulated

	// Fetch pages until we have enough matching results (accounting for glob filtering)
	for len(entries) < validPageSize && driverPageToken != nil {
		retval, nextDriverPageToken, err := b.bucket.ListPage(ctx, driverPageToken, validPageSize, &blob.ListOptions{
			Prefix: prefix,
			BeforeList: func(as func(interface{}) bool) error {
				// Handle GCS
				var q *storage.Query
				if as(&q) {
					// Only fetch the fields we need.
					_ = q.SetAttrSelection([]string{"Name", "Size", "Updated"})
					if startAfter != "" {
						q.StartOffset = startAfter
					}
				}
				// Handle S3
				var s3Input *s3.ListObjectsV2Input
				if as(&s3Input) {
					if startAfter != "" {
						s3Input.StartAfter = aws.String(startAfter)
					}
				}
				// Handle Azure Blob Storage
				var azOpts *container.ListBlobsHierarchyOptions
				if as(&azOpts) {
					if startAfter != "" {
						azOpts.StartFrom = &startAfter
					}
				}
				return nil
			},
		})
		if err != nil {
			return nil, "", err
		}

		// Filter by glob pattern and skip startAfter entries
		lastProcessedIdx := -1
		for i, obj := range retval {
			// Skip entries until we're past startAfter
			if startAfter != "" {
				// error out here because we already have StartOffset/StartAfter/StartFrom pass in api
				if obj.Key < startAfter {
					return nil, "", fmt.Errorf("blob: entry with key < startAfter (%q)", startAfter)
				}
				if obj.Key == startAfter {
				continue
			}

			fileLevel := fileutil.PathLevel(obj.Key, delimiter)

			// Match directory if the glob is not double-star ("**")
			// and the file level is greater than the glob level.
			if !hasDoubleStar && fileLevel > globLevel {
				dirPath := fileutil.PrefixUntilLevel(obj.Key, globLevel, delimiter)

				// If this is a different directory, finalize the previous one
				if currentDir != nil && currentDir.Path != dirPath {
					entries = append(entries, *currentDir)
					currentDir = nil
					if len(entries) >= validPageSize {
						break
					}
				}

				globForDir := fileutil.EnsureTrailingDelim(glob, delimiter)
				lastProcessedIdx = i
				// Match against glob pattern
				ok, err := doublestar.Match(globForDir, dirPath)
				if err != nil {
					return nil, "", err
				}
				if !ok {
					continue
				}

				// Initialize current directory
				if currentDir == nil {
					currentDir = &drivers.ObjectStoreEntry{
						Path:      dirPath,
						IsDir:     true,
						Size:      0,
						UpdatedOn: obj.ModTime,
					}
				}

				// Accumulate size and update timestamp
				currentDir.Size += obj.Size
				if obj.ModTime.After(currentDir.UpdatedOn) {
					currentDir.UpdatedOn = obj.ModTime
				}
				continue
			}

			// finalizing the current dir
			if currentDir != nil {
				entries = append(entries, *currentDir)
				currentDir = nil
			}

			if len(entries) >= validPageSize {
				break
			}

			lastProcessedIdx = i
			ok, err := doublestar.Match(glob, obj.Key)
			if err != nil {
				return nil, "", err
			}
			if !ok {
				continue
			}

			entries = append(entries, drivers.ObjectStoreEntry{
				Path:      obj.Key,
				IsDir:     obj.IsDir,
				Size:      obj.Size,
				UpdatedOn: obj.ModTime,
			})
			if len(entries) == validPageSize {
				break
			}
		}

		if len(entries) == validPageSize {
			if lastProcessedIdx == len(retval)-1 {
				driverPageToken = nextDriverPageToken
				startAfter = ""
			} else if lastProcessedIdx != -1 {
				startAfter = retval[lastProcessedIdx].Key
			}
			break
		}

		driverPageToken = nextDriverPageToken
		startAfter = ""
	}

	if driverPageToken == nil {
		// finalizing the current dir, if no object left to process
		if currentDir != nil {
			entries = append(entries, *currentDir)
			currentDir = nil
		}
		return entries, "", nil
	}
	return entries, pagination.MarshalPageToken(driverPageToken, startAfter), nil
}

func (b *Bucket) ListObjects(ctx context.Context, path, delimiter string, pageSize uint32, pageToken string) ([]drivers.ObjectStoreEntry, string, error) {
	validPageSize := pagination.ValidPageSize(pageSize, drivers.DefaultPageSizeForObjects)
	driverPageToken := blob.FirstPageToken
	if pageToken != "" {
		if err := pagination.UnmarshalPageToken(pageToken, &driverPageToken); err != nil {
			return nil, "", fmt.Errorf("invalid page token: %w", err)
		}
	}

	retval, nextDriverPageToken, err := b.bucket.ListPage(ctx, driverPageToken, validPageSize, &blob.ListOptions{
		Prefix:    path,
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
		return nil, "", err
	}
	entries := make([]drivers.ObjectStoreEntry, 0, len(retval))
	for _, obj := range retval {
		entries = append(entries, drivers.ObjectStoreEntry{
			Path:      obj.Key,
			IsDir:     strings.HasSuffix(obj.Key, "/"), // Workaround for some object stores not marking IsDir correctly
			Size:      obj.Size,
			UpdatedOn: obj.ModTime,
		})
	}
	nextToken := ""
	if nextDriverPageToken != nil {
		nextToken = pagination.MarshalPageToken(nextDriverPageToken)
	}
	return entries, nextToken, nil
}
