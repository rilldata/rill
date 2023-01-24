package blob

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/c2h5oh/datasize"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/connectors"
	"github.com/rilldata/rill/runtime/pkg/container"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"gocloud.dev/blob"
	"golang.org/x/sync/errgroup"
)

// increasing this limit can increase speed ingestion
// but may increase bottleneck at duckdb or network/db IO
// set without any benchamarks
const concurrentBlobDownloadLimit = 8

type Strategy string

const (
	TAIL Strategy = "tail"
	HEAD Strategy = "head"
	NONE Strategy = "none"
)

type ExtractConfig struct {
	Strategy Strategy
	Size     int64
}

type ExtractPolicy struct {
	Row       ExtractConfig
	Partition ExtractConfig
}

type BlobIterator struct {
	bucket           *blob.Bucket
	objectPaths      []*blob.ListObject // object path in cloud storage
	index            int
	rowExtractConfig *ExtractConfig
	lastObjectSize   int64
}

func NewExtractConfigs(input *runtimev1.Source_ExtractPolicy) (*ExtractPolicy, error) {
	config := &ExtractPolicy{
		Row:       ExtractConfig{Strategy: NONE},
		Partition: ExtractConfig{Strategy: NONE},
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
	Extract               *ExtractPolicy
}

func ContainerForParitionStrategy(option ExtractConfig) (container.Container, error) {
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

func withinSize(option ExtractConfig, size int64) bool {
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
func NewIterator(ctx context.Context, bucket *blob.Bucket, config FetchConfigs, globPattern, bucketPath string) (connectors.Iterator, error) {
	validateConfigs(&config)
	iterator := &BlobIterator{
		bucket:           bucket,
		rowExtractConfig: &config.Extract.Row,
	}

	var (
		size, fetched int64
		matchCount    int
		containerFull bool
	)
	c, err := ContainerForParitionStrategy(config.Extract.Partition)
	if err != nil {
		return nil, err
	}

	token := blob.FirstPageToken
	for token != nil && !containerFull {
		objs, nextToken, err := bucket.ListPage(ctx, token, config.GlobPageSize, listOptions(globPattern))
		if err != nil {
			return nil, err
		}

		token = nextToken
		fetched += int64(len(objs))
		for _, obj := range objs {
			if matched, _ := doublestar.Match(globPattern, obj.Key); matched {
				size += obj.Size
				matchCount++
				// container stops consuming once parition strategy limits are crossed
				// withinSize keeps track of whether size of files matched so far are within row size limits
				if !c.Add(obj) && withinSize(config.Extract.Row, size) {
					// don't add more items
					containerFull = true
					break
				}
			}
		}
		if err := validateLimits(size, matchCount, fetched, config); err != nil {
			iterator.Close()
			return nil, err
		}
	}

	items := c.Items()
	if len(items) == 0 {
		return nil, fmt.Errorf("no files found for glob pattern %q", globPattern)
	}

	size = 0
	for _, val := range items {
		obj := val.(*blob.ListObject)
		iterator.objectPaths = append(iterator.objectPaths, obj)
		if config.Extract.Row.Strategy != NONE && size+obj.Size > config.Extract.Row.Size {
			iterator.lastObjectSize = config.Extract.Row.Size - size
			break
		}
		size += obj.Size
	}
	return iterator, nil
}

func (iter *BlobIterator) Close() error {
	return iter.bucket.Close()
}

func (iter *BlobIterator) HasNext() bool {
	return iter.index < len(iter.objectPaths)
}

// NextBatch downloads next n files and copies to local directory
// Callers responsibility to delete files once done
// Thread unsafe
func (iter *BlobIterator) NextBatch(ctx context.Context, n int) ([]string, error) {
	if !iter.HasNext() {
		return nil, io.EOF
	}

	start := iter.index
	end := iter.index + n
	if end > len(iter.objectPaths) {
		end = len(iter.objectPaths)
	}
	iter.index = end

	localPaths := make([]string, n)
	g, grpCtx := errgroup.WithContext(ctx)
	g.SetLimit(concurrentBlobDownloadLimit)
	for i, item := range iter.objectPaths[start:end] {
		obj := item
		index := start + i // with repect to object slices
		g.Go(func() error {
			file, err := fileutil.TempFile(os.TempDir(), obj.Key)
			if err != nil {
				return err
			}

			defer file.Close()
			localPaths[index-start] = file.Name()
			fmt.Println(file.Name())
			if index == len(iter.objectPaths)-1 && iter.lastObjectSize > int64(0) {
				// download partial file
				// todo :: add csv
				// todo :: parquet reader seems to be making too many calls for small files
				// check if for smaller size we can download entire file
				err = Download(grpCtx, iter.bucket, obj, ExtractConfig{Size: iter.lastObjectSize, Strategy: iter.rowExtractConfig.Strategy}, file)
			} else {
				// download full file
				err = downloadObject(grpCtx, iter.bucket, obj.Key, file)
			}
			if err != nil {
				return err
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return localPaths, nil
}

// listOptions for page listing api
func listOptions(globPattern string) *blob.ListOptions {
	listOptions := &blob.ListOptions{BeforeList: func(as func(interface{}) bool) error {
		// Access storage.Query via q here.
		var q *storage.Query
		if as(&q) {
			// we only need name and size, adding only required attributes to reduce data fetched
			_ = q.SetAttrSelection([]string{"Name", "Size"})
		}
		return nil
	}}

	prefix, glob := doublestar.SplitPattern(globPattern)
	if !fileutil.IsGlob(glob) {
		// single file
		listOptions.Prefix = globPattern
	} else if prefix != "." {
		listOptions.Prefix = prefix
	}

	return listOptions
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
