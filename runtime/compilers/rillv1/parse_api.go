package rillv1

import (
	"fmt"

	"google.golang.org/protobuf/types/known/structpb"
)

// APIYAML is the raw structure of a API resource defined in YAML (does not include common fields)
type APIYAML struct {
	commonYAML `yaml:",inline" mapstructure:",squash"` // Only to avoid loading common fields into Properties
	Metrics    *struct {
		SQL string `yaml:"sql"`
	} `yaml:"metrics"`
}

// parseAPI parses an API definition and adds the resulting resource to p.Resources.
func (p *Parser) parseAPI(node *Node) error {
	// Parse YAML
	tmp := &APIYAML{}
	err := p.decodeNodeYAML(node, true, tmp)
	if err != nil {
		return err
	}

	// Parse the resolver and its properties
	var resolver string
	var count int
	resolverProps := make(map[string]any)

	// Handle basic SQL resolver
	if node.SQL != "" {
		count++
		resolver = "SQL" // TODO: Replace with a constant when the resolver abstractions are implemented
		resolverProps["connector"] = node.Connector
		resolverProps["sql"] = node.SQL
	}

	// Handle metrics resolver
	if tmp.Metrics != nil {
		if !node.ConnectorInferred && node.Connector != "" {
			return fmt.Errorf(`can't set "connector" for the metrics resolver (it will use the connector of the metrics view)`)
		}

		count++
		resolver = "MetricsSQL" // TODO: Replace with a constant when the resolver abstractions are implemented
		resolverProps["sql"] = tmp.Metrics.SQL
		// NOTE: May add support for outright dimensions:, measures:, etc. here
	}

	// Validate there was exactly one resolver
	if count == 0 {
		return fmt.Errorf(`the API definition does not specify a resolver (for example, "sql:", "metrics:", ...)`)
	}
	if count > 1 {
		return fmt.Errorf(`the API definition specifies more than one resolver`)
	}

	// Convert resolver properties to structpb.Struct before inserting the resource (since we can't return errors after that point)
	resolverPropsPB, err := structpb.NewStruct(resolverProps)
	if err != nil {
		return fmt.Errorf("encountered invalid property type: %w", err)
	}

	r, err := p.insertResource(ResourceKindAPI, node.Name, node.Paths, node.Refs...)
	if err != nil {
		return err
	}
	// NOTE: After calling insertResource, an error must not be returned. Any validation should be done before calling it.

	r.APISpec.Resolver = resolver
	r.APISpec.ResolverProperties = resolverPropsPB

	return nil
}

// DataYAML is the raw YAML structure of a sub-property for defining a data resolver and properties.
// It is used across multiple resources, usually under "data:", but inlined for APIss.
type DataYAML struct {
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

	// TODO: Handle basic SQL resolver

	// Handle metrics resolver
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
		return "", nil, nil, fmt.Errorf(`the API definition does not specify a resolver (for example, "sql:", "metrics:", ...)`)
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
