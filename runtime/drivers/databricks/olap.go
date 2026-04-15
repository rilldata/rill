package databricks

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
	return drivers.DialectDatabricks
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
	if c.config.LogQueries {
		fields := []zap.Field{
			zap.String("sql", c.Dialect().SanitizeQueryForLogging(stmt.Query)),
			zap.Any("args", stmt.Args),
			observability.ZapCtx(ctx),
		}
		if len(stmt.QueryAttributes) > 0 {
			fields = append(fields, zap.Any("query_attributes", stmt.QueryAttributes))
		}
		c.logger.Info("databricks query", fields...)
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

	return &drivers.Result{Rows: rows, Schema: schema}, nil
}

// QuerySchema implements drivers.OLAPStore.
func (c *connection) QuerySchema(ctx context.Context, query string, args []any) (*runtimev1.StructType, error) {
	if c.config.LogQueries {
		c.logger.Info("databricks query schema", zap.String("sql", c.Dialect().SanitizeQueryForLogging(query)), zap.Any("args", args), observability.ZapCtx(ctx))
	}

	db, err := c.getDB(ctx)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, drivers.DefaultQuerySchemaTimeout)
	defer cancel()

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

// All implements drivers.OLAPInformationSchema.
func (c *connection) All(ctx context.Context, like string, pageSize uint32, pageToken string) ([]*drivers.OlapTable, string, error) {
	return drivers.AllFromInformationSchema(ctx, like, pageSize, pageToken, c)
}

// LoadPhysicalSize implements drivers.OLAPInformationSchema.
func (c *connection) LoadPhysicalSize(ctx context.Context, tables []*drivers.OlapTable) error {
	return nil
}

// LoadDDL implements drivers.OLAPInformationSchema.
func (c *connection) LoadDDL(ctx context.Context, table *drivers.OlapTable) error {
	db, err := c.getDB(ctx)
	if err != nil {
		return err
	}

	fqn := drivers.DialectDatabricks.EscapeTable(table.Database, table.DatabaseSchema, table.Name)

	objectType := "TABLE"
	if table.View {
		objectType = "VIEW"
	}

	var ddl string
	err = db.QueryRowContext(ctx, fmt.Sprintf("SHOW CREATE %s %s", objectType, fqn)).Scan(&ddl)
	if err != nil {
		return err
	}
	table.DDL = ddl
	return nil
}

// Lookup implements drivers.OLAPInformationSchema.
func (c *connection) Lookup(ctx context.Context, db, schema, name string) (*drivers.OlapTable, error) {
	meta, err := c.GetTable(ctx, db, schema, name)
	if err != nil {
		return nil, err
	}

	rtSchema := &runtimev1.StructType{}
	for colName, colType := range meta.Schema {
		t := databaseTypeToPB(colType)
		rtSchema.Fields = append(rtSchema.Fields, &runtimev1.StructType_Field{
			Name: colName,
			Type: t,
		})
	}
	return &drivers.OlapTable{
		Database:       db,
		DatabaseSchema: schema,
		Name:           name,
		View:           meta.View,
		Schema:         rtSchema,
	}, nil
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
		t := databaseTypeToPB(ct.DatabaseTypeName())
		fields[i] = &runtimev1.StructType_Field{
			Name: ct.Name(),
			Type: t,
		}
	}
	return &runtimev1.StructType{Fields: fields}, nil
}

func databaseTypeToPB(dbt string) *runtimev1.Type {
	t := &runtimev1.Type{Nullable: true}
	switch strings.ToUpper(dbt) {
	case "BOOLEAN":
		t.Code = runtimev1.Type_CODE_BOOL
	case "TINYINT", "SMALLINT", "INT", "BIGINT":
		t.Code = runtimev1.Type_CODE_INT64
	case "FLOAT":
		t.Code = runtimev1.Type_CODE_FLOAT64
	case "DOUBLE":
		t.Code = runtimev1.Type_CODE_FLOAT64
	case "DECIMAL", "DEC", "NUMERIC":
		t.Code = runtimev1.Type_CODE_DECIMAL
	case "STRING", "VARCHAR", "CHAR":
		t.Code = runtimev1.Type_CODE_STRING
	case "BINARY", "VARBINARY":
		t.Code = runtimev1.Type_CODE_BYTES
	case "DATE":
		t.Code = runtimev1.Type_CODE_DATE
	case "TIMESTAMP", "TIMESTAMP_NTZ":
		t.Code = runtimev1.Type_CODE_TIMESTAMP
	case "ARRAY", "MAP", "STRUCT", "VARIANT":
		t.Code = runtimev1.Type_CODE_JSON
	case "INTERVAL":
		t.Code = runtimev1.Type_CODE_INTERVAL
	default:
		t.Code = runtimev1.Type_CODE_UNSPECIFIED
	}
	return t
}
