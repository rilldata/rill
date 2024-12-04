package bigquery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"os"
	"regexp"
	"strings"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/apache/arrow/go/v14/parquet"
	"github.com/apache/arrow/go/v14/parquet/compress"
	"github.com/apache/arrow/go/v14/parquet/pqarrow"
	"github.com/c2h5oh/datasize"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/zap"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// recommended size is 512MB - 1GB, entire data is buffered in memory before its written to disk
const rowGroupBufferSize = int64(datasize.MB) * 512

const _jsonDownloadLimitBytes = 100 * int64(datasize.MB)

// Regex to parse BigQuery SELECT ALL statement: SELECT * FROM `project_id.dataset.table`
var selectQueryRegex = regexp.MustCompile("(?i)^\\s*SELECT\\s+\\*\\s+FROM\\s+(`?[a-zA-Z0-9_.-]+`?)\\s*$")

var _ drivers.Warehouse = &Connection{}

// QueryAsFiles implements drivers.SQLStore
func (c *Connection) QueryAsFiles(ctx context.Context, props map[string]any) (drivers.FileIterator, error) {
	srcProps, err := parseSourceProperties(props)
	if err != nil {
		return nil, err
	}

	opts, err := c.clientOption(ctx)
	if err != nil {
		return nil, err
	}

	var client *bigquery.Client
	var it *bigquery.RowIterator
	var fallbackToQueryExecution bool

	match := selectQueryRegex.FindStringSubmatch(srcProps.SQL)
	queryIsSelectAll := match != nil
	if queryIsSelectAll {
		// "SELECT * FROM `project_id.dataset.table`" statement so storage api might be used
		// project_id and backticks are optional
		fullTableName := match[1]
		fullTableName = strings.Trim(fullTableName, "`")

		var projectID, dataset, tableID string

		parts := strings.Split(fullTableName, ".")
		switch len(parts) {
		case 2:
			dataset, tableID = parts[0], parts[1]
			projectID = srcProps.ProjectID
		case 3:
			projectID, dataset, tableID = parts[0], parts[1], parts[2]
		default:
			return nil, fmt.Errorf("invalid table format, `project_id.dataset.table` is expected")
		}

		client, err = createClient(ctx, srcProps.ProjectID, opts)
		if err != nil {
			return nil, err
		}

		if err = client.EnableStorageReadClient(ctx, opts...); err != nil {
			client.Close()
			return nil, err
		}

		table := client.DatasetInProject(projectID, dataset).Table(tableID)
		// extract source metadata to ensure the source is a regular table or a snapshot
		// as storage api doesn't support other types
		metadata, err := table.Metadata(ctx)
		if err != nil {
			client.Close()
			return nil, fmt.Errorf("source metadata cannot be extracted: %w", err)
		}
		if metadata.Type == bigquery.RegularTable || metadata.Type == bigquery.Snapshot {
			it = table.Read(ctx)
		} else {
			c.logger.Debug("source is not a regular table or a snapshot, falling back to a query execution")
			fallbackToQueryExecution = true
			client.Close()
		}
	}

	if !queryIsSelectAll || fallbackToQueryExecution {
		// storage api cannot be used, switching to a query execution
		now := time.Now()

		client, err = createClient(ctx, srcProps.ProjectID, opts)
		if err != nil {
			return nil, err
		}

		if err := client.EnableStorageReadClient(ctx, opts...); err != nil {
			client.Close()
			return nil, err
		}

		q := client.Query(srcProps.SQL)
		it, err = q.Read(ctx)

		if err != nil && strings.Contains(err.Error(), "Response too large to return") {
			// https://cloud.google.com/knowledge/kb/bigquery-response-too-large-to-return-consider-setting-allowlargeresults-to-true-in-your-job-configuration-000004266
			client.Close()
			return nil, fmt.Errorf("response too large, consider converting the source to a table and " +
				"ingesting the entire table with 'select * from `project_id.dataset.tablename`'")
		}

		if err != nil && !strings.Contains(err.Error(), "Syntax error") {
			// close the read storage API client
			client.Close()
			c.logger.Debug("query failed, retrying without storage api", zap.Error(err))
			// the query results are always cached in a temporary table that storage api can use
			// there are some exceptions when results aren't cached
			// so we also try without storage api
			client, err = createClient(ctx, srcProps.ProjectID, opts)
			if err != nil {
				return nil, err
			}

			q := client.Query(srcProps.SQL)
			it, err = q.Read(ctx)
		}

		if err != nil {
			client.Close()
			return nil, err
		}

		c.logger.Debug("query took", zap.Duration("duration", time.Since(now)), observability.ZapCtx(ctx))
	}

	tempDir, err := c.storage.RandomTempDir("bigquery-*")
	if err != nil {
		return nil, err
	}
	return &fileIterator{
		client:       client,
		bqIter:       it,
		logger:       c.logger,
		totalRecords: int64(it.TotalRows),
		ctx:          ctx,
		tempDir:      tempDir,
	}, nil
}

func createClient(ctx context.Context, projectID string, opts []option.ClientOption) (*bigquery.Client, error) {
	client, err := bigquery.NewClient(ctx, projectID, opts...)
	if err != nil {
		if strings.Contains(err.Error(), "unable to detect projectID") {
			return nil, fmt.Errorf("projectID not detected in credentials. Please set `project_id` in source yaml")
		}
		return nil, fmt.Errorf("failed to create bigquery client: %w", err)
	}
	return client, nil
}

type fileIterator struct {
	client  *bigquery.Client
	bqIter  *bigquery.RowIterator
	logger  *zap.Logger
	tempDir string

	totalRecords int64
	downloaded   bool

	ctx context.Context // TODO :: refatcor NextBatch to take context on NextBatch
}

