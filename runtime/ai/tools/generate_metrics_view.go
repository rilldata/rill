package tools

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/tool"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

type generateMetricsViewInput struct {
	ModelName       string `mapstructure:"model_name"`
	MetricsViewName string `mapstructure:"metrics_view_name"`
	DashboardName   string `mapstructure:"dashboard_name"`
}

func newGenerateMetricsViewInput(in map[string]any) (*generateMetricsViewInput, error) {
	var input generateMetricsViewInput
	if err := mapstructure.Decode(in, &input); err != nil {
		return nil, fmt.Errorf("failed to decode input: %w", err)
	}
	if err := input.validate(); err != nil {
		return nil, fmt.Errorf("input validation failed: %w", err)
	}
	return &input, nil
}

func (i *generateMetricsViewInput) validate() error {
	if i.ModelName == "" {
		return fmt.Errorf("model_name parameter is required and must be a string")
	}
	if i.MetricsViewName == "" {
		return fmt.Errorf("metrics_view_name parameter is required and must be a string")
	}
	if i.DashboardName == "" {
		return fmt.Errorf("dashboard_name parameter is required and must be a string")
	}
	return nil
}

func GenerateMetricsViewYAML(instanceID string, rt *runtime.Runtime) *tool.FunctionTool {
	tool := tool.NewFunctionTool(
		"generate_metrics_view_yaml",
		"Generates a YAML configuration for a metrics view based on the provided model name",
		func(ctx context.Context, params map[string]any) (any, error) {
			input, err := newGenerateMetricsViewInput(params)
			if err != nil {
				return nil, err
			}

			olap, release, err := rt.OLAP(ctx, instanceID, "")
			if err != nil {
				return nil, fmt.Errorf("failed to get OLAP connection: %w", err)
			}
			defer release()

			tbl, err := olap.InformationSchema().Lookup(ctx, "", "", input.ModelName)
			if err != nil {
				return generateMetricsViewYAMLResult("", fmt.Errorf("failed to lookup table schema: %w", err)), nil
			}

			yaml, err := generateMetricsViewYAMLSimple("", tbl, true, true)
			if err != nil {
				return generateMetricsViewYAMLResult("", fmt.Errorf("failed to generate metrics view YAML: %w", err)), nil
			}

			_, err = putResourceAndWaitForReconcile(ctx, rt, instanceID, fmt.Sprintf("metrics_views/%s.yaml", input.MetricsViewName), yaml, &runtimev1.ResourceName{
				Kind: runtime.ResourceKindMetricsView,
				Name: input.ModelName,
			})
			if err != nil {
				return map[string]any{
					"error": fmt.Sprintf("Failed to create or reconcile metrics view resource: %s", err.Error()),
				}, nil
			}
			// also create a dashboard for the metrics view
			dashboardYAML := fmt.Sprintf(dashboardYAML, input.MetricsViewName, input.MetricsViewName)
			_, err = putResourceAndWaitForReconcile(ctx, rt, instanceID, fmt.Sprintf("dashboards/%s.yaml", input.DashboardName), dashboardYAML, &runtimev1.ResourceName{
				Kind: runtime.ResourceKindExplore,
				Name: input.DashboardName,
			})
			if err != nil {
				return map[string]any{
					"error": fmt.Sprintf("Failed to create or reconcile dashboard resource: %s", err.Error()),
				}, nil
			}
			return map[string]any{
				"result": fmt.Sprintf("Metrics view '%s' and dashboard '%s' created successfully", input.MetricsViewName, input.DashboardName),
			}, nil
		},
	)

	tool.WithSchema(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"model_name": map[string]any{
				"type":        "string",
				"description": "The name of the model to generate the metrics view for",
			},
			"metrics_view_name": map[string]any{
				"type":        "string",
				"description": "The name of the metrics view to create",
			},
			"dashboard_name": map[string]any{
				"type":        "string",
				"description": "The name of the dashboard to create",
			},
		},
		"required": []string{"model_name", "metrics_view_name", "dashboard_name"},
	})
	return tool
}

