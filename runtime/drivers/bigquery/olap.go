package bigquery

import (
	"context"
	sqldriver "database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"time"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/civil"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/pkg/sqlconvert"
	"go.uber.org/zap"
	"google.golang.org/api/iterator"
)

var _ drivers.OLAPStore = (*Connection)(nil)

// Dialect implements drivers.OLAPStore.
func (c *Connection) Dialect() drivers.Dialect {
	return drivers.DialectBigQuery
}

// Exec implements drivers.OLAPStore.
func (c *Connection) Exec(ctx context.Context, stmt *drivers.Statement) error {
	res, err := c.Query(ctx, stmt)
	if err != nil {
		return err
	}
	if res != nil {
		return res.Close()
	}
	return nil
}

// InformationSchema implements drivers.OLAPStore.
func (c *Connection) InformationSchema() drivers.OLAPInformationSchema {
	return c
}

// MayBeScaledToZero implements drivers.OLAPStore.
func (c *Connection) MayBeScaledToZero(ctx context.Context) bool {
	return true
}

// Query implements drivers.OLAPStore.
func (c *Connection) Query(ctx context.Context, stmt *drivers.Statement) (res *drivers.Result, resErr error) {
	if c.config.LogQueries {
		c.logger.Info("bigquery query", zap.String("sql", c.Dialect().SanitizeQueryForLogging(stmt.Query)), zap.Any("args", stmt.Args), observability.ZapCtx(ctx))
	}
	client, err := c.getClient(ctx)
	if err != nil {
		return nil, err
	}

	q := client.Query(stmt.Query)
	q.Parameters = make([]bigquery.QueryParameter, len(stmt.Args))
	for i, arg := range stmt.Args {
		q.Parameters[i] = bigquery.QueryParameter{
			Value: arg,
		}
	}
	if stmt.DryRun {
		q.DryRun = true
		// Can not use q.Read for dry run so must trigger the job and check status
		j, err := q.Run(ctx)
		if err != nil {
			return nil, err
		}
		// Dry run is not asynchronous so no need to call Wait
		status := j.LastStatus()
		return nil, status.Err()
	}
	it, err := q.Read(ctx)
	if err != nil {
		return nil, err
	}

	// We use query.Read which can also use fast path when results are small.
	// In fast path schema is only available after fetching the first row.
	var firstRow []bigquery.Value
	for i := 0; i < len(it.Schema); i++ {
		firstRow = append(firstRow, new(bigquery.Value))
	}
	rowErr := it.Next(&firstRow)
	if rowErr != nil && !errors.Is(rowErr, iterator.Done) {
		return nil, err
	}
	// schema is returned even if there are no rows
	schema, err := fromBQSchema(it.Schema)
	if err != nil {
		return nil, err
	}
	row := newRows(it, firstRow, errors.Is(rowErr, iterator.Done))
	res = &drivers.Result{
		Rows:   row,
		Schema: schema,
	}
	return res, nil
}

// QuerySchema implements drivers.OLAPStore.
func (c *Connection) QuerySchema(ctx context.Context, query string, args []any) (*runtimev1.StructType, error) {
	return nil, drivers.ErrNotImplemented
}

// WithConnection implements drivers.OLAPStore.
func (c *Connection) WithConnection(ctx context.Context, priority int, fn drivers.WithConnectionFunc) error {
	return drivers.ErrNotImplemented
}

// All implements drivers.OLAPInformationSchema.
func (c *Connection) All(ctx context.Context, like string, pageSize uint32, pageToken string) ([]*drivers.OlapTable, string, error) {
	return drivers.AllFromInformationSchema(ctx, like, pageSize, pageToken, c)
}

// LoadPhysicalSize implements drivers.OLAPInformationSchema.
func (c *Connection) LoadPhysicalSize(ctx context.Context, tables []*drivers.OlapTable) error {
	return nil
}

