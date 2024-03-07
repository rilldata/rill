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
	var resolvers int
	resolverProps := make(map[string]any)

	// Handle basic SQL resolver
	if node.SQL != "" {
		resolvers++
		resolver = "SQL" // TODO: Replace with a constant when the resolver abstractions are implemented
		resolverProps["connector"] = node.Connector
		resolverProps["sql"] = node.SQL
	}

	// Handle metrics resolver
	if tmp.Metrics != nil {
		if !node.ConnectorInferred && node.Connector != "" {
			return fmt.Errorf(`can't set "connector" for the metrics resolver (it will use the connector of the metrics view)`)
		}

		resolvers++
		resolver = "Metrics" // TODO: Replace with a constant when the resolver abstractions are implemented
		resolverProps["sql"] = tmp.Metrics.SQL
		// NOTE: May add support for outright dimensions:, measures:, etc. here
	}

	// Validate there was exactly one resolver
	if resolvers == 0 {
		return fmt.Errorf(`the API definition does not specify a resolver (for example, "sql:", "metrics:", ...)`)
	}
	if resolvers > 1 {
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
