package blob

import (
	"context"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/c2h5oh/datasize"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/container"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"gocloud.dev/blob"
	"golang.org/x/sync/errgroup"
)

type Strategy string

const (
	TAIL Strategy = "tail"
	HEAD Strategy = "head"
	NONE Strategy = "none"
)

type ExtractOptions struct {
	Strategy Strategy
	Size     int64
}

type ExtractConfigs struct {
	Row       ExtractOptions
	Partition ExtractOptions
}

func NewExtractConfigs(input *runtimev1.Source_ExtractPolicy) (*ExtractConfigs, error) {
	config := &ExtractConfigs{
		Row:       ExtractOptions{Strategy: NONE, Size: math.MaxInt64},
		Partition: ExtractOptions{Strategy: NONE},
	}

	if input == nil {
		return config, nil
	}

	// parse partition
	if input.Partition != nil {
		// parse strategy
		strategy, err := parseStrategy(input.Partition.Strategy)
		if err != nil {
			return nil, err
		}

		config.Partition.Strategy = strategy

		// parse size
		size, err := strconv.ParseInt(input.Partition.Size, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid size, parse failed with error %w", err)
		}
		if size <= 0 {
			return nil, fmt.Errorf("invalid size %q", size)
		}

		config.Partition.Size = size
	}

	// parse rows
	if input.Row != nil {
		// parse strategy
		strategy, err := parseStrategy(input.Row.Strategy)
		if err != nil {
			return nil, err
		}

		config.Row.Strategy = strategy

		// parse size
		// todo :: add support for number of rows
		size, err := getBytes(input.Row.Size)
		if err != nil {
			return nil, fmt.Errorf("invalid size, parse failed with error %w", err)
		}
		if size <= 0 {
			return nil, fmt.Errorf("invalid size %q", size)
		}

		config.Row.Size = size
	}

	return config, nil
}

func parseStrategy(s string) (Strategy, error) {
	switch strings.ToLower(s) {
	case "tail":
		return TAIL, nil
	case "head":
		return HEAD, nil
	default:
		return "", fmt.Errorf("invalid extract strategy %q", s)
	}
}

func getBytes(size string) (int64, error) {
	var s datasize.ByteSize
	if err := s.UnmarshalText([]byte(size)); err != nil {
		return 0, err
	}

	return int64(s.Bytes()), nil
}

type FetchConfigs struct {
	GlobMaxTotalSize      int64
	GlobMaxObjectsMatched int
	GlobMaxObjectsListed  int64
	GlobPageSize          int
	Extract               *ExtractConfigs
}

func ContainerForParitionStrategy(option ExtractOptions) (container.Container, error) {
	switch option.Strategy {
	case TAIL:
		return container.NewTailContainer(int(option.Size))
	case HEAD:
		return container.NewBoundedContainer(int(option.Size))
	default:
		// No option selected
		return container.NewUnboundedContainer()
	}
}

func withinSize(option ExtractOptions, size int64) bool {
	switch option.Strategy {
	case TAIL:
		return true
	case HEAD:
		return size < option.Size
	default:
		// No option selected
		return true
	}
}

// downloads file to local paths
// todo :: return blob handler as iterator
func FetchFileNames(ctx context.Context, bucket *blob.Bucket, config FetchConfigs, globPattern, bucketPath string) ([]string, error) {
	validateConfigs(&config)
	c, err := ContainerForParitionStrategy(config.Extract.Partition)
	if err != nil {
		return nil, err
	}

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

	var (
		size, fetched int64
		matchCount    int
		containerFull bool
	)

	token := blob.FirstPageToken
	for token != nil && !containerFull {
		objs, nextToken, err := bucket.ListPage(ctx, token, config.GlobPageSize, listOptions)
		if err != nil {
			return nil, err
		}

		token = nextToken
		fetched += int64(len(objs))
		for _, obj := range objs {
			if matched, _ := doublestar.Match(globPattern, obj.Key); matched {
				size += obj.Size
				matchCount++
				if !c.Add(obj) && withinSize(config.Extract.Row, size) {
					// don't add more items
					containerFull = true
					break
				}
			}
		}
		if err := validateLimits(size, matchCount, fetched, config); err != nil {
			return nil, err
		}
	}

	items := c.Items()
	if len(items) == 0 {
		return nil, fmt.Errorf("no files found for glob pattern %q", globPattern)
	}

	size = 0
	applicableItems := make([]*blob.ListObject, 0)
	for _, val := range items {
		obj := val.(*blob.ListObject)
		applicableItems = append(applicableItems, obj)
		size += obj.Size
		if size > config.Extract.Row.Size {
			break
		}
	}

	return DownloadAll(ctx, bucket, &config.Extract.Row, applicableItems)
}

// object path is relative to bucket
func DownloadAll(ctx context.Context, bucket *blob.Bucket, options *ExtractOptions, items []*blob.ListObject) ([]string, error) {
	totalSize := int64(0)
	for _, item := range items {
		totalSize += item.Size
	}

	tempDir, err := os.MkdirTemp(os.TempDir(), "blob*ingestion")
	if err != nil {
		return nil, err
	}
	localPaths := make([]string, len(items))

	g, grpCtx := errgroup.WithContext(ctx)
	g.SetLimit(concurrentBlobDownloadLimit)
	for i, item := range items {
		obj := item
		index := i
		g.Go(func() error {
			file, err := fileutil.TempFile(tempDir, obj.Key)
			if err != nil {
				return err
			}

			localPaths[index] = file.Name()
			fmt.Println(file.Name())
			defer file.Close()
			if index == len(items)-1 && totalSize > options.Size {
				sizeReqd := options.Size - (totalSize - obj.Size)
				// download partial file
				err = Download(grpCtx, bucket, obj, ExtractOptions{Size: sizeReqd, Strategy: options.Strategy}, file)
			} else {
				// download full file
				err = downloadObject(grpCtx, bucket, obj.Key, file)
			}
			if err != nil {
				return err
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		// one of the download failed
		// remove the files
		fileutil.ForceRemoveFiles(localPaths)
		// remove the empty temp directory
		os.Remove(tempDir)
		return nil, err
	}

	return localPaths, nil
}

func downloadObject(ctx context.Context, bucket *blob.Bucket, objpath string, file *os.File) error {
	rc, err := bucket.NewReader(ctx, objpath, nil)
	if err != nil {
		return fmt.Errorf("Object(%q).NewReader: %w", objpath, err)
	}
	defer rc.Close()

	_, err = io.Copy(file, rc)
	return err
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
		fetchConfigs.GlobMaxObjectsMatched = 100
	}
	if fetchConfigs.GlobMaxObjectsListed == 0 {
		fetchConfigs.GlobMaxObjectsListed = 10 * 1000
	}
	if fetchConfigs.GlobMaxTotalSize == 0 {
		// 10 GB
		fetchConfigs.GlobMaxTotalSize = 10 * 1024 * 1024 * 1024
	}
	if fetchConfigs.GlobPageSize == 0 {
		fetchConfigs.GlobPageSize = 1000
	}
}
