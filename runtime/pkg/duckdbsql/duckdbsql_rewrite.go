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

func (sn *selectNode) rewriteLimit(limit, offset int) error {
	modifiersNode := sn.ast.MustGet("modifiers")
	updated := false
	for _, v := range modifiersNode.ForRangeArr() {
		if v.MustGet("type").String() != "LIMIT_MODIFIER" {
			continue
		}

		modifierType := v.MustGet("limit").MustGet("class").String()
		switch modifierType {
		case "CONSTANT":
			v.MustGet("limit").MustGet("value").MustSetInt(limit).At("value")
			updated = true
		case "PARAMETER":
			err := v.Delete("limit")
			if err != nil {
				return err
			}

			limitObject, err := createConstantLimit(limit)
			if err != nil {
				return err
			}

			v.MustSet(limitObject).At("limit")
			updated = true
		}
	}

	if !updated {
		v, err := createLimitModifier(limit)
		if err != nil {
			return err
		}

		_, err = modifiersNode.Append(v).InTheEnd()
		if err != nil {
			return err
		}
	}

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

// TODO: offsets
func createConstantLimit(limit int) (*jsonvalue.V, error) {
	return jsonvalue.Unmarshal([]byte(fmt.Sprintf(`
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
`, limit)))
}

func createLimitModifier(limit int) (*jsonvalue.V, error) {
	return jsonvalue.Unmarshal([]byte(fmt.Sprintf(`
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
`, limit)))
}
