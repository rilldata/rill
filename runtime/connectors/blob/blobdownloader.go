package blob

import (
	"context"
	"fmt"

	"cloud.google.com/go/storage"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"gocloud.dev/blob"
)

type FetchConfigs struct {
	GlobMaxTotalSize      int64
	GlobMaxObjectsMatched int
	GlobMaxObjectsListed  int64
	GlobPageSize          int
}

// downloads file to local paths
// todo :: return blob handler as iterator
func FetchFileNames(ctx context.Context, bucket *blob.Bucket, config FetchConfigs, globPattern, bucketPath string) ([]string, error) {
	validateConfigs(&config)
	prefix, glob := doublestar.SplitPattern(globPattern)

	handler := &BlobHandler{
		prefix: prefix,
		bucket: bucket,
		path:   bucketPath,
	}

	if !fileutil.IsGlob(glob) {
		// glob represent plain object
		handler.FileNames = []string{globPattern}
		err := handler.DownloadAll(ctx)
		if err != nil {
			return nil, err
		}
		return handler.LocalPaths, nil
	}

	listOptions := &blob.ListOptions{BeforeList: func(as func(interface{}) bool) error {
		// Access storage.Query via q here.
		var q *storage.Query
		if as(&q) {
			// we only need name and size, adding only required attributes to reduce data fetched
			_ = q.SetAttrSelection([]string{"Name", "Size"})
		}
		return nil
	}}

	if prefix != "." {
		listOptions.Prefix = prefix
	}

	var size, fetched int64
	var matchCount int
	var fileNames []string

	token := blob.FirstPageToken
	for token != nil {
		objs, nextToken, err := bucket.ListPage(ctx, token, config.GlobPageSize, listOptions)
		if err != nil {
			return nil, err
		}
		token = nextToken

		for _, obj := range objs {
			if matched, _ := doublestar.Match(globPattern, obj.Key); matched {
				size += obj.Size
				matchCount++
				fileNames = append(fileNames, obj.Key)
			}
		}

		fetched += int64(len(objs))

		if err := validateLimits(size, matchCount, fetched, config); err != nil {
			return nil, err
		}
	}

	if len(fileNames) == 0 {
		return nil, fmt.Errorf("no files found for glob pattern %q", globPattern)
	}

	handler.FileNames = fileNames
	if err := handler.DownloadAll(ctx); err != nil {
		return nil, err
	}
	return handler.LocalPaths, nil
}

func validateLimits(size int64, matchCount int, fetched int64, config FetchConfigs) error {
	if size > config.GlobMaxTotalSize {
		return fmt.Errorf("glob pattern exceeds limits: would fetch more than %d bytes", config.GlobMaxTotalSize)
	}
	if matchCount > config.GlobMaxObjectsMatched {
		return fmt.Errorf("glob pattern exceeds limits: matched more than %d files", config.GlobMaxObjectsMatched)
	}
	if fetched > config.GlobMaxObjectsListed {
		return fmt.Errorf("glob pattern exceeds limits: listed more than %d files", config.GlobMaxObjectsListed)
	}
	return nil
}

func validateConfigs(fetchConfigs *FetchConfigs) {
	if fetchConfigs.GlobMaxObjectsMatched == 0 {
		fetchConfigs.GlobMaxObjectsMatched = 1000
	}
	if fetchConfigs.GlobMaxObjectsListed == 0 {
		fetchConfigs.GlobMaxObjectsListed = 1000 * 1000
	}
	if fetchConfigs.GlobMaxTotalSize == 0 {
		// 10 GB
		fetchConfigs.GlobMaxTotalSize = 10 * 1024 * 1024 * 1024
	}
	if fetchConfigs.GlobPageSize == 0 {
		fetchConfigs.GlobPageSize = 1000
	}
}
