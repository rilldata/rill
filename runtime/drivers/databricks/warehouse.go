package databricks

import (
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
	"github.com/apache/arrow-go/v18/arrow/memory"
	"github.com/apache/arrow-go/v18/parquet"
	"github.com/apache/arrow-go/v18/parquet/compress"
	"github.com/apache/arrow-go/v18/parquet/pqarrow"
	"github.com/c2h5oh/datasize"
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

	var dsn string
	if srcProps.DSN != "" { // get from src properties
		dsn = srcProps.DSN
	} else {
		dsn = c.config.resolveDSN()
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

	ipcStreams, err := rows.(dbsqlrows.Rows).GetArrowIPCStreams(ctx)
	if err != nil {
		return nil, err
	}

	if !ipcStreams.HasNext() {
		rows.Close()
		conn.Close()
		db.Close()
		return nil, drivers.ErrNoRows
	}

	tempDir, err := c.storage.RandomTempDir("databricks")
	if err != nil {
		return nil, err
	}

	return &fileIterator{
		db:         db,
		conn:       conn,
		rows:       rows,
		ipcStreams: ipcStreams,
		logger:     c.logger,
		tempDir:    tempDir,
	}, nil
}

type fileIterator struct {
	db         *sql.DB
	conn       *sql.Conn
	rows       sqld.Rows
	ipcStreams dbsqlrows.ArrowIPCStreamIterator
	logger     *zap.Logger
	tempDir    string

	totalRecords int64
	downloaded   bool
}

var _ drivers.FileIterator = &fileIterator{}

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

		rdr, err := ipc.NewReader(stream, ipc.WithAllocator(allocator), ipc.WithSchema(schema))
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
	DSN string `mapstructure:"dsn"`
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
