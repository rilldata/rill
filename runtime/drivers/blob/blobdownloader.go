package blob

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"cloud.google.com/go/storage"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/zap"
	"gocloud.dev/blob"
	"golang.org/x/sync/errgroup"
)

// Number of concurrent file downloads.
// 23-11-13: Experimented with increasing the value to 16. It caused network saturation errors on macOS.
const _concurrentBlobDownloadLimit = 8

// map of supoprted extensions for partial downloads vs readers
// zipped csv files can't be partialled downloaded
// parquet files with compression has extension in format .<compression>.parquet eg: .gz.parquet
var _partialDownloadReaders = map[string]string{
	".parquet": "parquet",
	".csv":     "csv",
	".tsv":     "csv",
	".txt":     "csv",
	".ndjson":  "json",
	".json":    "json",
}

// implements connector.FileIterator
type blobIterator struct {
	ctx        context.Context
	bucket     *blob.Bucket
	objects    []*objectWithPlan
	index      int
	localFiles []string
	// all localfiles are created in this dir
	tempDir string
	opts    *Options
	logger  *zap.Logger
	// data is already fetched during planning stage itself for single file cases
	// TODO :: refactor this to return a different iterator maybe ?
	nextPaths []string
}

var _ drivers.FileIterator = &blobIterator{}

