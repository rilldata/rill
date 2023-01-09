package blob

import (
	"context"
	"fmt"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/rilldata/rill/runtime/pkg/fileutil"

	"gocloud.dev/blob"
)

type FetchConfigs struct {
	MaxTotalSize       int64
	MaxDownloadObjetcs int
	MaxObjectsListed   int64
	PageSize           int
}

func blobType(path string) BlobType {
	if strings.Contains(path, "file") {
		return File
	} else if strings.Contains(path, "gs") {
		return GCS
	} else if strings.Contains(path, "s3") {
		return S3
	}
	return File
}

// downloads file to local paths
// todo :: return blob handler as iterator
func FetchFileNames(ctx context.Context, bucket *blob.Bucket, config FetchConfigs, globPattern, bucketPath string) ([]string, error) {
	defer bucket.Close()
	validateConfigs(&config)
	prefix, glob := doublestar.SplitPattern(globPattern)
	handler := &BlobHandler{
		prefix:   prefix,
		bucket:   bucket,
		BlobType: blobType(bucketPath),
		path:     bucketPath,
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
	fileNames := make([]string, 0)
	// list max matched files or 100 in one API listing

	token := blob.FirstPageToken
	for token != nil {
		objs, nextToken, err := bucket.ListPage(ctx, token, config.PageSize, listOptions)
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
		return nil, fmt.Errorf("no filenames matching glob pattern")
	}
	handler.FileNames = fileNames
	if err := handler.DownloadAll(ctx); err != nil {
		return nil, err
	}
	return handler.LocalPaths, nil
}

func validateLimits(size int64, matchCount int, fetched int64, config FetchConfigs) error {
	if size > config.MaxTotalSize {
		return fmt.Errorf("glob pattern exceeds limits: size fetched %v, max size %v", size, config.MaxTotalSize)
	}
	if matchCount > config.MaxDownloadObjetcs {
		return fmt.Errorf("glob pattern exceeds limits: files matched %v, max matches allowed %v", size, config.MaxDownloadObjetcs)
	}
	if fetched > config.MaxObjectsListed {
		return fmt.Errorf("glob pattern exceeds limits: files listed %v, max file listing allowed %v", size, config.MaxObjectsListed)
	}
	return nil
}

func validateConfigs(fetchConfigs *FetchConfigs) {
	if fetchConfigs.MaxDownloadObjetcs == 0 {
		fetchConfigs.MaxDownloadObjetcs = 100
	}
	if fetchConfigs.MaxObjectsListed == 0 {
		fetchConfigs.MaxObjectsListed = 10 * 1000
	}
	if fetchConfigs.MaxTotalSize == 0 {
		// 10 GB
		fetchConfigs.MaxTotalSize = 10 * 1024 * 1024 * 1024
	}
	if fetchConfigs.PageSize == 0 {
		fetchConfigs.PageSize = 1000
	}
}
