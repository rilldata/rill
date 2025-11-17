package postgres

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/zap"
)

var _ drivers.OLAPStore = (*connection)(nil)

// Dialect implements drivers.OLAPStore.
func (c *connection) Dialect() drivers.Dialect {
	return drivers.DialectPostgres
}

// Exec implements drivers.OLAPStore.
func (c *connection) Exec(ctx context.Context, stmt *drivers.Statement) error {
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
func (c *connection) InformationSchema() drivers.OLAPInformationSchema {
	return c
}

// MayBeScaledToZero implements drivers.OLAPStore.
func (c *connection) MayBeScaledToZero(ctx context.Context) bool {
	return false
}

// Query implements drivers.OLAPStore.
func (c *connection) Query(ctx context.Context, stmt *drivers.Statement) (*drivers.Result, error) {
	if c.logger != nil {
		c.logger.Info("Postgres query", zap.String("sql", c.Dialect().SanitizeQueryForLogging(stmt.Query)), zap.Any("args", stmt.Args), observability.ZapCtx(ctx))
	}
	db, err := c.getDB(ctx)
	if err != nil {
		return nil, err
	}

	if stmt.DryRun {
		_, err = db.ExecContext(ctx, fmt.Sprintf("EXPLAIN %s", stmt.Query), stmt.Args...)
		return nil, err
	}

	rows, err := db.QueryxContext(ctx, stmt.Query, stmt.Args...)
	if err != nil {
		return nil, err
	}

	schema, err := rowsToSchema(rows)
	if err != nil {
		rows.Close()
		return nil, err
	}

	res := &drivers.Result{Rows: rows, Schema: schema}
	return res, nil
}

// QuerySchema implements drivers.OLAPStore.
func (c *connection) QuerySchema(ctx context.Context, query string, args []any) (*runtimev1.StructType, error) {
	return nil, drivers.ErrNotImplemented
}

// WithConnection implements drivers.OLAPStore.
func (c *connection) WithConnection(ctx context.Context, priority int, fn drivers.WithConnectionFunc) error {
	return drivers.ErrNotImplemented
}

// All implements drivers.OLAPInformationSchema.
func (c *connection) All(ctx context.Context, like string, pageSize uint32, pageToken string) ([]*drivers.OlapTable, string, error) {
	return drivers.AllFromInformationSchema(ctx, like, pageSize, pageToken, c)
}

// LoadPhysicalSize implements drivers.OLAPInformationSchema.
func (c *connection) LoadPhysicalSize(ctx context.Context, tables []*drivers.OlapTable) error {
	return nil
}

// Lookup implements drivers.OLAPInformationSchema.
func (c *connection) Lookup(ctx context.Context, db, schema, name string) (*drivers.OlapTable, error) {
	meta, err := c.GetTable(ctx, db, schema, name)
	if err != nil {
		return nil, err
	}

	rtSchema := &runtimev1.StructType{}
	for name, typ := range meta.Schema {
		t := databaseTypeToPB(typ)
		rtSchema.Fields = append(rtSchema.Fields, &runtimev1.StructType_Field{
			Name: name,
			Type: t,
		})
	}
	return &drivers.OlapTable{
		Database:          db,
		DatabaseSchema:    schema,
		Name:              name,
		View:              meta.View,
		Schema:            rtSchema,
		UnsupportedCols:   nil,
		PhysicalSizeBytes: 0,
	}, nil
}

func rowsToSchema(r *sqlx.Rows) (*runtimev1.StructType, error) {
	fds, err := r.ColumnTypes()
	if err != nil {
		return nil, err
	}

	fields := make([]*runtimev1.StructType_Field, len(fds))
	for i, fd := range fds {
		rt := databaseTypeToPB(fd.DatabaseTypeName())
		fields[i] = &runtimev1.StructType_Field{
			Name: fd.Name(),
			Type: rt,
		}
	}
	return &runtimev1.StructType{Fields: fields}, nil
}

func databaseTypeToPB(dbt string) *runtimev1.Type {
	t := &runtimev1.Type{Nullable: true}

	// Handle array types (prefixed with underscore)
	if dbt != "" && dbt[0] == '_' {
		t.Code = runtimev1.Type_CODE_ARRAY
		return t
	}

	switch dbt {
	case "NUMERIC", "DECIMAL":
		t.Code = runtimev1.Type_CODE_DECIMAL
	case "INT2", "SMALLINT", "SMALLSERIAL":
		t.Code = runtimev1.Type_CODE_INT64 // sql driver returns int64 for smallint
	case "INT4", "INTEGER", "SERIAL":
		t.Code = runtimev1.Type_CODE_INT64 // sql driver returns int64 for INT4
	case "INT8", "BIGINT", "BIGSERIAL":
		t.Code = runtimev1.Type_CODE_INT64
	case "FLOAT4", "REAL":
		t.Code = runtimev1.Type_CODE_FLOAT64 // sql driver returns float64 for FLOAT4
	case "FLOAT8", "DOUBLE PRECISION":
		t.Code = runtimev1.Type_CODE_FLOAT64
	case "VARCHAR", "CHAR", "CHARACTER", "CHARACTER VARYING", "TEXT", "BPCHAR", "NAME":
		t.Code = runtimev1.Type_CODE_STRING
	case "BYTEA":
		t.Code = runtimev1.Type_CODE_BYTES
	case "BOOL", "BOOLEAN":
		t.Code = runtimev1.Type_CODE_BOOL
	case "DATE":
		t.Code = runtimev1.Type_CODE_DATE
	case "TIME", "TIME WITHOUT TIME ZONE":
		t.Code = runtimev1.Type_CODE_STRING // TIME is returned as string by pgx
	case "TIMESTAMP", "TIMESTAMP WITHOUT TIME ZONE":
		t.Code = runtimev1.Type_CODE_TIMESTAMP
	case "TIMESTAMPTZ", "TIMESTAMP WITH TIME ZONE":
		t.Code = runtimev1.Type_CODE_TIMESTAMP
	case "INTERVAL":
		t.Code = runtimev1.Type_CODE_INTERVAL
	case "UUID":
		t.Code = runtimev1.Type_CODE_UUID
	case "JSON", "JSONB":
		t.Code = runtimev1.Type_CODE_JSON
	case "ARRAY":
		t.Code = runtimev1.Type_CODE_ARRAY
	case "INET", "CIDR", "MACADDR", "MACADDR8":
		t.Code = runtimev1.Type_CODE_STRING
	case "BIT", "BIT VARYING", "VARBIT":
		t.Code = runtimev1.Type_CODE_STRING
	case "POINT", "LINE", "LSEG", "BOX", "PATH", "POLYGON", "CIRCLE":
		t.Code = runtimev1.Type_CODE_STRING
	case "MONEY":
		t.Code = runtimev1.Type_CODE_DECIMAL
	case "XML":
		t.Code = runtimev1.Type_CODE_STRING
	case "TSVECTOR", "TSQUERY":
		t.Code = runtimev1.Type_CODE_STRING
	default:
		t.Code = runtimev1.Type_CODE_UNSPECIFIED
		return t
	}
	return t
}
