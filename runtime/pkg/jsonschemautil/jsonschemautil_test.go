package jsonschemautil

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtractDefAsSchema_SimpleDefinitionWithoutRefs(t *testing.T) {
	jsonSchema := `{
		"type": "object",
		"$defs": {
			"Person": {
				"type": "object",
				"properties": {
					"name": {"type": "string"},
					"age": {"type": "integer"}
				},
				"required": ["name"]
			},
			"Address": {
				"type": "object",
				"properties": {
					"street": {"type": "string"}
				}
			}
		}
	}`

	result, err := ExtractDefAsSchema(jsonSchema, "Person")
	require.NoError(t, err)

	// Parse result to verify structure
	var resultMap map[string]any
	err = json.Unmarshal([]byte(result), &resultMap)
	require.NoError(t, err)

	// Verify the result has the expected structure
	require.Equal(t, "object", resultMap["type"])
	require.Contains(t, resultMap, "properties")
	require.Contains(t, resultMap, "required")
	require.NotContains(t, resultMap, "$defs") // No refs, so no $defs needed

	// Verify properties
	props := resultMap["properties"].(map[string]any)
	require.Contains(t, props, "name")
	require.Contains(t, props, "age")
}

func TestExtractDefAsSchema_DefinitionWithSingleRef(t *testing.T) {
	jsonSchema := `{
		"type": "object",
		"$defs": {
			"Person": {
				"type": "object",
				"properties": {
					"name": {"type": "string"},
					"address": {"$ref": "#/$defs/Address"}
				}
			},
			"Address": {
				"type": "object",
				"properties": {
					"street": {"type": "string"},
					"city": {"type": "string"}
				}
			},
			"UnusedDef": {
				"type": "string"
			}
		}
	}`

	result, err := ExtractDefAsSchema(jsonSchema, "Person")
	require.NoError(t, err)

	// Parse result to verify structure
	var resultMap map[string]any
	err = json.Unmarshal([]byte(result), &resultMap)
	require.NoError(t, err)

	// Verify the result has the expected structure
	require.Equal(t, "object", resultMap["type"])
	require.Contains(t, resultMap, "properties")
	require.Contains(t, resultMap, "$defs")

	// Verify $defs contains only the referenced Address
	defs := resultMap["$defs"].(map[string]any)
	require.Contains(t, defs, "Address")
	require.NotContains(t, defs, "UnusedDef")
	require.Len(t, defs, 1)
}

func TestExtractDefAsSchema_DefinitionWithNestedRefs(t *testing.T) {
	jsonSchema := `{
		"type": "object",
		"$defs": {
			"Person": {
				"type": "object",
				"properties": {
					"name": {"type": "string"},
					"address": {"$ref": "#/$defs/Address"}
				}
			},
			"Address": {
				"type": "object",
				"properties": {
					"street": {"type": "string"},
					"country": {"$ref": "#/$defs/Country"}
				}
			},
			"Country": {
				"type": "object",
				"properties": {
					"name": {"type": "string"},
					"code": {"type": "string"}
				}
			}
		}
	}`

	result, err := ExtractDefAsSchema(jsonSchema, "Person")
	require.NoError(t, err)

	// Parse result to verify structure
	var resultMap map[string]any
	err = json.Unmarshal([]byte(result), &resultMap)
	require.NoError(t, err)

	// Verify $defs contains both Address and Country
	defs := resultMap["$defs"].(map[string]any)
	require.Contains(t, defs, "Address")
	require.Contains(t, defs, "Country")
	require.Len(t, defs, 2)
}

func TestExtractDefAsSchema_DefinitionWithCircularRefs(t *testing.T) {
	jsonSchema := `{
		"type": "object",
		"$defs": {
			"Node": {
				"type": "object",
				"properties": {
					"value": {"type": "string"},
					"children": {
						"type": "array",
						"items": {"$ref": "#/$defs/Node"}
					}
				}
			}
		}
	}`

	result, err := ExtractDefAsSchema(jsonSchema, "Node")
	require.NoError(t, err)

	// Parse result to verify structure
	var resultMap map[string]any
	err = json.Unmarshal([]byte(result), &resultMap)
	require.NoError(t, err)

	// Verify the self-reference is preserved
	props := resultMap["properties"].(map[string]any)
	children := props["children"].(map[string]any)
	items := children["items"].(map[string]any)
	require.Equal(t, "#/$defs/Node", items["$ref"])
}

