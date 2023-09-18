package bigquery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/apache/arrow/go/v13/parquet"
	"github.com/apache/arrow/go/v13/parquet/compress"
	"github.com/apache/arrow/go/v13/parquet/pqarrow"
	"github.com/c2h5oh/datasize"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/zap"
	"google.golang.org/api/iterator"
)

// recommended size is 512MB - 1GB, entire data is buffered in memory before its written to disk
const rowGroupBufferSize = int64(datasize.MB) * 512

const _jsonDownloadLimitBytes = 100 * int64(datasize.MB)

// Query implements drivers.SQLStore
func (c *Connection) Query(ctx context.Context, props map[string]any) (drivers.RowIterator, error) {
	return nil, fmt.Errorf("not implemented")
}

// QueryAsFiles implements drivers.SQLStore
func (c *Connection) QueryAsFiles(ctx context.Context, props map[string]any, opt *drivers.QueryOption, p drivers.Progress) (drivers.FileIterator, error) {
	srcProps, err := parseSourceProperties(props)
	if err != nil {
		return nil, err
	}

	opts, err := c.clientOption(ctx)
	if err != nil {
		return nil, err
	}

	client, err := bigquery.NewClient(ctx, srcProps.ProjectID, opts...)
	if err != nil {
		if strings.Contains(err.Error(), "unable to detect projectID") {
			return nil, fmt.Errorf("projectID not detected in credentials. Please set `project_id` in source yaml")
		}
		return nil, fmt.Errorf("failed to create bigquery client: %w", err)
	}

	if err := client.EnableStorageReadClient(ctx, opts...); err != nil {
		client.Close()
		return nil, err
	}

	now := time.Now()
	q := client.Query(srcProps.SQL)
	it, err := q.Read(ctx)
	if err != nil && !strings.Contains(err.Error(), "Syntax error") {
		// close the read storage API client
		client.Close()
		c.logger.Info("query failed, retrying without storage api", zap.Error(err))
		// the query results are always cached in a temporary table that storage api can use
		// there are some exceptions when results aren't cached
		// so we also try without storage api
		client, err = bigquery.NewClient(ctx, srcProps.ProjectID, opts...)
		if err != nil {
			return nil, fmt.Errorf("failed to create bigquery client: %w", err)
		}

		q := client.Query(srcProps.SQL)
		it, err = q.Read(ctx)
	}
	if err != nil {
		client.Close()
		return nil, err
	}
	c.logger.Info("query took", zap.Duration("duration", time.Since(now)), observability.ZapCtx(ctx))

	p.Target(int64(it.TotalRows), drivers.ProgressUnitRecord)
	return &fileIterator{
		client:       client,
		bqIter:       it,
		logger:       c.logger,
		limitInBytes: opt.TotalLimitInBytes,
		progress:     p,
		totalRecords: int64(it.TotalRows),
		ctx:          ctx,
	}, nil
}

type fileIterator struct {
	client       *bigquery.Client
	bqIter       *bigquery.RowIterator
	logger       *zap.Logger
	limitInBytes int64
	progress     drivers.Progress

	totalRecords int64
	tempFilePath string
	downloaded   bool

	ctx context.Context // TODO :: refatcor NextBatch to take context on NextBatch
}

// Close implements drivers.FileIterator.
func (f *fileIterator) Close() error {
	return os.Remove(f.tempFilePath)
}

// HasNext implements drivers.FileIterator.
func (f *fileIterator) HasNext() bool {
	return !f.downloaded
}

// KeepFilesUntilClose implements drivers.FileIterator.
func (f *fileIterator) KeepFilesUntilClose(keepFilesUntilClose bool) {
}

func (f *fileIterator) NextBatchSize(sizeInBytes int64) ([]string, error) {
	return f.NextBatch(1)
}

