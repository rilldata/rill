package openapiutil

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

// ParseJSONParameters parses a JSON string representing OpenAPI parameters.
func ParseJSONParameters(jsonData string) (openapi3.Parameters, error) {
	var parameters openapi3.Parameters
	err := json.Unmarshal([]byte(jsonData), &parameters)
	if err != nil {
		return nil, err
	}
	return parameters, nil
}

// ParseJSONSchema parses a JSON schema into an OpenAPI schema.
//
// Notably, it also extracts the `$defs` from the JSON schema and rewrites its `$ref` paths.
// This is necessary because OpenAPI 3.0 and JSON schema differ in how they handle definitions.
// Where JSON schema uses `$defs` inside the schema, OpenAPI uses a global `components.schemas` section that is shared by all schemas in the document.
// We therefore need to extract the `$defs` into independent OpenAPI schemas, and then rewrite the `$ref` paths in the JSON schema to reference `#/components/schemas“ instead of `#/$defs`.
//
// Since the OpenAPI definitions have global scope, we also prefix the definition names with `namePrefix` to avoid collisions.
func ParseJSONSchema(namePrefix, jsonSchema string) (*openapi3.Schema, map[string]*openapi3.SchemaRef, error) {
	// Validate it is a valid JSON schema
	_, err := jsonschema.CompileString("schema", jsonSchema)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid JSON schema: %w", err)
	}

	// Parse it into a map
	var schemaMap map[string]any
	if err := json.Unmarshal([]byte(jsonSchema), &schemaMap); err != nil {
		return nil, nil, err
	}

	// Extract $defs into components
	components := make(map[string]*openapi3.SchemaRef)
	if defs, ok := schemaMap["$defs"].(map[string]any); ok {
		for defName, defSchema := range defs {
			err := rewriteJSONSchemaRefs(namePrefix, defSchema)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to rewrite refs for def %q: %w", defName, err)
			}

			defSchemaMap, ok := defSchema.(map[string]any)
			if !ok {
				return nil, nil, fmt.Errorf("failed to rewrite def %q: expected map, got %T", defName, defSchema)
			}

			openapiSchema, err := mapToSchema(defSchemaMap)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to convert def %q to OpenAPI: %w", defName, err)
			}

			components[namePrefix+defName] = &openapi3.SchemaRef{
				Value: openapiSchema,
			}
		}

		// Remove $defs from the main schema
		delete(schemaMap, "$defs")
	}

	// Rewrite $ref paths in the main schema
	err = rewriteJSONSchemaRefs(namePrefix, schemaMap)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to rewrite refs in main schema: %w", err)
	}

	// Convert the main schema to OpenAPI format
	mainSchema, err := mapToSchema(schemaMap)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to convert main schema to OpenAPI: %w", err)
	}

	return mainSchema, components, nil
}

// rewriteJSONSchemaRefs recursively rewrites $ref paths from "#/$defs/Name" to "#/components/schemas/<namePrefix>Name".
// This is necessary to convert a JSON schema to an OpenAPI schema. See the docstring for ParseJSONSchema for more details.
// This is an in-place operation that modifies the data structure directly.
func rewriteJSONSchemaRefs(namePrefix string, data any) error {
	switch v := data.(type) {
	case map[string]any:
		for key, value := range v {
			// If it's not a $ref, recursively process the value.
			if key != "$ref" {
				if err := rewriteJSONSchemaRefs(namePrefix, value); err != nil {
					return err
				}
				continue
			}

			refStr, ok := value.(string)
			if !ok {
				return fmt.Errorf("expected string value for $ref, got %T", value)
			}

			// Convert #/$defs/Name to #/components/schemas/<namePrefix>Name
			if strings.HasPrefix(refStr, "#/$defs/") {
				defName := strings.TrimPrefix(refStr, "#/$defs/")
				v[key] = fmt.Sprintf("#/components/schemas/%s%s", namePrefix, defName)
			}
		}
	case []any:
		// Recursively process each item
		for _, item := range v {
			if err := rewriteJSONSchemaRefs(namePrefix, item); err != nil {
				return err
			}
		}
	default:
		// For primitive types (string, number, bool, nil), nothing to do
		// Since this is in-place, we don't need to return anything
	}
	return nil
}

func mapToSchema(schema map[string]any) (*openapi3.Schema, error) {
	specSchema := openapi3.Schema{}

	jsonData, err := json.Marshal(schema)
	if err != nil {
		return nil, err
	}

	err = specSchema.UnmarshalJSON(jsonData)
	if err != nil {
		return nil, err
	}

	return &specSchema, nil
}
