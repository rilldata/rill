package bigquery

import (
	"context"
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
	"go.uber.org/zap"
)

// Query implements drivers.SQLStore
func (c *Connection) Query(ctx context.Context, props map[string]any, sql string) (drivers.RowIterator, error) {
	return nil, drivers.ErrNotImplemented
}

// QueryAsFiles implements drivers.SQLStore
func (c *Connection) QueryAsFiles(ctx context.Context, props map[string]any, sql string) (drivers.FileIterator, error) {
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

	return &fileIterator{
		client: client,
		bqIter: it,
		logger: c.logger,
	}, nil
}

type fileIterator struct {
	client       *bigquery.Client
	bqIter       *bigquery.RowIterator
	logger       *zap.Logger
	tempFilePath string
	downloaded   bool
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
func (f *fileIterator) NextBatch(limit int) ([]string, error) {
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
	writer, err := pqarrow.NewFileWriter(rdr.Schema(), fw,
		parquet.NewWriterProperties(
			parquet.WithCompression(compress.Codecs.Snappy),
			parquet.WithRootRepetition(parquet.Repetitions.Required),
			// duckdb has issues reading statistics of string type generated with this write
			// column statistics are not useful if full file need to be ingested so better to disable to save computations
			parquet.WithStats(false),
		),
		pqarrow.NewArrowWriterProperties(pqarrow.WithStoreSchema()))
	if err != nil {
		return nil, err
	}
	defer writer.Close()

	// write arrow records to parquet file
	for rdr.Next() {
		if err := writer.WriteBuffered(rdr.Record()); err != nil {
			return nil, err
		}
	}
	writer.Close()
	fw.Close()
	f.logger.Info("time taken to write arrow records in parquet file", zap.Duration("duration", time.Since(tf)))

	fileInfo, err := os.Stat(fw.Name())
	if err != nil {
		return nil, err
	}
	f.logger.Info("size of file", zap.String("size", datasize.ByteSize(fileInfo.Size()).HumanReadable()))
	return []string{fw.Name()}, nil
}

// Size implements drivers.FileIterator.
func (*fileIterator) Size(unit drivers.ProgressUnit) (int64, bool) {
	panic("unimplemented")
}

var _ drivers.FileIterator = &fileIterator{}