func generateMetricsViewYAMLResult(yaml string, err error) map[string]any {
	if err != nil {
		return map[string]any{
			"message": "metrics view YAML generation failed",
			"error":   err.Error(),
		}
	}
	return map[string]any{
		"message": "YAML generated successfully",
		"yaml":    yaml,
	}
}

// generateMetricsViewYAMLSimple generates a simple metrics view YAML definition from a table schema.
func generateMetricsViewYAMLSimple(connector string, tbl *drivers.OlapTable, isDefaultConnector, isModel bool) (string, error) {
	doc := &metricsViewYAML{
		Version:       1,
		Type:          "metrics_view",
		DisplayName:   identifierToDisplayName(tbl.Name),
		TimeDimension: generateMetricsViewYAMLSimpleTimeDimension(tbl.Schema),
		Dimensions:    generateMetricsViewYAMLSimpleDimensions(tbl.Schema),
		Measures:      generateMetricsViewYAMLSimpleMeasures(tbl),
	}

	if isModel {
		doc.Model = tbl.Name
	} else {
		if !isDefaultConnector {
			doc.Connector = connector
		}
		if tbl.Database != "" && !tbl.IsDefaultDatabase {
			doc.Database = tbl.Database
		}
		if tbl.DatabaseSchema != "" && !tbl.IsDefaultDatabaseSchema {
			doc.DatabaseSchema = tbl.DatabaseSchema
		}
		doc.Model = tbl.Name // Note: We also reference externally managed tables with `model:`. This is supported in the metrics view YAML.
	}

	return marshalMetricsViewYAML(doc, false)
}

func generateMetricsViewYAMLSimpleTimeDimension(schema *runtimev1.StructType) string {
	for _, f := range schema.Fields {
		switch f.Type.Code {
		case runtimev1.Type_CODE_TIMESTAMP, runtimev1.Type_CODE_TIME, runtimev1.Type_CODE_DATE:
			return f.Name
		}
	}
	return ""
}

func generateMetricsViewYAMLSimpleDimensions(schema *runtimev1.StructType) []*metricsViewDimensionYAML {
	var dims []*metricsViewDimensionYAML
	for _, f := range schema.Fields {
		switch f.Type.Code {
		case runtimev1.Type_CODE_BOOL, runtimev1.Type_CODE_STRING, runtimev1.Type_CODE_BYTES, runtimev1.Type_CODE_UUID, runtimev1.Type_CODE_DATE, runtimev1.Type_CODE_TIMESTAMP:
			dims = append(dims, &metricsViewDimensionYAML{
				Name:        f.Name,
				DisplayName: identifierToDisplayName(f.Name),
				Column:      f.Name,
			})
		}
	}
	return dims
}

func generateMetricsViewYAMLSimpleMeasures(tbl *drivers.OlapTable) []*metricsViewMeasureYAML {
	// Add a count measure
	var measures []*metricsViewMeasureYAML
	measures = append(measures, &metricsViewMeasureYAML{
		Name:         "total_records",
		DisplayName:  "Total records",
		Expression:   "COUNT(*)",
		Description:  "",
		FormatPreset: "humanize",
	})

	// Add sum measures for float columns
	for _, f := range tbl.Schema.Fields {
		switch f.Type.Code {
		case runtimev1.Type_CODE_FLOAT32, runtimev1.Type_CODE_FLOAT64:
			measures = append(measures, &metricsViewMeasureYAML{
				Name:         fmt.Sprintf("%s_sum", f.Name),
				DisplayName:  fmt.Sprintf("Sum of %s", identifierToDisplayName(f.Name)),
				Expression:   fmt.Sprintf("SUM(%s)", safeSQLName(f.Name)),
				Description:  "",
				FormatPreset: "humanize",
			})
		}
	}

	// Create a map of column names, which are used to ensure the generated measure names do not collide with column names.
	columns := make(map[string]struct{})
	for _, f := range tbl.Schema.Fields {
		columns[f.Name] = struct{}{}
	}

	// If a measure name collides with a table column name, append `_measure` until it's unique
	for _, m := range measures {
		for i := 0; i < 10; i++ {
			if _, ok := columns[m.Name]; !ok {
				break
			}
			m.Name += "_measure"
		}
	}

	return measures
}

