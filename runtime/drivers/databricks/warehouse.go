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

	"github.com/apache/arrow-go/v18/arrow/ipc"
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

const rowGroupBufferSize = int64(datasize.MB) * 64

// QueryAsFiles implements drivers.Warehouse.
// Executes SQL against Databricks, retrieves results as Arrow IPC streams, and writes them to a Parquet file.
// The IPC stream approach bridges the Databricks driver's arrow/go/v12 with Rill's arrow-go/v18.
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

	db, err := c.openRawDB()
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
	err = conn.Raw(func(x interface{}) error {
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

	// Use the ArrowIPCStreamIterator to get IPC byte streams.
	// The IPC wire format is cross-version compatible, so v18's ipc.NewReader
	// can read streams produced by the Databricks driver's v12 Arrow library.
	streams, err := rows.(dbsqlrows.Rows).GetArrowIPCStreams(ctx)
	if err != nil {
		return nil, fmt.Errorf("databricks: failed to get arrow IPC streams: %w", err)
	}

	if !streams.HasNext() {
		streams.Close()
		return nil, drivers.ErrNoRows
	}

	tempDir, err := c.storage.RandomTempDir("databricks")
	if err != nil {
		return nil, err
	}

	return &fileIterator{
		db:      db,
		conn:    conn,
		rows:    rows,
		streams: streams,
		logger:  c.logger,
		tempDir: tempDir,
	}, nil
}

type fileIterator struct {
	db      *sql.DB
	conn    *sql.Conn
	rows    sqld.Rows
	streams dbsqlrows.ArrowIPCStreamIterator
	logger  *zap.Logger
	tempDir string

	downloaded bool
}

var _ drivers.FileIterator = &fileIterator{}

// Close implements drivers.FileIterator.
func (f *fileIterator) Close() error {
	if f.streams != nil {
		f.streams.Close()
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
	return ""
}

// SetKeepFilesUntilClose implements drivers.FileIterator.
func (f *fileIterator) SetKeepFilesUntilClose() {
	// No-op: files are already kept until Close.
}

// Next implements drivers.FileIterator.
// Reads all Arrow IPC streams and writes them to a single Parquet file.
func (f *fileIterator) Next(ctx context.Context) ([]string, error) {
	if f.downloaded {
		return nil, io.EOF
	}

	ctx, span := tracer.Start(ctx, "fileIterator.Next")
	defer span.End()

	// Release DB resources after reading all data.
	defer func() {
		f.streams.Close()
		f.streams = nil
		f.rows.Close()
		f.conn.Close()
		f.db.Close()
		f.rows = nil
	}()

	f.logger.Debug("downloading databricks results to parquet file", observability.ZapCtx(ctx))

	fw, err := os.CreateTemp(f.tempDir, "temp*.parquet")
	if err != nil {
		return nil, err
	}
	defer fw.Close()
	f.downloaded = true

	tf := time.Now()
	defer func() {
		f.logger.Debug("time taken to write arrow records to parquet file",
			zap.Duration("duration", time.Since(tf)), observability.ZapCtx(ctx))
	}()

	var writer *pqarrow.FileWriter
	totalRows := int64(0)

	for f.streams.HasNext() {
		reader, err := f.streams.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("databricks: failed to get next IPC stream: %w", err)
		}

		// Read the IPC stream using v18's ipc.NewReader.
		ipcReader, err := ipc.NewReader(reader)
		if err != nil {
			return nil, fmt.Errorf("databricks: failed to create IPC reader: %w", err)
		}

		// Initialize the Parquet writer on the first stream (which provides the schema).
		if writer == nil {
			schema := ipcReader.Schema()
			writer, err = pqarrow.NewFileWriter(schema, fw,
				parquet.NewWriterProperties(
					parquet.WithCompression(compress.Codecs.Snappy),
					parquet.WithRootRepetition(parquet.Repetitions.Required),
					parquet.WithStats(false),
				),
				pqarrow.NewArrowWriterProperties(pqarrow.WithStoreSchema()))
			if err != nil {
				ipcReader.Release()
				return nil, err
			}
		}

		for ipcReader.Next() {
			rec := ipcReader.RecordBatch()
			if writer.RowGroupTotalBytesWritten() >= rowGroupBufferSize {
				writer.NewBufferedRowGroup()
			}
			if err := writer.WriteBuffered(rec); err != nil {
				ipcReader.Release()
				return nil, err
			}
			totalRows += rec.NumRows()
		}

		if ipcReader.Err() != nil {
			ipcReader.Release()
			return nil, fmt.Errorf("databricks: error reading IPC stream: %w", ipcReader.Err())
		}
		ipcReader.Release()
	}

	if writer == nil {
		return nil, drivers.ErrNoRows
	}

	writer.Close()
	fw.Close()

	fileInfo, err := os.Stat(fw.Name())
	if err != nil {
		return nil, err
	}

	f.logger.Debug("databricks parquet file written",
		zap.String("size", datasize.ByteSize(fileInfo.Size()).HumanReadable()),
		zap.Int64("rows", totalRows),
		observability.ZapCtx(ctx))

	return []string{fw.Name()}, nil
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
