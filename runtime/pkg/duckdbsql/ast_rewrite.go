package duckdbsql

import (
	"encoding/json"
	"fmt"
	"math"
)

func (fn *fromNode) rewriteToBaseTable(name string) error {
	baseTable, err := createBaseTable(name, fn.ast)
	if err != nil {
		return err
	}
	fn.parent[fn.childKey] = baseTable
	return nil
}

func (fn *fromNode) rewriteToReadTableFunction(name string, paths []string, props map[string]any) error {
	baseTable, err := createTableFunction(name, paths, props, fn.ast)
	if err != nil {
		return err
	}
	fn.parent[fn.childKey] = baseTable
	return nil
}

func (sn *selectNode) rewriteLimit(limit int) error {
	// If this is a CTE_NODE, traverse down to find the final SELECT/SET_OPERATION node
	if toString(sn.ast, astKeyType) == "CTE_NODE" {
		finalNode := sn.ast
		for {
			child := toNode(finalNode, astKeyChild)
			if child == nil {
				break
			}
			// If the child is still a CTE_NODE, continue traversing
			if toString(child, astKeyType) == "CTE_NODE" {
				finalNode = child
				continue
			}
			// Found the final SELECT or SET_OPERATION node
			finalNode = child
			break
		}

		// Apply limit to the final node
		sn := &selectNode{ast: finalNode}
		return sn.rewriteLimit(limit)
	}

	modifiersNode := toNodeArray(sn.ast, astKeyModifiers)
	updated := false
	for _, v := range modifiersNode {
		if toString(v, astKeyType) != "LIMIT_MODIFIER" {
			continue
		}

		modifierType := toString(toNode(v, astKeyLimit), astKeyClass)
		switch modifierType {
		case "CONSTANT":
			toNode(toNode(v, astKeyLimit), astKeyValue)[astKeyValue] = limit
			updated = true
		case "PARAMETER":
			delete(v, astKeyLimit)

			limitObject, err := createConstantLimit(limit)
			if err != nil {
				return err
			}

			v[astKeyLimit] = limitObject
			updated = true
		}
	}

	if !updated {
		v, err := createLimitModifier(limit)
		if err != nil {
			return err
		}

		sn.ast[astKeyModifiers] = append(sn.ast[astKeyModifiers].([]interface{}), v)
	}

	return nil
}

func (fn *fromNode) rewriteToSqliteScanFunction(params []string) error {
	baseTable, err := createSqliteScanTableFunction(params)
	if err != nil {
		return err
	}
	fn.parent[fn.childKey] = baseTable
	return nil
}

func createBaseTable(name string, ast astNode) (astNode, error) {
	// TODO: validation and fill in other fields from ast
	var n astNode
	err := json.Unmarshal([]byte(fmt.Sprintf(`{
	 "type": "BASE_TABLE",
	 "alias": "%s",
	 "sample": null,
	 "schema_name": "",
	 "table_name": "%s",
	 "column_name_alias": [],
	 "catalog_name": ""
	}`, toString(ast, astKeyAlias), name)), &n)
	if err != nil {
		return nil, err
	}

	n[astKeySample] = ast[astKeySample]
	n[astKeyColumnNameAlias] = ast[astKeyColumnNameAlias]
	return n, nil
}

func createTableFunction(name string, paths []string, props map[string]any, ast astNode) (astNode, error) {
	var n astNode
	err := json.Unmarshal([]byte(fmt.Sprintf(`{
  "type": "TABLE_FUNCTION",
  "alias": "%s",
  "sample": null,
  "function": {},
  "column_name_alias": []
}`, toString(ast, astKeyAlias))), &n)
	if err != nil {
		return nil, err
	}

	fn, err := createFunctionCall("", name, "")
	if err != nil {
		return nil, err
	}
	n[astKeyFunction] = fn

	pa, err := createListValue[string]("", paths)
	if err != nil {
		return nil, err
	}
	// create the list of args with the 1st arg being the list of path
	args := []astNode{pa}

	for k, v := range props {
		vn, err := createKeyedFunctionArg(k, v)
		if err != nil {
			return nil, err
		}
		if vn == nil {
			continue
		}
		args = append(args, vn)
	}

	fn[astKeyChildren] = args

	return n, nil
}

