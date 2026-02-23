package starrocks

import (
	"context"
	"fmt"
	"time"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/flight"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/sqlconvert"
)

// queryFlightSQL executes a query using Arrow Flight SQL and returns a drivers.Result.
// If stmt.Args is non-empty, falls back to MySQL because Flight SQL does not support
// parameterized queries. Concurrency is limited by flightSem to prevent exhausting
// StarRocks FE's per-user connection limit.
func (c *connection) queryFlightSQL(ctx context.Context, stmt *drivers.Statement) (*drivers.Result, error) {
	// Flight SQL does not support parameterized queries; fall back to MySQL.
	if len(stmt.Args) > 0 {
		return c.queryMySQL(ctx, stmt)
	}

	// Acquire semaphore to limit concurrent Flight SQL queries
	if err := c.flightSem.Acquire(ctx, 1); err != nil {
		return nil, err
	}
	// NOTE: We cannot simply "defer c.flightSem.Release(1)" because the rows
	// are consumed after this function returns. Release happens in Result.Close()
	// via SetCleanupFunc, following the same pattern as DuckDB.

	info, err := c.flightClient.Execute(ctx, stmt.Query)
	if err != nil {
		c.flightSem.Release(1)
		return nil, fmt.Errorf("flight sql execute: %w", err)
	}

	if len(info.Endpoint) == 0 {
		c.flightSem.Release(1)
		return &drivers.Result{
			Schema: &runtimev1.StructType{Fields: []*runtimev1.StructType_Field{}},
			Rows:   &emptyRows{},
		}, nil
	}

	// Route DoGet to the correct node (FE or BE) based on endpoint Location.
	// StarRocks returns data from BE nodes; the FE client cannot serve DoGet.
	reader, err := c.doGetFromEndpoint(ctx, info.Endpoint[0])
	if err != nil {
		c.flightSem.Release(1)
		return nil, err
	}

	arrowSchema := reader.Schema()
	schema, err := arrowSchemaToRuntimeSchema(arrowSchema)
	if err != nil {
		reader.Release()
		c.flightSem.Release(1)
		return nil, fmt.Errorf("flight sql schema conversion: %w", err)
	}

	rows := &flightRows{
		reader:      reader,
		arrowSchema: arrowSchema,
	}

	res := &drivers.Result{
		Schema: schema,
		Rows:   rows,
	}
	res.SetCleanupFunc(func() error {
		c.flightSem.Release(1)
		return nil
	})

	return res, nil
}

// querySchemaFlightSQL returns the schema of a query using Arrow Flight SQL.
// If args is non-empty, falls back to MySQL because Flight SQL does not support
// parameterized queries.
func (c *connection) querySchemaFlightSQL(ctx context.Context, query string, args []any) (*runtimev1.StructType, error) {
	// Flight SQL does not support parameterized queries; fall back to MySQL.
	if len(args) > 0 {
		return c.querySchemaMySQL(ctx, query, args)
	}

	// Acquire semaphore to limit concurrent Flight SQL queries
	if err := c.flightSem.Acquire(ctx, 1); err != nil {
		return nil, err
	}
	defer c.flightSem.Release(1)

	schemaQuery := fmt.Sprintf("SELECT * FROM (%s) AS _schema_query LIMIT 0", query)

	info, err := c.flightClient.Execute(ctx, schemaQuery)
	if err != nil {
		return nil, fmt.Errorf("flight sql schema query: %w", err)
	}

	if len(info.Endpoint) == 0 {
		return &runtimev1.StructType{Fields: []*runtimev1.StructType_Field{}}, nil
	}

	reader, err := c.doGetFromEndpoint(ctx, info.Endpoint[0])
	if err != nil {
		return nil, err
	}
	defer reader.Release()

	return arrowSchemaToRuntimeSchema(reader.Schema())
}

// flightRows implements the Rows interface for Arrow Flight SQL results.
type flightRows struct {
	reader       *flight.Reader
	arrowSchema  *arrow.Schema
	currentBatch arrow.Record
	batchIdx     int64 // current row index within the current batch
	err          error
	closed       bool
}

