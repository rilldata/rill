package pgwire

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rilldata/rill/runtime"
	base "github.com/rilldata/rill/runtime/pkg/pgwire"

	// Register the in-memory DuckDB database/sql driver used for catalog emulation.
	_ "github.com/duckdb/duckdb-go/v2"
)

var (
	catalogNames = regexp.MustCompile(`(?i)\b(pg_catalog|pg_attribute|pg_class|pg_type|pg_namespace|pg_index|pg_constraint|information_schema)\b`)
	fromClause   = regexp.MustCompile(`(?i)\bFROM\b`)
	showVariable = regexp.MustCompile(`(?is)^\s*SHOW\s+(.+?)\s*;?\s*$`)

	privilegeFunctions = regexp.MustCompile(`pg_catalog\.(has_any_column_privilege|has_column_privilege|has_database_privilege|has_foreign_data_wrapper_privilege|has_function_privilege|has_language_privilege|has_parameter_privilege|has_schema_privilege|has_sequence_privilege|has_server_privilege|has_table_privilege|has_tablespace_privilege|has_type_privilege|pg_has_role)\(([^,]+), ([^,]+), ([^)]+)\)`)
	pgBackendPID       = regexp.MustCompile(`(?i)(pg_catalog\.)?pg_backend_pid\([^)]*\)`)
	pgGetIndexDef      = regexp.MustCompile(`(?i)(pg_catalog\.)?pg_get_indexdef\([^)]*\)`)
	pgVersion          = regexp.MustCompile(`(?i)(pg_catalog\.)?version\(\)`)
	serialSequence     = regexp.MustCompile(`(?i)pg_catalog\.pg_get_serial_sequence\([^)]*\)`)
	identityOptions    = regexp.MustCompile(`(?is)\(SELECT\s+json_build_object\([^)]*\)\s*FROM[^)]*\)\s+as\s+identity_options`)
)

func isCatalogQuery(query string) bool {
	trimmed := strings.TrimSpace(query)
	if trimmed == "-- ping" || showVariable.MatchString(trimmed) || catalogNames.MatchString(trimmed) {
		return true
	}
	upper := strings.ToUpper(trimmed)
	return strings.HasPrefix(upper, "SELECT") && !fromClause.MatchString(trimmed)
}

func queryCatalog(ctx context.Context, rt *runtime.Runtime, instanceID string, claims *runtime.SecurityClaims, query string, resultFormats []int16, types *pgtype.Map) (base.Rows, error) {
	if strings.TrimSpace(query) == "-- ping" {
		return &memoryRows{tag: "SELECT 0"}, nil
	}
	if matches := showVariable.FindStringSubmatch(query); len(matches) == 2 {
		return showResult(strings.TrimSpace(matches[1]), resultFormats, types)
	}

	db, err := sql.Open("duckdb", "")
	if err != nil {
		return nil, err
	}
	closeDB := true
	defer func() {
		if closeDB {
			_ = db.Close()
		}
	}()

	if _, err := db.ExecContext(ctx, "CREATE SCHEMA IF NOT EXISTS public"); err != nil {
		return nil, err
	}
	if _, err := db.ExecContext(ctx, "USE public"); err != nil {
		return nil, err
	}
	if err := createMetricsViewTables(ctx, db, rt, instanceID, claims); err != nil {
		return nil, err
	}
	// DuckDB lacks this PostgreSQL catalog relation, which is queried by both
	// Superset and Metabase during introspection. A temporary relation is used
	// because DuckDB does not permit creating objects in pg_catalog.
	if _, err := db.ExecContext(ctx, "CREATE TEMP TABLE pg_matviews(schemaname VARCHAR, matviewname VARCHAR, matviewowner VARCHAR, tablespace VARCHAR, hasindexes BOOLEAN, ispopulated BOOLEAN, definition VARCHAR)"); err != nil {
		return nil, err
	}

	rows, err := db.QueryContext(ctx, rewriteCatalogSQL(query))
	if err != nil {
		return nil, err
	}
	fields, err := fieldsForSQLRows(rows, resultFormats)
	if err != nil {
		_ = rows.Close()
		return nil, err
	}
	closeDB = false
	return &sqlRows{rows: rows, db: db, fields: fields, types: types}, nil
}