// TODO: offsets
func createConstantLimit(limit int) (astNode, error) {
	var n astNode
	err := json.Unmarshal([]byte(fmt.Sprintf(`
	{
	   "class":"CONSTANT",
	   "type":"VALUE_CONSTANT",
	   "alias":"",
	   "value":{
		  "type":{
			 "id":"INTEGER",
			 "type_info":null
		  },
		  "is_null":false,
		  "value":%d
	   }
	}
`, limit)), &n)
	return n, err
}

func createLimitModifier(limit int) (astNode, error) {
	var n astNode
	err := json.Unmarshal([]byte(fmt.Sprintf(`
{
	"type":"LIMIT_MODIFIER",
	"limit":{
	   "class":"CONSTANT",
	   "type":"VALUE_CONSTANT",
	   "alias":"",
	   "value":{
		  "type":{
			 "id":"INTEGER",
			 "type_info":null
		  },
		  "is_null":false,
		  "value":%d
	   }
	},
	"offset":null
 }
`, limit)), &n)
	return n, err
}

// Creates a blank statement from an expression
func createExpressionStatement(exprNode astNode) (string, error) {
	jsonNode, err := json.Marshal(exprNode)
	if err != nil {
		return "", err
	}
	baseJSON := map[string]interface{}{
		"error": false,
		"statements": []map[string]interface{}{
			{
				"node": map[string]interface{}{
					"type":        "SELECT_NODE",
					"modifiers":   []interface{}{},
					"cte_map":     map[string]interface{}{"map": []interface{}{}},
					"select_list": []json.RawMessage{jsonNode},
					"from_table": map[string]interface{}{
						"type":              "BASE_TABLE",
						"alias":             "",
						"sample":            nil,
						"schema_name":       "",
						"table_name":        "Dummy",
						"column_name_alias": []interface{}{},
						"catalog_name":      "",
					},
					"where_clause":       nil,
					"group_expressions":  []interface{}{},
					"group_sets":         []interface{}{},
					"aggregate_handling": "STANDARD_HANDLING",
					"having":             nil,
					"sample":             nil,
					"qualify":            nil,
				},
			},
		},
	}
	finalJSON, err := json.Marshal(baseJSON)
	if err != nil {
		return "", err
	}
	return string(finalJSON), nil
}

// createKeyedFunctionArg creates an arg with a key.
// EG: read_csv("/path", delim='|') - delim='|' is an arg with key, whereas "/path" a plain constant.
func createKeyedFunctionArg(key string, val any) (astNode, error) {
	var n astNode
	err := json.Unmarshal([]byte(fmt.Sprintf(`{
  "class": "COMPARISON",
  "type": "COMPARE_EQUAL",
  "alias": "",
  "left": {
    "class": "COLUMN_REF",
    "type": "COLUMN_REF",
    "alias": "",
    "column_names": ["%s"]
  },
  "right": {}
}`, key)), &n)
	if err != nil {
		return nil, err
	}
	rvn, err := createGenericValue("", val)
	if err != nil {
		return nil, err
	}
	n[astKeyRight] = rvn
	return n, nil
}

func createGenericValue(key string, val any) (astNode, error) {
	var t string
	switch vt := val.(type) {
	case map[string]any:
		return createStructValue(key, vt)
	case []any:
		return createListValue[any](key, vt)
	case bool:
		t = "BOOLEAN"
	case int, int32:
		t = "INTEGER"
	case int64:
		t = "BIGINT"
	case uint, uint32:
		t = "UINTEGER"
	case uint64:
		t = "UBIGINT"
	case float32:
		t = "FLOAT"
		// Temporary fix since duckdb is not converting to sql properly
		if math.Floor(float64(vt)) == float64(vt) {
			val = forceConvertToNum[int32](vt)
			t = "INTEGER"
		}
	case float64:
		t = "DOUBLE"
		// Temporary fix since duckdb is not converting to sql properly
		if math.Floor(vt) == vt {
			val = forceConvertToNum[int64](vt)
			t = "BIGINT"
		}
	case string:
		t = "VARCHAR"
		val = fmt.Sprintf(`%q`, vt)
	// TODO: others
	default:
		return nil, nil
	}

	var n astNode
	err := json.Unmarshal([]byte(fmt.Sprintf(`{
  "class": "CONSTANT",
  "type": "VALUE_CONSTANT",
  "alias": "%s",
  "value": {
    "type": {
      "id": "%s",
      "type_info": null
    },
    "is_null": false,
    "value": %v
  }
}`, key, t, val)), &n)
	return n, err
}

