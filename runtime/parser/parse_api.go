package parser

import (
	"fmt"

	"github.com/rilldata/rill/runtime/pkg/openapiutil"
	"google.golang.org/protobuf/types/known/structpb"
)

// APIYAML is the raw structure of a API resource defined in YAML (does not include common fields)
type APIYAML struct {
	DataYAML           `yaml:",inline" mapstructure:",squash"`
	OpenAPI            *OpenAPIYAML        `yaml:"openapi"`
	Security           *SecurityPolicyYAML `yaml:"security"`
	SkipNestedSecurity bool                `yaml:"skip_nested_security"`
}

type OpenAPIYAML struct {
	Summary string `yaml:"summary"`
	Request struct {
		Parameters []map[string]any `yaml:"parameters"`
	} `yaml:"request"`
	Response struct {
		Schema map[string]any `yaml:"schema"`
	} `yaml:"response"`
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
	var openapiSummary string
	var openapiParams []*structpb.Struct
	var openapiSchema *structpb.Struct
	if tmp.OpenAPI != nil {
		openapiSummary = tmp.OpenAPI.Summary

		_, err := openapiutil.MapToParameters(tmp.OpenAPI.Request.Parameters)
		if err != nil {
			return fmt.Errorf("encountered invalid parameter type: %w", err)
		}
		for _, param := range tmp.OpenAPI.Request.Parameters {
			paramPB, err := structpb.NewStruct(param)
			if err != nil {
				return fmt.Errorf("encountered invalid parameter type: %w", err)
			}
			openapiParams = append(openapiParams, paramPB)
		}

		_, err = openapiutil.MapToSchema(tmp.OpenAPI.Response.Schema)
		if err != nil {
			return fmt.Errorf("encountered invalid schema type: %w", err)
		}
		openapiSchema, err = structpb.NewStruct(tmp.OpenAPI.Response.Schema)
		if err != nil {
			return fmt.Errorf("encountered invalid schema type: %w", err)
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
	r.APISpec.OpenapiParameters = openapiParams
	r.APISpec.OpenapiResponseSchema = openapiSchema
	r.APISpec.SecurityRules = securityRules
	r.APISpec.SkipNestedSecurity = tmp.SkipNestedSecurity

	return nil
}
