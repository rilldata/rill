package bigquery

import (
	"context"
	"fmt"

	"cloud.google.com/go/bigquery"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"google.golang.org/api/iterator"
)

var _ drivers.OLAPStore = (*Connection)(nil)

// Dialect implements drivers.OLAPStore.
func (c *Connection) Dialect() drivers.Dialect {
	return drivers.DialectBigQuery
}

// Exec implements drivers.OLAPStore.
func (c *Connection) Exec(ctx context.Context, stmt *drivers.Statement) error {
	return drivers.ErrNotImplemented
}

// InformationSchema implements drivers.OLAPStore.
func (c *Connection) InformationSchema() drivers.OLAPInformationSchema {

}

// MayBeScaledToZero implements drivers.OLAPStore.
func (c *Connection) MayBeScaledToZero(ctx context.Context) bool {
	return false
}

// Query implements drivers.OLAPStore.
func (c *Connection) Query(ctx context.Context, stmt *drivers.Statement) (*drivers.Result, error) {
	client, err := c.createClient(ctx, "") // project id detected from configs
	if err != nil {
		return nil, err
	}
	q := client.Query(stmt.Query)
	j, err := q.Run(ctx)
	if err != nil {
		return nil, err
	}
	it, err := j.Read(ctx)
	if err != nil {
		return nil, err
	}

	var firstRow []any
	for i := 0; i < len(it.Schema); i++ {
		firstRow = append(firstRow, new(any))
	}
	err = it.Next(firstRow)
	if err != nil {
		if err == iterator.Done {
			return nil, drivers.ErrNoRows
		}
		return nil, err
	}

	schema, err := fromBQSchema(it.Schema)
	if err != nil {
		return nil, err
	}
	row := newRows(it, firstRow)
	return &drivers.Result{
		Rows:   row,
		Schema: schema,
	}, nil
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
	schemas, token, err := c.ListDatabaseSchemas(ctx, pageSize, pageToken)
	if err != nil {
		return nil, "", err
	}
	tables := make([]*drivers.OlapTable, 0)
	for _, schema := range schemas {
		ts, token, err := c.ListTables(ctx, schema.Database, schema.DatabaseSchema, 1000, "")
		if err != nil {
			return nil, "", err
		}
		if token != "" {
			// we don't support pagination across multiple schemas
			return nil, "", fmt.Errorf("schema has more than 1000 tables can not list all")
		}
		for _, t := range ts {
			table := &drivers.OlapTable{
				Database:          schema.Database,
				DatabaseSchema:    schema.DatabaseSchema,
				Name:              t.Name,
				View:              t.View,
				Schema:            nil, // todo: load schema
				UnsupportedCols:   nil,
				PhysicalSizeBytes: 0,
			}
			tables = append(tables, table)
		}

	}
	return tables, token, nil
}

// LoadPhysicalSize implements drivers.OLAPInformationSchema.
func (c *Connection) LoadPhysicalSize(ctx context.Context, tables []*drivers.OlapTable) error {
	panic("unimplemented")
}

// Lookup implements drivers.OLAPInformationSchema.
func (c *Connection) Lookup(ctx context.Context, db string, schema string, name string) (*drivers.OlapTable, error) {
	panic("unimplemented")
}

type rows struct {
	ri *bigquery.RowIterator

	firstRow       []any
	lastRow        []any // last scanned row from ri in Next
	lastErr        error
	canScanLastRow bool
}

func newRows(ri *bigquery.RowIterator, firstRow []any) *rows {
	r := &rows{
		ri:       ri,
		firstRow: firstRow,
	}
	r.lastRow = make([]any, len(firstRow))
	for i := range len(firstRow) {
		r.lastRow[i] = new(any)
	}
	return r
}

var _ drivers.Rows = &rows{}

// Close implements drivers.Rows.
func (r *rows) Close() error {
	return nil
}

// Err implements drivers.Rows.
func (r *rows) Err() error {
	panic("unimplemented")
}

// MapScan implements drivers.Rows.
func (r *rows) MapScan(dest map[string]any) error {
	panic("unimplemented")
}

// Next implements drivers.Rows.
func (r *rows) Next() bool {
	if r.firstRow != nil {
		r.canScanLastRow = true
		return true
	}
	err := r.ri.Next(r.lastRow)
	if err != nil {
		if err == iterator.Done {
			return false
		}
		r.lastErr = err
		return false
	}
	r.canScanLastRow = true
	return true
}

// Scan implements drivers.Rows.
func (r *rows) Scan(dest ...any) error {
	if len(dest) != len(r.lastRow) {
		return fmt.Errorf("expected %d destination arguments in Scan, got %d", len(r.lastRow), len(dest))
	}
	if !r.canScanLastRow {
		return fmt.Errorf("must call Next before Scan")
	}

	var row []any
	if r.firstRow != nil {
		row = r.firstRow
		r.firstRow = nil
	} else {
		row = r.lastRow
	}
	r.canScanLastRow = false

	for i := range dest {
		dest[i] = *(row[i].(*interface{}))
	}
	return nil
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
	// big numeric can have width upto 76 digits which can't be converted to DECIMAL type in duckdb
	// which supports width upto 38 digits only.
	case bigquery.BigNumericFieldType:
		t.Code = runtimev1.Type_CODE_STRING
	case bigquery.TimestampFieldType:
		t.Code = runtimev1.Type_CODE_TIMESTAMP
	case bigquery.DateTimeFieldType:
		t.Code = runtimev1.Type_CODE_TIMESTAMP
	case bigquery.TimeFieldType:
		t.Code = runtimev1.Type_CODE_TIME
	case bigquery.DateFieldType:
		t.Code = runtimev1.Type_CODE_DATE
	case bigquery.BooleanFieldType:
		t.Code = runtimev1.Type_CODE_BOOL
	case bigquery.IntegerFieldType:
		t.Code = runtimev1.Type_CODE_INT64
	case bigquery.BytesFieldType:
		t.Code = runtimev1.Type_CODE_BYTES
	case bigquery.RecordFieldType:
		return nil, fmt.Errorf("record type not supported")
	default:
		return nil, fmt.Errorf("type %s not supported", field.Type)
	}
	return t, nil
}
