package duckdbsql

import (
	"errors"
	"math"
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
		node[astKeySample] = a.correctSampleClause(toNode(node, astKeySample))

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

	case "PIVOT":
		a.traverseFromTable(node, astKeySource)
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
	// TODO: add to local alias

	switch functionName {
	case "sqlite_scan":
		a.newFromNode(node, parent, childKey, ref)
		ref.Params = make([]string, 0)
		for _, argument := range arguments {
			typ := toString(argument, astKeyType)
			switch typ {
			case "VALUE_CONSTANT":
				ref.Params = append(ref.Params, getListOfValues[string](argument)...)
			case "COLUMN_REF":
				columnNames := toArray(argument, astKeyColumnNames)
				for _, column := range columnNames {
					ref.Params = append(ref.Params, column.(string))
				}
			default:
			}
		}
		if len(ref.Params) >= 1 {
			// first param is path to local db file
			ref.Paths = ref.Params[:1]
		}
		return
	case "read_csv_auto", "read_csv",
		"read_parquet",
		"read_json", "read_json_auto", "read_json_objects", "read_json_objects_auto",
		"read_ndjson_objects", "read_ndjson", "read_ndjson_auto":
		ref.Paths = getListOfValues[string](arguments[0])
	default:
		// only read_... are supported for now
		return
	}

	// adding the node here will make sure other types of table functions are ignored
	a.newFromNode(node, parent, childKey, ref)

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
		ref.Properties[columnNames[0].(string)] = valueToGoValue(right)
	}
}

func valueToGoValue(v astNode) any {
	switch toString(v, astKeyType) {
	case "VALUE_CONSTANT":
		return constantValueToGoValue(toNode(v, astKeyValue))
	case "FUNCTION":
		if toString(v, astKeySchema) == "main" {
			switch toString(v, astKeyFunctionName) {
			case "struct_pack":
				return structValueToGoValue(v)
			case "list_value":
				return arrayValueToGoValue(v)
			}
		}
	case "OPERATOR_CAST":
		return castValueToGoValue(v)
	}
	return nil
}

func constantValueToGoValue(v astNode) any {
	if toBoolean(v, astKeyIsNull) {
		return nil
	}

	t := toNode(v, astKeyType)
	val := v[astKeyValue]
	switch toString(t, astKeyID) {
	case "BOOLEAN":
		return val.(bool)
	case "TINYINT", "SMALLINT", "INTEGER":
		return forceConvertToNum[int32](val)
	case "BIGINT":
		return forceConvertToNum[int64](val)
	case "UTINYINT", "USMALLINT", "UINTEGER":
		return forceConvertToNum[uint32](val)
	case "UBIGINT":
		return forceConvertToNum[uint64](val)
	case "FLOAT":
		return forceConvertToNum[float32](val)
	case "DOUBLE":
		return forceConvertToNum[float64](val)
	case "DECIMAL":
		ti := toNode(t, astKeyTypeInfo)
		if ti == nil {
			return 0.0
		}
		return forceConvertToNum[float64](val) / math.Pow(10, forceConvertToNum[float64](ti[astKeyScale]))
	case "VARCHAR":
		return val.(string)
		// TODO: others
	}
	return nil
}

func structValueToGoValue(v astNode) map[string]any {
	structVal := map[string]any{}

	for _, child := range toNodeArray(v, astKeyChildren) {
		structVal[toString(child, astKeyAlias)] = valueToGoValue(child)
	}

	return structVal
}

func arrayValueToGoValue(v astNode) []any {
	arr := make([]any, 0)
	for _, child := range toNodeArray(v, astKeyChildren) {
		arr = append(arr, valueToGoValue(child))
	}
	return arr
}

func castValueToGoValue(v astNode) any {
	val := valueToGoValue(toNode(v, astKeyChild))
	if toString(toNode(v, astKeyCastType), astKeyID) == "BOOLEAN" {
		return castToBoolean(val)
	}
	// TODO: other types
	return nil
}

func forceConvertToNum[N int32 | int64 | uint32 | uint64 | float32 | float64](v any) N {
	switch vt := v.(type) {
	case int:
		return N(vt)
	case int32:
		return N(vt)
	case int64:
		return N(vt)
	case float32:
		return N(vt)
	case float64:
		return N(vt)
	}
	return 0
}
