package duckdbsql

import (
	"context"
	databasesql "database/sql"
	"database/sql/driver"
	"sync"

	"github.com/marcboeker/go-duckdb"
)

// Format normalizes a DuckDB SQL statement
func Format(sql string) (string, error) {
	return queryString("SELECT json_deserialize_sql(json_serialize_sql(?::VARCHAR))", sql)
}

// Sanitize strips comments and normalizes a DuckDB SQL statement
func Sanitize(sql string) (string, error) {
	panic("not implemented")
}

// RewriteLimit rewrites a DuckDB SQL statement to limit the result size
func RewriteLimit(sql string, limit, offset int) (string, error) {
	panic("not implemented")
}

// TableRef has information extracted about a DuckDB table or table function reference
type TableRef struct {
	Name       string
	Function   string
	Path       string
	Properties map[string]any
}

// ExtractTableRefs extracts table references from a DuckDB SQL query
func ExtractTableRefs(sql string) ([]*TableRef, error) {
	panic("not implemented")
}

// RewriteTableRefs replaces table references in a DuckDB SQL query
func RewriteTableRefs(sql string, fn func(table *TableRef) (*TableRef, bool)) (string, error) {
	panic("not implemented")
}

// Annotation is key-value annotation extracted from a DuckDB SQL comment
type Annotation struct {
	Key   string
	Value string
}

// ExtractAnnotations extracts annotations from comments prefixed with '@', and optionally a value after a ':'.
// Examples: "-- @materialize" and "-- @materialize: true".
func ExtractAnnotations() ([]*Annotation, error) {
	panic("not implemented")
}

// ColumnRef has information about a column in the select list of a DuckDB SQL statement
type ColumnRef struct {
	Name      string
	Expr      string
	IsAggr    bool
	IsStar    bool
	IsExclude bool
}

// ExtractColumnRefs extracts column references from the outermost SELECT of a DuckDB SQL statement
func ExtractColumnRefs(sql string) ([]*ColumnRef, error) {
	panic("not implemented")
}

// queryString runs a DuckDB query and returns the result as a scalar string
func queryString(qry string, args ...any) (string, error) {
	rows, err := query(qry, args...)
	if err != nil {
		return "", err
	}

	var res string
	if rows.Next() {
		err := rows.Scan(&res)
		if err != nil {
			return "", err
		}
	}

	err = rows.Close()
	if err != nil {
		return "", err
	}

	return res, nil
}

// Use a global in-memory DuckDB connection for invoking DuckDB's json_serialize_sql and json_deserialize_sql
var (
	db     *databasesql.DB
	dbOnce sync.Once
)

// query runs a DuckDB query
func query(qry string, args ...any) (*databasesql.Rows, error) {
	// Lazily initialize db global as an in-memory DuckDB connection
	dbOnce.Do(func() {
		// Using NewConnector since DuckDB requires extensions to be loaded separately on each connection
		connector, err := duckdb.NewConnector("", func(conn driver.ExecerContext) error {
			// Load JSON extension
			_, err := conn.ExecContext(context.Background(), "INSTALL 'json'; LOAD 'json';", nil)
			return err
		})
		if err != nil {
			panic(err)
		}

		// Set global
		db = databasesql.OpenDB(connector)
		db.SetMaxOpenConns(1)
	})

	return db.Query(qry, args...)
}
