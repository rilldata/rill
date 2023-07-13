package duckdbsql

import (
	"errors"
	"regexp"
)

// TODO: handle parameters in values

func (a *AST) traverse() error {
	if toBoolean(a.ast, astKeyError) {
		return errors.New(toString(a.ast, astKeyErrorMessage))
	}

	statements := toNodeArray(a.ast, astKeyStatements)
	if len(statements) == 0 {
		return errors.New("no statement found")
	}

	// TODO: validation
	// TODO: CTEs and SET operations
	for _, statement := range statements {
		a.traverseSelectQueryStatement(toNode(statement, astKeyNode), true)
	}

	return nil
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
			// only get the select list from the root select.
			// if there is a star, get the actual columns from TableColumns query
			a.traverseSelectList(toNodeArray(node, astKeySelectColumnList))
		}
		a.traverseCTEMap(toNode(node, astKeyCTE))
		a.traverseFromTable(node, astKeyFromTable)

	case "SET_OPERATION_NODE":
		a.traverseCTEMap(toNode(node, astKeyCTE))
		a.traverseSelectQueryStatement(toNode(node, astKeyLeft), isRoot)
		a.traverseSelectQueryStatement(toNode(node, astKeyRight), isRoot)
	}
}

func (a *AST) traverseSelectList(colNodes []astNode) {
	for _, col := range colNodes {
		a.traverseColumnNode(col)
	}
}

func (a *AST) traverseColumnNode(node astNode) {
	switch toString(node, astKeyType) {
	case "COLUMN_REF":
		a.newColumnNode(node, &ColumnRef{
			Name: getColumnName(node),
		})

	case "STAR":
		a.newColumnNode(node, &ColumnRef{
			RelationName: toString(node, astKetRelationName),
			IsStar:       true,
		})

	case "FUNCTION":
		funcName := toString(node, astKeyFunctionName)
		if funcName == "exclude" {
			a.traverseExcludeColumnNode(node)
		} else {
			a.traverseExpressionColumnNode(node)
		}
	}
}

func (a *AST) traverseExcludeColumnNode(node astNode) {
	for _, child := range toNodeArray(node, astKeyChildren) {
		if toString(child, astKeyType) != "COLUMN_REF" {
			continue
		}

		a.newColumnNode(node, &ColumnRef{
			Name:      getColumnName(child),
			IsExclude: true,
		})
	}
}

var selectExpressionIsolation = regexp.MustCompile(`^SELECT (.*?)(?: AS .*)? FROM Dummy$`)

func (a *AST) traverseExpressionColumnNode(node astNode) {
	exprStatement, err := createExpressionStatement(node)
	if err != nil {
		return
	}

	exprSQL, err := queryString("SELECT json_deserialize_sql(?::JSON)", exprStatement)
	if err != nil {
		return
	}

	subMatches := selectExpressionIsolation.FindStringSubmatch(string(exprSQL))
	if len(subMatches) == 0 {
		return
	}

	a.newColumnNode(node, &ColumnRef{
		Name: toString(node, astKeyAlias),
		Expr: subMatches[1],
	})
	// TODO: fill in isAggr
}

func (a *AST) traverseFromTable(parent astNode, childKey string) {
	node := toNode(parent, childKey)
	switch toString(node, astKeyType) {
	case "JOIN":
		a.traverseFromTable(node, "left")
		a.traverseFromTable(node, "right")

	case "SUBQUERY":
		a.traverseSelectQueryStatement(toNode(toNode(node, astKeySubQuery), astKeyNode), false)

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
	a.newFromNode(node, parent, childKey, &TableRef{
		Name:       name,
		LocalAlias: a.aliases[name],
	})
	a.added[name] = true
	// TODO: add to local alias
}

func (a *AST) traverseTableFunction(parent astNode, childKey string) {
	node := toNode(parent, childKey)
	functionNode := toNode(node, astKeyFunction)
	functionName := toString(functionNode, astKeyFunctionName)
	arguments := toNodeArray(functionNode, astKeyChildren)

	ref := &TableRef{
		Function:   functionName,
		Properties: map[string]any{},
	}
	a.newFromNode(node, parent, childKey, ref)
	// TODO: add to local alias

	switch functionName {
	case "read_csv_auto", "read_csv", "read_parquet", "read_json_auto", "read_json":
		ref.Path = toString(toNode(arguments[0], astKeyValue), astKeyValue)
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
			ref.Properties[columnNames[0].(string)] = constantValueToGoValue(toNode(right, astKeyValue))
		case "FUNCTION":
			if toString(right, astKeyFunctionName) == "struct_pack" {
				ref.Properties[columnNames[0].(string)] = structValueToGoValue(right)
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