func TestExtractDefAsSchema_DefinitionWithArrayOfRefs(t *testing.T) {
	jsonSchema := `{
		"type": "object",
		"$defs": {
			"Team": {
				"type": "object",
				"properties": {
					"name": {"type": "string"},
					"members": {
						"type": "array",
						"items": {"$ref": "#/$defs/Person"}
					}
				}
			},
			"Person": {
				"type": "object",
				"properties": {
					"name": {"type": "string"},
					"email": {"type": "string"}
				}
			}
		}
	}`

	result, err := ExtractDefAsSchema(jsonSchema, "Team")
	require.NoError(t, err)

	// Parse result to verify structure
	var resultMap map[string]any
	err = json.Unmarshal([]byte(result), &resultMap)
	require.NoError(t, err)

	// Verify $defs contains Person
	defs := resultMap["$defs"].(map[string]any)
	require.Contains(t, defs, "Person")
	require.Len(t, defs, 1)
}

func TestExtractDefAsSchema_NonExistentDefinition(t *testing.T) {
	jsonSchema := `{
		"type": "object",
		"$defs": {
			"Person": {
				"type": "object"
			}
		}
	}`

	_, err := ExtractDefAsSchema(jsonSchema, "NonExistent")
	require.Error(t, err)
	require.Contains(t, err.Error(), "definition \"NonExistent\" not found")
}

func TestExtractDefAsSchema_NoDefs(t *testing.T) {
	jsonSchema := `{
		"type": "object",
		"properties": {
			"name": {"type": "string"}
		}
	}`

	_, err := ExtractDefAsSchema(jsonSchema, "Person")
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

func TestExtractDefAsSchema_InvalidJSON(t *testing.T) {
	jsonSchema := `{invalid json`

	_, err := ExtractDefAsSchema(jsonSchema, "Person")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse JSON schema")
}

func TestExtractDefAsSchema_DefNotMap(t *testing.T) {
	jsonSchema := `{
		"type": "object",
		"$defs": {
			"Person": "not a map"
		}
	}`

	_, err := ExtractDefAsSchema(jsonSchema, "Person")
	require.Error(t, err)
	require.Contains(t, err.Error(), "not a valid object")
}

func TestExtractDefAsSchema_ComplexNestedStructure(t *testing.T) {
	jsonSchema := `{
		"type": "object",
		"$defs": {
			"Company": {
				"type": "object",
				"properties": {
					"name": {"type": "string"},
					"employees": {
						"type": "array",
						"items": {"$ref": "#/$defs/Employee"}
					},
					"headquarters": {"$ref": "#/$defs/Address"}
				}
			},
			"Employee": {
				"type": "object",
				"properties": {
					"name": {"type": "string"},
					"role": {"$ref": "#/$defs/Role"},
					"address": {"$ref": "#/$defs/Address"}
				}
			},
			"Role": {
				"type": "object",
				"properties": {
					"title": {"type": "string"},
					"department": {"type": "string"}
				}
			},
			"Address": {
				"type": "object",
				"properties": {
					"street": {"type": "string"},
					"city": {"type": "string"}
				}
			},
			"UnusedDef": {
				"type": "string"
			}
		}
	}`

	result, err := ExtractDefAsSchema(jsonSchema, "Company")
	require.NoError(t, err)

	// Parse result to verify structure
	var resultMap map[string]any
	err = json.Unmarshal([]byte(result), &resultMap)
	require.NoError(t, err)

	// Verify $defs contains all transitively referenced definitions
	defs := resultMap["$defs"].(map[string]any)
	require.Contains(t, defs, "Employee")
	require.Contains(t, defs, "Role")
	require.Contains(t, defs, "Address")
	require.NotContains(t, defs, "UnusedDef")
	require.Len(t, defs, 3)
}
