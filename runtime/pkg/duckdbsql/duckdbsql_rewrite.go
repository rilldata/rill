package duckdbsql

import (
	"fmt"

	jsonvalue "github.com/Andrew-M-C/go.jsonvalue"
)

func (fn *fromNode) rewriteToBaseTable(name string) error {
	baseTable, err := createBaseTable(name, fn.ast)
	if err != nil {
		return err
	}

	fn.parent.MustSet(baseTable).At(fn.childKey)
	return nil
}

func createBaseTable(name string, ast *jsonvalue.V) (*jsonvalue.V, error) {
	// TODO: validation and fill in other fields from ast
	v, err := jsonvalue.Unmarshal([]byte(fmt.Sprintf(`{
	 "type": "BASE_TABLE",
	 "alias": "%s",
	 "sample": null,
	 "schema_name": "",
	 "table_name": "%s",
	 "column_name_alias": [],
	 "catalog_name": ""
	}`, ast.MustGet("alias").String(), name)))
	if err != nil {
		return nil, err
	}

	v.MustSet(ast.MustGet("sample")).At("sample")
	v.MustSet(ast.MustGet("column_name_alias")).At("column_name_alias")
	return v, nil
}
