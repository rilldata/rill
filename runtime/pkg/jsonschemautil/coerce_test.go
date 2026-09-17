package jsonschemautil

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/stretchr/testify/require"
)

// coerceTestSchema is a trimmed version of metricsview.QueryJSONSchema.
// It is embedded here instead of imported because metricsview depends on this package.
const coerceTestSchema = `{
	"type": "object",
	"properties": {
		"metrics_view": {"type": "string"},
		"contents": {"type": "string"},
		"dimensions": {
			"type": "array",
			"items": {"$ref": "#/$defs/Dimension"}
		},
		"pivot_on": {
			"type": "array",
			"items": {"type": "string"}
		},
		"time_range": {"$ref": "#/$defs/TimeRange"},
		"where": {"$ref": "#/$defs/Expression"},
		"color": {
			"anyOf": [
				{"$ref": "#/$defs/TimeRange"},
				{"type": "string"}
			]
		},
		"limit": {"type": "integer"}
	},
	"$defs": {
		"Dimension": {
			"type": "object",
			"properties": {
				"name": {"type": "string"}
			}
		},
		"TimeRange": {
			"type": "object",
			"properties": {
				"start": {"type": "string"},
				"end": {"type": "string"}
			}
		},
		"Expression": {
			"type": "object",
			"properties": {
				"name": {"type": "string"},
				"val": {},
				"cond": {"$ref": "#/$defs/Condition"}
			}
		},
		"Condition": {
			"type": "object",
			"properties": {
				"op": {"type": "string"},
				"exprs": {
					"type": "array",
					"items": {"$ref": "#/$defs/Expression"}
				}
			}
		}
	}
}`

