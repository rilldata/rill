package databricks

import (
	"context"
	"fmt"
	"strings"

	"github.com/databricks/databricks-sql-go/driverctx"
	"github.com/jmoiron/sqlx"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/zap"
)

var _ drivers.OLAPStore = (*connection)(nil)

// queryTagsConf is the statement configuration the Databricks driver uses to carry
// query tags. It appears in the error returned by workspaces that don't support them.
const queryTagsConf = "query_tags"

// Dialect implements drivers.OLAPStore.
func (c *connection) Dialect() drivers.Dialect {
	return DialectDatabricks
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

	// Send the query attributes as Databricks query tags, which are recorded in the
	// query_tags column of system.query.history. The driver attaches them per statement
	// (as a ConfOverlay on the ExecuteStatement request) rather than per session, so they
	// stay correct even though connections are pooled and shared across users.
	taggedCtx := c.contextWithQueryTags(ctx, stmt.QueryAttributes)

	db, err := c.getDB(ctx)
	if err != nil {
		return nil, err
	}

	if stmt.DryRun {
		_, err = db.ExecContext(taggedCtx, fmt.Sprintf("EXPLAIN %s", stmt.Query), stmt.Args...)
		if c.queryTagsRejected(err) {
			_, err = db.ExecContext(ctx, fmt.Sprintf("EXPLAIN %s", stmt.Query), stmt.Args...)
		}
		return nil, err
	}

	rows, err := db.QueryxContext(taggedCtx, stmt.Query, stmt.Args...)
	if c.queryTagsRejected(err) {
		// The workspace doesn't support query tags. Retry untagged so a metrics view
		// that sets query_attributes still works there.
		rows, err = db.QueryxContext(ctx, stmt.Query, stmt.Args...)
	}
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

// Head implements drivers.OLAPStore.
func (c *connection) Head(ctx context.Context, db, schema, table string, limit int64) (*drivers.Result, error) {
	tbl, err := c.InformationSchema().Lookup(ctx, db, schema, table)
	if err != nil {
		return nil, err
	}

	var columns []string
	for _, field := range tbl.Schema.Fields {
		columns = append(columns, c.Dialect().EscapeIdentifier(field.Name))
	}

	q := fmt.Sprintf("SELECT %s FROM %s", strings.Join(columns, ", "), c.Dialect().EscapeTable(db, schema, table))
	if limit > 0 {
		q += fmt.Sprintf(" LIMIT %d", limit)
	}
	return c.Query(ctx, &drivers.Statement{Query: q})
}

// contextWithQueryTags returns a context carrying attrs as Databricks query tags.
// The context is returned unchanged when there are no attributes, or when this
// connection has already seen the workspace reject query tags.
func (c *connection) contextWithQueryTags(ctx context.Context, attrs map[string]string) context.Context {
	if len(attrs) == 0 || c.queryTagsUnsupported.Load() {
		return ctx
	}
	// Copy so a later mutation of the statement's map can't race with the driver
	// reading it while the query is in flight.
	tags := make(map[string]string, len(attrs))
	for k, v := range attrs {
		tags[k] = v
	}
	return driverctx.NewContextWithQueryTags(ctx, tags)
}

// queryTagsRejected reports whether err is the workspace refusing the query_tags
// configuration, meaning the query should be retried without tags.
//
// Query tags are sent as a statement configuration, and Databricks fails the whole
// statement when it doesn't recognise a configuration rather than ignoring it:
//
//	[CONFIG_NOT_AVAILABLE.WITHOUT_SUGGESTION] Configuration query_tags is not available.
//
// Without this, enabling query_attributes on a metrics view would break every query
// against a workspace that doesn't have the (Public Preview) query tags feature. The
// result is latched on the connection so only the first query pays for the retry.
func (c *connection) queryTagsRejected(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	if !strings.Contains(msg, "CONFIG_NOT_AVAILABLE") || !strings.Contains(msg, queryTagsConf) {
		return false
	}
	if c.queryTagsUnsupported.CompareAndSwap(false, true) {
		c.logger.Warn("databricks: workspace rejected query tags, dropping query_attributes for this connection", zap.Error(err))
	}
	return true
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
	t := &runtimev1.Type{Nullable: true, RawType: dbt}
	// Strip type parameters: e.g. "DECIMAL(18,6)" → "DECIMAL", "ARRAY<INT>" → "ARRAY"
	upper := strings.ToUpper(dbt)
	if i := strings.IndexAny(upper, "(<"); i != -1 {
		upper = upper[:i]
	}
	switch upper {
	case "BOOLEAN":
		t.Code = runtimev1.Type_CODE_BOOL
	case "TINYINT", "BYTE":
		t.Code = runtimev1.Type_CODE_INT8
	case "SMALLINT", "SHORT":
		t.Code = runtimev1.Type_CODE_INT16
	case "INT":
		t.Code = runtimev1.Type_CODE_INT32
	case "BIGINT", "LONG":
		t.Code = runtimev1.Type_CODE_INT64
	case "FLOAT":
		t.Code = runtimev1.Type_CODE_FLOAT32
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