// NextBatch implements drivers.FileIterator.
// TODO :: currently it downloads all records in a single file. Need to check if it is efficient to ingest a single file with size in tens of GBs or more.
func (f *fileIterator) NextBatch(limit int) ([]string, error) {
	// storage API not available so can't read as arrow records. Read results row by row and dump in a json file.
	if !f.bqIter.IsAccelerated() {
		f.logger.Info("downloading results in json file", observability.ZapCtx(f.ctx))
		if err := f.downloadAsJSONFile(); err != nil {
			return nil, err
		}
		return []string{f.tempFilePath}, nil
	}
	f.logger.Info("downloading results in parquet file", observability.ZapCtx(f.ctx))

	// create a temp file
	fw, err := os.CreateTemp("", "temp*.parquet")
	if err != nil {
		return nil, err
	}
	defer fw.Close()
	f.tempFilePath = fw.Name()
	f.downloaded = true

	rdr, err := f.AsArrowRecordReader()
	if err != nil {
		return nil, err
	}
	defer rdr.Release()

	tf := time.Now()
	defer func() {
		f.logger.Info("time taken to write arrow records in parquet file", zap.Duration("duration", time.Since(tf)), observability.ZapCtx(f.ctx))
	}()
	writer, err := pqarrow.NewFileWriter(rdr.Schema(), fw,
		parquet.NewWriterProperties(
			parquet.WithCompression(compress.Codecs.Snappy),
			parquet.WithRootRepetition(parquet.Repetitions.Required),
			// duckdb has issues reading statistics of string type generated with this write
			// column statistics may not be useful if full file need to be ingested so better to disable to save computations
			parquet.WithStats(false),
		),
		pqarrow.NewArrowWriterProperties(pqarrow.WithStoreSchema()))
	if err != nil {
		if strings.Contains(err.Error(), "not implemented: support for DECIMAL256") {
			return nil, fmt.Errorf("BIGNUMERIC datatype is not supported. Consider casting to STRING or NUMERIC (if loss of precision is acceptable) in the submitted query")
		}
		return nil, err
	}
	defer writer.Close()

	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()
	// write arrow records to parquet file
	for rdr.Next() {
		select {
		case <-f.ctx.Done():
			return nil, f.ctx.Err()
		case <-ticker.C:
			fileInfo, err := os.Stat(fw.Name())
			if err == nil { // ignore error
				if fileInfo.Size() > f.limitInBytes {
					return nil, drivers.ErrIngestionLimitExceeded
				}
			}

		default:
			rec := rdr.Record()
			f.progress.Observe(rec.NumRows(), drivers.ProgressUnitRecord)
			if writer.RowGroupTotalBytesWritten() >= rowGroupBufferSize {
				writer.NewBufferedRowGroup()
			}
			if err := writer.WriteBuffered(rec); err != nil {
				return nil, err
			}
		}
	}
	if rdr.Err() != nil {
		return nil, fmt.Errorf("file write failed with error: %w", rdr.Err())
	}
	writer.Close()
	fw.Close()

	fileInfo, err := os.Stat(fw.Name())
	if err != nil {
		return nil, err
	}
	f.logger.Info("size of file", zap.String("size", datasize.ByteSize(fileInfo.Size()).HumanReadable()), observability.ZapCtx(f.ctx))
	return []string{fw.Name()}, nil
}

// Size implements drivers.FileIterator.
func (f *fileIterator) Size(unit drivers.ProgressUnit) (int64, bool) {
	switch unit {
	case drivers.ProgressUnitRecord:
		return f.totalRecords, true
	case drivers.ProgressUnitFile:
		return 1, true
	default:
		return 0, false
	}
}

func (f *fileIterator) downloadAsJSONFile() error {
	tf := time.Now()
	defer func() {
		f.logger.Info("time taken to write row in json file", zap.Duration("duration", time.Since(tf)), observability.ZapCtx(f.ctx))
	}()

	// create a temp file
	fw, err := os.CreateTemp("", "temp*.ndjson")
	if err != nil {
		return err
	}
	defer fw.Close()
	f.tempFilePath = fw.Name()
	f.downloaded = true

	init := false
	rows := 0
	enc := json.NewEncoder(fw)
	enc.SetEscapeHTML(false)
	for {
		row := make(map[string]bigquery.Value)
		err := f.bqIter.Next(&row)
		if err != nil {
			if errors.Is(err, iterator.Done) {
				if !init {
					return fmt.Errorf("no results found for the query")
				}
				return nil
			}
			return err
		}

		// schema and total rows is available after first call to next only
		if !init {
			init = true
			f.progress.Target(int64(f.bqIter.TotalRows), drivers.ProgressUnitRecord)
			if hasBigNumericType(f.bqIter.Schema) {
				return fmt.Errorf("BIGNUMERIC datatype is not supported. Consider casting to STRING or NUMERIC (if loss of precision is acceptable) in the submitted query")
			}
		}

		err = enc.Encode(row)
		if err != nil {
			return fmt.Errorf("conversion of row to json failed with error: %w", err)
		}

		// If we don't have storage API access, BigQuery may return massive JSON results. (But even with storage API access, it may return JSON for small results.)
		// We want to avoid JSON for massive results. Currently, the only way to do so is to error at a limit.
		rows++
		if rows != 0 && rows%10000 == 0 { // Check file size every 10k rows
			fileInfo, err := os.Stat(fw.Name())
			if err != nil {
				return fmt.Errorf("bigquery: failed to poll json file size: %w", err)
			}
			if fileInfo.Size() >= _jsonDownloadLimitBytes {
				return fmt.Errorf("bigquery: json download exceeded limit of %d bytes (enable and provide access to the BigQuery Storage Read API to read larger results)", _jsonDownloadLimitBytes)
			}
		}
	}
}

func hasBigNumericType(s bigquery.Schema) bool {
	for _, f := range s {
		if f.Type == bigquery.BigNumericFieldType {
			return true
		} else if f.Type == bigquery.RecordFieldType && hasBigNumericType(f.Schema) {
			return true
		}
	}
	return false
}

var _ drivers.FileIterator = &fileIterator{}