// metricsViewYAML is a struct for generating a metrics view YAML file.
// We do not use the parser's structs since they are not suitable for generating pretty output YAML.
type metricsViewYAML struct {
	Version        int                         `yaml:"version,omitempty"`
	Type           string                      `yaml:"type,omitempty"`
	DisplayName    string                      `yaml:"display_name,omitempty"`
	Connector      string                      `yaml:"connector,omitempty"`
	Database       string                      `yaml:"database,omitempty"`
	DatabaseSchema string                      `yaml:"database_schema,omitempty"`
	Model          string                      `yaml:"model,omitempty"`
	TimeDimension  string                      `yaml:"timeseries,omitempty"`
	Dimensions     []*metricsViewDimensionYAML `yaml:"dimensions,omitempty"`
	Measures       []*metricsViewMeasureYAML   `yaml:"measures,omitempty"`
}

type metricsViewDimensionYAML struct {
	Name        string `yaml:"name"`
	DisplayName string `yaml:"display_name"`
	Column      string `yaml:"column"`
}

type metricsViewMeasureYAML struct {
	Name         string `yaml:"name"`
	DisplayName  string `yaml:"display_name"`
	Expression   string `yaml:"expression"`
	Description  string `yaml:"description"`
	FormatPreset string `yaml:"format_preset,omitempty"`
}

func marshalMetricsViewYAML(doc *metricsViewYAML, aiPowered bool) (string, error) {
	buf := new(bytes.Buffer)

	buf.WriteString("# Metrics view YAML\n")
	buf.WriteString("# Reference documentation: https://docs.rilldata.com/reference/project-files/dashboards\n")
	if aiPowered {
		buf.WriteString("# This file was generated using AI.\n")
	}
	buf.WriteString("\n")

	yamlBytes, err := yaml.Marshal(doc)
	if err != nil {
		return "", err
	}

	var rootNode yaml.Node
	if err := yaml.Unmarshal(yamlBytes, &rootNode); err != nil {
		return "", err
	}

	insertEmptyLinesInYaml(&rootNode)

	enc := yaml.NewEncoder(buf)
	enc.SetIndent(2)
	if err := enc.Encode(&rootNode); err != nil {
		return "", err
	}

	if err := enc.Close(); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func insertEmptyLinesInYaml(node *yaml.Node) {
	for i := 0; i < len(node.Content); i++ {
		if node.Content[i].Kind == yaml.MappingNode {
			for j := 0; j < len(node.Content[i].Content); j += 2 {
				keyNode := node.Content[i].Content[j]
				valueNode := node.Content[i].Content[j+1]

				if keyNode.Value == "dimensions" || keyNode.Value == "measures" {
					keyNode.HeadComment = "\n"
				}
				if keyNode.Value == "type" {
					valueNode.LineComment = "\n\n"
				}
				insertEmptyLinesInYaml(valueNode)
			}
		} else if node.Content[i].Kind == yaml.SequenceNode {
			for j := 0; j < len(node.Content[i].Content); j++ {
				if node.Content[i].Content[j].Kind == yaml.MappingNode {
					node.Content[i].Content[j].HeadComment = "\n"
				}
			}
		}
	}
}

func identifierToDisplayName(s string) string {
	return strings.TrimLeft(cases.Title(language.English).String(strings.ReplaceAll(s, "_", " ")), " ")
}

var alphanumericUnderscoreRegexp = regexp.MustCompile("^[_a-zA-Z0-9]+$")

// safeSQLName escapes a SQL column identifier.
// If the name is simple (only contains alphanumeric characters and underscores), it does not escape the string.
// This is because the output is user-facing, so we want to return as simple names as possible.
func safeSQLName(name string) string {
	if name == "" {
		return name
	}
	if alphanumericUnderscoreRegexp.MatchString(name) {
		return name
	}
	return drivers.DialectDuckDB.EscapeIdentifier(name)
}

var dashboardYAML = `
# Explore YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/explore-dashboards

type: explore

display_name: "%s dashboard"
metrics_view: %s

dimensions: '*'
measures: '*'
`
