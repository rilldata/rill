package snowflake

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/zap"
)

var _ drivers.OLAPStore = (*connection)(nil)

// Dialect implements drivers.OLAPStore.
func (c *connection) Dialect() drivers.Dialect {
	return DialectSnowflake
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
func (c *connection) InformationSchema() drivers.InformationSchema {
	return c
}

// EstimateSize implements drivers.OLAPStore.
func (c *connection) EstimateSize(ctx context.Context) (int64, error) {
	return -1, nil
}

// MayBeScaledToZero implements drivers.OLAPStore.
func (c *connection) MayBeScaledToZero(ctx context.Context) bool {
	return true
}

// Query implements drivers.OLAPStore.
func (c *connection) Query(ctx context.Context, stmt *drivers.Statement) (*drivers.Result, error) {
	if c.configProperties.LogQueries {
		fields := []zap.Field{
			zap.String("sql", c.Dialect().SanitizeQueryForLogging(stmt.Query)),
			zap.Any("args", stmt.Args),
			observability.ZapCtx(ctx),
		}
		if len(stmt.QueryAttributes) > 0 {
			fields = append(fields, zap.Any("query_attributes", stmt.QueryAttributes))
		}
		c.logger.Info("Snowflake query", fields...)
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
		_ = rows.Close()
		return nil, err
	}

	res := &drivers.Result{Rows: rows, Schema: schema}
	return res, nil
}

func (c *connection) Head(ctx context.Context, db, schema, table string, limit int64) (*drivers.Result, error) {
	tbl, err := c.InformationSchema().Lookup(ctx, db, schema, table)
	if err != nil {
		return nil, err
	}

	var columns []string
	for _, field := range tbl.Schema.Fields {
		columns = append(columns, c.Dialect().EscapeIdentifier(field.Name))
	}

	limitClause := ""
	if limit > 0 {
		limitClause = fmt.Sprintf(" LIMIT %d", limit)
	}

	return c.Query(ctx, &drivers.Statement{
		Query: fmt.Sprintf("SELECT %s FROM %s%s", strings.Join(columns, ", "), c.Dialect().EscapeTable(db, schema, table), limitClause),
	})
}

// QuerySchema implements drivers.OLAPStore.
func (c *connection) QuerySchema(ctx context.Context, query string, args []any) (*runtimev1.StructType, error) {
	if c.configProperties.LogQueries {
		c.logger.Info("snowflake query", zap.String("sql", c.Dialect().SanitizeQueryForLogging(query)), zap.Any("args", args), observability.ZapCtx(ctx))
	}

	db, err := c.getDB(ctx)
	if err != nil {
		return nil, err
	}

	ctx, cancelFunc := context.WithTimeout(ctx, drivers.DefaultQuerySchemaTimeout)
	defer cancelFunc()

	rows, err := db.QueryxContext(ctx, fmt.Sprintf("SELECT * FROM (%s) LIMIT 0", query), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return rowsToSchema(rows)
}

// WithConnection implements drivers.OLAPStore.
func (c *connection) WithConnection(ctx context.Context, priority int, fn drivers.WithConnectionFunc) error {
	return drivers.ErrNotImplemented
}

func rowsToSchema(r *sqlx.Rows) (*runtimev1.StructType, error) {
	if r == nil {
		return nil, nil
	}

	cts, err := r.ColumnTypes()
	if err != nil {
		return nil, err
	}

	fields := make([]*runtimev1.StructType_Field, len(cts))
	for i, ct := range cts {
		_, scale, _ := ct.DecimalSize()
		nullable, ok := ct.Nullable()
		if !ok {
			nullable = true
		}

		t, err := databaseTypeToPB(ct.DatabaseTypeName(), scale, nullable)
		if err != nil {
			return nil, err
		}

		fields[i] = &runtimev1.StructType_Field{
			Name: ct.Name(),
			Type: t,
		}
	}

	return &runtimev1.StructType{Fields: fields}, nil
}

func databaseTypeToPB(dbt string, scale int64, nullable bool) (*runtimev1.Type, error) {
	t := &runtimev1.Type{Nullable: nullable, RawType: dbt}
	switch dbt {
	case "NUMBER", "DECIMAL", "NUMERIC":
		t.Code = runtimev1.Type_CODE_DECIMAL
	case "INT", "INTEGER", "BIGINT", "SMALLINT", "TINYINT", "BYTEINT": // All integers have same range in Snowflake
		t.Code = runtimev1.Type_CODE_INT256
	case "FLOAT", "FLOAT4", "FLOAT8": // All floats have same range in Snowflake
		t.Code = runtimev1.Type_CODE_FLOAT64
	case "DOUBLE", "DOUBLE PRECISION", "REAL":
		t.Code = runtimev1.Type_CODE_FLOAT64
	case "FIXED":
		if scale == 0 {
			t.Code = runtimev1.Type_CODE_INT256
		} else {
			t.Code = runtimev1.Type_CODE_DECIMAL
		}
	case "VARCHAR", "STRING", "TEXT", "CHAR", "CHARACTER":
		t.Code = runtimev1.Type_CODE_STRING
	case "BINARY", "VARBINARY":
		t.Code = runtimev1.Type_CODE_BYTES
	case "BOOLEAN":
		t.Code = runtimev1.Type_CODE_BOOL
	case "DATE":
		t.Code = runtimev1.Type_CODE_DATE
	case "DATETIME", "TIMESTAMP_NTZ": // ideally there should be a separate type signifying no timezone but runtime doesn't have one
		t.Code = runtimev1.Type_CODE_TIMESTAMP
	case "TIME":
		t.Code = runtimev1.Type_CODE_TIME
	case "TIMESTAMP_LTZ", "TIMESTAMP_TZ", "TIMESTAMP":
		t.Code = runtimev1.Type_CODE_TIMESTAMP
	case "INTERVAL":
		t.Code = runtimev1.Type_CODE_INTERVAL
	case "HUGEINT":
		t.Code = runtimev1.Type_CODE_INT128
	case "ENUM":
		t.Code = runtimev1.Type_CODE_STRING // TODO - Consider how to handle enums
	case "UUID":
		t.Code = runtimev1.Type_CODE_UUID
	case "VARIANT", "OBJECT", "ARRAY", "STRUCT":
		t.Code = runtimev1.Type_CODE_JSON
	case "GEOMETRY", "GEOGRAPHY":
		t.Code = runtimev1.Type_CODE_STRING
	case "NULL":
		t.Code = runtimev1.Type_CODE_UNSPECIFIED
	default:
		return nil, fmt.Errorf("unhandled snowflake type: %s", dbt)
	}
	return t, nil
}
