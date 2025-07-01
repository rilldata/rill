package parser

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode"

	"github.com/rilldata/rill/runtime/pkg/openapiutil"
)

// APIYAML is the raw structure of a API resource defined in YAML (does not include common fields)
type APIYAML struct {
	DataYAML           `yaml:",inline" mapstructure:",squash"`
	OpenAPI            *OpenAPIYAML        `yaml:"openapi"`
	Security           *SecurityPolicyYAML `yaml:"security"`
	SkipNestedSecurity bool                `yaml:"skip_nested_security"`
}

type OpenAPIYAML struct {
	Summary        string           `yaml:"summary"`
	Parameters     []map[string]any `yaml:"parameters"`
	RequestSchema  map[string]any   `yaml:"request_schema"`
	ResponseSchema map[string]any   `yaml:"response_schema"`
	DefsPrefix     *string          `yaml:"defs_prefix,omitempty"` // Optional prefix for definitions in the OpenAPI spec. Defaults to the API name in Pascal case.
}

// parseAPI parses an API definition and adds the resulting resource to p.Resources.
func (p *Parser) parseAPI(node *Node) error {
	// Parse YAML
	tmp := &APIYAML{}
	err := p.decodeNodeYAML(node, false, tmp)
	if err != nil {
		return err
	}

	// Validate
	var openapiSummary, openapiParams, openapiRequestSchema, openapiResponseSchema, openapiDefsPrefix string
	if tmp.OpenAPI != nil {
		openapiSummary = tmp.OpenAPI.Summary

		if len(tmp.OpenAPI.Parameters) != 0 {
			paramsJSON, err := json.Marshal(tmp.OpenAPI.Parameters)
			if err != nil {
				return fmt.Errorf("invalid openapi.parameters: %w", err)
			}
			_, err = openapiutil.ParseJSONParameters(string(paramsJSON))
			if err != nil {
				return fmt.Errorf("invalid openapi.parameters: %w", err)
			}
			openapiParams = string(paramsJSON)
		}

		if len(tmp.OpenAPI.RequestSchema) != 0 {
			requestSchemaJSON, err := json.Marshal(tmp.OpenAPI.RequestSchema)
			if err != nil {
				return fmt.Errorf("invalid openapi.request_schema: %w", err)
			}
			_, _, err = openapiutil.ParseJSONSchema(node.Name, string(requestSchemaJSON))
			if err != nil {
				return fmt.Errorf("invalid openapi.request_schema: %w", err)
			}
			openapiRequestSchema = string(requestSchemaJSON)
		}

		if len(tmp.OpenAPI.ResponseSchema) != 0 {
			responseSchemaJSON, err := json.Marshal(tmp.OpenAPI.ResponseSchema)
			if err != nil {
				return fmt.Errorf("invalid openapi.response_schema: %w", err)
			}
			_, _, err = openapiutil.ParseJSONSchema(node.Name, string(responseSchemaJSON))
			if err != nil {
				return fmt.Errorf("invalid openapi.response_schema: %w", err)
			}
			openapiResponseSchema = string(responseSchemaJSON)
		}

		if tmp.OpenAPI.DefsPrefix == nil {
			// Default to the API name in PascalCase
			openapiDefsPrefix = toPascalCase(node.Name)
		} else {
			openapiDefsPrefix = *tmp.OpenAPI.DefsPrefix
		}
	}

	// Map common node properties to DataYAML
	if !node.ConnectorInferred && node.Connector != "" {
		tmp.DataYAML.Connector = node.Connector
	}
	if node.SQL != "" {
		tmp.DataYAML.SQL = node.SQL
	}

	// Parse the resolver and its properties from the DataYAML
	resolver, resolverProps, resolverRefs, err := p.parseDataYAML(&tmp.DataYAML, node.Connector)
	if err != nil {
		return err
	}
	node.Refs = append(node.Refs, resolverRefs...)

	securityRules, err := tmp.Security.Proto()
	if err != nil {
		return fmt.Errorf("failed to parse security rules: %w", err)
	}
	for _, rule := range securityRules {
		if rule.GetAccess() == nil {
			return fmt.Errorf("the 'api' resource type only supports 'access' security rules")
		}
	}

	r, err := p.insertResource(ResourceKindAPI, node.Name, node.Paths, node.Refs...)
	if err != nil {
		return err
	}
	// NOTE: After calling insertResource, an error must not be returned. Any validation should be done before calling it.

	r.APISpec.Resolver = resolver
	r.APISpec.ResolverProperties = resolverProps
	r.APISpec.OpenapiSummary = openapiSummary
	r.APISpec.OpenapiParametersJson = openapiParams
	r.APISpec.OpenapiRequestSchemaJson = openapiRequestSchema
	r.APISpec.OpenapiResponseSchemaJson = openapiResponseSchema
	r.APISpec.OpenapiDefsPrefix = openapiDefsPrefix
	r.APISpec.SecurityRules = securityRules
	r.APISpec.SkipNestedSecurity = tmp.SkipNestedSecurity

	return nil
}

// toPascalCase converts a string to PascalCase.
// The string may contain underscores and dashes, which will be treated as word separators.
func toPascalCase(s string) string {
	if s == "" {
		return s
	}

	words := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-' || r == ' '
	})

	for i, word := range words {
		if word != "" {
			words[i] = string(unicode.ToUpper(rune(word[0]))) + word[1:]
		}
	}

	return strings.Join(words, "")
}
