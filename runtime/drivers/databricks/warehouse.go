package databricks

import (
	"bytes"
	"context"
	"database/sql"
	sqld "database/sql/driver"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/ipc"
	"github.com/apache/arrow-go/v18/parquet"
	"github.com/apache/arrow-go/v18/parquet/compress"
	"github.com/apache/arrow-go/v18/parquet/pqarrow"
	// v12 Arrow IPC writer (kernel exports v12 records); aliased to avoid clashing
	// with the v18 arrow/ipc import above.
	dbipc "github.com/apache/arrow/go/v12/arrow/ipc"
	"github.com/c2h5oh/datasize"
	dbsqlerr "github.com/databricks/databricks-sql-go/errors"
	dbsqlrows "github.com/databricks/databricks-sql-go/rows"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

var tracer = otel.Tracer("github.com/rilldata/rill/runtime/drivers/databricks")

// entire data is buffered in memory before its written to disk so keeping it small reduces memory usage
// but keeping it too small can lead to bad ingestion performance
// 64MB seems to be a good balance
const rowGroupBufferSize = int64(datasize.MB) * 64

// QueryAsFiles implements drivers.Warehouse.
// Fetches query result as Arrow IPC streams, converts them to Parquet.
func (c *connection) QueryAsFiles(ctx context.Context, props map[string]any) (outIt drivers.FileIterator, outErr error) {
	ctx, span := tracer.Start(ctx, "Connection.QueryAsFiles")
	defer func() {
		if outErr != nil {
			span.SetStatus(codes.Error, outErr.Error())
		}
		span.End()
	}()

	srcProps, err := parseSourceProperties(props)
	if err != nil {
		return nil, err
	}

	// effectiveDSN auto-detects Lakehouse//RT and selects the SEA backend if needed,
	// shared with the OLAP path so ingest works without configuration. On a
	// non-definitive result, defer like getDB (surface the probe error / retry) rather
	// than ingest over a possibly-wrong backend.
	dsn, definitive, probeErr := c.effectiveDSN(ctx)
	if !definitive {
		if probeErr != nil {
			return nil, probeErr
		}
		return nil, errors.New("databricks: could not determine warehouse protocol")
	}
	db, err := sql.Open("databricks", dsn)
	if err != nil {
		return nil, err
	}
	defer func() {
		if outErr != nil {
			db.Close()
		}
	}()

	conn, err := db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if outErr != nil {
			conn.Close()
		}
	}()

	var rows sqld.Rows
	err = conn.Raw(func(x any) error {
		rows, err = x.(sqld.QueryerContext).QueryContext(ctx, srcProps.SQL, nil)
		return err
	})
	if err != nil {
		return nil, err
	}
	defer func() {
		if outErr != nil {
			rows.Close()
		}
	}()

	// Thrift/DBSQL uses native IPC streams. The SEA backend (Lakehouse//RT) doesn't
	// implement them and returns ErrNotSupportedByKernel, so fall back to
	// GetArrowBatches via kernelIPCStreams; the parquet path is identical for both.
	dr := rows.(dbsqlrows.Rows)
	selfDescribing := false // whether each stream carries its own schema (kernel adapter)
	ipcStreams, err := dr.GetArrowIPCStreams(ctx)
	if errors.Is(err, dbsqlerr.ErrNotSupportedByKernel) {
		var batches dbsqlrows.ArrowBatchIterator
		batches, err = dr.GetArrowBatches(ctx)
		if err == nil {
			ipcStreams = &kernelIPCStreams{batches: batches}
			selfDescribing = true
		}
	}
	if err != nil {
		return nil, err
	}
	defer func() {
		if outErr != nil {
			ipcStreams.Close()
		}
	}()

	if !ipcStreams.HasNext() {
		return nil, drivers.ErrNoRows
	}

	tempDir, err := c.storage.RandomTempDir("databricks")
	if err != nil {
		return nil, err
	}

	return &fileIterator{
		db:             db,
		conn:           conn,
		rows:           rows,
		ipcStreams:     ipcStreams,
		selfDescribing: selfDescribing,
		logger:         c.logger,
		tempDir:        tempDir,
	}, nil
}

type fileIterator struct {
	db         *sql.DB
	conn       *sql.Conn
	rows       sqld.Rows
	ipcStreams dbsqlrows.ArrowIPCStreamIterator
	// selfDescribing is true when each stream carries its own schema message (the
	// kernel adapter). For native Thrift streams it's false, and subsequent streams
	// are read with ipc.WithSchema to validate cross-stream schema consistency.
	selfDescribing bool
	logger         *zap.Logger
	tempDir        string

	totalRecords int64
	downloaded   bool
}

var _ drivers.FileIterator = &fileIterator{}

// kernelIPCStreams adapts the SEA backend's ArrowBatchIterator to the
// ArrowIPCStreamIterator this file consumes, re-serializing each batch to a
// self-contained IPC stream with the driver's v12 writer.
type kernelIPCStreams struct {
	batches dbsqlrows.ArrowBatchIterator
}

var _ dbsqlrows.ArrowIPCStreamIterator = &kernelIPCStreams{}

func (k *kernelIPCStreams) HasNext() bool { return k.batches.HasNext() }

