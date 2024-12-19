package rillv1

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/duckdbsql"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"google.golang.org/protobuf/types/known/structpb"
)

// ModelYAML is the raw structure of a Model resource defined in YAML (does not include common fields)
type ModelYAML struct {
	commonYAML            `yaml:",inline" mapstructure:",squash"` // Only to avoid loading common fields into InputProperties
	Refresh               *ScheduleYAML                           `yaml:"refresh"`
	Timeout               string                                  `yaml:"timeout"`
	Incremental           bool                                    `yaml:"incremental"`
	State                 *DataYAML                               `yaml:"state"`
	Partitions            *DataYAML                               `yaml:"partitions"`
	Splits                *DataYAML                               `yaml:"splits"` // Deprecated: use "partitions" instead
	PartitionsWatermark   string                                  `yaml:"partitions_watermark"`
	PartitionsConcurrency uint                                    `yaml:"partitions_concurrency"`
	InputProperties       map[string]any                          `yaml:",inline" mapstructure:",remain"`
	Stage                 struct {
		Connector  string         `yaml:"connector"`
		Properties map[string]any `yaml:",inline" mapstructure:",remain"`
	} `yaml:"stage"`
	Output struct {
		Connector  string         `yaml:"connector"`
		Properties map[string]any `yaml:",inline" mapstructure:",remain"`
	} `yaml:"output"`
	Materialize *bool `yaml:"materialize"`
}

// parseModel parses a model definition and adds the resulting resource to p.Resources.
func (p *Parser) parseModel(ctx context.Context, node *Node) error {
	// Parse YAML
	tmp := &ModelYAML{}
	err := p.decodeNodeYAML(node, false, tmp)
	if err != nil {
		return err
	}

	// Parse refresh schedule
	schedule, err := p.parseScheduleYAML(tmp.Refresh)
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
	}

	// special handling to mark model as updated when local file changes
	if inputConnector == "local_file" {
		err = p.trackResourceNamesForDataPaths(ctx, ResourceName{Name: node.Name, Kind: ResourceKindModel}.Normalized(), inputProps)
		if err != nil {
			return err
		}
	}

	inputPropsPB, err := structpb.NewStruct(inputProps)
	if err != nil {
		return fmt.Errorf(`found invalid input property type: %w`, err)
	}

	// Stage details are optional
	var stagePropsPB *structpb.Struct
	if len(tmp.Stage.Properties) > 0 {
		stagePropsPB, err = structpb.NewStruct(tmp.Stage.Properties)
		if err != nil {
			return fmt.Errorf(`found invalid input property type: %w`, err)
		}
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

	// Parse incremental state resolver
	var incrementalStateResolver string
	var incrementalStateResolverProps *structpb.Struct
	if tmp.State != nil {
		var refs []ResourceName
		incrementalStateResolver, incrementalStateResolverProps, refs, err = p.parseDataYAML(tmp.State, outputConnector)
		if err != nil {
			return fmt.Errorf(`failed to parse "state": %w`, err)
		}
		node.Refs = append(node.Refs, refs...)
	}

	// Parse partitions resolver
	var partitionsResolver string
	var partitionsResolverProps *structpb.Struct
	if tmp.Splits != nil { // Backwards compatibility: "splits" is deprecated and has been renamed to "partitions"
		if tmp.Partitions != nil {
			return fmt.Errorf(`"partitions" and "splits" are mutually exclusive`)
		}
		tmp.Partitions = tmp.Splits
	}
	if tmp.Partitions != nil {
		var refs []ResourceName
		partitionsResolver, partitionsResolverProps, refs, err = p.parseDataYAML(tmp.Partitions, inputConnector)
		if err != nil {
			return fmt.Errorf(`failed to parse "partitions": %w`, err)
		}
		node.Refs = append(node.Refs, refs...)

		// As a small convenience, automatically set the watermark field for resolvers where we know a good default
		if tmp.PartitionsWatermark == "" {
			if partitionsResolver == "glob" {
				tmp.PartitionsWatermark = "updated_on"
			}
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

	r.ModelSpec.IncrementalStateResolver = incrementalStateResolver
	r.ModelSpec.IncrementalStateResolverProperties = incrementalStateResolverProps

	r.ModelSpec.PartitionsResolver = partitionsResolver
	r.ModelSpec.PartitionsResolverProperties = partitionsResolverProps
	r.ModelSpec.PartitionsWatermarkField = tmp.PartitionsWatermark
	r.ModelSpec.PartitionsConcurrencyLimit = uint32(tmp.PartitionsConcurrency)

	r.ModelSpec.InputConnector = inputConnector
	r.ModelSpec.InputProperties = inputPropsPB

	r.ModelSpec.StageConnector = tmp.Stage.Connector
	r.ModelSpec.StageProperties = stagePropsPB

	r.ModelSpec.OutputConnector = outputConnector
	r.ModelSpec.OutputProperties = outputPropsPB

	return nil
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

func (p *Parser) trackResourceNamesForDataPaths(ctx context.Context, name ResourceName, inputProps map[string]any) error {
	c, ok := inputProps["invalidate_on_change"].(bool)
	if ok && !c {
		return nil
	}
	path, ok := inputProps["path"].(string)
	if !ok {
		return nil
	}

	var localPaths []string
	if fileutil.IsGlob(path) {
		entries, err := p.Repo.ListRecursive(ctx, path, true)
		if err != nil || len(entries) == 0 {
			// The actual error will be returned by the model reconciler
			return nil
		}

		for _, entry := range entries {
			localPaths = append(localPaths, entry.Path)
		}
	} else {
		localPaths = []string{normalizePath(path)}
	}

	// Update parser's resourceNamesForDataPaths map to track which resources depend on the local file
	for _, path := range localPaths {
		resources := p.resourceNamesForDataPaths[path]
		if !slices.Contains(resources, name) {
			resources = append(resources, name)
			p.resourceNamesForDataPaths[path] = resources
		}
	}

	// Calculate hash of local files
	hash, err := p.Repo.FileHash(ctx, localPaths)
	if err != nil {
		return err
	}
	// Add hash to input properties so that the model spec is considered updated when the local file changes
	inputProps["local_files_hash"] = hash
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