// Lookup implements drivers.OLAPInformationSchema.
func (c *Connection) Lookup(ctx context.Context, db, schema, name string) (*drivers.OlapTable, error) {
	client, err := c.getClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get BigQuery client: %w", err)
	}

	table := client.Dataset(schema).Table(name)
	meta, err := table.Metadata(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get table metadata: %w", err)
	}
	runtimeSchema, err := fromBQSchema(meta.Schema)
	if err != nil {
		return nil, err
	}
	return &drivers.OlapTable{
		Database:          db,
		DatabaseSchema:    schema,
		Name:              name,
		View:              meta.Type == bigquery.ViewTable,
		Schema:            runtimeSchema,
		UnsupportedCols:   nil, // all columns are currently being mapped though may not be as specific as in BigQuery
		PhysicalSizeBytes: 0,
	}, nil
}

type rows struct {
	ri *bigquery.RowIterator

	firstRow        []bigquery.Value
	canScanFirstRow bool

	lastRow    []bigquery.Value // last scanned row from ri in Next
	lastErr    error
	canScanRow bool
}

func newRows(ri *bigquery.RowIterator, firstRow []bigquery.Value, noRows bool) *rows {
	if noRows {
		return &rows{
			lastErr: iterator.Done,
		}
	}
	r := &rows{
		ri:              ri,
		firstRow:        firstRow,
		canScanFirstRow: true,
	}
	r.lastRow = make([]bigquery.Value, len(firstRow))
	for i := range len(firstRow) {
		r.lastRow[i] = new(bigquery.Value)
	}
	return r
}

var _ drivers.Rows = &rows{}

// Close implements drivers.Rows.
func (r *rows) Close() error {
	r.firstRow = nil
	r.lastRow = nil
	return nil
}

// Err implements drivers.Rows.
func (r *rows) Err() error {
	if errors.Is(r.lastErr, iterator.Done) {
		return nil
	}
	return r.lastErr
}

// MapScan implements drivers.Rows.
func (r *rows) MapScan(dest map[string]any) error {
	if dest == nil {
		return fmt.Errorf("nil destination map in MapScan")
	}
	row, err := r.nextRow()
	if err != nil {
		return err
	}
	for i, col := range r.ri.Schema {
		dest[col.Name], err = convertValue(r.ri.Schema[i], row[i])
		if err != nil {
			return err
		}
	}
	return nil
}

// Next implements drivers.Rows.
func (r *rows) Next() bool {
	if r.lastErr != nil {
		return false
	}

	// first row was already fetched during query execution to get schema
	if r.canScanFirstRow {
		r.canScanRow = true
		r.canScanFirstRow = false
		return true
	}

	err := r.ri.Next(&r.lastRow)
	if err != nil {
		if errors.Is(err, iterator.Done) {
			return false
		}
		r.lastErr = err
		return false
	}
	r.canScanRow = true
	return true
}

