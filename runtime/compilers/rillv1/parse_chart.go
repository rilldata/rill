package rillv1

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

type ChartYaml struct {
	commonYAML `yaml:",inline"` // Not accessed here, only setting it so we can use KnownFields for YAML parsing
	Title      string `yaml:"title"`
	Data       struct {
		Name     string         `yaml:"name"`
		Args     map[string]any `yaml:"args"`
		ArgsJSON string         `yaml:"args_json"`
	} `yaml:"data"`
	VegaLite string `yaml:"vega_lite"`
}

func (p *Parser) parseChart(ctx context.Context, node *Node) error {
	// Parse YAML
	tmp := &ChartYaml{}
	err := p.decodeNodeYAML(node, true, tmp)
	if err != nil {
		return err
	}

	// Validate SQL or connector isn't set
	if node.SQL != "" {
		return fmt.Errorf("charts cannot have SQL")
	}
	if !node.ConnectorInferred && node.Connector != "" {
		return fmt.Errorf("charts cannot have a connector")
	}

	// Query name
	if tmp.Data.Name == "" {
		return fmt.Errorf(`invalid value %q for property "data.name"`, tmp.Data.Name)
	}

	// Query args
	if tmp.Data.ArgsJSON != "" {
		// Validate JSON
		if !json.Valid([]byte(tmp.Data.ArgsJSON)) {
			return errors.New(`failed to parse "data.args_json" as JSON`)
		}
	} else {
		// Fall back to Data.args if data.args_json is not set
		data, err := json.Marshal(tmp.Data.Args)
		if err != nil {
			return fmt.Errorf(`failed to serialize "data.args" to JSON: %w`, err)
		}
		tmp.Data.ArgsJSON = string(data)
	}
	if tmp.Data.ArgsJSON == "" {
		return errors.New(`missing query args (must set either "data.args" or "data.args_json")`)
	}

	if tmp.VegaLite == "" {
		return errors.New(`missing vega_lite configuration`)
	}
	if !json.Valid([]byte(tmp.VegaLite)) {
		return errors.New(`failed to parse "vega_lite" as JSON`)
	}

	// Track chart
	r, err := p.insertResource(ResourceKindChart, node.Name, node.Paths, node.Refs...)
	if err != nil {
		return err
	}
	// NOTE: After calling insertResource, an error must not be returned. Any validation should be done before calling it.

	r.ChartSpec.Title = tmp.Title
	r.ChartSpec.QueryName = tmp.Data.Name
	r.ChartSpec.QueryArgsJson = tmp.Data.ArgsJSON
	r.ChartSpec.VegaLiteSpec = tmp.VegaLite

	return nil
}
