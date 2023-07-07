package duckdbsql

// TODO: handle parameters in values

func (a *AST) traverse() {
	if toBoolean(a.ast, astKeyError) {
		return
	}

	statements := toNodeArray(a.ast, astKeyStatements)
	if len(statements) == 0 {
		return
	}

	// TODO: validation
	// TODO: CTEs and SET operations
	for _, statement := range statements {
		a.traverseSelectQueryStatement(toNode(statement, astKeyNode), true)
	}
}

func (a *AST) traverseSelectQueryStatement(node astNode, isRoot bool) {
	if isRoot {
		sn := &selectNode{
			ast: node,
		}
		a.rootNodes = append(a.rootNodes, sn)
	}

	switch toString(node, astKeyType) {
	case "SELECT_NODE":
		if isRoot {
			a.traverseSelectList(toNodeArray(node, astKeySelectColumnList))
		}
		a.traverseCTEMap(toNode(node, astKeyCTE))
		a.traverseFromTable(node, astKeyFromTable)

	case "SET_OPERATION_NODE":
		a.traverseCTEMap(toNode(node, astKeyCTE))
		a.traverseSelectQueryStatement(toNode(node, astKeyLeft), false)
		a.traverseSelectQueryStatement(toNode(node, astKeyRight), false)
	}
}

func (a *AST) traverseSelectList(colNodes []astNode) {
	for _, col := range colNodes {
		a.traverseColumnNode(col)
	}
}

func (a *AST) traverseColumnNode(node astNode) {
	cn := &columnNode{
		ast: node,
		ref: &ColumnRef{},
	}
	a.columns = append(a.columns, cn)

	// TODO
}

func (a *AST) traverseFromTable(parent astNode, childKey string) {
	node := toNode(parent, childKey)
	switch toString(node, astKeyType) {
	case "JOIN":
		a.traverseFromTable(node, "left")
		a.traverseFromTable(node, "right")

	case "BASE_TABLE":
		a.traverseBaseTable(parent, childKey)

	case "TABLE_FUNCTION":
		a.traverseTableFunction(parent, childKey)
	}
}

func (a *AST) traverseCTEMap(node astNode) {
	mapEntries := toNodeArray(node, astKeyMap)
	for _, mapEntry := range mapEntries {
		a.aliases[toString(mapEntry, astKeyKey)] = true
		a.traverseSelectQueryStatement(toNode(toNode(toNode(mapEntry, astKeyValue), astKeyQuery), astKeyNode), false)
	}
}

func (a *AST) traverseBaseTable(parent astNode, childKey string) {
	node := toNode(parent, childKey)
	name := toString(node, astKeyTableName)
	if a.added[name] {
		return
	}
	fn := &fromNode{
		ast:      node,
		parent:   parent,
		childKey: childKey,
		ref: &TableRef{
			Name:       name,
			LocalAlias: a.aliases[name],
		},
	}
	a.fromNodes = append(a.fromNodes, fn)
	a.added[name] = true
	// TODO: add to local alias
}

func (a *AST) traverseTableFunction(parent astNode, childKey string) {
	node := toNode(parent, childKey)
	functionNode := toNode(node, astKeyFunction)
	functionName := toString(functionNode, astKeyFunctionName)
	arguments := toNodeArray(functionNode, astKeyChildren)

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
	// TODO: add to local alias

	switch functionName {
	case "read_csv_auto", "read_csv", "read_parquet", "read_json_auto", "read_json":
		fn.ref.Path = toString(toNode(arguments[0], astKeyValue), astKeyValue)
	default:
		// only read_... are supported for now
		return
	}

	for _, argument := range arguments[1:] {
		if toString(argument, astKeyType) != "COMPARE_EQUAL" {
			continue
		}

		left := toNode(argument, astKeyLeft)
		if toString(left, astKeyType) != "COLUMN_REF" {
			continue
		}
		columnNames := toArray(left, astKeyColumnNames)
		if len(columnNames) == 0 {
			return
		}

		right := toNode(argument, astKeyRight)
		switch toString(right, astKeyType) {
		case "VALUE_CONSTANT":
			fn.ref.Properties[columnNames[0].(string)] = constantValueToGoValue(toNode(right, astKeyValue))
		case "FUNCTION":
			if toString(right, astKeyFunctionName) == "struct_pack" {
				fn.ref.Properties[columnNames[0].(string)] = structValueToGoValue(right)
			}
		}
	}
}

func structValueToGoValue(v astNode) map[string]any {
	structVal := map[string]any{}

	for _, child := range toNodeArray(v, astKeyChildren) {
		structVal[toString(child, astKeyAlias)] = constantValueToGoValue(toNode(child, astKeyValue))
	}

	return structVal
}

func constantValueToGoValue(v astNode) any {
	val := v[astKeyValue]
	switch toString(toNode(v, astKeyType), astKeyID) {
	case "BOOLEAN":
		return val.(bool)
	case "TINYINT", "SMALLINT", "INTEGER":
		return val.(int32)
	case "BIGINT":
		return val.(int64)
	case "UTINYINT", "USMALLINT", "UINTEGER":
		return val.(uint32)
	case "UBIGINT":
		return val.(uint64)
	case "FLOAT":
		return val.(float32)
	case "DOUBLE":
		return val.(float64)
	case "VARCHAR":
		return val.(string)
		// TODO: others
	}
	return nil
}
