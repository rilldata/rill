package blob

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// Number of concurrent file downloads.
// 23-11-13: Experimented with increasing the value to 16. It caused network saturation errors on macOS.
const _concurrentBlobDownloadLimit = 8

// Default batch size
const _defaultBatchSizeBytes = 1024 * 1024 * 1024 // 1 GB

// Metrics
var (
	tracer                = otel.Tracer("github.com/rilldata/rill/runtime/pkg/blob")
	meter                 = otel.Meter("github.com/rilldata/rill/runtime/pkg/blob")
	downloadTimeHistogram = observability.Must(meter.Float64Histogram("download.time", metric.WithUnit("s")))
	downloadSizeCounter   = observability.Must(meter.Int64UpDownCounter("download.size", metric.WithUnit("bytes")))
	downloadSpeedCounter  = observability.Must(meter.Float64UpDownCounter("download.speed", metric.WithUnit("bytes/s")))
)

// Options for Download
type DownloadOptions struct {
	// Glob is the pattern to match files in the bucket
	Glob string
	// Format is the format of the files (e.g. "csv", "json", "parquet")
	Format string
	// TempDir where temporary files should be stored
	TempDir string
	// KeepFilesUntilClose if true, files will not be deleted until Close() is called.
	KeepFilesUntilClose bool
	// BatchSizeBytes is the size of the batch to download before sending it to the client.
	BatchSizeBytes int64
	// CloseBucket will close the bucket when the iterator is closed.
	CloseBucket bool
}

// Download returns an iterator for downloading objects matching a glob pattern.
// The downloaded objects will be stored in a temporary directory with the same file hierarchy the bucket, enabling parsing of hive partitioning on the downloaded files.
// The client should call Close() once done to release all resources.
// Calling Close() on the iterator will also close the bucket.
func (b *Bucket) Download(ctx context.Context, opts *DownloadOptions) (res drivers.FileIterator, resErr error) {
	defer func() {
		if resErr != nil && opts.CloseBucket {
			_ = b.Close()
		}
	}()

	if opts.BatchSizeBytes <= 0 {
		opts.BatchSizeBytes = _defaultBatchSizeBytes
	}

	tempDir, err := os.MkdirTemp(opts.TempDir, "blob_ingestion")
	if err != nil {
		return nil, err
	}

	entries, err := b.ListObjectsForGlob(ctx, opts.Glob)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(ctx)

	it := &blobIterator{
		bucket:      b,
		opts:        opts,
		objects:     entries,
		tempDir:     tempDir,
		ctx:         ctx,
		cancel:      cancel,
		batchCh:     make(chan []string),
		downloadsCh: make(chan downloadResult),
	}

	// Start download of individual files in the background.
	go it.downloadFiles()

	// Start batching of downloaded files in the background.
	// By batching in the background, we can background fetch a dynamic number of files until BatchSizeBytes is reached.
	go it.batchDownloads()

	return it, nil
}

// blobIterator implements connector.FileIterator
type blobIterator struct {
	bucket  *Bucket
	opts    *DownloadOptions
	objects []drivers.ObjectStoreEntry
	tempDir string

	ctx         context.Context
	cancel      func()
	batchCh     chan []string       // Channel for batches of downloaded files (buffers up to BatchSizeBytes)
	downloadsCh chan downloadResult // Channel for individual downloaded files
	downloadErr error
	lastBatch   []string
}

var _ drivers.FileIterator = &blobIterator{}

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
	if it.opts.CloseBucket {
		err := it.bucket.Close()
		if err != nil {
			closeErr = errors.Join(closeErr, err)
		}
	}

	return closeErr
}

func (it *blobIterator) Format() string {
	return it.opts.Format
}

func (it *blobIterator) SetKeepFilesUntilClose() {
	// Set the flag to keep files until Close() is called.
	it.opts.KeepFilesUntilClose = true
}

func (it *blobIterator) Next(ctx context.Context) (res []string, resErr error) {
	// Even though the download happens in a goroutine, adding a trace here is still a good approximation of how long it waits for a download.
	_, span := tracer.Start(ctx, "blobIterator.Next")
	defer func() {
		if resErr != nil {
			span.SetStatus(codes.Error, resErr.Error())
		}
		span.End()
	}()

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

func (it *blobIterator) downloadFiles() {
	// Ensure the downloadsCh is closed when the function returns.
	// This unblocks waiting calls to Next() or Close().
	defer close(it.downloadsCh)

	// Create an errgroup for background downloads with limited concurrency.
	g, ctx := errgroup.WithContext(it.ctx)
	g.SetLimit(_concurrentBlobDownloadLimit)

	var loopErr error
	for i := 0; i < len(it.objects); i++ {
		obj := it.objects[i]
		if obj.IsDir {
			continue // skip directories
		}

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
		filename := filepath.Join(it.tempDir, obj.Path)
		if err := os.MkdirAll(filepath.Dir(filename), os.ModePerm); err != nil {
			loopErr = err
			it.cancel() // cancel the context to cancel the errgroup
			break
		}

		// Download the file and send it on downloadsCh.
		// NOTE: Errors returned here will be assigned to it.downloadErr after the loop.
		g.Go(func() error {
			startTime := time.Now()
			err := retry(ctx, 5, 10*time.Second, func() error {
				file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
				if err != nil {
					return err
				}
				defer file.Close()

				rc, err := it.bucket.Underlying().NewReader(ctx, obj.Path, nil)
				if err != nil {
					return err
				}
				defer rc.Close()

				_, err = io.Copy(file, rc)
				return err
			})
			if err != nil {
				return err
			}

			// Send downloaded file
			it.downloadsCh <- downloadResult{path: filename, bytes: obj.Size}

			// Collect metrics of download size and time
			attrs := attribute.NewSet(attribute.String("path", it.opts.Glob))
			duration := time.Since(startTime)
			downloadTimeHistogram.Record(ctx, duration.Seconds(), metric.WithAttributeSet(attrs))
			downloadSizeCounter.Add(ctx, obj.Size, metric.WithAttributeSet(attrs))
			if duration.Seconds() != 0 {
				downloadSpeedCounter.Add(ctx, float64(obj.Size)/duration.Seconds(), metric.WithAttributeSet(attrs))
			}
			it.bucket.logger.Debug("object store: downloaded object", zap.String("object", obj.Path), zap.Duration("duration", duration), observability.ZapCtx(it.ctx))

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

// downloadResult represents a successfully downloaded file
type downloadResult struct {
	path  string
	bytes int64
}

func retry(ctx context.Context, maxRetries int, delay time.Duration, fn func() error) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		err = fn()
		if err == nil {
			return nil // success
		} else if !strings.Contains(err.Error(), "stream error: stream ID") {
			break // break and return error
		}

		select {
		case <-ctx.Done():
			return ctx.Err() // return on context cancellation
		case <-time.After(delay):
		}
	}
	return err
}
