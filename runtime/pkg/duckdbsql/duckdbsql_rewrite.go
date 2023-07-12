package duckdbsql

import (
	"encoding/json"
	"fmt"
)

func (fn *fromNode) rewriteToBaseTable(name string) error {
	baseTable, err := createBaseTable(name, fn.ast)
	if err != nil {
		return err
	}
	fn.parent[fn.childKey] = baseTable
	return nil
}

func (sn *selectNode) rewriteLimit(limit, offset int) error {
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
	return fmt.Sprintf(`
{
  "error": false,
  "statements": [{
    "node": {
      "type": "SELECT_NODE",
      "modifiers": [],
      "cte_map": {
        "map": []
      },
      "select_list": [%s],
      "from_table": {
        "type": "BASE_TABLE",
        "alias": "",
        "sample": null,
        "schema_name": "",
        "table_name": "Dummy",
        "column_name_alias": [],
        "catalog_name": ""
      },
      "where_clause": null,
      "group_expressions": [],
      "group_sets": [],
      "aggregate_handling": "STANDARD_HANDLING",
      "having": null,
      "sample": null,
      "qualify": null
    }
  }],
}
`, jsonNode), nil
}
