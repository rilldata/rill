package blob

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
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

type fileBatch []string

// implements connector.FileIterator
type blobIterator struct {
	ctx    context.Context
	cancel func()
	opts   *Options
	logger *zap.Logger
	bucket *blob.Bucket

	// all localfiles are created in this dir
	tempDir   string
	objects   []*objectWithPlan
	lastBatch []string
	grp       *errgroup.Group
	files     chan fileBatch
	err       error // any error in download

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
	// BatchSizeBytes is the combined size of all files returned in one call to next()
	BatchSizeBytes int64
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
	if opts.BatchSizeBytes == 0 {
		// 5 GB
		opts.BatchSizeBytes = 5 * 1024 * 1024 * 1024
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

	ctxWithCancel, cancel := context.WithCancel(ctx)
	it := &blobIterator{
		ctx:    ctxWithCancel,
		cancel: cancel,
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
		it.nextPaths, err = it.Next()
		it.err = nil
		if err != nil {
			it.Close()
			return nil, err
		}
	}

	return it, nil
}

// Close frees the resources
func (it *blobIterator) Close() error {
	it.cancel()
	it.err = fmt.Errorf("closed iterator")
	// drain the channel
	if it.files != nil {
		select {
		case <-it.files:
		default:
		}
	}
	if it.grp != nil {
		_ = it.grp.Wait() // wait for background calls to complete and then delete temp directory
	}
	// remove temp dir since recursive paths created have to be removed as well
	var err error
	if it.tempDir != "" {
		err = os.RemoveAll(it.tempDir)
	}
	if bucketCloseErr := it.bucket.Close(); bucketCloseErr != nil {
		return bucketCloseErr
	}
	return err
}

// Next downloads next n files and copies to local directory
func (it *blobIterator) Next() ([]string, error) {
	if it.err != nil {
		return nil, it.err
	}
	if len(it.nextPaths) != 0 { // single file path
		paths := it.nextPaths
		it.nextPaths = nil
		it.err = io.EOF
		return paths, nil
	}

	if it.files == nil {
		// start downloading files
		it.files = make(chan fileBatch)
		go func() {
			it.init()
		}()
	}

	if !it.opts.KeepFilesUntilClose {
		// delete files created in previous iteration
		fileutil.ForceRemoveFiles(it.lastBatch)
	}
	files, ok := <-it.files
	if !ok {
		if it.err == nil {
			// downloaded successfully
			it.err = io.EOF
		}
		return nil, it.err
	}
	it.lastBatch = files
	// clients can make changes to slice if passing the same slice that iterator holds
	// creating a copy since we want to delete all these files on next batch/close
	result := make([]string, len(it.lastBatch))
	copy(result, it.lastBatch)
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

func (it *blobIterator) init() {
	g, grpCtx := errgroup.WithContext(it.ctx)
	it.grp = g
	g.SetLimit(_concurrentBlobDownloadLimit)
	var stop bool
	files := make([]string, 0)
	var sizeInBytes int64
	var wg sync.WaitGroup
	for i := 0; i < len(it.objects) && !stop; i++ {
		select {
		case <-it.ctx.Done():
			close(it.files)
			return
		default:
			obj := it.objects[i]
			// we can download partial file as well but its okay to use full object size since its okay to not have exact size
			sizeInBytes += obj.obj.Size
			// need to create file by maintaining same dir path as in glob for hivepartition support
			filename := filepath.Join(it.tempDir, obj.obj.Key)
			files = append(files, filename)
			wg.Add(1)
			g.Go(func() error {
				defer wg.Done()
				if err := os.MkdirAll(filepath.Dir(filename), os.ModePerm); err != nil {
					return err
				}

				file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, os.ModePerm)
				if err != nil {
					return err
				}
				defer file.Close()

				ext := filepath.Ext(obj.obj.Key)
				partialReader, isPartialDownloadSupported := _partialDownloadReaders[ext]
				downloadFull := obj.full || !isPartialDownloadSupported

				// Collect metrics of download size and time
				startTime := time.Now()
				if downloadFull {
					// download full file
					err = downloadObject(grpCtx, it.bucket, obj.obj.Key, file)
				} else {
					// download partial file
					// check if, for smaller size we can download entire file
					switch partialReader {
					case "parquet":
						err = downloadParquet(grpCtx, it.bucket, obj.obj, obj.extractOption, file)
					case "csv":
						err = downloadText(grpCtx, it.bucket, obj.obj, &textExtractOption{extractOption: obj.extractOption, hasCSVHeader: true}, file)
					case "json":
						err = downloadText(grpCtx, it.bucket, obj.obj, &textExtractOption{extractOption: obj.extractOption, hasCSVHeader: false}, file)
					default:
						// should not reach here
						panic(fmt.Errorf("partial download not supported for extension %q", ext))
					}
				}

				if err != nil {
					// found an error stop triggering download of other files
					stop = true
					return err
				}

				duration := time.Since(startTime)
				size := obj.obj.Size
				st, err := file.Stat()
				if err == nil {
					size = st.Size()
				}
				it.logger.Info("download complete", zap.String("object", obj.obj.Key), zap.Duration("duration", duration), observability.ZapCtx(it.ctx))
				drivers.RecordDownloadMetrics(grpCtx, &drivers.DownloadMetrics{
					Connector: "blob",
					Ext:       ext,
					Partial:   !downloadFull,
					Duration:  duration,
					Size:      size,
				})
				return nil
			})
			// check if we have downloaded enough data and can be pushed to the channel
			if sizeInBytes >= it.opts.BatchSizeBytes {
				wg.Wait()
				if stop { // download failed, exit
					it.err = g.Wait()
					close(it.files)
					return
				}
				it.files <- files
				files = make([]string, 0)
				sizeInBytes = 0
			}
		}
	}
	// final batch
	it.err = g.Wait()
	if it.err == nil && len(files) != 0 {
		it.files <- files
	}
	close(it.files)
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

	items := planner.items()
	if len(items) == 0 {
		return nil, fmt.Errorf("no files found for glob pattern %q", it.opts.GlobPattern)
	}

	it.logger.Info("planner completed", zap.String("glob", it.opts.GlobPattern), zap.Int64("listed_objects", fetched),
		zap.Int("matched", matchCount), zap.Int64("bytes_matched", size), zap.Int64("batch_size", it.opts.BatchSizeBytes),
		observability.ZapCtx(it.ctx))
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
