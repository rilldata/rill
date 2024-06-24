package rillv1

import (
	"encoding/json"
	"fmt"

	"github.com/go-openapi/spec"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

// APIYAML is the raw structure of a API resource defined in YAML (does not include common fields)
type APIYAML struct {
	DataYAML `yaml:",inline" mapstructure:",squash"`
	OpenAPI  *OpenAPIYAML `yaml:"openapi"`
}

type OpenAPIYAML struct {
	Info     map[string]any `yaml:"info"`
	Request  *ReqSpecYAML   `yaml:"request"`
	Response *RespSpecYAML  `yaml:"response"`
}

type ReqSpecYAML struct {
	Summary    string           `yaml:"summary"`
	Parameters []map[string]any `yaml:"parameters"`
}

type RespSpecYAML struct {
	Description string         `yaml:"description"`
	Schema      map[string]any `yaml:"schema"`
}

func ConvertParameters(params []map[string]any) ([]spec.Parameter, error) {
	var parameters []spec.Parameter
	for _, param := range params {
		var specParam spec.Parameter
		jsonData, err := json.Marshal(param)
		if err != nil {
			return nil, err
		}
		err = specParam.UnmarshalJSON(jsonData)
		if err != nil {
			return nil, err
		}
		parameters = append(parameters, specParam)
	}
	return parameters, nil
}

func ConvertSchema(schema map[string]any) (*spec.Schema, error) {
	specSchema := spec.Schema{}
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

// parseAPI parses an API definition and adds the resulting resource to p.Resources.
func (p *Parser) parseAPI(node *Node) error {
	// Parse YAML
	tmp := &APIYAML{}
	err := p.decodeNodeYAML(node, false, tmp)
	if err != nil {
		return err
	}

	// Validate
	var reqSummary string
	var reqParams []*structpb.Struct
	if tmp.OpenAPI != nil && tmp.OpenAPI.Request != nil {
		reqSummary = tmp.OpenAPI.Request.Summary
		_, err := ConvertParameters(tmp.OpenAPI.Request.Parameters)
		if err != nil {
			return fmt.Errorf("encountered invalid parameter type: %w", err)
		}
		for _, param := range tmp.OpenAPI.Request.Parameters {
			paramPB, err := structpb.NewStruct(param)
			if err != nil {
				return fmt.Errorf("encountered invalid parameter type: %w", err)
			}
			reqParams = append(reqParams, paramPB)
		}
	}
	var resDescription string
	var resSchema *structpb.Struct
	if tmp.OpenAPI != nil && tmp.OpenAPI.Response != nil {
		resDescription = tmp.OpenAPI.Response.Description
		_, err := ConvertSchema(tmp.OpenAPI.Response.Schema)
		if err != nil {
			return fmt.Errorf("encountered invalid schema type: %w", err)
		}
		resSchema, err = structpb.NewStruct(tmp.OpenAPI.Response.Schema)
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
	resolver, resolverProps, resolverRefs, err := p.parseDataYAML(&tmp.DataYAML)
	if err != nil {
		return err
	}
	node.Refs = append(node.Refs, resolverRefs...)

	r, err := p.insertResource(ResourceKindAPI, node.Name, node.Paths, node.Refs...)
	if err != nil {
		return err
	}
	// NOTE: After calling insertResource, an error must not be returned. Any validation should be done before calling it.

	r.APISpec.Resolver = resolver
	r.APISpec.ResolverProperties = resolverProps
	r.APISpec.OpenApiSpec = &runtimev1.OpenAPISpec{
		ReqSummary:     reqSummary,
		ReqParams:      reqParams,
		ResDescription: resDescription,
		ResSchema:      resSchema,
	}

	return nil
}

// DataYAML is the raw YAML structure of a sub-property for defining a data resolver and properties.
// It is used across multiple resources, usually under "data:", but inlined for APIs.
type DataYAML struct {
	Connector  string         `yaml:"connector"`
	SQL        string         `yaml:"sql"`
	MetricsSQL string         `yaml:"metrics_sql"`
	API        string         `yaml:"api"`
	Args       map[string]any `yaml:"args"`
}

// parseDataYAML parses a data resolver and its properties from a DataYAML.
// It returns the resolver name, its properties, and refs found in the resolver props.
func (p *Parser) parseDataYAML(raw *DataYAML) (string, *structpb.Struct, []ResourceName, error) {
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
		resolverProps["connector"] = raw.Connector
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