type Options struct {
	GlobMaxTotalSize      int64
	GlobMaxObjectsMatched int
	GlobMaxObjectsListed  int64
	GlobPageSize          int
	ExtractPolicy         *ExtractPolicy
	GlobPattern           string
	// Although at this point GlobMaxTotalSize and StorageLimitInBytes have same impl but
	// this is total size the source should consume on disk and is calculated upstream basis how much data one instance has already consumed
	// across other sources and the instance level limits
	StorageLimitInBytes int64
	// Retain files and only delete during close
	KeepFilesUntilClose bool
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

// NewIterator returns new instance of blobIterator
// the iterator keeps list of blob objects eagerly planned as per user's glob pattern and extract policies
// clients should call close once done to release all resources like closing the bucket
// the iterator takes responsibility of closing the bucket
func NewIterator(ctx context.Context, bucket *blob.Bucket, opts Options, l *zap.Logger) (drivers.FileIterator, error) {
	opts.validate()

	it := &blobIterator{
		ctx:    ctx,
		bucket: bucket,
		opts:   &opts,
		logger: l,
	}

	tempDir, err := os.MkdirTemp(os.TempDir(), "blob_ingestion")
	if err != nil {
		it.Close()
		return nil, err
	}
	it.tempDir = tempDir

	objects, err := it.plan()
	if err != nil {
		it.Close()
		return nil, err
	}
	it.objects = objects
	if len(objects) == 1 {
		it.nextPaths, err = it.NextBatch(1)
		it.index = 0
		if err != nil {
			it.Close()
			return nil, err
		}
	}

	return it, nil
}

// Close frees the resources
func (it *blobIterator) Close() error {
	// remove temp dir since recursive paths created have to be removed as well
	err := os.RemoveAll(it.tempDir)
	if bucketCloseErr := it.bucket.Close(); bucketCloseErr != nil {
		return bucketCloseErr
	}
	return err
}

// HasNext returns true if iterator has more data
func (it *blobIterator) HasNext() bool {
	return it.index < len(it.objects)
}

// NextBatchSize downloads next n files and copies to local directory
func (it *blobIterator) NextBatchSize(sizeInBytes int64) ([]string, error) {
	if !it.HasNext() {
		return nil, io.EOF
	}
	if len(it.nextPaths) != 0 {
		paths := it.nextPaths
		it.index += len(paths)
		it.nextPaths = nil
		return paths, nil
	}

	if !it.opts.KeepFilesUntilClose {
		// delete files created in last iteration
		fileutil.ForceRemoveFiles(it.localFiles)
	}

	// new slice creation is not necessary on every iteration
	// but there may be cases where n in first batch is different from n in next batch
	// to keep things easy creating a new slice every time
	it.localFiles = make([]string, 0)
	g, grpCtx := errgroup.WithContext(it.ctx)
	g.SetLimit(_concurrentBlobDownloadLimit)

	var totalSizeInBytes int64
	start := it.index
	for ; it.index < len(it.objects) && totalSizeInBytes < sizeInBytes; it.index++ {
		obj := it.objects[it.index]
		totalSizeInBytes += obj.obj.Size
		g.Go(func() error {
			// need to create file by maintaining same dir path as in glob for hivepartition support
			filename := filepath.Join(it.tempDir, obj.obj.Key)
			if err := os.MkdirAll(filepath.Dir(filename), os.ModePerm); err != nil {
				return err
			}

			file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, os.ModePerm)
			if err != nil {
				return err
			}
			defer file.Close()

			it.localFiles = append(it.localFiles, file.Name())
			ext := filepath.Ext(obj.obj.Key)
			partialReader, isPartialDownloadSupported := _partialDownloadReaders[ext]
			downloadFull := obj.full || !isPartialDownloadSupported

			// Collect metrics of download size and time
			startTime := time.Now()
			defer func() {
				size := obj.obj.Size
				st, err := file.Stat()
				if err == nil {
					size = st.Size()
				}

				duration := time.Since(startTime)
				it.logger.Info("download complete", zap.String("object", obj.obj.Key), zap.Duration("duration", duration), observability.ZapCtx(it.ctx))
				drivers.RecordDownloadMetrics(grpCtx, &drivers.DownloadMetrics{
					Connector: "blob",
					Ext:       ext,
					Partial:   !downloadFull,
					Duration:  duration,
					Size:      size,
				})
			}()

			// download full file
			if downloadFull {
				return downloadObject(grpCtx, it.bucket, obj.obj.Key, file)
			}
			// download partial file
			// check if, for smaller size we can download entire file
			switch partialReader {
			case "parquet":
				return downloadParquet(grpCtx, it.bucket, obj.obj, obj.extractOption, file)
			case "csv":
				return downloadText(grpCtx, it.bucket, obj.obj, &textExtractOption{extractOption: obj.extractOption, hasCSVHeader: true}, file)
			case "json":
				return downloadText(grpCtx, it.bucket, obj.obj, &textExtractOption{extractOption: obj.extractOption, hasCSVHeader: false}, file)
			default:
				// should not reach here
				panic(fmt.Errorf("partial download not supported for extension %q", ext))
			}
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	// clients can make changes to slice if passing the same slice that iterator holds
	// creating a copy since we want to delete all these files on next batch/close
	result := make([]string, it.index-start)
	copy(result, it.localFiles)
	return result, nil
}

// NextBatch downloads next n files and copies to local directory
func (it *blobIterator) NextBatch(n int) ([]string, error) {
	if !it.HasNext() {
		return nil, io.EOF
	}
	if len(it.nextPaths) != 0 {
		paths := it.nextPaths
		it.index += len(paths)
		it.nextPaths = nil
		return paths, nil
	}

	if !it.opts.KeepFilesUntilClose {
		// delete files created in last iteration
		fileutil.ForceRemoveFiles(it.localFiles)
	}
	start := it.index
	end := it.index + n
	if end > len(it.objects) {
		end = len(it.objects)
	}
	it.index = end

	// new slice creation is not necessary on every iteration
	// but there may be cases where n in first batch is different from n in next batch
	// to keep things easy creating a new slice every time
	it.localFiles = make([]string, end-start)
	g, grpCtx := errgroup.WithContext(it.ctx)
	g.SetLimit(_concurrentBlobDownloadLimit)
	for i, obj := range it.objects[start:end] {
		obj := obj
		index := start + i // with repect to object slice
		g.Go(func() error {
			// need to create file by maintaining same dir path as in glob for hivepartition support
			filename := filepath.Join(it.tempDir, obj.obj.Key)
			if err := os.MkdirAll(filepath.Dir(filename), os.ModePerm); err != nil {
				return err
			}

			file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, os.ModePerm)
			if err != nil {
				return err
			}
			defer file.Close()

			it.localFiles[index-start] = file.Name()
			ext := filepath.Ext(obj.obj.Key)
			partialReader, isPartialDownloadSupported := _partialDownloadReaders[ext]
			downloadFull := obj.full || !isPartialDownloadSupported

			// Collect metrics of download size and time
			startTime := time.Now()
			defer func() {
				size := obj.obj.Size
				st, err := file.Stat()
				if err == nil {
					size = st.Size()
				}

				duration := time.Since(startTime)
				it.logger.Info("download complete", zap.String("object", obj.obj.Key), zap.Duration("duration", duration), observability.ZapCtx(it.ctx))
				drivers.RecordDownloadMetrics(grpCtx, &drivers.DownloadMetrics{
					Connector: "blob",
					Ext:       ext,
					Partial:   !downloadFull,
					Duration:  duration,
					Size:      size,
				})
			}()

			// download full file
			if downloadFull {
				return downloadObject(grpCtx, it.bucket, obj.obj.Key, file)
			}
			// download partial file
			// check if, for smaller size we can download entire file
			switch partialReader {
			case "parquet":
				return downloadParquet(grpCtx, it.bucket, obj.obj, obj.extractOption, file)
			case "csv":
				return downloadText(grpCtx, it.bucket, obj.obj, &textExtractOption{extractOption: obj.extractOption, hasCSVHeader: true}, file)
			case "json":
				return downloadText(grpCtx, it.bucket, obj.obj, &textExtractOption{extractOption: obj.extractOption, hasCSVHeader: false}, file)
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
	// clients can make changes to slice if passing the same slice that iterator holds
	// creating a copy since we want to delete all these files on next batch/close
	copy(result, it.localFiles)
	return result, nil
}

func (it *blobIterator) Size(unit drivers.ProgressUnit) (int64, bool) {
	switch unit {
	case drivers.ProgressUnitByte:
		var size int64
		for _, obj := range it.objects {
			if obj.full {
				size += obj.obj.Size
			} else {
				// TODO :: make it more accurate considering more data can be downloaded
				size += int64(obj.extractOption.limitInBytes)
			}
		}
		return size, true
	case drivers.ProgressUnitFile:
		return int64(len(it.objects)), true
	default:
		return 0, false
	}
}

func (it *blobIterator) KeepFilesUntilClose(keepFilesUntilClose bool) {
	it.opts.KeepFilesUntilClose = keepFilesUntilClose
}

// todo :: ideally planner should take ownership of the bucket and return an iterator with next returning objectWithPlan
func (it *blobIterator) plan() ([]*objectWithPlan, error) {
	var (
		size, fetched int64
		matchCount    int
	)
	planner, err := newPlanner(it.opts.ExtractPolicy)
	if err != nil {
		return nil, err
	}

	listOpts, ok := listOptions(it.opts.GlobPattern)
	if !ok {
		it.logger.Info("glob pattern corresponds to single object", zap.String("glob", it.opts.GlobPattern))
		// required to fetch size to enforce disk limits
		attr, err := it.bucket.Attributes(it.ctx, it.opts.GlobPattern)
		if err != nil {
			// can fail due to permission not available
			it.logger.Info("failed to fetch attributes of the object", zap.Error(err))
		} else {
			size = attr.Size
		}

		planner.add(&blob.ListObject{Key: it.opts.GlobPattern, Size: size})
		if err := it.opts.validateLimits(size, 1, 1); err != nil {
			return nil, err
		}
		return planner.items(), nil
	}
	it.logger.Info("planner started", zap.String("glob", it.opts.GlobPattern), zap.String("prefix", listOpts.Prefix), observability.ZapCtx(it.ctx))
	token := blob.FirstPageToken
	for token != nil && !planner.done() {
		objs, nextToken, err := it.bucket.ListPage(it.ctx, token, it.opts.GlobPageSize, listOpts)
		if err != nil {
			return nil, err
		}

		token = nextToken
		fetched += int64(len(objs))
		for _, obj := range objs {
			if matched, _ := doublestar.Match(it.opts.GlobPattern, obj.Key); matched {
				size += obj.Size
				matchCount++
				if !planner.add(obj) {
					break
				}
			}
		}
		if err := it.opts.validateLimits(size, matchCount, fetched); err != nil {
			return nil, err
		}
	}

	it.logger.Info("planner completed", zap.String("glob", it.opts.GlobPattern), zap.Int64("listed_objects", fetched), zap.Int("matched", matchCount), zap.Int64("bytes_matched", size), observability.ZapCtx(it.ctx))

	items := planner.items()
	if len(items) == 0 {
		return nil, fmt.Errorf("no files found for glob pattern %q", it.opts.GlobPattern)
	}
	return items, nil
}

// listOptions for page listing api
func listOptions(globPattern string) (*blob.ListOptions, bool) {
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
		return nil, false
	} else if prefix != "." {
		listOptions.Prefix = prefix
	}

	return listOptions, true
}

// download full object
func downloadObject(ctx context.Context, bucket *blob.Bucket, objpath string, file *os.File) error {
	rc, err := bucket.NewReader(ctx, objpath, nil)
	if err != nil {
		return fmt.Errorf("Object(%q).NewReader: %w", objpath, err)
	}
	defer rc.Close()

	_, err = io.Copy(file, rc)
	return err
}
