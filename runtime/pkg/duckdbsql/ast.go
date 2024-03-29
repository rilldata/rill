package duckdbsql

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type AST struct {
	sql       string
	ast       astNode
	rootNodes []*selectNode
	aliases   map[string]bool
	added     map[string]bool
	fromNodes []*fromNode
	columns   []*columnNode
}

type selectNode struct {
	ast astNode
}

type columnNode struct {
	ast astNode
	ref *ColumnRef
}

type fromNode struct {
	ast      astNode
	parent   astNode
	childKey string
	ref      *TableRef
}

func Parse(sql string) (*AST, error) {
	sqlAst, err := queryString("select json_serialize_sql(?::VARCHAR)", sql)
	if err != nil {
		return nil, err
	}

	nativeAst := astNode{}

	decoder := json.NewDecoder(bytes.NewReader(sqlAst))
	// DuckDB uses uint64 for query_location and
	// inner queries may have query_location equals to max value of uint64 that cannot fit into float64
	decoder.UseNumber()

	err = decoder.Decode(&nativeAst)
	if err != nil {
		return nil, err
	}

	ast := &AST{
		sql:       sql,
		ast:       nativeAst,
		rootNodes: make([]*selectNode, 0),
		aliases:   map[string]bool{},
		added:     map[string]bool{},
		fromNodes: make([]*fromNode, 0),
		columns:   make([]*columnNode, 0),
	}

	err = ast.traverse()
	if err != nil {
		return nil, err
	}
	return ast, nil
}

// Format normalizes a DuckDB SQL statement
func (a *AST) Format() (string, error) {
	if a.ast == nil {
		return "", fmt.Errorf("calling format on failed parse")
	}

	sql, err := json.Marshal(a.ast)
	if err != nil {
		return "", err
	}
	res, err := queryString("SELECT json_deserialize_sql(?::JSON)", string(sql))
	return string(res), err
}

// RewriteTableRefs replaces table references in a DuckDB SQL query. Only replacing with a base table reference is supported right now.
func (a *AST) RewriteTableRefs(fn func(table *TableRef) (*TableRef, bool)) error {
	if a.ast == nil {
		return fmt.Errorf("calling rewrite on failed parse")
	}

	for _, node := range a.fromNodes {
		if node.ast == nil {
			continue
		}

		newRef, shouldReplace := fn(node.ref)
		if !shouldReplace {
			continue
		}

		if newRef.Name != "" {
			err := node.rewriteToBaseTable(newRef.Name)
			if err != nil {
				return err
			}
		} else if newRef.Function != "" {
			switch newRef.Function {
			case "sqlite_scan":
				newRef.Params[0] = newRef.Paths[0]
				err := node.rewriteToSqliteScanFunction(newRef.Params)
				if err != nil {
					return err
				}
			case "read_csv_auto", "read_csv",
				"read_parquet",
				"read_json", "read_json_auto", "read_json_objects", "read_json_objects_auto",
				"read_ndjson_objects", "read_ndjson", "read_ndjson_auto":
				err := node.rewriteToReadTableFunction(newRef.Function, newRef.Paths, newRef.Properties)
				if err != nil {
					return err
				}
				// non read_ functions are not supported right now
			}
		}
	}

	return nil
}

// RewriteLimit rewrites a DuckDB SQL statement to limit the result size
func (a *AST) RewriteLimit(limit, offset int) error {
	if a.ast == nil {
		return fmt.Errorf("calling rewrite on failed parse")
	}

	if len(a.rootNodes) == 0 {
		return nil
	}

	// We only need to add limit to the top level query
	err := a.rootNodes[0].rewriteLimit(limit)
	if err != nil {
		return err
	}

	return nil
}

// ExtractColumnRefs extracts column references from the outermost SELECT of a DuckDB SQL statement
func (a *AST) ExtractColumnRefs() []*ColumnRef {
	columnRefs := make([]*ColumnRef, 0)
	for _, node := range a.columns {
		columnRefs = append(columnRefs, node.ref)
	}
	return columnRefs
}

func (a *AST) GetTableRefs() []*TableRef {
	tableRefs := make([]*TableRef, 0)
	for _, node := range a.fromNodes {
		tableRefs = append(tableRefs, node.ref)
	}
	return tableRefs
}

func (a *AST) newFromNode(node, parent astNode, childKey string, ref *TableRef) {
	fn := &fromNode{
		ast:      node,
		parent:   parent,
		childKey: childKey,
		ref:      ref,
	}
	a.fromNodes = append(a.fromNodes, fn)
}

func (a *AST) newColumnNode(node astNode, ref *ColumnRef) {
	cn := &columnNode{
		ast: node,
		ref: ref,
	}
	a.columns = append(a.columns, cn)
}

// TableRef has information extracted about a DuckDB table or table function reference
type TableRef struct {
	Name       string
	Function   string
	Paths      []string
	Properties map[string]any
	LocalAlias bool
	// Params passed to sqlite_scan
	Params []string
}

// ColumnRef has information about a column in the select list of a DuckDB SQL statement
type ColumnRef struct {
	Name         string
	RelationName string
	Expr         string
	IsAggr       bool
	IsStar       bool
	IsExclude    bool
}
