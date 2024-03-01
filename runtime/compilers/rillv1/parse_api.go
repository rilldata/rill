package rillv1

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/types/known/structpb"
)

// APIYAML is the raw structure of a API resource defined in YAML (does not include common fields)
type APIYAML struct {
	commonYAML `yaml:",inline" mapstructure:",squash"` // Only to avoid loading common fields into Properties
	Resolver   string                                  `yaml:"resolver"`
	Properties map[string]any                          `yaml:",inline" mapstructure:",remain"`
}

// parseAPI parses an API definition and adds the resulting resource to p.Resources.
func (p *Parser) parseAPI(ctx context.Context, node *Node) error {
	tmp := &APIYAML{}
	err := p.decodeNodeYAML(node, true, tmp)
	if err != nil {
		return err
	}

	_, hasMetrics := tmp.Properties["metrics"]
	if node.SQL == "" && !hasMetrics && tmp.Resolver == "" {
		return fmt.Errorf("missing resolver %v", tmp.Resolver)
	}

	r, err := p.insertResource(ResourceKindAPI, node.Name, node.Paths, node.Refs...)
	if err != nil {
		return err
	}

	if tmp.Properties == nil {
		tmp.Properties = make(map[string]any)
	}
	if node.SQL != "" {
		if tmp.Resolver != "" && tmp.Resolver != "SQLResolver" {
			return fmt.Errorf("resolver must be empty or SQLResolver")
		}
		r.APISpec.Resolver = "SQLResolver"
		tmp.Properties["sql"] = node.SQL
	} else if hasMetrics {
		if tmp.Resolver != "" && tmp.Resolver != "MetricSQLResolver" {
			return fmt.Errorf("resolver must be empty or MetricSQLResolver")
		}
		r.APISpec.Resolver = "MetricSQLResolver"
	} else {
		r.APISpec.Resolver = tmp.Resolver
	}

	props, err := structpb.NewStruct(tmp.Properties)
	if err != nil {
		return fmt.Errorf("encountered invalid property type: %w", err)
	}
	r.APISpec.ResolverProperties = mergeStructPB(r.APISpec.ResolverProperties, props)
	return nil
}
