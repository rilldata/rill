package rillv1

import (
	"fmt"

	"github.com/rilldata/rill/runtime/pkg/openapiutil"
	"google.golang.org/protobuf/types/known/structpb"
	"gopkg.in/yaml.v3"
)

// APIYAML is the raw structure of a API resource defined in YAML (does not include common fields)
type APIYAML struct {
	DataYAML `yaml:",inline" mapstructure:",squash"`
	OpenAPI  *OpenAPIYAML `yaml:"openapi"`
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
			return fmt.Errorf("the 'custom api' resource type only supports 'access' security rules")
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

	return nil
}

// DataYAML is the raw YAML structure of a sub-property for defining a data resolver and properties.
// It is used across multiple resources, usually under "data:", but inlined for APIs.
type DataYAML struct {
	Connector      string             `yaml:"connector"`
	SQL            string             `yaml:"sql"`
	MetricsSQL     string             `yaml:"metrics_sql"`
	API            string             `yaml:"api"`
	Args           map[string]any     `yaml:"args"`
	Glob           yaml.Node          `yaml:"glob"` // Path (string) or properties (map[string]any)
	ResourceStatus map[string]any     `yaml:"resource_status"`
	Security       SecurityPolicyYAML `yaml:"security"`
}

// parseDataYAML parses a data resolver and its properties from a DataYAML.
// The contextualConnector argument is optional; if provided and the resolver supports a connector, it becomes the default connector for the resolver.
// It returns the resolver name, its properties, and refs found in the resolver props.
func (p *Parser) parseDataYAML(raw *DataYAML, contextualConnector string) (string, *structpb.Struct, []ResourceName, error) {
	// Parse the resolver and its properties
	var count int
	var resolver string
	var refs []ResourceName
	resolverProps := make(map[string]any)

	// Handle basic SQL resolver
	if raw.SQL != "" {
		count++
		resolver = "sql"
		resolverProps["sql"] = raw.SQL
		if raw.Connector != "" {
			resolverProps["connector"] = raw.Connector
		} else if contextualConnector != "" {
			resolverProps["connector"] = contextualConnector
		}
	}

	// Handle metrics SQL resolver
	if raw.MetricsSQL != "" {
		count++
		resolver = "metrics_sql"
		resolverProps["sql"] = raw.MetricsSQL
	}

	// Handle API resolver
	if raw.API != "" {
		count++
		resolver = "api"
		resolverProps["api"] = raw.API
		refs = append(refs, ResourceName{Kind: ResourceKindAPI, Name: raw.API})
		if raw.Args != nil {
			resolverProps["args"] = raw.Args
		}
	}

	// Handle glob resolver
	if !raw.Glob.IsZero() {
		var props map[string]any
		switch raw.Glob.Kind {
		case yaml.ScalarNode:
			props = map[string]any{"path": raw.Glob.Value}
		default:
			props = make(map[string]any)
			err := raw.Glob.Decode(props)
			if err != nil {
				return "", nil, nil, fmt.Errorf("failed to parse glob properties: %w", err)
			}
		}

		count++
		resolver = "glob"
		resolverProps = props
	}

	// Handle resource_status resolver
	if raw.ResourceStatus != nil {
		count++
		resolver = "resource_status"
		resolverProps = raw.ResourceStatus
	}

	// Validate there was exactly one resolver
	if count == 0 {
		return "", nil, nil, fmt.Errorf(`the API definition does not specify a resolver (for example, "sql:", "metrics_sql:", ...)`)
	}
	if count > 1 {
		return "", nil, nil, fmt.Errorf(`the API definition specifies more than one resolver`)
	}

	// Convert resolver properties to structpb.Struct
	resolverPropsPB, err := structpb.NewStruct(resolverProps)
	if err != nil {
		return "", nil, nil, fmt.Errorf("encountered invalid property type: %w", err)
	}

	return resolver, resolverPropsPB, refs, nil
}