func createMetricsViewTables(ctx context.Context, db *sql.DB, rt *runtime.Runtime, instanceID string, claims *runtime.SecurityClaims) error {
	ctrl, err := rt.Controller(ctx, instanceID)
	if err != nil {
		return err
	}
	resources, err := ctrl.List(ctx, runtime.ResourceKindMetricsView, "", false)
	if err != nil {
		return err
	}
	sort.Slice(resources, func(i, j int) bool {
		return resources[i].Meta.Name.Name < resources[j].Meta.Name.Name
	})
	for _, resource := range resources {
		state := resource.GetMetricsView()
		if state == nil || state.State == nil || state.State.ValidSpec == nil {
			continue
		}
		security, err := rt.ResolveSecurity(ctx, instanceID, claims, resource)
		if err != nil {
			return err
		}
		if !security.CanAccess() {
			continue
		}

		spec := state.State.ValidSpec
		columns := make([]struct{ name, typ string }, 0, len(spec.Dimensions)+len(spec.Measures)+1)
		seen := make(map[string]bool)
		addColumn := func(name, typ string) {
			if name == "" || seen[name] || !security.CanAccessField(name) {
				return
			}
			seen[name] = true
			columns = append(columns, struct{ name, typ string }{name: name, typ: typ})
		}
		addColumn(spec.TimeDimension, "TIMESTAMPTZ")
		for _, dimension := range spec.Dimensions {
			addColumn(dimension.Name, "VARCHAR")
		}
		for _, measure := range spec.Measures {
			addColumn(measure.Name, "DOUBLE PRECISION")
		}
		if len(columns) == 0 {
			continue
		}

		var statement strings.Builder
		statement.WriteString("CREATE TABLE ")
		statement.WriteString(quoteIdentifier(resource.Meta.Name.Name))
		statement.WriteString(" (")
		for i, column := range columns {
			if i != 0 {
				statement.WriteString(", ")
			}
			statement.WriteString(quoteIdentifier(column.name))
			statement.WriteByte(' ')
			statement.WriteString(column.typ)
		}
		statement.WriteByte(')')
		if _, err := db.ExecContext(ctx, statement.String()); err != nil {
			return err
		}
	}
	return nil
}

func rewriteCatalogSQL(query string) string {
	query = strings.TrimSpace(strings.TrimSuffix(query, ";"))

	// Superset compatibility.
	query = strings.ReplaceAll(query, "ix.indrelid = c.conrelid and\n                                ix.indexrelid = c.conindid and\n                                c.contype in ('p', 'u', 'x')", "ix.indrelid = c.conrelid")
	query = strings.ReplaceAll(query, "t.oid = a.attrelid and a.attnum = ANY(ix.indkey)", "t.oid = a.attrelid")
	query = strings.ReplaceAll(query, "pg_get_constraintdef(cons.oid)", "pg_get_constraintdef(cons.oid, false)")
	query = strings.ReplaceAll(query, "pg_catalog.format_type(a.atttypid, a.atttypmod)", `CASE pg_catalog.format_type(a.atttypid, a.atttypmod)
		WHEN 'bool' THEN 'boolean'
		WHEN 'float4' THEN 'real'
		WHEN 'float8' THEN 'double precision'
		WHEN 'hugeint' THEN 'bigint'
		WHEN 'int2' THEN 'smallint'
		WHEN 'int4' THEN 'integer'
		WHEN 'int8' THEN 'bigint'
		WHEN 'timestamptz' THEN 'timestamp with time zone'
		WHEN 'timetz' THEN 'time with time zone'
		WHEN 'varchar' THEN 'character varying'
		ELSE pg_catalog.format_type(a.atttypid, a.atttypmod)
	END`)

	if strings.EqualFold(query, "SELECT nspname FROM pg_namespace WHERE nspname NOT LIKE 'pg_%' ORDER BY nspname") {
		query = "SELECT nspname FROM pg_namespace WHERE nspname NOT IN ('pg_catalog', 'information_schema', 'main') ORDER BY nspname"
	}

	// Metabase compatibility.
	query = strings.ReplaceAll(query, "t.schemaname <> 'information_schema'", "t.schemaname <> 'information_schema' AND t.schemaname <> 'pg_catalog' AND t.schemaname <> 'main'")
	query = strings.ReplaceAll(query, "(information_schema._pg_expandarray(i.indkey)).n", "generate_subscripts(i.indkey, 1)")
	query = privilegeFunctions.ReplaceAllString(query, `(select pg_catalog.$1($3, $4))`)
	query = strings.ReplaceAll(query, "pg_catalog.pg_matviews", "pg_matviews")
	query = serialSequence.ReplaceAllString(query, "NULL")
	query = pgBackendPID.ReplaceAllString(query, `(SELECT 1234) AS pg_backend_pid`)
	query = pgGetIndexDef.ReplaceAllString(query, "NULL")
	query = pgVersion.ReplaceAllString(query, `(SELECT 'PostgreSQL 16.3 (Rill pgwire)') AS version`)
	query = strings.ReplaceAll(query, `'pg_class'::regclass`, `(SELECT oid FROM pg_class WHERE relname = 'pg_class')`)
	query = identityOptions.ReplaceAllString(query, "NULL AS identity_options")
	return query
}