func (k *kernelIPCStreams) Next() (io.Reader, error) {
	rec, err := k.batches.Next()
	if err != nil {
		return nil, err // propagates io.EOF
	}
	defer rec.Release()

	var buf bytes.Buffer
	w := dbipc.NewWriter(&buf, dbipc.WithSchema(rec.Schema()))
	if err := w.Write(rec); err != nil {
		_ = w.Close()
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return &buf, nil
}

func (k *kernelIPCStreams) Close() { k.batches.Close() }

func (k *kernelIPCStreams) SchemaBytes() ([]byte, error) {
	sc, err := k.batches.Schema()
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	w := dbipc.NewWriter(&buf, dbipc.WithSchema(sc))
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Close implements drivers.FileIterator.
func (f *fileIterator) Close() error {
	if f.ipcStreams != nil {
		f.ipcStreams.Close()
	}
	if f.rows != nil {
		f.rows.Close()
		f.conn.Close()
		f.db.Close()
	}
	return os.RemoveAll(f.tempDir)
}

// Format implements drivers.FileIterator.
func (f *fileIterator) Format() string {
	return "parquet"
}

// SetKeepFilesUntilClose implements drivers.FileIterator.
func (f *fileIterator) SetKeepFilesUntilClose() {
	// No-op because it already does this.
}

// Next implements drivers.FileIterator.
// Query result is written to a single parquet file.
func (f *fileIterator) Next(ctx context.Context) ([]string, error) {
	if f.downloaded {
		return nil, io.EOF
	}

	ctx, span := tracer.Start(ctx, "fileIterator.Next")
	defer span.End()

	// Close db resources early
	defer func() {
		f.ipcStreams.Close()
		f.rows.Close()
		f.conn.Close()
		f.db.Close()
		// Mark rows as nil to prevent double close
		f.rows = nil
		f.ipcStreams = nil
	}()

	f.logger.Debug("downloading results in parquet file", observability.ZapCtx(ctx))

	fw, err := os.CreateTemp(f.tempDir, "temp*.parquet")
	if err != nil {
		return nil, err
	}
	defer fw.Close()
	f.downloaded = true

	tf := time.Now()
	defer func() {
		f.logger.Debug("time taken to write arrow records in parquet file", zap.Duration("duration", time.Since(tf)), observability.ZapCtx(ctx))
	}()

	// Read the first IPC stream to get the schema and initial records
	firstStream, err := f.ipcStreams.Next()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, drivers.ErrNoRows
		}
		return nil, err
	}

	rdr, err := ipc.NewReader(firstStream)
	if err != nil {
		return nil, err
	}

	schema := rdr.Schema()
	for _, field := range schema.Fields() {
		if field.Type.ID() == arrow.TIME32 || field.Type.ID() == arrow.TIME64 {
			rdr.Release()
			return nil, fmt.Errorf("TIME data type (column %q) is not currently supported, "+
				"consider excluding it or casting it to another data type", field.Name)
		}
	}

	writer, err := pqarrow.NewFileWriter(schema, fw,
		parquet.NewWriterProperties(
			parquet.WithCompression(compress.Codecs.Snappy),
			parquet.WithRootRepetition(parquet.Repetitions.Required),
			// duckdb has issues reading statistics of string type generated with this write;
			// column statistics may not be useful if full file needs to be ingested so better to disable to save computations
			parquet.WithStats(false),
		),
		pqarrow.NewArrowWriterProperties(pqarrow.WithStoreSchema()))
	if err != nil {
		rdr.Release()
		return nil, err
	}
	defer writer.Close()

	// Write records from the first stream
	if err := f.writeRecords(ctx, rdr, writer); err != nil {
		rdr.Release()
		return nil, err
	}
	rdr.Release()

	// Process remaining IPC streams
	for f.ipcStreams.HasNext() {
		stream, err := f.ipcStreams.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}

		// Native (Thrift) subsequent streams are validated against the first stream's
		// schema; the kernel adapter's streams are self-describing, so skip WithSchema
		// there (each carries its own schema message).
		var opts []ipc.Option
		if !f.selfDescribing {
			opts = append(opts, ipc.WithSchema(schema))
		}
		rdr, err := ipc.NewReader(stream, opts...)
		if err != nil {
			return nil, err
		}

		if err := f.writeRecords(ctx, rdr, writer); err != nil {
			rdr.Release()
			return nil, err
		}
		rdr.Release()
	}

	writer.Close()
	fw.Close()

	fileInfo, err := os.Stat(fw.Name())
	if err != nil {
		return nil, err
	}
	f.logger.Debug("size of file", zap.String("size", datasize.ByteSize(fileInfo.Size()).HumanReadable()), zap.Int64("total_records", f.totalRecords), observability.ZapCtx(ctx))
	return []string{fw.Name()}, nil
}

// writeRecords reads all records from the IPC reader and writes them to the parquet writer.
func (f *fileIterator) writeRecords(ctx context.Context, rdr *ipc.Reader, writer *pqarrow.FileWriter) error {
	for rdr.Next() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		rec := rdr.RecordBatch()
		if writer.RowGroupTotalBytesWritten() >= rowGroupBufferSize {
			writer.NewBufferedRowGroup()
		}
		if err := writer.WriteBuffered(rec); err != nil {
			return err
		}
		f.totalRecords += rec.NumRows()
	}
	return rdr.Err()
}

type sourceProperties struct {
	SQL string `mapstructure:"sql"`
}

func parseSourceProperties(props map[string]any) (*sourceProperties, error) {
	conf := &sourceProperties{}
	err := mapstructure.Decode(props, conf)
	if err != nil {
		return nil, err
	}
	if conf.SQL == "" {
		return nil, fmt.Errorf("property 'sql' is mandatory for connector \"databricks\"")
	}
	return conf, err
}
