package blob

import (
	"container/list"
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
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"gocloud.dev/blob"
	"golang.org/x/sync/errgroup"
)

// increasing this limit can increase speed ingestion
// but may increase bottleneck at duckdb or network/db IO
// set without any benchamarks
const concurrentBlobDownloadLimit = 8

var partialDownloadExtensions = map[string]bool{".parquet": true, ".csv": true, ".tsv": true, ".txt": true, ".parquet.gz": true}

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
	Row  ExtractConfig
	File ExtractConfig
}

type BlobIterator struct {
	bucket  *blob.Bucket
	objects []*blobObject
	index   int
}

func NewExtractConfigs(input *runtimev1.Source_ExtractPolicy) (*ExtractPolicy, error) {
	config := &ExtractPolicy{
		Row:  ExtractConfig{Strategy: NONE},
		File: ExtractConfig{Strategy: NONE},
	}

	if input == nil {
		return config, nil
	}

	// parse file
	if input.File != nil {
		// parse strategy
		strategy, err := parseStrategy(input.File.Strategy)
		if err != nil {
			return nil, err
		}

		config.File.Strategy = strategy

		// parse size
		size, err := strconv.ParseInt(input.File.Size, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid size, parse failed with error %w", err)
		}
		if size <= 0 {
			return nil, fmt.Errorf("invalid size %q", size)
		}

		config.File.Size = size
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

// container encapsulates all extract policy
type container struct {
	policy *ExtractPolicy
	items  *list.List
	size   int64
}

func (s *container) Add(item *blob.ListObject) bool {
	if s.IsFull() {
		return false
	}

	s.size += item.Size
	switch s.policy.File.Strategy {
	case TAIL:
		// keep latest item at front
		s.items.PushFront(item)
		if s.items.Len() > int(s.policy.File.Size) {
			// remove oldest item
			s.items.Remove(s.items.Back())
		}
	case HEAD:
		s.items.PushBack(item)
	default:
		s.items.PushBack(item)
	}
	return true
}

func (s *container) hasCapacity() bool {
	switch s.policy.File.Strategy {
	case TAIL:
		return true
	case HEAD:
		return s.items.Len() < int(s.policy.File.Size)
	default:
		return true
	}
}

func (s *container) IsFull() bool {
	if s.policy.File.Strategy != NONE {
		// file strategy present, row policy is per file
		// only need to limit number of files by file strategy
		return !s.hasCapacity()
	}

	if s.policy.Row.Strategy != NONE {
		// file policy absent
		// row policy present, limit size across all files
		return s.size > s.policy.Row.Size
	}

	// no policy
	return false
}

func (s *container) Items() []*blobObject {
	result := make([]*blobObject, s.items.Len())

	var cumSize int64
	for front, i := s.items.Front(), 0; front != nil; front, i = s.items.Front(), i+1 {
		item := s.items.Remove(front)
		obj := &blobObject{obj: item.(*blob.ListObject), full: true, stratety: NONE}
		if s.policy.File.Strategy != NONE {
			// file strategy present
			if s.policy.Row.Strategy != NONE {
				// row policy is per file
				obj.full = false
				obj.size = s.policy.Row.Size
				obj.stratety = s.policy.Row.Strategy
			}
		} else {
			if s.policy.Row.Strategy != NONE {
				// file strategy absent row policy is global
				if i == len(result)-1 {
					obj.full = false
					obj.size = s.policy.Row.Size - cumSize
					obj.stratety = s.policy.Row.Strategy
				}
				cumSize += obj.size
			}
		}
		result[i] = obj
	}
	return result
}

type blobObject struct {
	obj      *blob.ListObject
	full     bool
	size     int64
	stratety Strategy
}

func newContainer(policy *ExtractPolicy) *container {
	return &container{policy: policy, items: list.New()}
}

// downloads file to local paths
func NewIterator(ctx context.Context, bucket *blob.Bucket, config FetchConfigs, globPattern, bucketPath string) (connectors.Iterator, error) {
	validateConfigs(&config)

	var (
		size, fetched int64
		matchCount    int
	)
	c := newContainer(config.Extract)

	token := blob.FirstPageToken
	for token != nil && !c.IsFull() {
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
				c.Add(obj)
			}
		}
		if err := validateLimits(size, matchCount, fetched, config); err != nil {
			bucket.Close()
			return nil, err
		}
	}

	items := c.Items()
	if len(items) == 0 {
		return nil, fmt.Errorf("no files found for glob pattern %q", globPattern)
	}

	return &BlobIterator{bucket: bucket, objects: items}, nil
}

func (iter *BlobIterator) Close() error {
	return iter.bucket.Close()
}

func (iter *BlobIterator) HasNext() bool {
	return iter.index < len(iter.objects)
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
	if end > len(iter.objects) {
		end = len(iter.objects)
	}
	iter.index = end

	localPaths := make([]string, n)
	g, grpCtx := errgroup.WithContext(ctx)
	g.SetLimit(concurrentBlobDownloadLimit)
	for i, item := range iter.objects[start:end] {
		obj := item
		index := start + i // with repect to object slice
		g.Go(func() error {
			file, err := fileutil.TempFile(os.TempDir(), obj.obj.Key)
			if err != nil {
				return err
			}

			defer file.Close()
			localPaths[index-start] = file.Name()
			fmt.Println(file.Name())
			if obj.full || !isPartialDownloadSupported(obj.obj.Key) {
				// download full file
				err = downloadObject(grpCtx, iter.bucket, obj.obj.Key, file)
			} else {
				// download partial file
				// check if, for smaller size we can download entire file
				err = Download(grpCtx, iter.bucket, obj.obj, ExtractConfig{Size: obj.size, Strategy: obj.stratety}, file)
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

func isPartialDownloadSupported(name string) bool {
	ext := fileutil.FullExt(name)
	// zipped csv, tsv files are not supported
	return partialDownloadExtensions[ext]
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