// Next advances to the next row. Returns false when no more rows available.
func (r *flightRows) Next() bool {
	if r.closed || r.err != nil {
		return false
	}

	// Try to advance within the current batch
	if r.currentBatch != nil {
		r.batchIdx++
		if r.batchIdx < r.currentBatch.NumRows() {
			return true
		}
		// Current batch exhausted, release it
		r.currentBatch.Release()
		r.currentBatch = nil
	}

	// Read next batch from the stream
	for r.reader.Next() {
		rec := r.reader.Record()
		if rec.NumRows() == 0 {
			continue
		}
		rec.Retain()
		r.currentBatch = rec
		r.batchIdx = 0
		return true
	}

	r.err = r.reader.Err()
	return false
}

// MapScan scans the current row into a map.
func (r *flightRows) MapScan(dest map[string]any) error {
	if r.currentBatch == nil {
		return fmt.Errorf("no current row")
	}

	for i := 0; i < int(r.currentBatch.NumCols()); i++ {
		col := r.currentBatch.Column(i)
		fieldName := r.arrowSchema.Field(i).Name

		if col.IsNull(int(r.batchIdx)) {
			dest[fieldName] = nil
			continue
		}

		dest[fieldName] = extractArrowValue(col, int(r.batchIdx), r.arrowSchema.Field(i).Type)
	}

	return nil
}

// Scan scans the current row values into dest.
func (r *flightRows) Scan(dest ...any) error {
	if r.currentBatch == nil {
		return fmt.Errorf("no current row")
	}
	if len(dest) != int(r.currentBatch.NumCols()) {
		return fmt.Errorf("expected %d columns, got %d scan targets", r.currentBatch.NumCols(), len(dest))
	}

	for i := 0; i < int(r.currentBatch.NumCols()); i++ {
		col := r.currentBatch.Column(i)
		var val any
		if !col.IsNull(int(r.batchIdx)) {
			val = extractArrowValue(col, int(r.batchIdx), r.arrowSchema.Field(i).Type)
		}
		if err := sqlconvert.ConvertAssign(dest[i], val); err != nil {
			return fmt.Errorf("scan column %d (%s): %w", i, r.arrowSchema.Field(i).Name, err)
		}
	}

	return nil
}

// Close releases all resources.
func (r *flightRows) Close() error {
	if r.closed {
		return nil
	}
	r.closed = true
	if r.currentBatch != nil {
		r.currentBatch.Release()
		r.currentBatch = nil
	}
	if r.reader != nil {
		r.reader.Release()
	}
	return nil
}

// Err returns any error encountered during iteration.
func (r *flightRows) Err() error {
	return r.err
}

// emptyRows is a Rows implementation for empty result sets.
type emptyRows struct{}

func (r *emptyRows) Next() bool                      { return false }
func (r *emptyRows) MapScan(dest map[string]any) error { return fmt.Errorf("no rows") }
func (r *emptyRows) Scan(dest ...any) error           { return fmt.Errorf("no rows") }
func (r *emptyRows) Close() error                     { return nil }
func (r *emptyRows) Err() error                       { return nil }

// extractArrowValue extracts a Go value from an Arrow column at the given row index.
// The returned types match those produced by the MySQL protocol path for consistency.
func extractArrowValue(col arrow.Array, idx int, dt arrow.DataType) any {
	switch col := col.(type) {
	case *array.Boolean:
		return col.Value(idx)

	case *array.Int8:
		return int16(col.Value(idx)) // Match sql.NullInt16

	case *array.Int16:
		return col.Value(idx)

	case *array.Int32:
		return col.Value(idx)

	case *array.Int64:
		return col.Value(idx)

	case *array.Uint8:
		return int16(col.Value(idx))

	case *array.Uint16:
		return int32(col.Value(idx))

	case *array.Uint32:
		return int64(col.Value(idx))

	case *array.Uint64:
		return int64(col.Value(idx))

	case *array.Float32:
		return float64(col.Value(idx)) // Match sql.NullFloat64

	case *array.Float64:
		return col.Value(idx)

	case *array.String:
		return col.Value(idx)

	case *array.LargeString:
		return col.Value(idx)

	case *array.Binary:
		return string(col.Value(idx))

	case *array.LargeBinary:
		return string(col.Value(idx))

	case *array.Date32:
		// Convert to "YYYY-MM-DD" string to match MySQL DATE behavior
		days := col.Value(idx)
		t := time.Unix(int64(days)*86400, 0).UTC()
		return t.Format("2006-01-02")

	case *array.Date64:
		ms := col.Value(idx)
		t := time.Unix(int64(ms)/1000, (int64(ms)%1000)*1e6).UTC()
		return t.Format("2006-01-02")

	case *array.Timestamp:
		// Convert to time.Time to match MySQL DATETIME behavior (parseTime=true)
		t := col.Value(idx)
		tsType := dt.(*arrow.TimestampType)
		return t.ToTime(tsType.Unit).UTC()

	case *array.Decimal128:
		return col.Value(idx).ToString(dt.(*arrow.Decimal128Type).Scale)

	case *array.Decimal256:
		return col.Value(idx).ToString(dt.(*arrow.Decimal256Type).Scale)

	case *array.List:
		return col.ValueStr(idx)

	case *array.Map:
		return col.ValueStr(idx)

	case *array.Struct:
		return col.ValueStr(idx)

	default:
		// Fallback: use string representation
		return col.ValueStr(idx)
	}
}

