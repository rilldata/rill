package rillv1

import (
	"fmt"

	"google.golang.org/protobuf/types/known/structpb"
)

// APIYAML is the raw structure of a API resource defined in YAML (does not include common fields)
type APIYAML struct {
	DataYAML `yaml:",inline" mapstructure:",squash"`
}

// parseAPI parses an API definition and adds the resulting resource to p.Resources.
func (p *Parser) parseAPI(node *Node) error {
	// Parse YAML
	tmp := &APIYAML{}
	err := p.decodeNodeYAML(node, false, tmp)
	if err != nil {
		return err
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
		resolver = "SQL"
		resolverProps["sql"] = raw.SQL
		resolverProps["connector"] = raw.Connector
	}

	// Handle metrics SQL resolver
	if raw.MetricsSQL != "" {
		count++
		resolver = "MetricsSQL"
		resolverProps["sql"] = raw.MetricsSQL
	}

	// Handle API resolver
	if raw.API != "" {
		count++
		resolver = "API"
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
