package pgwire

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/duckdb/duckdb-go/v2"
	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/jackc/pgx/v5/pgtype"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	base "github.com/rilldata/rill/runtime/pkg/pgwire"
)

var showVariable = regexp.MustCompile(`(?is)^\s*SHOW\s+(.+?)\s*;?\s*$`)

func (s *Session) queryCatalog(ctx context.Context, parsed *parsedSQL, parameters []base.Parameter, resultFormats []int16, describe bool) (base.Rows, error) {
	if strings.TrimSpace(parsed.text) == "-- ping" {
		return &memoryRows{tag: "SELECT 0"}, nil
	}
	if matches := showVariable.FindStringSubmatch(parsed.text); len(matches) == 2 {
		return showResult(strings.TrimSpace(matches[1]), resultFormats, s.types)
	}
	query, args, err := parsed.bindCatalogParameters(parameters, s.types)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("duckdb", "?enable_external_access=false")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
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
	if err := s.createMetricsViewTables(ctx, db); err != nil {
		return nil, err
	}
	// DuckDB lacks this PostgreSQL catalog relation, which is queried by both
	// Superset and Metabase during introspection. A temporary relation is used
	// because DuckDB does not permit creating objects in pg_catalog.
	if _, err := db.ExecContext(ctx, "CREATE TEMP TABLE pg_matviews(schemaname VARCHAR, matviewname VARCHAR, matviewowner VARCHAR, tablespace VARCHAR, hasindexes BOOLEAN, ispopulated BOOLEAN, definition VARCHAR)"); err != nil {
		return nil, err
	}

	// Lock configuration before accepting client SQL. Wrapping the query also
	// lets the SQL parser enforce a read-only SELECT, including for CTEs.
	if _, err := db.ExecContext(ctx, "SET lock_configuration=true"); err != nil {
		return nil, err
	}
	query, err = rewriteCatalogSQL(query)
	if err != nil {
		return nil, err
	}
	query = "SELECT * FROM (\n" + query + "\n) AS catalog_query"
	if describe {
		query += " LIMIT 0"
	}
	conn, err := db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	err = conn.Raw(func(raw any) error {
		// Prepare (without Context) rejects multiple statements without executing
		// any of them. PrepareContext executes preceding statements in this driver.
		stmt, err := raw.(*duckdb.Conn).Prepare(query)
		if err != nil {
			return err
		}
		defer stmt.Close()
		typ, err := stmt.(*duckdb.Stmt).StatementType()
		if err != nil {
			return err
		}
		if typ != duckdb.STATEMENT_TYPE_SELECT {
			return &base.Error{Code: "0A000", Message: "only catalog SELECT statements are supported"}
		}
		return nil
	})
	_ = conn.Close()
	if err != nil {
		return nil, err
	}
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	fields, err := fieldsForSQLRows(rows, resultFormats)
	if err != nil {
		_ = rows.Close()
		return nil, err
	}
	closeDB = false
	return &sqlRows{rows: rows, db: db, fields: fields, types: s.types}, nil
}

