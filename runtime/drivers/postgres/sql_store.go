package postgres

import (
	"context"
	sqldriver "database/sql/driver"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"golang.org/x/exp/slices"

	// load pgx driver
	_ "github.com/jackc/pgx/v4/stdlib"
)

// Query implements drivers.SQLStore
func (c *connection) Query(ctx context.Context, props map[string]any) (drivers.RowIterator, error) {
	sql, ok := props["sql"].(string)
	if !ok {
		return nil, fmt.Errorf("property \"sql\" is mandatory for connector \"motherduck\"")
	}

	res, err := c.db.QueryxContext(ctx, sql)
	if err != nil {
		return nil, err
	}

	schema, err := rowsToSchema(res)
	if err != nil {
		return nil, err
	}

	slices.SortFunc(schema.Fields, func(a, b *runtimev1.StructType_Field) bool {
		return a.Name < b.Name
	})

	return &rowIterator{
		rows:   res,
		schema: schema,
		row:    make([]sqldriver.Value, len(schema.Fields)),
		rowMap: make(map[string]any, len(schema.Fields)),
	}, nil
}

// QueryAsFiles implements drivers.SQLStore
func (c *connection) QueryAsFiles(ctx context.Context, props map[string]any, opt *drivers.QueryOption, p drivers.Progress) (drivers.FileIterator, error) {
	return nil, drivers.ErrNotImplemented
}

type rowIterator struct {
	rows   *sqlx.Rows
	schema *runtimev1.StructType

	row    []sqldriver.Value
	rowMap map[string]any
}

// Close implements drivers.RowIterator.
func (r *rowIterator) Close() error {
	return r.rows.Close()
}

// Next implements drivers.RowIterator.
func (r *rowIterator) Next(ctx context.Context) ([]sqldriver.Value, error) {
	if !r.rows.Next() {
		if r.rows.Err() == nil {
			return nil, drivers.ErrIteratorDone
		}
		return nil, r.rows.Err()
	}

	err := r.rows.MapScan(r.rowMap)
	if err != nil {
		return nil, err
	}

	for i, field := range r.schema.Fields {
		r.row[i] = r.rowMap[field.Name]
	}
	return r.row, nil
}

// Schema implements drivers.RowIterator.
func (r *rowIterator) Schema(ctx context.Context) (*runtimev1.StructType, error) {
	return r.schema, nil
}

// Size implements drivers.RowIterator.
func (r *rowIterator) Size(unit drivers.ProgressUnit) (uint64, bool) {
	return 0, false
}

var _ drivers.RowIterator = &rowIterator{}

func rowsToSchema(r *sqlx.Rows) (*runtimev1.StructType, error) {
	if r == nil {
		return nil, drivers.ErrIteratorDone
	}

	cts, err := r.ColumnTypes()
	if err != nil {
		return nil, err
	}

	fields := make([]*runtimev1.StructType_Field, len(cts))
	for i, ct := range cts {
		nullable, ok := ct.Nullable()
		if !ok {
			nullable = true
		}

		fields[i] = &runtimev1.StructType_Field{
			Name: ct.Name(),
			Type: databaseTypeToPB(ct.DatabaseTypeName(), nullable),
		}
	}

	return &runtimev1.StructType{Fields: fields}, nil
}

// Refer table for superset of types https://www.postgresql.org/docs/current/datatype.html
func databaseTypeToPB(dbt string, nullable bool) *runtimev1.Type {
	t := &runtimev1.Type{Nullable: nullable}

	// type of array of base types being with _ like _FLOAT8
	if strings.HasPrefix(dbt, "_") {
		// TODO :: use lists once appender supports it
		t.Code = runtimev1.Type_CODE_JSON
		return t
	}

	switch dbt {
	case "BIGINT", "INT8", "BIGSERIAL", "SERIAL8":
		t.Code = runtimev1.Type_CODE_INT64
	case "BIT", "BIT VARYING", "VARBIT":
		t.Code = runtimev1.Type_CODE_STRING // TODO bitstring type once appender supports it
	case "BOOLEAN", "BOOL":
		t.Code = runtimev1.Type_CODE_BOOL
	case "BYTEA":
		t.Code = runtimev1.Type_CODE_BYTES
	case "CHARACTER", "CHARACTER VARYING", "BPCHAR":
		t.Code = runtimev1.Type_CODE_STRING // TODO separate datatypes for fixed length and variable length string
	case "DATE":
		t.Code = runtimev1.Type_CODE_DATE
	case "DOUBLE PRECISION", "FLOAT8":
		t.Code = runtimev1.Type_CODE_FLOAT64
	case "INTEGER", "INT", "INT4":
		t.Code = runtimev1.Type_CODE_INT32
	case "INTERVAL":
		t.Code = runtimev1.Type_CODE_STRING // TODO - Consider adding interval type
	case "JSON":
		t.Code = runtimev1.Type_CODE_JSON
	case "JSONB":
		t.Code = runtimev1.Type_CODE_BYTES
	case "NUMERIC", "DECIMAL":
		t.Code = runtimev1.Type_CODE_STRING
	case "REAL", "FLOAT4":
		t.Code = runtimev1.Type_CODE_FLOAT32
	case "SMALLINT", "INT2", "SMALLSERIAL", "SERIAL2":
		t.Code = runtimev1.Type_CODE_INT16
	case "SERIAL", "SERIAL4":
		t.Code = runtimev1.Type_CODE_INT32
	case "TEXT":
		t.Code = runtimev1.Type_CODE_STRING
	case "TIME":
		t.Code = runtimev1.Type_CODE_TIME
	case "TIME WITH TIME ZONE":
		t.Code = runtimev1.Type_CODE_TIMESTAMP
	case "TIMESTAMP":
		t.Code = runtimev1.Type_CODE_TIME
	case "TIMESTAMP WITH TIME ZONE":
		t.Code = runtimev1.Type_CODE_TIMESTAMP
	case "TIMESTAMPTZ":
		t.Code = runtimev1.Type_CODE_TIMESTAMP
	case "UUID":
		t.Code = runtimev1.Type_CODE_UUID
	case "VARCHAR":
		t.Code = runtimev1.Type_CODE_STRING
	case "POINT", "LINE", "LSEG", "BOX", "PATH", "POLYGON", "CIRCLE":
		t.Code = runtimev1.Type_CODE_JSON // postgres predefined struct types, move to struct once appender supports structs
	default:
		// There are many datatypes in postgres, convert all to string
		t.Code = runtimev1.Type_CODE_STRING
	}

	// Done
	return t
}
