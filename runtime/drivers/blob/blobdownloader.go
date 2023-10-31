package blob

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
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
	// General blob format (json, csv, parquet, etc)
	Format string
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
		// 2 GB
		opts.BatchSizeBytes = 2 * 1024 * 1024 * 1024
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

// blobIterator implements connector.FileIterator
type blobIterator struct {
	opts      *Options
	logger    *zap.Logger
	bucket    *blob.Bucket
	objects   []*objectWithPlan
	tempDir   string
	lastBatch []string

	ctx         context.Context
	cancel      func()
	batchCh     chan []string       // Channel for batches of downloaded files (buffers up to BatchSizeBytes)
	downloadsCh chan downloadResult // Channel for individual downloaded files
	downloadErr error
}

var _ drivers.FileIterator = &blobIterator{}

// NewIterator returns an iterator for downloading objects matching a glob pattern and extract policy.
// The downloaded objects will be stored in a temporary directory with the same file hierarchy as in the bucket, enabling parsing of hive partitioning on the downloaded files.
// The client should call Close() once done to release all resources.
// Calling Close() on the iterator will also close the bucket.
func NewIterator(ctx context.Context, bucket *blob.Bucket, opts Options, l *zap.Logger) (drivers.FileIterator, error) {
	opts.validate()

	tempDir, err := os.MkdirTemp(os.TempDir(), "blob_ingestion")
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(ctx)

	it := &blobIterator{
		opts:        &opts,
		logger:      l,
		bucket:      bucket,
		tempDir:     tempDir,
		ctx:         ctx,
		cancel:      cancel,
		batchCh:     make(chan []string),
		downloadsCh: make(chan downloadResult),
	}

	objects, err := it.plan()
	if err != nil {
		// close batchCh since it.Close waits on batchCh
		close(it.batchCh)
		it.Close()
		return nil, err
	}
	it.objects = objects

	// Start download of individual files in the background.
	go it.downloadFiles()

	// Start batching of downloaded files in the background.
	// By batching in the background, we can background fetch a dynamic number of files until BatchSizeBytes is reached.
	go it.batchDownloads()

	// For cases where there's only one file, we want to prefetch it to return the error early (from NewIterator instead of Next)
	if len(objects) == 1 {
		it.opts.KeepFilesUntilClose = true
		batch, err := it.Next()
		if err != nil {
			it.Close()
			return nil, err
		}
		return &prefetchedIterator{batch: batch, underlying: it}, nil
	}

	return it, nil
}

func (it *blobIterator) Close() error {
	// Cancel the background downloads (this will eventually close downloadsCh, which eventually closes batchCh)
	it.cancel()

	// Drain batchCh until it's closed (to avoid the background goroutine hanging forever)
	var stop bool
	for !stop {
		_, ok := <-it.batchCh
		if !ok {
			stop = true
		}
	}

	var closeErr error

	// Remove any lingering temporary files
	if it.tempDir != "" {
		err := os.RemoveAll(it.tempDir)
		if err != nil {
			closeErr = errors.Join(closeErr, err)
		}
	}

	// Close the bucket
	err := it.bucket.Close()
	if err != nil {
		closeErr = errors.Join(closeErr, err)
	}

	return closeErr
}

