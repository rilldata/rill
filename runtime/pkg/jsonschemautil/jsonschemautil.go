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
	// Extract the defs from the JSON schema
	defs, err := ExtractReferencedDefs(jsonSchema, defName)
	if err != nil {
		return "", err
	}

	// Build the result schema
	jsonData, err := json.Marshal(map[string]any{
		"$ref":  fmt.Sprintf("#/$defs/%s", defName),
		"$defs": defs,
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal result schema: %w", err)
	}

	return string(jsonData), nil
}

// MustExtractReferencedDefs wraps ExtractReferencedDefs and panics if an error occurs.
func MustExtractReferencedDefs(jsonSchema, defName string) map[string]any {
	result, err := ExtractReferencedDefs(jsonSchema, defName)
	if err != nil {
		panic(fmt.Sprintf("failed to extract referenced definitions for %q: %v", defName, err))
	}
	return result
}

// ExtractReferencedDefs extracts all definitions referenced by a specific definition in the $defs property of JSON schema.
// It returns a map of definition names to their corresponding JSON schema definitions.
// The resulting map will include the specified definition and any other definitions that are referenced by it.
func ExtractReferencedDefs(jsonSchema, defName string) (map[string]any, error) {
	// Extract the defs from the JSON schema
	var schema struct {
		Defs map[string]any `json:"$defs"`
	}
	if err := json.Unmarshal([]byte(jsonSchema), &schema); err != nil {
		return nil, fmt.Errorf("failed to parse JSON schema: %w", err)
	}
	defs := schema.Defs

	// Check the def exists
	if _, ok := defs[defName]; !ok {
		return nil, fmt.Errorf("definition %q not found in $defs", defName)
	}

	// Visit all the definitions
	visited := map[string]bool{defName: true}
	visitRefs(defs, defs[defName], visited)

	// Filter out defs that were not visited
	for defName := range defs {
		if !visited[defName] {
			delete(defs, defName)
		}
	}

	return defs, nil
}

// visitRefs recursively visits all definitions referenced in the given node.
func visitRefs(allDefs map[string]any, node any, visited map[string]bool) {
	switch v := node.(type) {
	case map[string]any:
		// Check for $ref
		if ref, ok := v["$ref"].(string); ok {
			if defName, ok := strings.CutPrefix(ref, "#/$defs/"); ok {
				// Skip if already visited (avoid infinite recursion)
				if visited[defName] {
					return
				}
				visited[defName] = true
				visitRefs(allDefs, allDefs[defName], visited)
			}
		}

		// Recursively process all values in the map
		for _, value := range v {
			visitRefs(allDefs, value, visited)
		}
	case []any:
		// Recursively process each item in the array
		for _, item := range v {
			visitRefs(allDefs, item, visited)
		}
	}
	// Primitive types can't contain a $ref, so nothing to do
}
