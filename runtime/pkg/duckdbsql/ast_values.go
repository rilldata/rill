package duckdbsql

import "strings"

// TODO: figure out a way to cast map[string]interface{} returned by json unmarshal to map[astNodeKey]interface{} and replace string in key to astNodeKey
type astNode map[string]interface{}

const (
	astKeyError            string = "error"
	astKeyErrorMessage     string = "error_message"
	astKeyStatements       string = "statements"
	astKeyNode             string = "node"
	astKeyType             string = "type"
	astKeyKey              string = "key"
	astKeyFromTable        string = "from_table"
	astKeySelectColumnList string = "select_list"
	astKeyTableName        string = "table_name"
	astKeyFunction         string = "function"
	astKeyFunctionName     string = "function_name"
	astKeyChildren         string = "children"
	astKeyChild            string = "child"
	astKeyValue            string = "value"
	astKeyLeft             string = "left"
	astKeyRight            string = "right"
	astKeyColumnNames      string = "column_names"
	astKeyAlias            string = "alias"
	astKeyID               string = "id"
	astKeySample           string = "sample"
	astKeySampleSize       string = "sample_size"
	astKeyColumnNameAlias  string = "column_name_alias"
	astKeyModifiers        string = "modifiers"
	astKeyLimit            string = "limit"
	astKeyClass            string = "class"
	astKeyCTE              string = "cte_map"
	astKeyCTEName          string = "cte_name"
	astKeyMap              string = "map"
	astKeyQuery            string = "query"
	astKeySubQuery         string = "subquery"
	astKetRelationName     string = "relation_name"
	astKeySchema           string = "schema"
	astKeyIsNull           string = "is_null"
	astKeyTypeInfo         string = "type_info"
	astKeyScale            string = "scale"
	astKeyCastType         string = "cast_type"
	astKeySource           string = "source"
	astKeyPosition         string = "position"
)

func toBoolean(a astNode, k string) bool {
	v, ok := a[k]
	if !ok {
		return false
	}
	return castToBoolean(v)
}

func toString(a astNode, k string) string {
	v, ok := a[k]
	if !ok {
		return ""
	}
	switch vt := v.(type) {
	case string:
		return vt
	default:
		return ""
	}
}

func toNode(a astNode, k string) astNode {
	v, ok := a[k]
	if !ok {
		return nil
	}
	switch vt := v.(type) {
	case map[string]interface{}:
		return vt
	default:
		return nil
	}
}

func toArray(a astNode, k string) []interface{} {
	v, ok := a[k]
	if !ok {
		return make([]interface{}, 0)
	}
	switch v.(type) {
	case interface{}:
		return v.([]interface{})
	default:
		return make([]interface{}, 0)
	}
}

func toNodeArray(a astNode, k string) []astNode {
	arr := toArray(a, k)
	nodeArr := make([]astNode, len(arr))
	for i, e := range arr {
		nodeArr[i] = e.(map[string]interface{})
	}
	return nodeArr
}

func toTypedArray[E interface{}](a astNode, k string) []E {
	arr := toArray(a, k)
	typedArr := make([]E, len(arr))
	for i, e := range arr {
		typedArr[i] = e.(E)
	}
	return typedArr
}

// getListOfValues converts a node that can have a single value or a list of values to a go array of a type
func getListOfValues[E interface{}](a astNode) []E {
	arr := make([]E, 0)
	switch toString(a, astKeyType) {
	case "VALUE_CONSTANT":
		if vt, ok := a[astKeyValue].(map[string]interface{})[astKeyValue].(E); ok {
			arr = append(arr, vt)
		}

	case "FUNCTION":
		if toString(a, astKeyFunctionName) == "list_value" {
			for _, child := range toNodeArray(a, astKeyChildren) {
				if toString(child, astKeyType) != "VALUE_CONSTANT" {
					continue
				}
				if vt, ok := child[astKeyValue].(map[string]interface{})[astKeyValue].(E); ok {
					arr = append(arr, vt)
				}
			}
		}
	}
	return arr
}

func getColumnName(node astNode) string {
	alias := toString(node, astKeyAlias)
	if alias != "" {
		return alias
	}
	return strings.Join(toTypedArray[string](node, astKeyColumnNames), ".")
}

func castToBoolean(val any) bool {
	switch vt := val.(type) {
	case bool:
		return vt
	case string:
		switch strings.ToLower(vt) {
		case "true", "t":
			return true
		case "false", "f":
			return false
		default:
			return false
		}
	default:
		return false
	}
}
