package duckdbsql

import (
	"context"
	databasesql "database/sql"
	"database/sql/driver"
	"sync"

	jsonvalue "github.com/Andrew-M-C/go.jsonvalue"
	"github.com/marcboeker/go-duckdb"
)

type AST struct {
	ast         *jsonvalue.V
	selectNodes []*selectNode
	fromNodes   []*fromNode
	columns     []*columnNode
}

type selectNode struct {
	ast *jsonvalue.V
}

type columnNode struct {
	ast *jsonvalue.V
	ref *ColumnRef
}

type fromNode struct {
	ast      *jsonvalue.V
	parent   *jsonvalue.V
	childKey string
	ref      *TableRef
}

func Parse(sql string) (*AST, error) {
	// TODO: optimise and parse into []byte
	serializedSQL, err := queryString("select json_serialize_sql(?::VARCHAR)", sql)
	if err != nil {
		return nil, err
	}

	v, err := jsonvalue.Unmarshal(serializedSQL)
	if err != nil {
		return nil, err
	}
	ast := &AST{
		ast:         v,
		selectNodes: make([]*selectNode, 0),
		fromNodes:   make([]*fromNode, 0),
		columns:     make([]*columnNode, 0),
	}

	ast.traverse()

	return ast, nil
}

// Format normalizes a DuckDB SQL statement
func (a *AST) Format() (string, error) {
	// TODO: cleanup this code
	res, err := queryString("SELECT json_deserialize_sql(?::JSON)", a.ast.MustMarshalString())
	return string(res), err
}

// RewriteTableRefs replaces table references in a DuckDB SQL query
func (a *AST) RewriteTableRefs(fn func(table *TableRef) (*TableRef, bool)) error {
	for _, node := range a.fromNodes {
		newRef, shouldReplace := fn(node.ref)
		if !shouldReplace {
			continue
		}

		if node.ref.Name == "" && newRef.Name != "" {
			err := node.rewriteToBaseTable(newRef.Name)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// RewriteLimit rewrites a DuckDB SQL statement to limit the result size
func (a *AST) RewriteLimit(limit, offset int) error {
	for _, node := range a.selectNodes {
		err := node.rewriteLimit(limit, offset)
		if err != nil {
			return err
		}
	}

	return nil
}

// TableRef has information extracted about a DuckDB table or table function reference
type TableRef struct {
	Name       string
	Function   string
	Path       string
	Properties map[string]any
}

// Annotation is key-value annotation extracted from a DuckDB SQL comment
type Annotation struct {
	Key   string
	Value string
}

// ExtractAnnotations extracts annotations from comments prefixed with '@', and optionally a value after a ':'.
// Examples: "-- @materialize" and "-- @materialize: true".
// TODO: duckdb's parser doesnt return comments. We need our own parser.
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
func queryString(qry string, args ...any) ([]byte, error) {
	rows, err := query(qry, args...)
	if err != nil {
		return nil, err
	}

	var res []byte
	if rows.Next() {
		err := rows.Scan(&res)
		if err != nil {
			return nil, err
		}
	}

	err = rows.Close()
	if err != nil {
		return nil, err
	}

	return res, nil
}

// Use a global in-memory DuckDB connection for invoking DuckDB's json_serialize_sql and json_deserialize_sql
// TODO: Why not get driver connection?
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
