package parser

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/google/jsonschema-go/jsonschema"
	"gopkg.in/yaml.v3"

	_ "embed"
)

//go:embed schema/rillyaml.schema.yaml
var rillYAMLSchema string

//go:embed schema/project.schema.yaml
var resourceYAMLSchema string

// Utils for parsing rillYAMLSchema and resourceYAMLSchema
var (
	parsedRillYAMLSchemaOnce sync.Once
	parsedRillYAMLSchema     *jsonschema.Schema

	parsedResourceYAMLSchemaOnce sync.Once
	parsedResourceYAMLSchema     *jsonschema.Schema
)

// resourceKindToDefinitionKey maps a ResourceKind to its key in project.schema.yaml.
var resourceKindToDefinitionKey = map[ResourceKind]string{
	ResourceKindSource:      "sources",
	ResourceKindModel:       "models",
	ResourceKindMetricsView: "metrics-views",
	ResourceKindExplore:     "explore-dashboards",
	ResourceKindAlert:       "alerts",
	ResourceKindTheme:       "themes",
	ResourceKindComponent:   "components",
	ResourceKindCanvas:      "canvas-dashboards",
	ResourceKindAPI:         "apis",
	ResourceKindConnector:   "connectors",
}

// JSONSchemaForRillYAML returns the JSON schema for validating rill.yaml files.
func JSONSchemaForRillYAML() (*jsonschema.Schema, error) {
	// Ensure the schema is parsed
	parsedRillYAMLSchemaOnce.Do(func() {
		var err error
		parsedRillYAMLSchema, err = parseSchemaFromYAML(rillYAMLSchema)
		if err != nil {
			panic(fmt.Sprintf("failed to parse schema: %v", err))
		}
	})

	return parsedRillYAMLSchema, nil
}

// JSONSchemaForResourceType returns a JSON schema for validating the properties of a given resource type.
// Note: You can use ParseResourceKind to get the ResourceKind from a string.
func JSONSchemaForResourceType(resourceType ResourceKind) (*jsonschema.Schema, error) {
	// Ensure the schema is parsed
	parsedResourceYAMLSchemaOnce.Do(func() {
		var err error
		parsedResourceYAMLSchema, err = parseSchemaFromYAML(resourceYAMLSchema)
		if err != nil {
			panic(fmt.Sprintf("failed to parse schema: %v", err))
		}
	})

	// Look up the definition key for this resource type
	defKey, ok := resourceKindToDefinitionKey[resourceType]
	if !ok {
		return nil, fmt.Errorf("no schema definition for resource type %v", resourceType)
	}

	// Get the definition from the schema
	defSchema, ok := parsedResourceYAMLSchema.Definitions[defKey]
	if !ok {
		return nil, fmt.Errorf("schema definition %q not found", defKey)
	}

	return defSchema, nil
}

// parseSchemaFromYAML parses a JSON schema from YAML content.
func parseSchemaFromYAML(yamlContent string) (*jsonschema.Schema, error) {
	// First parse YAML to a generic interface
	var yamlData any
	if err := yaml.Unmarshal([]byte(yamlContent), &yamlData); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Convert to JSON (this handles YAML-specific types like map[string]any)
	jsonBytes, err := json.Marshal(yamlData)
	if err != nil {
		return nil, fmt.Errorf("failed to convert YAML to JSON: %w", err)
	}

	// Parse into jsonschema.Schema
	var schema jsonschema.Schema
	if err := json.Unmarshal(jsonBytes, &schema); err != nil {
		return nil, fmt.Errorf("failed to parse JSON schema: %w", err)
	}

	return &schema, nil
}
