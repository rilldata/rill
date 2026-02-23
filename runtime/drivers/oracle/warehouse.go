package oracle

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/memory"
	"github.com/apache/arrow-go/v18/parquet"
	"github.com/apache/arrow-go/v18/parquet/compress"
	"github.com/apache/arrow-go/v18/parquet/pqarrow"
	"github.com/c2h5oh/datasize"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/zap"
)

const rowGroupBufferSize = int64(datasize.MB) * 64
const recordBatchSize = 10000

var _ drivers.Warehouse = (*connection)(nil)

// QueryAsFiles implements drivers.Warehouse.
func (c *connection) QueryAsFiles(ctx context.Context, props map[string]any) (drivers.FileIterator, error) {
	srcProps, err := parseWarehouseProps(props)
	if err != nil {
		return nil, err
	}

	db, err := c.getDB(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := db.QueryContext(ctx, srcProps.SQL)
	if err != nil {
		return nil, fmt.Errorf("oracle query failed: %w", err)
	}

	tempDir, err := c.storage.RandomTempDir("oracle")
	if err != nil {
		rows.Close()
		return nil, err
	}

	return &warehouseIterator{
		rows:    rows,
		logger:  c.logger,
		tempDir: tempDir,
	}, nil
}

type warehouseProps struct {
	SQL string `mapstructure:"sql"`
	DSN string `mapstructure:"dsn"`
}

func parseWarehouseProps(props map[string]any) (*warehouseProps, error) {
	conf := &warehouseProps{}
	if err := mapstructure.Decode(props, conf); err != nil {
		return nil, err
	}
	if conf.SQL == "" {
		return nil, fmt.Errorf("property 'sql' is mandatory for connector \"oracle\"")
	}
	return conf, nil
}

type warehouseIterator struct {
	rows       *sql.Rows
	logger     *zap.Logger
	tempDir    string
	downloaded bool
}

var _ drivers.FileIterator = &warehouseIterator{}

func (f *warehouseIterator) Close() error {
	if f.rows != nil {
		f.rows.Close()
	}
	return os.RemoveAll(f.tempDir)
}

func (f *warehouseIterator) Format() string {
	return ""
}

func (f *warehouseIterator) SetKeepFilesUntilClose() {
	// No-op: already keeps files until Close.
}

func (f *warehouseIterator) Next(ctx context.Context) ([]string, error) {
	if f.downloaded {
		return nil, io.EOF
	}
	f.downloaded = true

	defer func() {
		f.rows.Close()
		f.rows = nil
	}()

	f.logger.Debug("downloading Oracle results to parquet file", observability.ZapCtx(ctx))
	tf := time.Now()
	defer func() {
		f.logger.Debug("time taken to write Oracle results to parquet", zap.Duration("duration", time.Since(tf)), observability.ZapCtx(ctx))
	}()

	// Determine schema from column types
	colTypes, err := f.rows.ColumnTypes()
	if err != nil {
		return nil, fmt.Errorf("failed to get column types: %w", err)
	}

	arrowSchema, err := sqlColTypesToArrowSchema(colTypes)
	if err != nil {
		return nil, err
	}

	// Create parquet file
	fw, err := os.CreateTemp(f.tempDir, "temp*.parquet")
	if err != nil {
		return nil, err
	}
	defer fw.Close()

	writer, err := pqarrow.NewFileWriter(arrowSchema, fw,
		parquet.NewWriterProperties(
			parquet.WithCompression(compress.Codecs.Snappy),
			parquet.WithRootRepetition(parquet.Repetitions.Required),
			parquet.WithStats(false),
		),
		pqarrow.NewArrowWriterProperties(pqarrow.WithStoreSchema()))
	if err != nil {
		return nil, err
	}
	defer writer.Close()

	alloc := memory.DefaultAllocator
	totalRows := int64(0)

	// Read rows in batches and write Arrow records to Parquet
	scanDest := makeWarehouseScanDest(colTypes)
	for {
		rec, n, err := readBatch(ctx, f.rows, colTypes, arrowSchema, alloc, scanDest)
		if err != nil {
			return nil, err
		}
		if n == 0 {
			break
		}

		if writer.RowGroupTotalBytesWritten() >= rowGroupBufferSize {
			writer.NewBufferedRowGroup()
		}
		if err := writer.WriteBuffered(rec); err != nil {
			rec.Release()
			return nil, err
		}
		totalRows += rec.NumRows()
		rec.Release()
	}

	if totalRows == 0 {
		return nil, drivers.ErrNoRows
	}

	writer.Close()
	fw.Close()

	fileInfo, err := os.Stat(fw.Name())
	if err != nil {
		return nil, err
	}
	f.logger.Debug("parquet file written",
		zap.String("size", datasize.ByteSize(fileInfo.Size()).HumanReadable()),
		zap.Int64("rows", totalRows),
		observability.ZapCtx(ctx))

	return []string{fw.Name()}, nil
}

// sqlColTypesToArrowSchema converts database/sql column types to an Arrow schema.
func sqlColTypesToArrowSchema(colTypes []*sql.ColumnType) (*arrow.Schema, error) {
	fields := make([]arrow.Field, len(colTypes))
	for i, ct := range colTypes {
		nullable, ok := ct.Nullable()
		if !ok {
			nullable = true
		}
		fields[i] = arrow.Field{
			Name:     ct.Name(),
			Type:     oracleTypeToArrow(ct.DatabaseTypeName()),
			Nullable: nullable,
		}
	}
	return arrow.NewSchema(fields, nil), nil
}

func oracleTypeToArrow(dbType string) arrow.DataType {
	switch dbType {
	case "NUMBER":
		return arrow.PrimitiveTypes.Float64
	case "FLOAT", "BINARY_FLOAT":
		return arrow.PrimitiveTypes.Float32
	case "BINARY_DOUBLE":
		return arrow.PrimitiveTypes.Float64
	case "INTEGER", "INT", "SMALLINT":
		return arrow.PrimitiveTypes.Int64
	case "VARCHAR2", "NVARCHAR2", "CHAR", "NCHAR", "LONG", "CLOB", "NCLOB", "ROWID", "UROWID", "XMLTYPE":
		return arrow.BinaryTypes.String
	case "BLOB", "RAW", "LONG RAW":
		return arrow.BinaryTypes.Binary
	case "DATE", "TIMESTAMP", "TIMESTAMP WITH TIME ZONE", "TIMESTAMP WITH LOCAL TIME ZONE":
		return arrow.FixedWidthTypes.Timestamp_us
	case "BOOLEAN", "BOOL":
		return arrow.FixedWidthTypes.Boolean
	default:
		return arrow.BinaryTypes.String
	}
}

// makeWarehouseScanDest creates reusable scan destinations for each column.
func makeWarehouseScanDest(colTypes []*sql.ColumnType) []any {
	dest := make([]any, len(colTypes))
	for i, ct := range colTypes {
		switch ct.DatabaseTypeName() {
		case "NUMBER", "FLOAT", "BINARY_FLOAT", "BINARY_DOUBLE":
			dest[i] = &sql.NullFloat64{}
		case "INTEGER", "INT", "SMALLINT":
			dest[i] = &sql.NullInt64{}
		case "BOOLEAN", "BOOL":
			dest[i] = &sql.NullBool{}
		case "DATE", "TIMESTAMP", "TIMESTAMP WITH TIME ZONE", "TIMESTAMP WITH LOCAL TIME ZONE":
			dest[i] = &sql.NullTime{}
		case "BLOB", "RAW", "LONG RAW":
			dest[i] = new(any)
		default:
			dest[i] = &sql.NullString{}
		}
	}
	return dest
}

// readBatch reads up to recordBatchSize rows and returns an Arrow record.
func readBatch(ctx context.Context, rows *sql.Rows, colTypes []*sql.ColumnType, schema *arrow.Schema, alloc memory.Allocator, scanDest []any) (arrow.Record, int, error) {
	builder := array.NewRecordBuilder(alloc, schema)
	defer builder.Release()

	n := 0
	for n < recordBatchSize && rows.Next() {
		select {
		case <-ctx.Done():
			return nil, 0, ctx.Err()
		default:
		}

		if err := rows.Scan(scanDest...); err != nil {
			return nil, 0, fmt.Errorf("failed to scan row: %w", err)
		}

		for i, ct := range colTypes {
			appendValue(builder.Field(i), ct.DatabaseTypeName(), scanDest[i])
		}
		n++
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	if n == 0 {
		return nil, 0, nil
	}

	rec := builder.NewRecord()
	return rec, n, nil
}

// appendValue appends a scanned value to the appropriate Arrow array builder.
func appendValue(fb array.Builder, dbType string, val any) {
	switch dbType {
	case "NUMBER", "FLOAT", "BINARY_FLOAT", "BINARY_DOUBLE":
		b := fb.(*array.Float64Builder)
		v := val.(*sql.NullFloat64)
		if v.Valid {
			b.Append(v.Float64)
		} else {
			b.AppendNull()
		}
	case "INTEGER", "INT", "SMALLINT":
		b := fb.(*array.Int64Builder)
		v := val.(*sql.NullInt64)
		if v.Valid {
			b.Append(v.Int64)
		} else {
			b.AppendNull()
		}
	case "BOOLEAN", "BOOL":
		b := fb.(*array.BooleanBuilder)
		v := val.(*sql.NullBool)
		if v.Valid {
			b.Append(v.Bool)
		} else {
			b.AppendNull()
		}
	case "DATE", "TIMESTAMP", "TIMESTAMP WITH TIME ZONE", "TIMESTAMP WITH LOCAL TIME ZONE":
		b := fb.(*array.TimestampBuilder)
		v := val.(*sql.NullTime)
		if v.Valid {
			b.Append(arrow.Timestamp(v.Time.UnixMicro()))
		} else {
			b.AppendNull()
		}
	case "BLOB", "RAW", "LONG RAW":
		b := fb.(*array.BinaryBuilder)
		v := val.(*any)
		if *v != nil {
			if bs, ok := (*v).([]byte); ok {
				b.Append(bs)
			} else {
				b.AppendNull()
			}
		} else {
			b.AppendNull()
		}
	default:
		b := fb.(*array.StringBuilder)
		v := val.(*sql.NullString)
		if v.Valid {
			b.Append(v.String)
		} else {
			b.AppendNull()
		}
	}
}
