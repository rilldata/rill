package rillv1

import (
	"context"
	"fmt"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

// APIYAML
type APIYAML struct {
	SQL        string         `yaml:"sql"`
	MetricsSQL *MetricSQLYAML `yaml:"metrics_sql"`
}

type MetricSQLYAML struct {
	SQL         string   `yaml:"sql"`
	MetricsView string   `yaml:"metrics_view"`
	Measures    []string `yaml:"measures"`
	Where       string   `yaml:"where"`
	Limit       string   `yaml:"limit"`
}

// parseAPI parses an API definition and adds the resulting resource to p.Resources.
func (p *Parser) parseAPI(ctx context.Context, node *Node) error {
	tmp := &APIYAML{}
	err := p.decodeNodeYAML(node, false, tmp)
	if err != nil {
		return err
	}

	if tmp.SQL == "" && tmp.MetricsSQL == nil {
		return fmt.Errorf("no SQL provided")
	}

	r, err := p.insertResource(ResourceKindAPI, node.Name, node.Paths, node.Refs...)
	if err != nil {
		return err
	}

	r.APISpec.Sql = strings.TrimSpace(tmp.SQL)
	if tmp.MetricsSQL != nil {
		r.APISpec.Metrics = new(runtimev1.MetricSQL)
		r.APISpec.Metrics.Sql = strings.TrimSpace(tmp.MetricsSQL.SQL)
		r.APISpec.Metrics.MetricsView = tmp.MetricsSQL.MetricsView
		r.APISpec.Metrics.Measures = tmp.MetricsSQL.Measures
		r.APISpec.Metrics.Where = strings.TrimSpace(tmp.MetricsSQL.Where)
		r.APISpec.Metrics.Limit = strings.TrimSpace(tmp.MetricsSQL.Limit)
	}

	return nil
}
