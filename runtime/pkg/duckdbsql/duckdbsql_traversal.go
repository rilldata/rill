package duckdbsql

import jsonvalue "github.com/Andrew-M-C/go.jsonvalue"

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

	a.traverseFromTable(node.MustGet("from_table"))
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

func (a *AST) traverseFromTable(node *jsonvalue.V) {
	switch node.MustGet("type").String() {
	case "JOIN":
		a.traverseFromTable(node.MustGet("left"))
		a.traverseFromTable(node.MustGet("right"))

	case "BASE_TABLE":
		a.traverseBaseTable(node)

	case "TABLE_FUNCTION":
		a.traverseTableFunction(node)
	}
}

func (a *AST) traverseBaseTable(node *jsonvalue.V) {
	fn := &fromNode{
		ast: node,
		ref: &TableRef{
			Name: node.MustGet("table_name").String(),
		},
	}
	a.fromNodes = append(a.fromNodes, fn)
}

func (a *AST) traverseTableFunction(node *jsonvalue.V) {
	functionNode := node.MustGet("function")
	functionName := functionNode.MustGet("function_name").String()
	arguments := functionNode.MustGet("children").ForRangeArr()

	fn := &fromNode{
		ast: node,
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

	// TODO
	//for _, argument := range arguments[1:] {
	//}
}