func createStandaloneValue(val any) (astNode, error) {
	var t string
	switch vt := val.(type) {
	// these are not supported for standalone value
	// case map[string]any:
	// case []any:
	case bool:
		t = "BOOLEAN"
	case int, int32:
		t = "INTEGER"
	case int64:
		t = "BIGINT"
	case uint, uint32:
		t = "UINTEGER"
	case uint64:
		t = "UBIGINT"
	case float32:
		t = "FLOAT"
		// Temporary fix since duckdb is not converting to sql properly
		if math.Floor(float64(vt)) == float64(vt) {
			val = forceConvertToNum[int32](vt)
			t = "INTEGER"
		}
	case float64:
		t = "DOUBLE"
		// Temporary fix since duckdb is not converting to sql properly
		if math.Floor(vt) == vt {
			val = forceConvertToNum[int64](vt)
			t = "BIGINT"
		}
	case json.Number:
		i, err := vt.Int64()
		if err == nil {
			return createStandaloneValue(i)
		}
		f, err := vt.Float64()
		if err == nil {
			return createStandaloneValue(f)
		}
		return nil, err
	case string:
		t = "VARCHAR"
		val = fmt.Sprintf(`%q`, vt)
	// TODO: others
	default:
		return nil, nil
	}

	var n astNode
	err := json.Unmarshal([]byte(fmt.Sprintf(`{
	"type": {
		"id": "%s",
		"type_info": null
	},
	"is_null": false,
	"value": %v
}`, t, val)), &n)
	return n, err
}

func createStructValue(key string, val map[string]any) (astNode, error) {
	n, err := createFunctionCall(key, "struct_pack", "main")
	if err != nil {
		return nil, err
	}

	var list []astNode
	for k, v := range val {
		vn, err := createGenericValue(k, v)
		if err != nil {
			return nil, err
		}
		if vn == nil {
			continue
		}
		list = append(list, vn)
	}

	n[astKeyChildren] = list
	return n, nil
}

func createListValue[E any](key string, val []E) (astNode, error) {
	n, err := createFunctionCall(key, "list_value", "main")
	if err != nil {
		return nil, err
	}

	var list []astNode
	for _, v := range val {
		vn, err := createGenericValue("", v)
		if err != nil {
			return nil, err
		}
		if vn == nil {
			continue
		}
		list = append(list, vn)
	}

	n[astKeyChildren] = list
	return n, nil
}

func createFunctionCall(key, name, schema string) (astNode, error) {
	var n astNode
	err := json.Unmarshal([]byte(fmt.Sprintf(`{
  "class": "FUNCTION",
  "type": "FUNCTION",
  "alias": "%s",
  "function_name": "%s",
  "schema": "%s",
  "children": [],
  "filter": null,
  "order_bys": {
    "type": "ORDER_MODIFIER",
    "orders": []
  },
  "distinct": false,
  "is_operator": false,
  "export_state": false,
  "catalog": ""
}`, key, name, schema)), &n)
	return n, err
}

func createSqliteScanTableFunction(params []string) (astNode, error) {
	var n astNode
	err := json.Unmarshal([]byte(`{
  "type": "TABLE_FUNCTION",
  "alias": "",
  "sample": null,
  "function": {},
  "column_name_alias": []
}`), &n)
	if err != nil {
		return nil, err
	}

	fn, err := createFunctionCall("", "sqlite_scan", "")
	if err != nil {
		return nil, err
	}
	n[astKeyFunction] = fn

	var list []astNode
	for _, v := range params {
		vn, err := createGenericValue("", v)
		if err != nil {
			return nil, err
		}
		if vn == nil {
			continue
		}
		list = append(list, vn)
	}
	fn[astKeyChildren] = list
	return n, nil
}
