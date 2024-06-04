package psql

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/jmoiron/sqlx"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers/duckdb"
	"go.uber.org/zap"

	// import duckdb driver
	_ "github.com/marcboeker/go-duckdb"
)

var (
	errForbidden = errors.New("access to metrics view is forbidden")

	functions = "has_any_column_privilege|has_column_privilege|has_database_privilege|has_foreign_data_wrapper_privilege|has_function_privilege|has_language_privilege|has_parameter_privilege|has_schema_privilege|has_sequence_privilege|has_server_privilege|has_table_privilege|has_tablespace_privilege|has_type_privilege|pg_has_role"

	re                = regexp.MustCompile(fmt.Sprintf(`pg_catalog.(%s)\(([^,]+), ([^,]+), ([^)]+)\)`, functions))
	dbRe              = regexp.MustCompile(`pg_catalog\.(\w+)`)
	regclassRe        = regexp.MustCompile(`'pg_class'::regclass`)
	versionRe         = regexp.MustCompile(`(?i)(pg_catalog\.)?version\(\)`)
	pgBackendPid      = regexp.MustCompile(`(?:pg_catalog\.)?pg_backend_pid\([^)]*\)`)
	indexRe           = regexp.MustCompile(`(?:pg_catalog\.)?pg_get_indexdef\([^)]*\)`)
	identifyOptionsRe = regexp.MustCompile(`(?is)\(SELECT\s+json_build_object\([^)]*\)\s*FROM[^)]*\)\s+as\s+identity_options`)
	serialSequenceRe  = regexp.MustCompile(`pg_catalog\.pg_get_serial_sequence\([^\)]*\)`)
	extraCharRe       = regexp.MustCompile(`[\n\t\r]`)
	showVarRe         = regexp.MustCompile(`(?i)SHOW\s+(.+)`)
)

type PSQLQueryOpts struct {
	SQL            string
	Runtime        *runtime.Runtime
	InstanceID     string
	UserAttributes map[string]any
	Priority       int
	Logger         *zap.Logger
}

func ResolvePSQLQuery(ctx context.Context, opts *PSQLQueryOpts) ([][]any, *runtimev1.StructType, error) {
	// various hacks to make postgres query compatible with a duckdb query
	sqlStr := rewriteSQL(opts.SQL)

	// check if it is a non catalog query like `SHOW variable`
	sqlStr = strings.TrimSuffix(sqlStr, ";")
	matches := showVarRe.FindStringSubmatch(sqlStr)
	if len(matches) > 1 {
		return handleShowVariableQuery(matches[1])
	}

	ctrl, err := opts.Runtime.Controller(ctx, opts.InstanceID)
	if err != nil {
		return nil, nil, err
	}

	resources, err := ctrl.List(ctx, runtime.ResourceKindMetricsView, "", false)
	if err != nil {
		return nil, nil, err
	}

	db, err := sqlx.Open("duckdb", "")
	if err != nil {
		return nil, nil, err
	}

	// postgres's default schema is public
	if _, err := db.ExecContext(ctx, "CREATE SCHEMA IF NOT EXISTS public"); err != nil {
		return nil, nil, err
	}

	if _, err := db.ExecContext(ctx, "USE public"); err != nil {
		return nil, nil, err
	}

	// loop over all resources and create corresponding table in duckdb so that these can be queried with information_schema
	for _, resource := range resources {
		cols, err := colsForMetricView(ctrl, opts, resource)
		if err != nil {
			if errors.Is(err, errForbidden) {
				continue
			}
			return nil, nil, err
		}

		olap, release, err := ctrl.Runtime.OLAP(ctx, opts.InstanceID, resource.GetMetricsView().Spec.Connector)
		if err != nil {
			return nil, nil, err
		}

		mvSpec := resource.GetMetricsView().Spec
		tbl, err := olap.InformationSchema().Lookup(ctx, mvSpec.Database, mvSpec.DatabaseSchema, mvSpec.Table)
		if err != nil {
			release()
			return nil, nil, err
		}
		release()

		// generate create table query
		var sb strings.Builder
		sb.WriteString("CREATE TABLE ")
		sb.WriteString(olap.Dialect().EscapeIdentifier(resource.Meta.Name.Name))
		sb.WriteString("(")
		colWritten := false
		for _, field := range tbl.Schema.Fields {
			if _, ok := cols[field.Name]; !ok { // columns is excluded from querying
				continue
			}
			if colWritten {
				sb.WriteString(", ")
			}
			colWritten = true
			sb.WriteString(olap.Dialect().EscapeIdentifier(field.Name))
			sb.WriteString(" ")
			typ, err := pbTypeToDuckDB(field.Type)
			if err != nil {
				release()
				return nil, nil, err
			}
			sb.WriteString(typ)
		}
		sb.WriteString(")")

		_, err = db.ExecContext(ctx, sb.String())
		if err != nil {
			return nil, nil, err
		}
	}

	// create postgres catalog tables missing in duckdb
	if err := populateCatalogTables(ctx, db); err != nil {
		return nil, nil, err
	}

	rows, err := db.QueryxContext(ctx, sqlStr)
	if err != nil {
		return nil, nil, err
	}

	schema, err := duckdb.RowsToSchema(rows)
	if err != nil {
		return nil, nil, err
	}

	var data [][]any
	for rows.Next() {
		row, err := rows.SliceScan()
		if err != nil {
			return nil, nil, err
		}

		data = append(data, row)
	}
	return data, schema, nil
}

