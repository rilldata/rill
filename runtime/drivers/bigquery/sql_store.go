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
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/zap"
	"google.golang.org/api/iterator"
)

// Query implements drivers.SQLStore
func (c *Connection) Query(ctx context.Context, props map[string]any, sql string) (drivers.RowIterator, error) {
	return nil, fmt.Errorf("not implemented")
}

// QueryAsFiles implements drivers.SQLStore
func (c *Connection) QueryAsFiles(ctx context.Context, props map[string]any, sql string, opt *drivers.QueryOption, p drivers.Progress) (drivers.FileIterator, error) {
	srcProps, err := parseSourceProperties(props)
	if err != nil {
		return nil, err
	}

	client, err := c.createClient(ctx, srcProps)
	if err != nil {
		if strings.Contains(err.Error(), "unable to detect projectID") {
			return nil, fmt.Errorf("projectID not detected in credentials. Please set `project_id` in source yaml")
		}
		return nil, fmt.Errorf("failed to create bigquery client: %w", err)
	}

	if err := client.EnableStorageReadClient(ctx); err != nil {
		client.Close()
		return nil, err
	}

	now := time.Now()
	q := client.Query(sql)
	it, err := q.Read(ctx)
	if err != nil && !strings.Contains(err.Error(), "Syntax error") {
		// close the read storage API client
		client.Close()
		c.logger.Info("query failed, retrying without storage api", zap.Error(err))
		// the query results are always cached in a temporary table that storage api can use
		// there are some exceptions when results aren't cached
		// so we also try without storage api
		client, err = c.createClient(ctx, srcProps)
		if err != nil {
			return nil, fmt.Errorf("failed to create bigquery client: %w", err)
		}

		q := client.Query(sql)
		it, err = q.Read(ctx)
	}
	if err != nil {
		client.Close()
		return nil, err
	}
	c.logger.Info("query took", zap.Duration("duration", time.Since(now)))

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

// NextBatch implements drivers.FileIterator.
// TODO :: currently it downloads all records in a single file. Need to check if it is efficient to ingest a single file with size in tens of GBs or more.
func (f *fileIterator) NextBatch(limit int) ([]string, error) {
	// storage API not available so can't read as arrow records. Read results row by row and dump in a json file.
	if !f.bqIter.IsAccelerated() {
		if err := f.downloadAsJSONFile(); err != nil {
			return nil, err
		}
		return []string{f.tempFilePath}, nil
	}

	// create a temp file
	fw, err := fileutil.OpenTempFileInDir("", "temp.parquet")
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
		return nil, err
	}
	defer writer.Close()

	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()
	// write arrow records to parquet file
	for rdr.Next() {
		select {
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
	fw, err := fileutil.OpenTempFileInDir("", "temp.ndjson")
	if err != nil {
		return err
	}
	defer fw.Close()
	f.tempFilePath = fw.Name()
	f.downloaded = true

	// not implementing size check since this flow is expected to be run for less data size only
	for {
		row := make(map[string]bigquery.Value)
		if err := f.bqIter.Next(&row); err != nil {
			if errors.Is(err, iterator.Done) {
				return nil
			}
			return err
		}

		bytes, err := json.Marshal(row)
		if err != nil {
			return fmt.Errorf("conversion of row to json failed with error: %w", err)
		}
		if _, err = fw.Write(bytes); err != nil {
			return err
		}
		if _, err = fw.WriteString("\n"); err != nil {
			return err
		}
	}
}

var _ drivers.FileIterator = &fileIterator{}
