package blob

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"cloud.google.com/go/storage"
	"github.com/bmatcuk/doublestar/v4"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/connectors"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"gocloud.dev/blob"
	"golang.org/x/sync/errgroup"
)

// increasing this limit can increase speed ingestion
// but may increase bottleneck at duckdb or network/db IO
// set without any benchamarks
const _concurrentBlobDownloadLimit = 8

// map of supoprted extensions for partial downloads vs readers
var _partialDownloadReaders = map[string]string{".parquet": "parquet", ".csv": "csv", ".tsv": "csv", ".txt": "csv", ".parquet.gz": "parquet"}

// implements connector.FileIterator
type blobIterator struct {
	ctx        context.Context
	bucket     *blob.Bucket
	objects    []*objectWithPlan
	index      int
	localFiles []string
	// all localfiles are created in this dir
	tempDir string
}

// NewIterator returns new instance of blobIterator
// the iterator keeps list of blob objects eagerly planned as per user's glob pattern and extract policies
// clients should call close once done to release all resources like closing the bucket
// the iterator takes responsibility of closing the bucket
func NewIterator(ctx context.Context, bucket *blob.Bucket, opts Options, globPattern string) (connectors.FileIterator, error) {
	opts.validate()

	items, err := plan(ctx, bucket, opts, globPattern)
	if err != nil {
		bucket.Close()
		return nil, err
	}

	tempDir, err := os.MkdirTemp(os.TempDir(), "blob*ingestion")
	if err != nil {
		return nil, err
	}

	return &blobIterator{
		ctx:     ctx,
		bucket:  bucket,
		objects: items,
		tempDir: tempDir,
	}, nil
}

func (iter *blobIterator) Close() error {
	fileutil.ForceRemoveFiles(iter.localFiles)
	// remove temp dir since recursive paths created have to be removed as well
	os.RemoveAll(iter.tempDir)
	return iter.bucket.Close()
}

func (iter *blobIterator) HasNext() bool {
	return iter.index < len(iter.objects)
}

// NextBatch downloads next n files and copies to local directory
// Callers responsibility to delete files once done
// Thread unsafe
func (iter *blobIterator) NextBatch(n int) ([]string, error) {
	if !iter.HasNext() {
		return nil, io.EOF
	}

	// delete files created in last iteration
	fileutil.ForceRemoveFiles(iter.localFiles)
	start := iter.index
	end := iter.index + n
	if end > len(iter.objects) {
		end = len(iter.objects)
	}
	iter.index = end

	// this is not necessary on every iteration
	// but there may be cases where n in first batch is different from n in next batch
	// to keep things easy creating a new slice every time
	iter.localFiles = make([]string, n)
	g, grpCtx := errgroup.WithContext(iter.ctx)
	g.SetLimit(_concurrentBlobDownloadLimit)
	for i, obj := range iter.objects[start:end] {
		obj := obj
		index := start + i // with repect to object slice
		g.Go(func() error {
			// need to create file by maintaining same dir path as in glob for hivepartition support
			dir := filepath.Join(iter.tempDir, filepath.Dir(obj.obj.Key))
			// filename
			objName := filepath.Base(obj.obj.Key)
			file, err := fileutil.OpenTempFileInDir(dir, objName)
			if err != nil {
				return err
			}
			defer file.Close()

			iter.localFiles[index-start] = file.Name()
			ext := fileutil.FullExt(obj.obj.Key)
			partialReader, isPartialDownloadSupported := _partialDownloadReaders[ext]
			if obj.full || !isPartialDownloadSupported {
				// download full file
				return downloadObject(grpCtx, iter.bucket, obj.obj.Key, file)
			}
			// download partial file
			// check if, for smaller size we can download entire file
			switch partialReader {
			case "parquet":
				return downloadParquet(grpCtx, iter.bucket, obj.obj, obj.extractOption, file)
			case "csv":
				return downloadCSV(grpCtx, iter.bucket, obj.obj, obj.extractOption, file)
			default:
				// should not reach here
				panic(fmt.Errorf("partial download not supported for extension %q", ext))
			}
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	result := make([]string, end-start)
	copy(result, iter.localFiles)
	return result, nil
}

// todo :: ideally planner should take ownership of the bucket and return an iterator with next returning objectWithPlan
func plan(ctx context.Context, bucket *blob.Bucket, opts Options, globPattern string) ([]*objectWithPlan, error) {
	var (
		size, fetched int64
		matchCount    int
	)
	planner, err := newPlanner(opts.ExtractPolicy)
	if err != nil {
		return nil, err
	}

	token := blob.FirstPageToken
	for token != nil && !planner.Done() {
		objs, nextToken, err := bucket.ListPage(ctx, token, opts.GlobPageSize, listOptions(globPattern))
		if err != nil {
			return nil, err
		}

		token = nextToken
		fetched += int64(len(objs))
		for _, obj := range objs {
			if matched, _ := doublestar.Match(globPattern, obj.Key); matched {
				size += obj.Size
				matchCount++
				if !planner.Add(obj) {
					break
				}
			}
		}
		if err := opts.validateLimits(size, matchCount, fetched); err != nil {
			return nil, err
		}
	}

	items := planner.Items()
	if len(items) == 0 {
		return nil, fmt.Errorf("no files found for glob pattern %q", globPattern)
	}
	return items, nil
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

type Options struct {
	GlobMaxTotalSize      int64
	GlobMaxObjectsMatched int
	GlobMaxObjectsListed  int64
	GlobPageSize          int
	ExtractPolicy         *ExtractPolicy
}

// sets defaults if not set by user
func (opts *Options) validate() {
	if opts.GlobMaxObjectsMatched == 0 {
		opts.GlobMaxObjectsMatched = 1000
	}
	if opts.GlobMaxObjectsListed == 0 {
		opts.GlobMaxObjectsListed = 1000 * 1000
	}
	if opts.GlobMaxTotalSize == 0 {
		// 10 GB
		opts.GlobMaxTotalSize = 10 * 1024 * 1024 * 1024
	}
	if opts.GlobPageSize == 0 {
		opts.GlobPageSize = 1000
	}
}

func (opts *Options) validateLimits(size int64, matchCount int, fetched int64) error {
	if size > opts.GlobMaxTotalSize {
		return fmt.Errorf("glob pattern exceeds limits: would fetch more than %d bytes", opts.GlobMaxTotalSize)
	}
	if matchCount > opts.GlobMaxObjectsMatched {
		return fmt.Errorf("glob pattern exceeds limits: matched more than %d files", opts.GlobMaxObjectsMatched)
	}
	if fetched > opts.GlobMaxObjectsListed {
		return fmt.Errorf("glob pattern exceeds limits: listed more than %d files", opts.GlobMaxObjectsListed)
	}
	return nil
}

type ExtractPolicy struct {
	RowsStrategy   runtimev1.Source_ExtractPolicy_Strategy
	RowsLimitBytes uint64
	FilesStrategy  runtimev1.Source_ExtractPolicy_Strategy
	FilesLimit     uint64
}

// todo :: add defaults if required
func NewExtractPolicy(extractPolicy *runtimev1.Source_ExtractPolicy) *ExtractPolicy {
	if extractPolicy == nil {
		return &ExtractPolicy{}
	}

	return &ExtractPolicy{
		FilesStrategy:  extractPolicy.FilesStrategy,
		FilesLimit:     extractPolicy.FilesLimit,
		RowsStrategy:   extractPolicy.RowsStrategy,
		RowsLimitBytes: extractPolicy.RowsLimitBytes,
	}
}