func showResult(variable string, resultFormats []int16, types *pgtype.Map) (base.Rows, error) {
	name := strings.ToLower(strings.TrimSpace(variable))
	column := name
	value := ""
	switch name {
	case "standard_conforming_string", "standard_conforming_strings":
		value = "on"
	case "transaction isolation level", "transaction_isolation":
		column = "transaction_isolation"
		value = "read committed"
	case "timezone", "time zone":
		column = "TimeZone"
		value = "Etc/UTC"
	case "server_version":
		value = "16.3"
	case "search_path":
		value = `"$user", public`
	default:
		value = ""
	}
	format, err := resultFormat(resultFormats, 0, 1)
	if err != nil {
		return nil, err
	}
	encoded, err := types.Encode(pgtype.TextOID, format, value, nil)
	if err != nil {
		return nil, err
	}
	return &memoryRows{
		fields: []pgproto3.FieldDescription{{Name: []byte(column), DataTypeOID: pgtype.TextOID, DataTypeSize: -1, TypeModifier: -1, Format: format}},
		rows:   [][][]byte{{encoded}},
		tag:    "SHOW",
	}, nil
}

func fieldsForSQLRows(rows *sql.Rows, resultFormats []int16) ([]pgproto3.FieldDescription, error) {
	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}
	fields := make([]pgproto3.FieldDescription, len(columnTypes))
	for i, column := range columnTypes {
		oid, size := postgresTypeForDatabaseName(column.DatabaseTypeName())
		format, err := resultFormat(resultFormats, i, len(columnTypes))
		if err != nil {
			return nil, err
		}
		fields[i] = pgproto3.FieldDescription{Name: []byte(column.Name()), DataTypeOID: oid, DataTypeSize: size, TypeModifier: -1, Format: format}
	}
	return fields, nil
}

func postgresTypeForDatabaseName(name string) (uint32, int16) {
	name = strings.ToUpper(name)
	switch {
	case strings.Contains(name, "BOOL"):
		return pgtype.BoolOID, 1
	case strings.Contains(name, "SMALLINT") || strings.Contains(name, "TINYINT"):
		return pgtype.Int2OID, 2
	case name == "INTEGER" || name == "INT":
		return pgtype.Int4OID, 4
	case strings.Contains(name, "BIGINT") || strings.Contains(name, "HUGEINT"):
		return pgtype.Int8OID, 8
	case strings.Contains(name, "DOUBLE") || strings.Contains(name, "DECIMAL") || strings.Contains(name, "NUMERIC"):
		return pgtype.Float8OID, 8
	case strings.Contains(name, "FLOAT") || strings.Contains(name, "REAL"):
		return pgtype.Float4OID, 4
	case strings.Contains(name, "TIMESTAMP"):
		return pgtype.TimestamptzOID, 8
	case name == "DATE":
		return pgtype.DateOID, 4
	case name == "TIME":
		return pgtype.TimeOID, 8
	case strings.Contains(name, "BLOB") || strings.Contains(name, "BYTE"):
		return pgtype.ByteaOID, -1
	default:
		return pgtype.TextOID, -1
	}
}

type sqlRows struct {
	rows   *sql.Rows
	db     *sql.DB
	fields []pgproto3.FieldDescription
	types  *pgtype.Map
	values [][]byte
	err    error
	count  int64
}

func (r *sqlRows) Fields() []pgproto3.FieldDescription { return r.fields }
func (r *sqlRows) Next() bool {
	if !r.rows.Next() {
		r.err = r.rows.Err()
		return false
	}
	values := make([]any, len(r.fields))
	dest := make([]any, len(values))
	for i := range values {
		dest[i] = &values[i]
	}
	if err := r.rows.Scan(dest...); err != nil {
		r.err = err
		return false
	}
	r.values = make([][]byte, len(values))
	for i, value := range values {
		var err error
		value, err = normalizeValue(value, r.fields[i].DataTypeOID)
		if err != nil {
			r.err = err
			return false
		}
		r.values[i], err = encodeValue(r.types, r.fields[i].DataTypeOID, r.fields[i].Format, value)
		if err != nil {
			r.err = err
			return false
		}
	}
	r.count++
	return true
}
func (r *sqlRows) Values() [][]byte { return r.values }
func (r *sqlRows) Err() error       { return r.err }
func (r *sqlRows) CommandTag() string {
	return fmt.Sprintf("SELECT %d", r.count)
}

func (r *sqlRows) Close() error {
	rowsErr := r.rows.Close()
	dbErr := r.db.Close()
	if rowsErr != nil {
		return rowsErr
	}
	return dbErr
}

type memoryRows struct {
	fields []pgproto3.FieldDescription
	rows   [][][]byte
	tag    string
	index  int
}

func (r *memoryRows) Fields() []pgproto3.FieldDescription { return r.fields }
func (r *memoryRows) Next() bool {
	if r.index >= len(r.rows) {
		return false
	}
	r.index++
	return true
}

func (r *memoryRows) Values() [][]byte {
	return r.rows[r.index-1]
}
func (r *memoryRows) Err() error         { return nil }
func (r *memoryRows) CommandTag() string { return r.tag }
func (r *memoryRows) Close() error       { return nil }

func quoteIdentifier(value string) string {
	return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
}