// arrowSchemaToRuntimeSchema converts an Arrow schema to a runtime StructType.
func arrowSchemaToRuntimeSchema(arrowSchema *arrow.Schema) (*runtimev1.StructType, error) {
	fields := make([]*runtimev1.StructType_Field, arrowSchema.NumFields())
	for i, field := range arrowSchema.Fields() {
		runtimeType, err := arrowTypeToRuntimeType(field.Type)
		if err != nil {
			return nil, fmt.Errorf("unsupported Arrow type %s for column %q: %w", field.Type, field.Name, err)
		}
		fields[i] = &runtimev1.StructType_Field{
			Name: field.Name,
			Type: runtimeType,
		}
	}
	return &runtimev1.StructType{Fields: fields}, nil
}

// arrowTypeToRuntimeType converts an Arrow data type to a runtime Type.
func arrowTypeToRuntimeType(dt arrow.DataType) (*runtimev1.Type, error) {
	switch dt.ID() {
	case arrow.BOOL:
		return &runtimev1.Type{Code: runtimev1.Type_CODE_BOOL}, nil

	case arrow.INT8, arrow.UINT8:
		return &runtimev1.Type{Code: runtimev1.Type_CODE_INT8}, nil

	case arrow.INT16, arrow.UINT16:
		return &runtimev1.Type{Code: runtimev1.Type_CODE_INT16}, nil

	case arrow.INT32, arrow.UINT32:
		return &runtimev1.Type{Code: runtimev1.Type_CODE_INT32}, nil

	case arrow.INT64, arrow.UINT64:
		return &runtimev1.Type{Code: runtimev1.Type_CODE_INT64}, nil

	case arrow.FLOAT16, arrow.FLOAT32:
		return &runtimev1.Type{Code: runtimev1.Type_CODE_FLOAT32}, nil

	case arrow.FLOAT64:
		return &runtimev1.Type{Code: runtimev1.Type_CODE_FLOAT64}, nil

	case arrow.DECIMAL128, arrow.DECIMAL256:
		return &runtimev1.Type{Code: runtimev1.Type_CODE_STRING}, nil

	case arrow.STRING, arrow.LARGE_STRING:
		return &runtimev1.Type{Code: runtimev1.Type_CODE_STRING}, nil

	case arrow.BINARY, arrow.LARGE_BINARY:
		return &runtimev1.Type{Code: runtimev1.Type_CODE_STRING}, nil

	case arrow.DATE32, arrow.DATE64:
		return &runtimev1.Type{Code: runtimev1.Type_CODE_DATE}, nil

	case arrow.TIMESTAMP:
		return &runtimev1.Type{Code: runtimev1.Type_CODE_TIMESTAMP}, nil

	case arrow.TIME32, arrow.TIME64:
		return &runtimev1.Type{Code: runtimev1.Type_CODE_TIME}, nil

	case arrow.LIST, arrow.LARGE_LIST, arrow.FIXED_SIZE_LIST:
		return &runtimev1.Type{Code: runtimev1.Type_CODE_ARRAY}, nil

	case arrow.MAP:
		return &runtimev1.Type{Code: runtimev1.Type_CODE_MAP}, nil

	case arrow.STRUCT:
		return &runtimev1.Type{Code: runtimev1.Type_CODE_STRUCT}, nil

	default:
		return nil, fmt.Errorf("unsupported Arrow type: %s", dt)
	}
}