// Close implements drivers.FileIterator.
func (f *fileIterator) Close() error {
	return os.RemoveAll(f.tempDir)
}

// Next implements drivers.FileIterator.
// TODO :: currently it downloads all records in a single file. Need to check if it is efficient to ingest a single file with size in tens of GBs or more.
func (f *fileIterator) Next() ([]string, error) {
	if f.downloaded {
		return nil, io.EOF
	}
	// storage API not available so can't read as arrow records. Read results row by row and dump in a json file.
	if !f.bqIter.IsAccelerated() {
		f.logger.Debug("downloading results in json file", observability.ZapCtx(f.ctx))
		file, err := f.downloadAsJSONFile()
		if err != nil {
			return nil, err
		}
		return []string{file}, nil
	}
	f.logger.Debug("downloading results in parquet file", observability.ZapCtx(f.ctx))

	// create a temp file
	fw, err := os.CreateTemp(f.tempDir, "temp*.parquet")
	if err != nil {
		return nil, err
	}
	defer fw.Close()
	f.downloaded = true

	rdr, err := f.AsArrowRecordReader()
	if err != nil {
		return nil, err
	}
	defer rdr.Release()

	tf := time.Now()
	defer func() {
		f.logger.Debug("time taken to write arrow records in parquet file", zap.Duration("duration", time.Since(tf)), observability.ZapCtx(f.ctx))
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

	rows := int64(0)
	// write arrow records to parquet file
	for rdr.Next() {
		select {
		case <-f.ctx.Done():
			return nil, f.ctx.Err()
		default:
			rec := rdr.Record()
			if writer.RowGroupTotalBytesWritten() >= rowGroupBufferSize {
				writer.NewBufferedRowGroup()
			}
			if err := writer.WriteBuffered(rec); err != nil {
				return nil, err
			}
			rows += rec.NumRows()
		}
	}
	if rdr.Err() != nil {
		return nil, fmt.Errorf("file write failed with error: %w", rdr.Err())
	}
	writer.Close()
	fw.Close()

	if uint64(rows) < f.bqIter.TotalRows {
		f.logger.Error("not all rows written to parquet file", zap.Int64("rows_written", rows), zap.Uint64("total_rows", f.bqIter.TotalRows), observability.ZapCtx(f.ctx))
	}

	fileInfo, err := os.Stat(fw.Name())
	if err != nil {
		return nil, err
	}
	f.logger.Info("parquet file written", zap.String("size", datasize.ByteSize(fileInfo.Size()).HumanReadable()), zap.Int64("rows", rows), observability.ZapCtx(f.ctx))
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

func (f *fileIterator) Format() string {
	return ""
}

func (f *fileIterator) downloadAsJSONFile() (string, error) {
	tf := time.Now()
	defer func() {
		f.logger.Debug("time taken to write row in json file", zap.Duration("duration", time.Since(tf)), observability.ZapCtx(f.ctx))
	}()

	// create a temp file
	fw, err := os.CreateTemp(f.tempDir, "temp*.ndjson")
	if err != nil {
		return "", err
	}
	defer fw.Close()
	f.downloaded = true

	init := false
	rows := int64(0)
	enc := json.NewEncoder(fw)
	enc.SetEscapeHTML(false)
	bigNumericFields := make([]string, 0)
	for {
		row := make(map[string]bigquery.Value)
		err := f.bqIter.Next(&row)
		if err != nil {
			if !errors.Is(err, iterator.Done) {
				return "", err
			}
			if !init {
				return "", drivers.ErrNoRows
			}
			fileInfo, err := os.Stat(fw.Name())
			if err != nil {
				return "", fmt.Errorf("bigquery: failed to poll json file size: %w", err)
			}

			if uint64(rows) < f.bqIter.TotalRows {
				f.logger.Error("not all rows written to json file", zap.Int64("rows_written", rows), zap.Uint64("total_rows", f.bqIter.TotalRows), observability.ZapCtx(f.ctx))
			}

			f.logger.Info("json file written", zap.String("size", datasize.ByteSize(fileInfo.Size()).HumanReadable()), zap.Int64("rows", rows), observability.ZapCtx(f.ctx))
			// all rows written successfully
			return fw.Name(), nil
		}

		// schema and total rows is available after first call to next only
		if !init {
			init = true
			for _, f := range f.bqIter.Schema {
				if f.Type == bigquery.BigNumericFieldType {
					bigNumericFields = append(bigNumericFields, f.Name)
				}
			}
		}

		// convert fields into a.b else fields are marshalled as a/b
		for _, f := range bigNumericFields {
			r, ok := row[f].(*big.Rat)
			if !ok {
				continue
			}
			num, exact := r.Float64()
			if exact {
				row[f] = num
			} else { // number doesn't fit in float so cast to string,
				row[f] = r.FloatString(38)
			}
		}

		err = enc.Encode(row)
		if err != nil {
			return "", fmt.Errorf("conversion of row to json failed with error: %w", err)
		}

		// If we don't have storage API access, BigQuery may return massive JSON results. (But even with storage API access, it may return JSON for small results.)
		// We want to avoid JSON for massive results. Currently, the only way to do so is to error at a limit.
		rows++
		if rows != 0 && rows%10000 == 0 { // Check file size every 10k rows
			fileInfo, err := os.Stat(fw.Name())
			if err != nil {
				return "", fmt.Errorf("bigquery: failed to poll json file size: %w", err)
			}
			if fileInfo.Size() >= _jsonDownloadLimitBytes {
				return "", fmt.Errorf("bigquery: json download exceeded limit of %d bytes (enable and provide access to the BigQuery Storage Read API to read larger results)", _jsonDownloadLimitBytes)
			}
		}
	}
}

var _ drivers.FileIterator = &fileIterator{}
