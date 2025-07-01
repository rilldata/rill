package jsonschemautil

import (
	"encoding/json"
	"fmt"
	"strings"
)

// MustExtractDefAsSchema wraps ExtractDefAsSchema and panics if an error occurs.
func MustExtractDefAsSchema(jsonSchema, defName string) string {
	result, err := ExtractDefAsSchema(jsonSchema, defName)
	if err != nil {
		panic(fmt.Sprintf("failed to extract definition %q: %v", defName, err))
	}
	return result
}

// ExtractDefAsSchema extracts a specific definition from the $defs property of JSON schema by its name.
// The resulting schema's $defs will contain any other definitions that are referenced by the extracted definition.
// It returns the definition as a JSON schema string.
func ExtractDefAsSchema(jsonSchema, defName string) (string, error) {
	// Parse the JSON schema into a map
	var schemaMap struct {
		Defs map[string]any `json:"$defs"`
	}
	if err := json.Unmarshal([]byte(jsonSchema), &schemaMap); err != nil {
		return "", fmt.Errorf("failed to parse JSON schema: %w", err)
	}

	// Find the requested definition
	targetDefAny, ok := schemaMap.Defs[defName]
	if !ok {
		return "", fmt.Errorf("definition %q not found in $defs", defName)
	}
	targetDef, ok := targetDefAny.(map[string]any)
	if !ok {
		return "", fmt.Errorf("definition %q is not a valid object", defName)
	}

	// Collect all referenced definitions
	referencedDefs := make(map[string]any)
	collectReferencedDefs(targetDef, schemaMap.Defs, referencedDefs, make(map[string]bool))

	// Build the result schema
	resultSchema := make(map[string]any)

	// Copy all properties from the target definition to the root
	for k, v := range targetDef {
		resultSchema[k] = v
	}

	// Add $defs if there are any referenced definitions
	if len(referencedDefs) > 0 {
		resultSchema["$defs"] = referencedDefs
	}

	// Convert back to JSON
	jsonData, err := json.Marshal(resultSchema)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result schema: %w", err)
	}

	return string(jsonData), nil
}

// collectReferencedDefs recursively collects all definitions referenced by the given data.
func collectReferencedDefs(data any, allDefs, collectedDefs map[string]any, visited map[string]bool) {
	switch v := data.(type) {
	case map[string]any:
		// Check for $ref
		if ref, ok := v["$ref"].(string); ok {
			if strings.HasPrefix(ref, "#/$defs/") {
				defName := strings.TrimPrefix(ref, "#/$defs/")

				// Skip if already visited (avoid infinite recursion)
				if visited[defName] {
					return
				}
				visited[defName] = true

				// Add the referenced definition if it exists
				if def, exists := allDefs[defName]; exists {
					collectedDefs[defName] = def
					// Recursively collect refs from this definition
					collectReferencedDefs(def, allDefs, collectedDefs, visited)
				}
			}
		}

		// Recursively process all values in the map
		for _, value := range v {
			collectReferencedDefs(value, allDefs, collectedDefs, visited)
		}

	case []any:
		// Recursively process each item in the array
		for _, item := range v {
			collectReferencedDefs(item, allDefs, collectedDefs, visited)
		}
	}
	// For primitive types (string, number, bool, nil), nothing to do
}
