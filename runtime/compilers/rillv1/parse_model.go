package rillv1

import (
	"errors"
	"fmt"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/duckdbsql"
	"google.golang.org/protobuf/types/known/structpb"
)

// ModelYAML is the raw structure of a Model resource defined in YAML (does not include common fields)
type ModelYAML struct {
	commonYAML      `yaml:",inline" mapstructure:",squash"` // Only to avoid loading common fields into Properties
	Refresh         *ScheduleYAML                           `yaml:"refresh"`
	Timeout         string                                  `yaml:"timeout"`
	Incremental     bool                                    `yaml:"incremental"`
	State           *DataYAML                               `yaml:"state"`
	InputProperties map[string]any                          `yaml:",inline" mapstructure:",remain"`
	Output          struct {
		Connector  string         `yaml:"connector"`
		Properties map[string]any `yaml:",inline" mapstructure:",remain"`
	} `yaml:"output"`
	Materialize *bool `yaml:"materialize"`
}

// parseModel parses a model definition and adds the resulting resource to p.Resources.
func (p *Parser) parseModel(node *Node) error {
	// Parse YAML
	tmp := &ModelYAML{}
	err := p.decodeNodeYAML(node, false, tmp)
	if err != nil {
		return err
	}

	// Parse refresh schedule
	schedule, err := parseScheduleYAML(tmp.Refresh)
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

	// Parse state resolver
	var stateResolver string
	var stateResolverProps *structpb.Struct
	if tmp.State != nil {
		var refs []ResourceName
		stateResolver, stateResolverProps, refs, err = p.parseDataYAML(tmp.State)
		if err != nil {
			return fmt.Errorf(`failed to parse "state": %w`, err)
		}
		node.Refs = append(node.Refs, refs...)
	}

	// Build input details
	inputConnector := node.Connector
	inputProps := tmp.InputProperties
	if inputProps == nil {
		inputProps = map[string]any{}
	}

	// Special handling for adding SQL to the input properties
	if sql := strings.TrimSpace(node.SQL); sql != "" {
		refs, err := p.inferSQLRefs(node)
		if err != nil {
			return err
		}
		node.Refs = append(node.Refs, refs...)

		inputProps["sql"] = sql
		if node.SQLUsesTemplating {
			inputProps["uses_templating"] = node.SQLUsesTemplating
		}
	}

	// Validate input details
	if len(inputProps) == 0 {
		return errors.New(`model does not identify any input properties (try entering a SQL query)`)
	}
	inputPropsPB, err := structpb.NewStruct(inputProps)
	if err != nil {
		return fmt.Errorf(`found invalid input property type: %w`, err)
	}

	// Build output details
	outputConnector := tmp.Output.Connector
	if outputConnector == "" {
		outputConnector = inputConnector
	}
	outputProps := tmp.Output.Properties

	// Backwards compatibility: materialize can be specified outside of the output properties
	if tmp.Materialize != nil {
		if outputProps == nil {
			outputProps = map[string]any{}
		}
		outputProps["materialize"] = *tmp.Materialize
	}

	// Validate output details
	var outputPropsPB *structpb.Struct
	if len(outputProps) > 0 {
		outputPropsPB, err = structpb.NewStruct(outputProps)
		if err != nil {
			return fmt.Errorf(`invalid property type in "output": %w`, err)
		}
	}

	// Insert the model
	r, err := p.insertResource(ResourceKindModel, node.Name, node.Paths, node.Refs...)
	if err != nil {
		return err
	}
	// NOTE: After calling insertResource, an error must not be returned. Any validation should be done before calling it.

	if schedule != nil {
		r.ModelSpec.RefreshSchedule = schedule
	}

	if timeout > 0 {
		r.ModelSpec.TimeoutSeconds = uint32(timeout.Seconds())
	}

	r.ModelSpec.Incremental = tmp.Incremental

	r.ModelSpec.StateResolver = stateResolver
	r.ModelSpec.StateResolverProperties = stateResolverProps

	r.ModelSpec.InputConnector = inputConnector
	r.ModelSpec.InputProperties = inputPropsPB

	r.ModelSpec.OutputConnector = outputConnector
	r.ModelSpec.OutputProperties = outputPropsPB

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

// inferSQLRefs attempts to infer table references from the node's SQL.
// The provided node must have a non-empty SQL field.
func (p *Parser) inferSQLRefs(node *Node) ([]ResourceName, error) {
	// Currently only supports DuckDB.
	driver, _, _ := p.driverForConnector(node.Connector)
	if driver != "duckdb" {
		return nil, nil
	}

	// Skip if the SQL is templated (because the DuckDB parser may choke on the template handlebars)
	if node.SQLUsesTemplating {
		return nil, nil
	}

	// Parse the SQL
	ast, err := duckdbsql.Parse(node.SQL)
	if err != nil {
		path := node.SQLPath
		if path == "" {
			path = node.YAMLPath
		}

		var posError duckdbsql.PositionError
		if errors.As(err, &posError) {
			return nil, pathError{
				path: path,
				err: locationError{
					err: posError.Err(),
					location: &runtimev1.CharLocation{
						Line: uint32(findLineNumber(node.SQL, posError.Position)),
					},
				},
			}
		}
		return nil, pathError{path: path, err: newDuckDBError(err)}
	}

	// Scan SQL for table references, tracking references in refs
	var refs []ResourceName
	for _, t := range ast.GetTableRefs() {
		if !t.LocalAlias && t.Name != "" && t.Function == "" && len(t.Paths) == 0 {
			refs = append(refs, ResourceName{Name: t.Name})
		}
	}

	return refs, nil
}
