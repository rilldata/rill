package rillv1

import (
	"errors"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/duckdbsql"
)

// ModelYAML is the raw structure of a Model resource defined in YAML (does not include common fields)
type ModelYAML struct {
	Materialize  *bool         `yaml:"materialize" mapstructure:"materialize"`
	Timeout      string        `yaml:"timeout" mapstructure:"timeout"`
	Refresh      *ScheduleYAML `yaml:"refresh" mapstructure:"refresh"`
	ParserConfig struct {
		DuckDB struct {
			InferRefs *bool `yaml:"infer_refs" mapstructure:"infer_refs"`
		} `yaml:"duckdb" mapstructure:"duckdb"`
	} `yaml:"parser"`
}

// parseModel parses a model definition and adds the resulting resource to p.Resources.
func (p *Parser) parseModel(node *Node) error {
	// Parse YAML
	tmp := &ModelYAML{}
	err := p.decodeNodeYAML(node, false, tmp)
	if err != nil {
		return err
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

	// Validate SQL
	if strings.TrimSpace(node.SQL) == "" {
		return errors.New("no SQL provided")
	}

	// If the connector is a DuckDB connector, extract info using DuckDB SQL parsing.
	// (If templating was used, we skip DuckDB inference because the DuckDB parser may not be able to parse the templated code.)
	driver, _, _ := p.driverForConnector(node.Connector)
	isDuckDB := driver == "duckdb"
	duckDBInferRefs := true
	if tmp.ParserConfig.DuckDB.InferRefs != nil {
		duckDBInferRefs = *tmp.ParserConfig.DuckDB.InferRefs
	}
	refs := node.Refs
	if isDuckDB && !node.SQLUsesTemplating && node.SQL != "" && duckDBInferRefs {
		// Parse the SQL
		ast, err := duckdbsql.Parse(node.SQL)
		if err != nil {
			var posError duckdbsql.PositionError
			if errors.As(err, &posError) {
				return pathError{
					path: node.SQLPath,
					err: locationError{
						err: posError.Err(),
						location: &runtimev1.CharLocation{
							Line: uint32(findLineNumber(node.SQL, posError.Position)),
						},
					},
				}
			}
			return pathError{path: node.SQLPath, err: newDuckDBError(err)}
		}

		// Scan SQL for table references, tracking references in refs
		for _, t := range ast.GetTableRefs() {
			if !t.LocalAlias && t.Name != "" && t.Function == "" && len(t.Paths) == 0 {
				refs = append(refs, ResourceName{Name: t.Name})
			}
		}
	}

	// Insert the model
	r, err := p.insertResource(ResourceKindModel, node.Name, node.Paths, refs...)
	if err != nil {
		return err
	}
	// NOTE: After calling insertResource, an error must not be returned. Any validation should be done before calling it.

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

// findLineNumber returns the line number of the pos in the given text.
// Lines are counted starting from 1, and positions start from 0.
func findLineNumber(text string, pos int) int {
	if pos < 0 || pos >= len(text) {
		return -1
	}

	lineNumber := 1
	for i, char := range text {
		if i == pos {
			break
		}
		if char == '\n' {
			lineNumber++
		}
	}

	return lineNumber
}
