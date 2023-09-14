package rillv1

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rilldata/rill/runtime/pkg/duckdbsql"
)

// modelYAML is the raw structure of a Model resource defined in YAML (does not include common fields)
type modelYAML struct {
	Materialize  *bool         `yaml:"materialize" mapstructure:"materialize"`
	Timeout      string        `yaml:"timeout" mapstructure:"timeout"`
	Refresh      *scheduleYAML `yaml:"refresh" mapstructure:"refresh"`
	ParserConfig struct {
		DuckDB struct {
			InferRefs *bool `yaml:"infer_refs" mapstructure:"infer_refs"`
		} `yaml:"duckdb" mapstructure:"duckdb"`
	} `yaml:"parser"`
}

// parseModel parses a model definition and adds the resulting resource to p.Resources.
func (p *Parser) parseModel(ctx context.Context, node *Node) error {
	// Parse YAML
	tmp := &modelYAML{}
	if node.YAML != nil {
		if err := node.YAML.Decode(tmp); err != nil {
			return pathError{path: node.YAMLPath, err: newYAMLError(err)}
		}
	}

	// Override YAML config with SQL annotations
	err := mapstructureUnmarshal(node.SQLAnnotations, tmp)
	if err != nil {
		return pathError{path: node.SQLPath, err: fmt.Errorf("invalid SQL annotations: %w", err)}
	}

	// Parse timeout
	var timeout time.Duration
	if tmp.Timeout != "" {
		timeout, err = parseDuration(tmp.Timeout)
		if err != nil {
			return err
		}
	}

	// Parse refresh schedule
	schedule, err := parseScheduleYAML(tmp.Refresh)
	if err != nil {
		return err
	}

	// If the connector is a DuckDB connector, extract info using DuckDB SQL parsing.
	// (If templating was used, we skip DuckDB inference because the DuckDB parser may not be able to parse the templated code.)
	isDuckDB := false
	for _, c := range p.DuckDBConnectors {
		if c == node.Connector {
			isDuckDB = true
			break
		}
	}
	duckDBInferRefs := true
	if tmp.ParserConfig.DuckDB.InferRefs != nil {
		duckDBInferRefs = *tmp.ParserConfig.DuckDB.InferRefs
	}
	refs := node.Refs
	if isDuckDB && !node.SQLUsesTemplating && node.SQL != "" && duckDBInferRefs {
		// Parse the SQL
		ast, err := duckdbsql.Parse(node.SQL)
		if err != nil {
			return pathError{path: node.SQLPath, err: newDuckDBError(err)}
		}

		// Scan SQL for table references, tracking references in refs
		for _, t := range ast.GetTableRefs() {
			if !t.LocalAlias && t.Name != "" && t.Function == "" && len(t.Paths) == 0 {
				refs = append(refs, ResourceName{Name: t.Name})
			}
		}
	}

	// NOTE: After calling upsertResource, an error must not be returned. Any validation should be done before calling it.

	// Upsert the model
	r := p.upsertResource(ResourceKindModel, node.Name, node.Paths, refs...)
	if node.SQL != "" {
		r.ModelSpec.Sql = strings.TrimSpace(node.SQL)
		r.ModelSpec.UsesTemplating = node.SQLUsesTemplating
	}
	if node.Connector != "" {
		r.ModelSpec.Connector = node.Connector
	}
	if tmp.Materialize != nil {
		r.ModelSpec.Materialize = tmp.Materialize
	}
	if timeout > 0 {
		r.ModelSpec.TimeoutSeconds = uint32(timeout.Seconds())
	}
	if schedule != nil {
		r.ModelSpec.RefreshSchedule = schedule
	}

	// parseSource calls parseModel for SQL sources without a connector. Materialize such models.
	if node.Kind == ResourceKindSource && r.ModelSpec.Materialize == nil {
		b := true
		r.ModelSpec.Materialize = &b
	}

	return nil
}
