package rillv1

import (
	"encoding/json"
	"errors"
	"fmt"
)

type ChartYaml struct {
	commonYAML `yaml:",inline"` // Not accessed here, only setting it so we can use KnownFields for YAML parsing
	Title      string `yaml:"title"`
	Data       struct {
		MetricsSql string            `yaml:"metrics_sql"`
		API        string            `yaml:"api"`
		Args       map[string]string `yaml:"args"`
	} `yaml:"data"`
	VegaLite string `yaml:"vega_lite"`
}

func (p *Parser) parseChart(node *Node) error {
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

	if tmp.VegaLite == "" {
		return errors.New(`missing vega_lite configuration`)
	}
	if !json.Valid([]byte(tmp.VegaLite)) {
		return errors.New(`failed to parse "vega_lite" as JSON`)
	}

	if (tmp.Data.MetricsSql == "" && tmp.Data.API == "") || (tmp.Data.MetricsSql != "" && tmp.Data.API != "") {
		return fmt.Errorf("exactly one of metrics_sql or api should be set")
	}

	if tmp.Data.API != "" {
		node.Refs = append(node.Refs, ResourceName{Kind: ResourceKindAPI, Name: tmp.Data.API})
	}

	// Track chart
	r, err := p.insertResource(ResourceKindChart, node.Name, node.Paths, node.Refs...)
	if err != nil {
		return err
	}
	// NOTE: After calling insertResource, an error must not be returned. Any validation should be done before calling it.

	r.ChartSpec.Title = tmp.Title
	r.ChartSpec.MetricsSql = tmp.Data.MetricsSql
	r.ChartSpec.Api = tmp.Data.API
	r.ChartSpec.Args = tmp.Data.Args
	r.ChartSpec.VegaLiteSpec = tmp.VegaLite

	return nil
}
