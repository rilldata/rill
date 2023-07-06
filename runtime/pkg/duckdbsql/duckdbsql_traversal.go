package duckdbsql

import (
	jsonvalue "github.com/Andrew-M-C/go.jsonvalue"
)

// TODO: handle parameters in values

func (a *AST) traverse() {
	if a.ast.MustGet("error").Bool() {
		return
	}

	// TODO: validation
	// TODO: CTEs and SET operations
	a.traverseSelectNode(a.ast.MustGet("statements").ForRangeArr()[0].MustGet("node"))
	a.traverseSelectList(a.ast.MustGet("statements").ForRangeArr()[0].MustGet("node").MustGet("select_list"))
}

func (a *AST) traverseSelectNode(node *jsonvalue.V) {
	sn := &selectNode{
		ast: node,
	}
	a.selectNodes = append(a.selectNodes, sn)

	a.traverseFromTable(node, "from_table")
}

func (a *AST) traverseSelectList(node *jsonvalue.V) {
	for _, col := range node.ForRangeArr() {
		a.traverseColumnNode(col)
	}
}

func (a *AST) traverseColumnNode(node *jsonvalue.V) {
	cn := &columnNode{
		ast: node,
		ref: &ColumnRef{},
	}
	a.columns = append(a.columns, cn)

	// TODO
}

func (a *AST) traverseFromTable(parent *jsonvalue.V, childKey string) {
	node := parent.MustGet(childKey)
	switch node.MustGet("type").String() {
	case "JOIN":
		a.traverseFromTable(node, "left")
		a.traverseFromTable(node, "right")

	case "BASE_TABLE":
		a.traverseBaseTable(parent, childKey)

	case "TABLE_FUNCTION":
		a.traverseTableFunction(parent, childKey)
	}
}

func (a *AST) traverseBaseTable(parent *jsonvalue.V, childKey string) {
	node := parent.MustGet(childKey)
	fn := &fromNode{
		ast:      node,
		parent:   parent,
		childKey: childKey,
		ref: &TableRef{
			Name: node.MustGet("table_name").String(),
		},
	}
	a.fromNodes = append(a.fromNodes, fn)
}

func (a *AST) traverseTableFunction(parent *jsonvalue.V, childKey string) {
	node := parent.MustGet(childKey)
	functionNode := node.MustGet("function")
	functionName := functionNode.MustGet("function_name").String()
	arguments := functionNode.MustGet("children").ForRangeArr()

	fn := &fromNode{
		ast:      node,
		parent:   parent,
		childKey: childKey,
		ref: &TableRef{
			Function:   functionName,
			Properties: map[string]any{},
		},
	}
	a.fromNodes = append(a.fromNodes, fn)

	switch functionName {
	case "read_csv_auto", "read_csv", "read_parquet", "read_json_auto", "read_json":
		fn.ref.Path = arguments[0].MustGet("value").MustGet("value").String()
	default:
		// only read_... are supported for now
		return
	}

	for _, argument := range arguments[1:] {
		if argument.MustGet("type").String() != "COMPARE_EQUAL" {
			continue
		}

		left := argument.MustGet("left")
		if left.MustGet("type").String() != "COLUMN_REF" {
			continue
		}

		right := argument.MustGet("right")
		switch right.MustGet("type").String() {
		case "VALUE_CONSTANT":
			fn.ref.Properties[left.MustGet("column_names").ForRangeArr()[0].String()] = constantValueToGoValue(right.MustGet("value"))
		case "FUNCTION":
			if right.MustGet("function_name").String() == "struct_pack" {
				fn.ref.Properties[left.MustGet("column_names").ForRangeArr()[0].String()] = structValueToGoValue(right)
			}
		}
	}
}

func structValueToGoValue(v *jsonvalue.V) map[string]any {
	structVal := map[string]any{}

	for _, child := range v.MustGet("children").ForRangeArr() {
		structVal[child.MustGet("alias").String()] = constantValueToGoValue(child.MustGet("value"))
	}

	return structVal
}

func constantValueToGoValue(v *jsonvalue.V) any {
	val := v.MustGet("value")
	switch v.MustGet("type").MustGet("id").String() {
	case "BOOLEAN":
		return val.Bool()
	case "TINYINT", "SMALLINT", "INTEGER":
		return val.Int32()
	case "BIGINT":
		return val.Int64()
	case "UTINYINT", "USMALLINT", "UINTEGER":
		return val.Uint32()
	case "UBIGINT":
		return val.Uint64()
	case "FLOAT":
		return val.Float32()
	case "DOUBLE":
		return val.Float64()
	case "VARCHAR":
		return val.String()
		// TODO: others
	}
	return nil
}