func populateCatalogTables(ctx context.Context, db *sqlx.DB) error {
	// duckdb redirects all catalog queries to system db by default so need to append db name which is memory for in-memory duckdb
	_, err := db.ExecContext(ctx, "CREATE TABLE memory.pg_catalog.pg_matviews(schemaname VARCHAR, matviewname VARCHAR, matviewowner VARCHAR, tablespace VARCHAR, hasindexes BOOLEAN, ispopulated BOOLEAN, definition VARCHAR)")
	return err
}

// colsForMetricView returns columns available for query taking security policies into account
func colsForMetricView(ctrl *runtime.Controller, opts *PSQLQueryOpts, mv *runtimev1.Resource) (map[string]any, error) {
	spec := mv.GetMetricsView().State.ValidSpec
	if spec == nil {
		return nil, fmt.Errorf("metrics view %q is not ready for querying, reconcile status: %q", mv.Meta.GetName(), mv.Meta.ReconcileStatus)
	}

	security, err := ctrl.Runtime.ResolveMetricsViewSecurity(opts.UserAttributes, opts.InstanceID, spec, mv.Meta.StateUpdatedOn.AsTime())
	if err != nil {
		return nil, err
	}

	cols := make(map[string]any)
	for _, measure := range spec.Measures {
		cols[measure.Name] = nil
	}
	for _, dim := range spec.Dimensions {
		cols[dim.Name] = nil
	}

	if security == nil {
		if spec.TimeDimension != "" {
			cols[spec.TimeDimension] = nil
		}
		return cols, nil
	}

	if !security.Access || security.ExcludeAll {
		return nil, errForbidden
	}

	final := make(map[string]any)
	if len(security.Include) != 0 {
		for measure := range cols {
			if slices.Contains(security.Include, measure) {
				final[measure] = nil
			}
		}
	}
	if len(final) > 0 { // only include the included cols if include is set
		cols = final
	}

	for col := range cols { // remove all excluded cols
		if !slices.Contains(security.Exclude, col) {
			final[col] = nil
		}
	}

	if spec.TimeDimension != "" {
		final[spec.TimeDimension] = nil
	}

	return final, nil
}

func rewriteSQL(sql string) string {
	sql = extraCharRe.ReplaceAllString(sql, "\n")

	// hacks for working with superset
	sql = strings.ReplaceAll(sql, "ix.indrelid = c.conrelid and\n                                ix.indexrelid = c.conindid and\n                                c.contype in ('p', 'u', 'x')", "ix.indrelid = c.conrelid")
	sql = strings.ReplaceAll(sql, "t.oid = a.attrelid and a.attnum = ANY(ix.indkey)", "t.oid = a.attrelid")
	sql = strings.ReplaceAll(sql, "pg_get_constraintdef(cons.oid)", "pg_get_constraintdef(cons.oid, false)")

	// duckdb reports type hugeint postgres supports bigint
	sql = strings.ReplaceAll(sql, "pg_catalog.format_type(a.atttypid, a.atttypmod)", "CASE WHEN pg_catalog.format_type(a.atttypid, a.atttypmod) == 'hugeint' THEN 'bigint' ELSE pg_catalog.format_type(a.atttypid, a.atttypmod) END")

	if sql == "SELECT nspname FROM pg_namespace WHERE nspname NOT LIKE 'pg_%' ORDER BY nspname" {
		sql = "SELECT nspname FROM pg_namespace WHERE nspname NOT IN ('pg_catalog', 'information_schema', 'main') ORDER BY nspname"
	}

	// hacks for working with metabase
	sql = strings.ReplaceAll(sql, "t.schemaname <> 'information_schema'", "t.schemaname <> 'information_schema' AND t.schemaname <> 'pg_catalog' AND t.schemaname <> 'main'")
	sql = strings.ReplaceAll(sql, "(information_schema._pg_expandarray(i.indkey)).n", "generate_subscripts(i.indkey, 1)")

	// DuckDB does not support user optional argument in `functions`. We need to remove that.
	sql = re.ReplaceAllString(sql, `(select memory.pg_catalog.$1($3, $4))`)
	// pg_get_serial_sequence not supported
	sql = serialSequenceRe.ReplaceAllString(sql, "NULL")
	// setting fixed pg_backend_pid
	sql = pgBackendPid.ReplaceAllString(sql, `(SELECT 1234) AS pg_backend_pid`)
	// pg_get_indexdef not supported
	sql = indexRe.ReplaceAllString(sql, "NULL")
	// postgres version
	sql = versionRe.ReplaceAllString(sql, `(SELECT 'PostgreSQL 16.3 (Debian 16.3-1.pgdg120+1) on aarch64-unknown-linux-gnu, compiled by gcc (Debian 12.2.0-14) 12.2.0, 64-bit') AS version`)
	// duckdb executes catalog queries in system schema by default. We want to execute in user database's public schema.
	sql = dbRe.ReplaceAllString(sql, `pg_catalog.$1`)
	// duckdb does not have `regclass` typecast
	sql = regclassRe.ReplaceAllString(sql, `(SELECT oid FROM pg_class WHERE relname = 'pg_class')`)
	// json_build_object is not supported. It is used in indexes for metabase so we directly set it as NULL.
	sql = identifyOptionsRe.ReplaceAllString(sql, " NULL AS identity_options")
	return sql
}

