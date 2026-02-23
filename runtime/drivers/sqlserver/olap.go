package sqlserver

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
	return drivers.DialectSQLServer
}

// Exec implements drivers.OLAPStore.
func (c *connection) Exec(ctx context.Context, stmt *drivers.Statement) error {
	if stmt.DryRun {
		return nil
	}
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
func (c *connection) Query(ctx context.Context, stmt *drivers.Statement) (res *drivers.Result, resErr error) {
	if c.logQueries {
		fields := []zap.Field{
			zap.String("sql", c.Dialect().SanitizeQueryForLogging(stmt.Query)),
			zap.Any("args", stmt.Args),
			observability.ZapCtx(ctx),
		}
		if len(stmt.QueryAttributes) > 0 {
			fields = append(fields, zap.Any("query_attributes", stmt.QueryAttributes))
		}
		c.logger.Info("SQL Server query", fields...)
	}

	db, err := c.getDB(ctx)
	if err != nil {
		return nil, err
	}

	if stmt.DryRun {
		// SQL Server does not support EXPLAIN; use SET FMTONLY ON as a dry run alternative
		_, err = db.ExecContext(ctx, fmt.Sprintf("SET FMTONLY ON; %s; SET FMTONLY OFF", stmt.Query), stmt.Args...)
		return nil, err
	}

	rows, err := db.QueryxContext(ctx, stmt.Query, stmt.Args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		if resErr != nil {
			_ = rows.Close()
		}
	}()

	schema, err := rowsToSchema(rows)
	if err != nil {
		return nil, err
	}

	res = &drivers.Result{Rows: rows, Schema: schema}
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

// LoadDDL implements drivers.OLAPInformationSchema.
func (c *connection) LoadDDL(ctx context.Context, table *drivers.OlapTable) error {
	// SQL Server does not have a simple SHOW CREATE TABLE equivalent.
	// Reconstructing DDL from metadata is complex; leaving unimplemented for now.
	return nil
}

// Lookup implements drivers.OLAPInformationSchema.
func (c *connection) Lookup(ctx context.Context, db, schema, name string) (*drivers.OlapTable, error) {
	meta, err := c.GetTable(ctx, db, schema, name)
	if err != nil {
		return nil, err
	}

	rtSchema := &runtimev1.StructType{}
	for colName, typ := range meta.Schema {
		rtSchema.Fields = append(rtSchema.Fields, &runtimev1.StructType_Field{
			Name: colName,
			Type: databaseTypeToPB(typ),
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
		fields[i] = &runtimev1.StructType_Field{
			Name: ct.Name(),
			Type: databaseTypeToPB(ct.DatabaseTypeName()),
		}
	}
	return &runtimev1.StructType{Fields: fields}, nil
}

func databaseTypeToPB(dbt string) *runtimev1.Type {
	t := &runtimev1.Type{Nullable: true}
	switch strings.ToUpper(dbt) {
	case "BIT":
		t.Code = runtimev1.Type_CODE_BOOL
	case "TINYINT":
		t.Code = runtimev1.Type_CODE_INT8
	case "SMALLINT":
		t.Code = runtimev1.Type_CODE_INT16
	case "INT":
		t.Code = runtimev1.Type_CODE_INT32
	case "BIGINT":
		t.Code = runtimev1.Type_CODE_INT64
	case "REAL":
		t.Code = runtimev1.Type_CODE_FLOAT64
	case "FLOAT":
		t.Code = runtimev1.Type_CODE_FLOAT64
	case "DECIMAL", "NUMERIC", "MONEY", "SMALLMONEY":
		t.Code = runtimev1.Type_CODE_DECIMAL
	case "CHAR", "VARCHAR", "TEXT", "NCHAR", "NVARCHAR", "NTEXT":
		t.Code = runtimev1.Type_CODE_STRING
	case "BINARY", "VARBINARY", "IMAGE":
		t.Code = runtimev1.Type_CODE_BYTES
	case "DATE":
		t.Code = runtimev1.Type_CODE_DATE
	case "TIME":
		t.Code = runtimev1.Type_CODE_TIME
	case "DATETIME", "DATETIME2", "SMALLDATETIME":
		t.Code = runtimev1.Type_CODE_TIMESTAMP
	case "DATETIMEOFFSET":
		t.Code = runtimev1.Type_CODE_TIMESTAMP
	case "UNIQUEIDENTIFIER":
		t.Code = runtimev1.Type_CODE_UUID
	case "XML":
		t.Code = runtimev1.Type_CODE_STRING
	default:
		t.Code = runtimev1.Type_CODE_UNSPECIFIED
	}
	return t
}