func (s *Session) createMetricsViewTables(ctx context.Context, db *sql.DB) error {
	ctrl, err := s.runtime.Controller(ctx, s.instanceID)
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
		security, err := s.runtime.ResolveSecurity(ctx, s.instanceID, s.claims, resource)
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
		for _, dimension := range spec.Dimensions {
			addColumn(dimension.Name, catalogColumnType(dimension.DataType))
		}
		addColumn(spec.TimeDimension, "TIMESTAMPTZ")
		for _, measure := range spec.Measures {
			addColumn(measure.Name, catalogColumnType(measure.DataType))
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

func (q *parsedSQL) isCatalog() bool {
	if strings.TrimSpace(q.text) == "-- ping" {
		return true
	}
	if len(q.tokens) == 0 {
		return false
	}
	first := q.text[q.tokens[0].start:q.tokens[0].end]
	if strings.EqualFold(first, "SHOW") {
		return true
	}
	if !strings.EqualFold(first, "SELECT") && !strings.EqualFold(first, "WITH") {
		return false
	}

	// Track FROM clauses at each parenthesis depth so projection names, aliases,
	// and function arguments cannot masquerade as catalog relations.
	type clause struct{ selection, from, relation bool }
	clauses := []clause{{}}
	hasFrom := false
	for i, token := range q.tokens {
		current := &clauses[len(clauses)-1]
		switch token.kind {
		case '(':
			current.relation = false
			clauses = append(clauses, clause{})
			continue
		case ')':
			if len(clauses) > 1 {
				clauses = clauses[:len(clauses)-1]
			}
			continue
		case ',':
			current.relation = current.from
			continue
		case 'w', 'i':
		default:
			continue
		}
		word := q.text[token.start:token.end]
		if token.kind == 'w' {
			switch strings.ToUpper(word) {
			case "SELECT":
				current.selection = true
			case "FROM", "JOIN":
				if current.selection {
					hasFrom = true
					current.from, current.relation = true, true
				}
				continue
			case "WHERE", "GROUP", "HAVING", "ORDER", "LIMIT", "OFFSET", "UNION", "EXCEPT", "INTERSECT", "QUALIFY":
				current.from, current.relation = false, false
			case "LATERAL", "ONLY":
				continue
			}
		} else {
			word = word[1 : len(word)-1]
		}
		if !current.relation {
			continue
		}
		current.relation = false
		qualified := i+1 < len(q.tokens) && q.tokens[i+1].kind == '.'
		switch strings.ToLower(word) {
		case "pg_catalog", "information_schema":
			if qualified {
				return true
			}
		case "pg_attribute", "pg_class", "pg_type", "pg_namespace", "pg_index", "pg_constraint", "pg_matviews":
			if !qualified {
				return true
			}
		}
	}
	return strings.EqualFold(first, "SELECT") && !hasFrom
}

// Keep catalog values as driver arguments. Casts preserve each parameter's
// declared type, including NULLs used during Describe and narrow integer types.
func (q *parsedSQL) bindCatalogParameters(parameters []base.Parameter, types *pgtype.Map) (string, []any, error) {
	var args []any
	replacements := make(map[int]string)
	query, err := q.replaceParameters(func(index, _ int) (string, error) {
		if replacement, ok := replacements[index]; ok {
			return replacement, nil
		}
		if index >= len(parameters) {
			return "", &base.Error{Code: "42P02", Message: fmt.Sprintf("there is no parameter $%d", index+1)}
		}
		parameter := parameters[index]
		if parameter.OID == 0 {
			parameter.OID = pgtype.TextOID
		}
		typ, ok := types.TypeForOID(parameter.OID)
		if !ok {
			return "", fmt.Errorf("unsupported parameter type OID %d", parameter.OID)
		}
		value, err := typ.Codec.DecodeDatabaseSQLValue(types, parameter.OID, parameter.Format, parameter.Value)
		if err != nil {
			return "", fmt.Errorf("invalid parameter $%d: %w", index+1, err)
		}
		cast := typ.Name
		switch parameter.OID {
		case pgtype.JSONBOID:
			cast = "json"
		case pgtype.OIDOID:
			cast = "uinteger"
		case pgtype.NumericOID:
			// DuckDB's default NUMERIC scale is 3. Match the supplied decimal
			// instead of silently rounding it; PostgreSQL numeric has no typemod here.
			precision, scale := 38, 0
			if value != nil {
				whole, fraction, _ := strings.Cut(strings.TrimLeft(value.(string), "+-"), ".")
				scale = len(fraction)
				precision = max(1, len(strings.TrimLeft(whole, "0"))+scale)
				if precision > 38 || (whole == "NaN" || whole == "Infinity") {
					return "", fmt.Errorf("numeric parameter $%d exceeds DuckDB decimal range", index+1)
				}
			}
			cast = fmt.Sprintf("DECIMAL(%d,%d)", precision, scale)
		}
		name := fmt.Sprintf("p%d", index+1)
		args = append(args, sql.Named(name, value))
		replacement := "CAST($" + name + " AS " + cast + ")"
		replacements[index] = replacement
		return replacement, nil
	})
	return query, args, err
}

// The reconciler already resolves these types using Executor.Schema.
// Reuse them instead of guessing or issuing schema queries for every catalog request.
func catalogColumnType(typ *runtimev1.Type) string {
	if typ != nil && typ.Code == runtimev1.Type_CODE_ARRAY {
		return catalogColumnType(typ.ArrayElementType) + "[]"
	}
	oid, _ := postgresType(typ)
	switch oid {
	case pgtype.BoolOID:
		return "BOOLEAN"
	case pgtype.Int2OID:
		return "SMALLINT"
	case pgtype.Int4OID:
		return "INTEGER"
	case pgtype.Int8OID:
		return "BIGINT"
	case pgtype.NumericOID:
		return "DECIMAL(38, 9)"
	case pgtype.Float4OID:
		return "REAL"
	case pgtype.Float8OID:
		return "DOUBLE PRECISION"
	case pgtype.TimestamptzOID:
		return "TIMESTAMPTZ"
	case pgtype.DateOID:
		return "DATE"
	case pgtype.TimeOID:
		return "TIME"
	case pgtype.ByteaOID:
		return "BLOB"
	case pgtype.JSONBOID:
		return "JSON"
	case pgtype.UUIDOID:
		return "UUID"
	default:
		return "VARCHAR"
	}
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
		return nil, &base.Error{Code: "42704", Message: fmt.Sprintf("unsupported configuration parameter %q", name)}
	}
	fields, err := base.ApplyResultFormats([]pgproto3.FieldDescription{{Name: []byte(column), DataTypeOID: pgtype.TextOID, DataTypeSize: -1, TypeModifier: -1}}, resultFormats)
	if err != nil {
		return nil, err
	}
	encoded, err := types.Encode(pgtype.TextOID, fields[0].Format, value, nil)
	if err != nil {
		return nil, err
	}
	return &memoryRows{
		fields: fields,
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
		fields[i] = pgproto3.FieldDescription{Name: []byte(column.Name()), DataTypeOID: oid, DataTypeSize: size, TypeModifier: -1}
	}
	return base.ApplyResultFormats(fields, resultFormats)
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
	case strings.Contains(name, "HUGEINT") || strings.Contains(name, "DECIMAL") || strings.Contains(name, "NUMERIC"):
		return pgtype.NumericOID, -1
	case strings.Contains(name, "BIGINT"):
		return pgtype.Int8OID, 8
	case strings.Contains(name, "DOUBLE"):
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
		if decimal, ok := value.(duckdb.Decimal); ok {
			value = pgtype.Numeric{Int: decimal.Value, Exp: -int32(decimal.Scale), Valid: true}
		}
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
	return true
}
func (r *sqlRows) Values() [][]byte   { return r.values }
func (r *sqlRows) Err() error         { return r.err }
func (r *sqlRows) CommandTag() string { return "" }

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