func handleShowVariableQuery(variable string) ([][]any, *runtimev1.StructType, error) {
	fields := make([]*runtimev1.StructType_Field, 1)
	fields[0] = &runtimev1.StructType_Field{
		Name: name(variable),
		Type: &runtimev1.Type{Code: runtimev1.Type_CODE_STRING, Nullable: false},
	}

	row := make([][]any, 1)
	row[0] = make([]any, 1)
	row[0][0] = value(variable)
	return row, &runtimev1.StructType{Fields: fields}, nil
}

func name(variable string) string {
	switch strings.ToLower(variable) {
	case "transaction isolation level":
		return "transaction_isolation"
	default:
		return variable
	}
}

func value(variable string) string {
	switch strings.ToLower(variable) {
	case "standard_conforming_string", "standard_conforming_strings":
		return "on"
	case "transaction isolation level":
		return "read committed"
	case "timezone":
		return "Etc/UTC"
	default:
		return "tbd"
	}
}

func pbTypeToDuckDB(t *runtimev1.Type) (string, error) {
	code := t.Code
	switch code {
	case runtimev1.Type_CODE_UNSPECIFIED:
		return "", fmt.Errorf("unspecified code")
	case runtimev1.Type_CODE_BOOL:
		return "BOOLEAN", nil
	case runtimev1.Type_CODE_INT8:
		return "TINYINT", nil
	case runtimev1.Type_CODE_INT16:
		return "SMALLINT", nil
	case runtimev1.Type_CODE_INT32:
		return "INTEGER", nil
	case runtimev1.Type_CODE_INT64:
		return "BIGINT", nil
	case runtimev1.Type_CODE_INT128:
		return "HUGEINT", nil
	case runtimev1.Type_CODE_UINT8:
		return "UTINYINT", nil
	case runtimev1.Type_CODE_UINT16:
		return "USMALLINT", nil
	case runtimev1.Type_CODE_UINT32:
		return "UINTEGER", nil
	case runtimev1.Type_CODE_UINT64:
		return "UBIGINT", nil
	case runtimev1.Type_CODE_FLOAT32:
		return "FLOAT", nil
	case runtimev1.Type_CODE_FLOAT64:
		return "DOUBLE", nil
	case runtimev1.Type_CODE_TIMESTAMP:
		return "TIMESTAMP", nil
	case runtimev1.Type_CODE_DATE:
		return "DATE", nil
	case runtimev1.Type_CODE_TIME:
		return "TIME", nil
	case runtimev1.Type_CODE_STRING:
		return "VARCHAR", nil
	case runtimev1.Type_CODE_BYTES:
		return "BLOB", nil
	case runtimev1.Type_CODE_ARRAY:
		return "", fmt.Errorf("array is not supported")
	case runtimev1.Type_CODE_STRUCT:
		return "", fmt.Errorf("struct is not supported")
	case runtimev1.Type_CODE_MAP:
		return "", fmt.Errorf("map is not supported")
	case runtimev1.Type_CODE_DECIMAL:
		return "DECIMAL", nil
	case runtimev1.Type_CODE_JSON:
		// keeping type as json but appending varchar using the appender API causes duckdb invalid vector error intermittently
		return "VARCHAR", nil
	case runtimev1.Type_CODE_UUID:
		return "UUID", nil
	default:
		return "", fmt.Errorf("unknown type_code %s", code)
	}
}
