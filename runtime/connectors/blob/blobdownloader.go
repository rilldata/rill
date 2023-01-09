package blob

import (
	"context"
	"fmt"
	"math"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/rilldata/rill/runtime/pkg/fileutil"

	"gocloud.dev/blob"
)

type FetchConfigs struct {
	MaxSize       int64
	MaxDownload   int
	MaxIterations int64
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

func FetchBlobHandler(ctx context.Context, bucket *blob.Bucket, config FetchConfigs, globPattern, bucketPath string) (*BlobHandler, error) {
	prefix, glob := doublestar.SplitPattern(globPattern)
	result := &BlobHandler{prefix: prefix, bucket: bucket, BlobType: blobType(bucketPath), path: bucketPath}
	if !fileutil.HasMeta(glob) {
		// glob represent plain object
		result.FileNames = []string{globPattern}
		return result, nil
	}
	before := func(as func(interface{}) bool) error {
		// Access storage.Query via q here.
		var q *storage.Query
		if as(&q) {
			// we only need name and size, adding only required attributes to reduce data fetched
			_ = q.SetAttrSelection([]string{"Name", "Size"})
		}
		return nil
	}

	listOptions := blob.ListOptions{BeforeList: before}
	if prefix != "." {
		listOptions.Prefix = prefix
	}

	var size int64
	matchCount := 0
	fileNames := make([]string, 0)
	// list max matched files or 100 in one API listing
	pageSize := int(math.Max(100, float64(config.MaxDownload)))
	fetched := int64(0)
	var returnErr error
	validateConfigs(&config)
	for token := blob.FirstPageToken; token != nil; {
		iter, nextToken, err := bucket.ListPage(ctx, token, pageSize, &listOptions)
		if err != nil {
			returnErr = err
			break
		}
		token = nextToken
		for _, obj := range iter {
			if match(glob, obj.Key) {
				size += obj.Size
				matchCount++
				fileNames = append(fileNames, obj.Key)
			}
		}

		if err := validateLimits(size, matchCount, fetched, config); err != nil {
			returnErr = err
			break
		}
	}
	if returnErr != nil {
		bucket.Close()
		return nil, returnErr
	}
	result.FileNames = fileNames
	return result, nil
}

func validateLimits(size int64, matchCount int, fetched int64, config FetchConfigs) error {
	if size > config.MaxSize {
		return fmt.Errorf("glob pattern exceeds limits: size fetched %v, max size %v", size, config.MaxSize)
	}
	if matchCount > config.MaxDownload {
		return fmt.Errorf("glob pattern exceeds limits: files matched %v, max matches allowed %v", size, config.MaxDownload)
	}
	if fetched > config.MaxIterations {
		return fmt.Errorf("glob pattern exceeds limits: files listed %v, max file listing allowed %v", size, config.MaxIterations)
	}
	return nil
}

func validateConfigs(fetchConfigs *FetchConfigs) {
	if fetchConfigs.MaxDownload == 0 {
		fetchConfigs.MaxDownload = 100
	}
	if fetchConfigs.MaxIterations == int64(0) {
		fetchConfigs.MaxIterations = 10 * 1000
	}
	if fetchConfigs.MaxSize == int64(0) {
		// 10 GB
		fetchConfigs.MaxDownload = 10 * 1024 * 1024 * 1024
	}
}

func match(glob, fileName string) bool {
	matched, _ := doublestar.Match(glob, fileName)
	return matched
}
