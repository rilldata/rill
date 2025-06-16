package openapiutil

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseJSONSchema(t *testing.T) {
	jsonSchema := `{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "object",
		"properties": {
			"expression": {
				"$ref": "#/$defs/Expression"
			},
			"metadata": {
				"$ref": "#/$defs/Metadata"
			}
		},
		"$defs": {
			"Expression": {
				"type": "object",
				"properties": {
					"operator": {
						"type": "string"
					},
					"left": {
						"$ref": "#/$defs/Expression"
					},
					"right": {
						"$ref": "#/$defs/Expression"
					}
				}
			},
			"Metadata": {
				"type": "object",
				"properties": {
					"name": {
						"type": "string"
					},
					"tags": {
						"type": "array",
						"items": {
							"type": "string"
						}
					}
				}
			}
		}
	}`

	// Convert the schema
	mainSchema, components, err := ParseJSONSchema("Prefix", jsonSchema)
	require.NoError(t, err)

	mainSchemaJSON, err := mainSchema.MarshalJSON()
	require.NoError(t, err)
	require.Equal(t, string(mainSchemaJSON), jsonRoundtrip(t, `{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"properties": {
			"expression": {
				"$ref": "#/components/schemas/PrefixExpression"
			},
			"metadata": {
				"$ref": "#/components/schemas/PrefixMetadata"
			}
		},
		"type": "object"
	}`))

	componentsJSON, err := json.Marshal(components)
	require.NoError(t, err)
	require.Equal(t, string(componentsJSON), jsonRoundtrip(t, `{
		"PrefixExpression": {
			"type": "object",
			"properties": {
				"left": {
					"$ref": "#/components/schemas/PrefixExpression"
				},
				"operator": {
					"type": "string"
				},
				"right": {
					"$ref": "#/components/schemas/PrefixExpression"
				}
			}
		},
		"PrefixMetadata": {
			"type": "object",
			"properties": {
				"name": {
					"type": "string"
				},
				"tags": {
					"type": "array",
					"items": {
						"type": "string"
					}
				}
			}
		}
	}`))
}

func jsonRoundtrip(t *testing.T, data string) string {
	var parsedData map[string]any
	err := json.Unmarshal([]byte(data), &parsedData)
	require.NoError(t, err)

	marshaledData, err := json.Marshal(parsedData)
	require.NoError(t, err)

	return string(marshaledData)
}