// Scan implements drivers.Rows.
func (r *rows) Scan(dest ...any) error {
	if len(dest) != len(r.lastRow) {
		return fmt.Errorf("expected %d destination arguments in Scan, got %d", len(r.lastRow), len(dest))
	}
	row, err := r.nextRow()
	if err != nil {
		return err
	}

	for i := range dest {
		v, err := convertValue(r.ri.Schema[i], row[i])
		if err != nil {
			return err
		}
		err = sqlconvert.ConvertAssign(dest[i], v)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *rows) nextRow() ([]bigquery.Value, error) {
	if r.lastErr != nil {
		return nil, r.lastErr
	}
	if !r.canScanRow {
		return nil, fmt.Errorf("must call Next before Scan")
	}

	var row []bigquery.Value
	if r.firstRow != nil {
		row = r.firstRow
		r.firstRow = nil
	} else {
		row = r.lastRow
	}
	r.canScanRow = false
	return row, nil
}

func fromBQSchema(bqSchema bigquery.Schema) (*runtimev1.StructType, error) {
	fields := make([]*runtimev1.StructType_Field, len(bqSchema))
	for i, s := range bqSchema {
		dbt, err := toPB(s)
		if err != nil {
			return nil, err
		}
		fields[i] = &runtimev1.StructType_Field{Name: s.Name, Type: dbt}
	}
	return &runtimev1.StructType{Fields: fields}, nil
}

func toPB(field *bigquery.FieldSchema) (*runtimev1.Type, error) {
	t := &runtimev1.Type{Nullable: !field.Required}
	if field.Repeated {
		t.Code = runtimev1.Type_CODE_ARRAY
		return t, nil
	}
	switch field.Type {
	case bigquery.StringFieldType:
		t.Code = runtimev1.Type_CODE_STRING
	case bigquery.JSONFieldType:
		t.Code = runtimev1.Type_CODE_JSON
	case bigquery.IntervalFieldType:
		t.Code = runtimev1.Type_CODE_STRING
	case bigquery.GeographyFieldType:
		t.Code = runtimev1.Type_CODE_STRING
	case bigquery.FloatFieldType:
		t.Code = runtimev1.Type_CODE_FLOAT64
	case bigquery.NumericFieldType:
		t.Code = runtimev1.Type_CODE_STRING
	case bigquery.BigNumericFieldType:
		t.Code = runtimev1.Type_CODE_STRING
	case bigquery.TimestampFieldType:
		t.Code = runtimev1.Type_CODE_TIMESTAMP
	case bigquery.DateTimeFieldType:
		t.Code = runtimev1.Type_CODE_STRING
	case bigquery.TimeFieldType:
		t.Code = runtimev1.Type_CODE_STRING
	case bigquery.DateFieldType:
		t.Code = runtimev1.Type_CODE_STRING
	case bigquery.BooleanFieldType:
		t.Code = runtimev1.Type_CODE_BOOL
	case bigquery.IntegerFieldType:
		t.Code = runtimev1.Type_CODE_INT64
	case bigquery.BytesFieldType:
		t.Code = runtimev1.Type_CODE_BYTES
	case bigquery.RecordFieldType:
		t.Code = runtimev1.Type_CODE_JSON
	case bigquery.RangeFieldType:
		t.Code = runtimev1.Type_CODE_STRING
	default:
		return nil, fmt.Errorf("type %s not supported", field.Type)
	}
	return t, nil
}

func convertValue(field *bigquery.FieldSchema, value bigquery.Value) (any, error) {
	val, err := convertValueHelper(field, value)
	if err != nil {
		return nil, err
	}

	if sqldriver.IsValue(val) {
		return val, nil
	}

	// Marshal ARRAY and RECORD types to JSON, since arrays/maps aren't
	// valid driver.Value types.
	out, err := json.Marshal(val)
	if err != nil {
		return nil, fmt.Errorf("error marshalling %s field of type %s to JSON: %w", field.Name, field.Type, err)
	}
	return string(out), nil
}

func convertValueHelper(field *bigquery.FieldSchema, value bigquery.Value) (any, error) {
	if field.Repeated {
		return convertRepeatedType(field, value)
	}
	return convertUnitType(field, value)
}

func convertUnitType(field *bigquery.FieldSchema, value bigquery.Value) (any, error) {
	switch field.Type {
	case bigquery.StringFieldType:
		return convertBasicType[string](field, value)
	case bigquery.BytesFieldType:
		return convertBasicType[[]byte](field, value)
	case bigquery.IntegerFieldType:
		return convertBasicType[int64](field, value)
	case bigquery.FloatFieldType:
		return convertBasicType[float64](field, value)
	case bigquery.BooleanFieldType:
		return convertBasicType[bool](field, value)
	case bigquery.TimestampFieldType:
		return convertBasicType[time.Time](field, value)
	case bigquery.DateFieldType:
		return convertStringerType[civil.Date](field, value)
	case bigquery.TimeFieldType:
		return convertStringerType[civil.Time](field, value)
	case bigquery.DateTimeFieldType:
		return convertStringerType[civil.DateTime](field, value)
	case bigquery.NumericFieldType:
		return convertRationalType(field, value, bigquery.NumericString)
	case bigquery.BigNumericFieldType:
		return convertRationalType(field, value, bigquery.BigNumericString)
	case bigquery.GeographyFieldType:
		return convertBasicType[string](field, value)
	case bigquery.IntervalFieldType:
		return convertStringerType[*bigquery.IntervalValue](field, value)
	case bigquery.RangeFieldType:
		return convertBigQueryRangeType(field, value)
	case bigquery.JSONFieldType:
		return convertBasicType[string](field, value)
	case bigquery.RecordFieldType:
		return convertRecordType(field, value)
	default:
		return nil, fmt.Errorf("type %s not supported", field.Type)
	}
}

func convertRepeatedType(field *bigquery.FieldSchema, value bigquery.Value) ([]any, error) {
	switch val := value.(type) {
	case nil:
		return nil, nil
	case []bigquery.Value:
		a := make([]any, len(val))
		for i, v := range val {
			av, err := convertUnitType(field, v)
			if err != nil {
				return nil, err
			}
			a[i] = av
		}
		return a, nil
	default:
		return nil, &unexpectedTypeError{
			FieldType: field.Type,
			Expected:  reflect.TypeFor[[]bigquery.Value](),
			Actual:    val,
		}
	}
}

func convertRecordType(field *bigquery.FieldSchema, value bigquery.Value) (map[string]any, error) {
	switch val := value.(type) {
	case nil:
		return nil, nil
	case []bigquery.Value:
		m := map[string]any{}
		for i, mf := range field.Schema {
			mv, err := convertValueHelper(mf, val[i])
			if err != nil {
				return nil, err
			}
			m[mf.Name] = mv
		}
		return m, nil
	default:
		return nil, &unexpectedTypeError{
			FieldType: field.Type,
			Expected:  reflect.TypeFor[[]bigquery.Value](),
			Actual:    val,
		}
	}
}

func convertBasicType[T any](field *bigquery.FieldSchema, value bigquery.Value) (any, error) {
	switch val := value.(type) {
	case nil:
		return nil, nil
	case T:
		return val, nil
	default:
		return nil, &unexpectedTypeError{
			FieldType: field.Type,
			Expected:  reflect.TypeFor[T](),
			Actual:    val,
		}
	}
}

func convertStringerType[T fmt.Stringer](field *bigquery.FieldSchema, value bigquery.Value) (any, error) {
	switch val := value.(type) {
	case nil:
		return nil, nil
	case T:
		return val.String(), nil
	default:
		return nil, &unexpectedTypeError{
			FieldType: field.Type,
			Expected:  reflect.TypeFor[T](),
			Actual:    val,
		}
	}
}

type ratToStr func(*big.Rat) string

func convertRationalType(field *bigquery.FieldSchema, value bigquery.Value, toStr ratToStr) (any, error) {
	switch val := value.(type) {
	case nil:
		return nil, nil
	case *big.Rat:
		// Attempt to use the minimum number of digits after the decimal point,
		// if the resulting number will be exact.
		if prec, exact := val.FloatPrec(); exact {
			return val.FloatString(prec), nil
		}

		// Otherwise, fallback to default string conversion function, which
		// uses the maximum number of digits supported by BigQuery.
		return toStr(val), nil
	default:
		return nil, &unexpectedTypeError{
			FieldType: field.Type,
			Expected:  reflect.TypeFor[*big.Rat](),
			Actual:    val,
		}
	}
}

func convertBigQueryRangeType(field *bigquery.FieldSchema, value bigquery.Value) (any, error) {
	switch val := value.(type) {
	case nil:
		return nil, nil
	case *bigquery.RangeValue:
		return fmt.Sprintf("[%v, %v)", val.Start, val.End), nil
	default:
		return nil, &unexpectedTypeError{
			FieldType: field.Type,
			Expected:  reflect.TypeFor[*bigquery.RangeValue](),
			Actual:    val,
		}
	}
}

type unexpectedTypeError struct {
	FieldType bigquery.FieldType
	Expected  reflect.Type
	Actual    bigquery.Value
}

func (e *unexpectedTypeError) Error() string {
	return fmt.Sprintf(
		"received unexpected type: %T for BigQuery field: %s (expected: %s)",
		e.Actual, e.FieldType, e.Expected,
	)
}
