package rillv1

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/types/known/structpb"
)

// APIYAML is the raw structure of a API resource defined in YAML (does not include common fields)
type APIYAML struct {
	commonYAML `yaml:",inline" mapstructure:",squash"` // Only to avoid loading common fields into Properties
	resolver   string                                  `yaml:"resolver"`
	Metrics    struct {
		SQL         string   `yaml:"sql"`
		MetricsView string   `yaml:"metrics_view"`
		Measures    []string `yaml:"measures"`
		Where       string   `yaml:"where"`
		Limit       string   `yaml:"limit"`
	} `yaml:"metrics"`
	Properties map[string]any `yaml:",inline" mapstructure:",remain"`
}

// parseAPI parses an API definition and adds the resulting resource to p.Resources.
func (p *Parser) parseAPI(ctx context.Context, node *Node) error {
	tmp := &APIYAML{}
	err := p.decodeNodeYAML(node, true, tmp)
	if err != nil {
		return err
	}

	if node.SQL == "" && tmp.Metrics.SQL == "" && tmp.Metrics.MetricsView == "" {
		return fmt.Errorf("missing resolver")
	}

	r, err := p.insertResource(ResourceKindAPI, node.Name, node.Paths, node.Refs...)
	if err != nil {
		return err
	}

	if tmp.Properties == nil {
		tmp.Properties = make(map[string]any)
	}
	if node.SQL != "" {
		if tmp.resolver != "" && tmp.resolver != "SQLResolver" {
			return fmt.Errorf("resolver must be empty or SQLResolver")
		}
		r.APISpec.Resolver = "SQLResolver"
		tmp.Properties["sql"] = node.SQL
	} else if tmp.Metrics.MetricsView != "" || tmp.Metrics.SQL != "" {
		if tmp.resolver != "" && tmp.resolver != "MetricSQLResolver" {
			return fmt.Errorf("resolver must be empty or MetricSQLResolver")
		}
		r.APISpec.Resolver = "MetricSQLResolver"
		tmp.Properties["metrics"] = tmp.Metrics
	}

	props, err := structpb.NewStruct(tmp.Properties)
	if err != nil {
		return fmt.Errorf("encountered invalid property type: %w", err)
	}
	r.APISpec.ResolverProperties = mergeStructPB(r.APISpec.ResolverProperties, props)
	return nil
}