func (it *blobIterator) Size(unit drivers.ProgressUnit) (int64, bool) {
	switch unit {
	case drivers.ProgressUnitByte:
		var size int64
		for _, obj := range it.objects {
			if obj.full {
				size += obj.obj.Size
			} else {
				// TODO: Make it more accurate considering more data can be downloaded
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

func (it *blobIterator) Next() ([]string, error) {
	// Delete files from the previous iteration
	if !it.opts.KeepFilesUntilClose {
		fileutil.ForceRemoveFiles(it.lastBatch)
	}

	// Get next batch
	batch, ok := <-it.batchCh
	if !ok {
		// The channel is closed, we're either done or a download errored.
		if it.downloadErr != nil {
			return nil, it.downloadErr
		}
		return nil, io.EOF
	}

	// Track the batch for cleanup in the next iteration
	it.lastBatch = batch

	// Clients may change the slice. Creating a copy to ensure we delete the files on next batch/close.
	result := make([]string, len(batch))
	copy(result, batch)
	return result, nil
}

func (it *blobIterator) Format() string {
	return it.opts.Format
}

// TODO: Ideally planner should take ownership of the bucket and return an iterator with next returning objectWithPlan
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

func (it *blobIterator) downloadFiles() {
	// Ensure the downloadsCh is closed when the function returns.
	// This unblocks waiting calls to Next() or Close().
	defer close(it.downloadsCh)

	// Create an errgroup for background downloads with limited concurrency.
	g, ctx := errgroup.WithContext(it.ctx)
	g.SetLimit(_concurrentBlobDownloadLimit)

	var loopErr error
	for i := 0; i < len(it.objects); i++ {
		// Stop the loop if the ctx was cancelled
		var stop bool
		select {
		case <-ctx.Done():
			stop = true
		default:
			// don't break
		}
		if stop {
			break // can't use break inside the select
		}

		// Create a path that maintains the same relative path as in the bucket (for hive partition support for globs)
		obj := it.objects[i]
		filename := filepath.Join(it.tempDir, obj.obj.Key)
		if err := os.MkdirAll(filepath.Dir(filename), os.ModePerm); err != nil {
			loopErr = err
			it.cancel() // cancel the context to cancel the errgroup
			break
		}

		// Download the file and send it on downloadsCh.
		// NOTE: Errors returned here will be assigned to it.downloadErr after the loop.
		g.Go(func() error {
			ext := filepath.Ext(obj.obj.Key)
			partialReader, isPartialDownloadSupported := _partialDownloadReaders[ext]
			downloadFull := obj.full || !isPartialDownloadSupported

			startTime := time.Now()
			var file *os.File
			err := retry(5, 10*time.Second, func() error {
				file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
				if err != nil {
					return err
				}
				defer file.Close()

				if downloadFull {
					return downloadObject(ctx, it.bucket, obj.obj.Key, file)
				}
				// download partial file
				switch partialReader {
				case "parquet":
					return downloadParquet(ctx, it.bucket, obj.obj, obj.extractOption, file)
				case "csv":
					return downloadText(ctx, it.bucket, obj.obj, &textExtractOption{extractOption: obj.extractOption, hasCSVHeader: true}, file)
				case "json":
					return downloadText(ctx, it.bucket, obj.obj, &textExtractOption{extractOption: obj.extractOption, hasCSVHeader: false}, file)
				default:
					// should not reach here
					panic(fmt.Errorf("partial download not supported for extension: %q", ext))
				}
			})
			// Returning the err will cancel the errgroup and propagate the error to it.downloadErr
			if err != nil {
				return err
			}

			// Send downloaded file
			// NOTE: Using full object size even for partial downloads. Its okay to not have exact size.
			it.downloadsCh <- downloadResult{path: filename, bytes: obj.obj.Size}

			// Collect metrics of download size and time
			duration := time.Since(startTime)
			size := obj.obj.Size
			st, err := file.Stat()
			if err == nil {
				size = st.Size()
			}
			it.logger.Info("download complete", zap.String("object", obj.obj.Key), zap.Duration("duration", duration), observability.ZapCtx(it.ctx))
			drivers.RecordDownloadMetrics(ctx, &drivers.DownloadMetrics{
				Connector: "blob",
				Ext:       ext,
				Partial:   !downloadFull,
				Duration:  duration,
				Size:      size,
			})

			return nil
		})
	}

	// Wait for all outstanding downloads to complete
	it.downloadErr = g.Wait()

	// If there was an error in the loop, it takes precedence (the error from g.Wait() is probably a ctx cancellation in that case)
	if loopErr != nil {
		it.downloadErr = loopErr
	}
}

func (it *blobIterator) batchDownloads() {
	// Ensure batchCh is closed when this function returns. This ensures Next() and Close() don't hang.
	defer close(it.batchCh)

	var batch []string
	var batchBytes int64
	for {
		// Get a new download
		res, ok := <-it.downloadsCh
		if !ok {
			// Channel closed means no more downloads.
			if it.downloadErr == nil && len(batch) > 0 {
				it.batchCh <- batch
			}
			return
		}

		// Append to batch
		batch = append(batch, res.path)
		batchBytes += res.bytes

		// Send batch if it's full
		if batchBytes >= it.opts.BatchSizeBytes {
			it.batchCh <- batch
			batch = nil
			batchBytes = 0
		}
	}
}

// prefetchedIterator is a lightweight wrapper around blobIterator for returning files that were prefetched during the call to NewIterator.
type prefetchedIterator struct {
	batch      []string
	done       bool
	underlying *blobIterator
}

func (it *prefetchedIterator) Close() error {
	return it.underlying.Close()
}

func (it *prefetchedIterator) Size(unit drivers.ProgressUnit) (int64, bool) {
	return it.underlying.Size(unit)
}

func (it *prefetchedIterator) Next() ([]string, error) {
	if it.done {
		return nil, io.EOF
	}
	it.done = true
	return it.batch, nil
}

func (it *prefetchedIterator) Format() string {
	return it.underlying.Format()
}

// downloadResult represents a successfully downloaded file
type downloadResult struct {
	path  string
	bytes int64
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

func retry(maxRetries int, delay time.Duration, fn func() error) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		err = fn()
		if err == nil {
			return nil // success
		} else if strings.Contains(err.Error(), "stream error: stream ID") {
			time.Sleep(delay) // retry
		} else {
			break // return error
		}
	}
	return err
}