func TestCoerceStringifiedJSON(t *testing.T) {
	tests := []struct {
		name        string
		schema      string
		args        string
		want        string
		wantChanged bool
	}{
		{
			name:        "stringified object via ref",
			schema:      coerceTestSchema,
			args:        `{"metrics_view": "mv", "time_range": "{\"start\": \"2024-01-01T00:00:00Z\", \"end\": \"2024-02-01T00:00:00Z\"}"}`,
			want:        `{"metrics_view": "mv", "time_range": {"start": "2024-01-01T00:00:00Z", "end": "2024-02-01T00:00:00Z"}}`,
			wantChanged: true,
		},
		{
			name:        "stringified expression with nested condition",
			schema:      coerceTestSchema,
			args:        `{"where": "{\"cond\": {\"op\": \"and\", \"exprs\": [{\"name\": \"country\"}]}}"}`,
			want:        `{"where": {"cond": {"op": "and", "exprs": [{"name": "country"}]}}}`,
			wantChanged: true,
		},
		{
			name:        "stringified array nested inside a real object",
			schema:      coerceTestSchema,
			args:        `{"where": {"cond": {"op": "and", "exprs": "[{\"name\": \"country\"}]"}}}`,
			want:        `{"where": {"cond": {"op": "and", "exprs": [{"name": "country"}]}}}`,
			wantChanged: true,
		},
		{
			name:        "stringified array of objects",
			schema:      coerceTestSchema,
			args:        `{"dimensions": "[{\"name\": \"country\"}]"}`,
			want:        `{"dimensions": [{"name": "country"}]}`,
			wantChanged: true,
		},
		{
			name:        "stringified element inside a real array",
			schema:      coerceTestSchema,
			args:        `{"dimensions": [{"name": "country"}, "{\"name\": \"device\"}"]}`,
			want:        `{"dimensions": [{"name": "country"}, {"name": "device"}]}`,
			wantChanged: true,
		},
		{
			name:        "untyped field is untouched",
			schema:      coerceTestSchema,
			args:        `{"where": {"name": "country", "val": "{\"looks\": \"like json\"}"}}`,
			want:        `{"where": {"name": "country", "val": "{\"looks\": \"like json\"}"}}`,
			wantChanged: false,
		},
		{
			name:        "string-typed field holding JSON text is untouched",
			schema:      coerceTestSchema,
			args:        `{"contents": "{\"type\": \"model\"}"}`,
			want:        `{"contents": "{\"type\": \"model\"}"}`,
			wantChanged: false,
		},
		{
			name:        "anyOf allowing string is untouched",
			schema:      coerceTestSchema,
			args:        `{"color": "{\"start\": \"2024-01-01T00:00:00Z\"}"}`,
			want:        `{"color": "{\"start\": \"2024-01-01T00:00:00Z\"}"}`,
			wantChanged: false,
		},
		{
			name:        "invalid JSON string is untouched",
			schema:      coerceTestSchema,
			args:        `{"time_range": "last 7 days"}`,
			want:        `{"time_range": "last 7 days"}`,
			wantChanged: false,
		},
		{
			name:        "wrong-kind JSON string is untouched",
			schema:      coerceTestSchema,
			args:        `{"time_range": "[1, 2]"}`,
			want:        `{"time_range": "[1, 2]"}`,
			wantChanged: false,
		},
		{
			name:        "JSON string with trailing garbage is untouched",
			schema:      coerceTestSchema,
			args:        `{"time_range": "{\"start\": \"2024-01-01T00:00:00Z\"} trailing"}`,
			want:        `{"time_range": "{\"start\": \"2024-01-01T00:00:00Z\"} trailing"}`,
			wantChanged: false,
		},
		{
			name:        "JSON string with trailing close brace is untouched",
			schema:      coerceTestSchema,
			args:        `{"time_range": "{\"start\": \"2024-01-01T00:00:00Z\"}}"}`,
			want:        `{"time_range": "{\"start\": \"2024-01-01T00:00:00Z\"}}"}`,
			wantChanged: false,
		},
		{
			name:        "JSON string with trailing close bracket is untouched",
			schema:      coerceTestSchema,
			args:        `{"dimensions": "[{\"name\": \"country\"}]]"}`,
			want:        `{"dimensions": "[{\"name\": \"country\"}]]"}`,
			wantChanged: false,
		},
		{
			name:        "well-formed args are unchanged",
			schema:      coerceTestSchema,
			args:        `{"metrics_view": "mv", "time_range": {"start": "2024-01-01T00:00:00Z"}, "dimensions": [{"name": "country"}], "limit": 100}`,
			want:        `{"metrics_view": "mv", "time_range": {"start": "2024-01-01T00:00:00Z"}, "dimensions": [{"name": "country"}], "limit": 100}`,
			wantChanged: false,
		},
		{
			name:        "entire arguments as one stringified object",
			schema:      coerceTestSchema,
			args:        `"{\"metrics_view\": \"mv\", \"time_range\": \"{\\\"start\\\": \\\"2024-01-01T00:00:00Z\\\"}\"}"`,
			want:        `{"metrics_view": "mv", "time_range": {"start": "2024-01-01T00:00:00Z"}}`,
			wantChanged: true,
		},
		{
			name: "nullable type union is coerced",
			schema: `{
				"type": "object",
				"properties": {
					"where": {"type": ["null", "object"], "properties": {"name": {"type": "string"}}}
				}
			}`,
			args:        `{"where": "{\"name\": \"country\"}"}`,
			want:        `{"where": {"name": "country"}}`,
			wantChanged: true,
		},
		{
			name: "type union allowing string is untouched",
			schema: `{
				"type": "object",
				"properties": {
					"where": {"type": ["string", "object"]}
				}
			}`,
			args:        `{"where": "{\"name\": \"country\"}"}`,
			want:        `{"where": "{\"name\": \"country\"}"}`,
			wantChanged: false,
		},
		{
			name: "nested defs scope without root defs",
			schema: `{
				"type": "object",
				"properties": {
					"where": {
						"$ref": "#/$defs/Expression",
						"$defs": {
							"Expression": {
								"type": "object",
								"properties": {
									"name": {"type": "string"},
									"cond": {"$ref": "#/$defs/Condition"}
								}
							},
							"Condition": {
								"type": "object",
								"properties": {
									"op": {"type": "string"}
								}
							}
						}
					}
				}
			}`,
			args:        `{"where": "{\"cond\": \"{\\\"op\\\": \\\"and\\\"}\"}"}`,
			want:        `{"where": {"cond": {"op": "and"}}}`,
			wantChanged: true,
		},
		{
			name: "additionalProperties map values are coerced",
			schema: `{
				"type": "object",
				"properties": {
					"where_per_metrics_view": {
						"type": "object",
						"additionalProperties": {
							"type": "object",
							"properties": {"name": {"type": "string"}}
						}
					}
				}
			}`,
			args:        `{"where_per_metrics_view": {"mv1": "{\"name\": \"country\"}"}}`,
			want:        `{"where_per_metrics_view": {"mv1": {"name": "country"}}}`,
			wantChanged: true,
		},
		{
			name: "property defined in an allOf branch is coerced",
			schema: `{
				"type": "object",
				"allOf": [
					{"properties": {"where": {"type": "object", "properties": {"name": {"type": "string"}}}}},
					{"properties": {"mode": {"type": "string"}}}
				]
			}`,
			args:        `{"where": "{\"name\": \"country\"}"}`,
			want:        `{"where": {"name": "country"}}`,
			wantChanged: true,
		},
		{
			name: "property defined in an anyOf branch is untouched",
			schema: `{
				"type": "object",
				"anyOf": [
					{"properties": {"where": {"type": "object", "properties": {"name": {"type": "string"}}}}},
					{"properties": {"mode": {"type": "string"}}}
				]
			}`,
			args:        `{"where": "{\"name\": \"country\"}"}`,
			want:        `{"where": "{\"name\": \"country\"}"}`,
			wantChanged: false,
		},
		{
			name: "items defined in a oneOf branch are untouched",
			schema: `{
				"type": "object",
				"properties": {
					"dimensions": {
						"type": "array",
						"oneOf": [
							{"items": {"type": "object", "properties": {"name": {"type": "string"}}}},
							{"maxItems": 10}
						]
					}
				}
			}`,
			args:        `{"dimensions": ["{\"name\": \"country\"}"]}`,
			want:        `{"dimensions": ["{\"name\": \"country\"}"]}`,
			wantChanged: false,
		},
		{
			name: "allOf-wrapped ref is coerced",
			schema: `{
				"type": "object",
				"properties": {
					"time_range": {"allOf": [{"$ref": "#/$defs/TimeRange"}]}
				},
				"$defs": {
					"TimeRange": {
						"type": "object",
						"properties": {"start": {"type": "string"}}
					}
				}
			}`,
			args:        `{"time_range": "{\"start\": \"2024-01-01T00:00:00Z\"}"}`,
			want:        `{"time_range": {"start": "2024-01-01T00:00:00Z"}}`,
			wantChanged: true,
		},
		{
			name: "allOf branches intersecting to object are coerced",
			schema: `{
				"type": "object",
				"properties": {
					"where": {"allOf": [{"type": ["object", "string"]}, {"type": "object"}]}
				}
			}`,
			args:        `{"where": "{\"name\": \"country\"}"}`,
			want:        `{"where": {"name": "country"}}`,
			wantChanged: true,
		},
		{
			name: "allOf branch allowing string is untouched",
			schema: `{
				"type": "object",
				"properties": {
					"where": {"allOf": [{"type": ["object", "string"]}]}
				}
			}`,
			args:        `{"where": "{\"name\": \"country\"}"}`,
			want:        `{"where": "{\"name\": \"country\"}"}`,
			wantChanged: false,
		},
		{
			name: "cyclic allOf ref is untouched instead of overflowing the stack",
			schema: `{
				"type": "object",
				"properties": {
					"where": {"$ref": "#/$defs/A"},
					"dimensions": {"$ref": "#/$defs/A"}
				},
				"$defs": {
					"A": {"allOf": [{"$ref": "#/$defs/A"}]}
				}
			}`,
			args:        `{"where": {"cond": "{\"op\": \"and\"}"}, "dimensions": ["{\"name\": \"country\"}"]}`,
			want:        `{"where": {"cond": "{\"op\": \"and\"}"}, "dimensions": ["{\"name\": \"country\"}"]}`,
			wantChanged: false,
		},
		{
			name: "ref with a string-permitting sibling type is untouched",
			schema: `{
				"type": "object",
				"properties": {
					"time_range": {"$ref": "#/$defs/TimeRange", "type": "string"}
				},
				"$defs": {
					"TimeRange": {
						"type": "object",
						"properties": {"start": {"type": "string"}}
					}
				}
			}`,
			args:        `{"time_range": "{\"start\": \"2024-01-01T00:00:00Z\"}"}`,
			want:        `{"time_range": "{\"start\": \"2024-01-01T00:00:00Z\"}"}`,
			wantChanged: false,
		},
		{
			name: "patternProperties-matched key is not coerced via additionalProperties",
			schema: `{
				"type": "object",
				"properties": {
					"where_per_metrics_view": {
						"type": "object",
						"patternProperties": {"^x_": {"type": "string"}},
						"additionalProperties": {"type": "object", "properties": {"name": {"type": "string"}}}
					}
				}
			}`,
			args:        `{"where_per_metrics_view": {"x_note": "{\"name\": \"country\"}", "mv1": "{\"name\": \"country\"}"}}`,
			want:        `{"where_per_metrics_view": {"x_note": "{\"name\": \"country\"}", "mv1": {"name": "country"}}}`,
			wantChanged: true,
		},
		{
			name: "array with prefixItems is untouched",
			schema: `{
				"type": "object",
				"properties": {
					"dimensions": {
						"type": "array",
						"prefixItems": [{"type": "string"}],
						"items": {"type": "object", "properties": {"name": {"type": "string"}}}
					}
				}
			}`,
			args:        `{"dimensions": ["{\"name\": \"country\"}", "{\"name\": \"state\"}"]}`,
			want:        `{"dimensions": ["{\"name\": \"country\"}", "{\"name\": \"state\"}"]}`,
			wantChanged: false,
		},
		{
			name: "unresolvable ref is untouched",
			schema: `{
				"type": "object",
				"properties": {
					"where": {"$ref": "#/$defs/Missing"}
				}
			}`,
			args:        `{"where": "{\"name\": \"country\"}"}`,
			want:        `{"where": "{\"name\": \"country\"}"}`,
			wantChanged: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var schema jsonschema.Schema
			require.NoError(t, json.Unmarshal([]byte(tt.schema), &schema))

			var args any
			require.NoError(t, json.Unmarshal([]byte(tt.args), &args))

			got, changed := CoerceStringifiedJSON(&schema, args)
			require.Equal(t, tt.wantChanged, changed)

			gotJSON, err := json.Marshal(got)
			require.NoError(t, err)
			require.JSONEq(t, tt.want, string(gotJSON))
		})
	}
}

func TestCoerceStringifiedJSONPreservesLargeIntegers(t *testing.T) {
	var schema jsonschema.Schema
	require.NoError(t, json.Unmarshal([]byte(coerceTestSchema), &schema))

	var args any
	dec := json.NewDecoder(strings.NewReader(`{"limit": 9007199254740993, "time_range": "{\"start\": \"2024-01-01T00:00:00Z\"}"}`))
	dec.UseNumber()
	require.NoError(t, dec.Decode(&args))

	got, changed := CoerceStringifiedJSON(&schema, args)
	require.True(t, changed)

	gotJSON, err := json.Marshal(got)
	require.NoError(t, err)
	require.Contains(t, string(gotJSON), "9007199254740993")
}
