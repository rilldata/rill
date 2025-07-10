package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/pontus-devoteam/agent-sdk-go/pkg/agent"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/tool"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"gopkg.in/yaml.v3"
)

// MetricsViewAgent is an AI agent that generates MetricsView YAML definitions from model schemas
type MetricsViewAgent struct {
	*agent.Agent
}

// NewMetricsViewAgent creates a new MetricsViewAgent
func NewMetricsViewAgent(modelName string) *MetricsViewAgent {
	a := agent.NewAgent("MetricsViewAgent")
	mva := &MetricsViewAgent{
		Agent: a,
	}
	mva.WithModel(modelName)
	mva.configure()

	return mva
}

func (m *MetricsViewAgent) configure() {
	m.SetSystemInstructions(`You are an agent specialized in generating MetricsView YAML definitions based on model schemas.

Your primary responsibility is to generate valid MetricsView YAML files that define relevant business metrics (KPIs) based on table schemas.

When generating MetricsView YAML:
1. Always use proper YAML syntax and structure
2. Generate relevant business metrics using only COUNT, SUM, MIN, MAX, and AVG functions
3. Do not use complex aggregations or WHERE/FILTER clauses in measure expressions
4. Create meaningful measure names in snake_case format
5. Provide descriptive display names for measures
6. Include appropriate format presets (typically "humanize")
7. Focus on metrics that would be valuable for business analysis

The output should be a complete MetricsView YAML definition that can be used directly in a Rill project.`)

	m.WithTools(
		m.createGenerateMetricsViewTool(),
		m.createValidateMetricsViewTool(),
	)
}

func (m *MetricsViewAgent) createGenerateMetricsViewTool() tool.Tool {
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"model_name": map[string]interface{}{
				"type":        "string",
				"description": "Name of the model to generate metrics for",
			},
			"table_schema": map[string]interface{}{
				"type":        "string",
				"description": "JSON representation of the table schema with field names and types",
			},
			"sql_dialect": map[string]interface{}{
				"type":        "string",
				"description": "SQL dialect to use (e.g., 'duckdb', 'postgres', 'bigquery')",
			},
		},
		"required": []string{"model_name", "table_schema", "sql_dialect"},
	}

	t := tool.NewFunctionTool(
		"generate_metrics_view_yaml",
		"Generates a MetricsView YAML definition with relevant business metrics based on the provided model schema",
		m.generateMetricsViewYAML,
	)

	t.WithSchema(schema)
	return t
}

func (m *MetricsViewAgent) createValidateMetricsViewTool() tool.Tool {
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"yaml_content": map[string]interface{}{
				"type":        "string",
				"description": "The MetricsView YAML content to validate",
			},
		},
		"required": []string{"yaml_content"},
	}

	t := tool.NewFunctionTool(
		"validate_metrics_view_yaml",
		"Validates MetricsView YAML syntax and structure",
		m.validateMetricsViewYAML,
	)

	t.WithSchema(schema)
	return t
}

func (m *MetricsViewAgent) generateMetricsViewYAML(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	modelName, ok := params["model_name"].(string)
	if !ok {
		return nil, fmt.Errorf("model_name parameter is required")
	}

	tableSchema, ok := params["table_schema"].(string)
	if !ok {
		return nil, fmt.Errorf("table_schema parameter is required")
	}

	sqlDialect, ok := params["sql_dialect"].(string)
	if !ok {
		return nil, fmt.Errorf("sql_dialect parameter is required")
	}

	// This is a placeholder implementation that would be replaced with actual LLM call
	// Using the same system prompt structure as the existing MetricsView generation
	// Create display name using proper title casing
	displayName := strings.ReplaceAll(modelName, "_", " ")
	// Simple title case for display name
	words := strings.Fields(displayName)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
		}
	}
	displayName = strings.Join(words, " ")

	yaml := fmt.Sprintf(`# MetricsView YAML generated for model: %s
# Schema: %s
# SQL Dialect: %s
# This is a placeholder implementation. In a real implementation, this would call an LLM to generate the YAML.

# Based on the existing system prompt from generate_metrics_view.go:
display_name: "%s"
measures:
  - name: total_records
    display_name: "Total Records"
    expression: "COUNT(*)"
    description: "Total number of records"
    format_preset: "humanize"
  - name: example_metric
    display_name: "Example Metric"
    expression: "SUM(example_column)"
    description: "Example sum metric"
    format_preset: "humanize"`, modelName, tableSchema, sqlDialect, displayName)

	return yaml, nil
}

func (m *MetricsViewAgent) validateMetricsViewYAML(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	yamlContent, ok := params["yaml_content"].(string)
	if !ok {
		return nil, fmt.Errorf("yaml_content parameter is required")
	}

	// Parse the YAML to validate syntax
	var doc map[string]interface{}
	if err := yaml.Unmarshal([]byte(yamlContent), &doc); err != nil {
		return map[string]interface{}{
			"valid":   false,
			"message": fmt.Sprintf("Invalid YAML syntax: %v", err),
			"details": map[string]interface{}{
				"yaml": yamlContent,
			},
		}, nil
	}

	// Basic validation - check for required fields
	requiredFields := []string{"display_name", "measures"}
	for _, field := range requiredFields {
		if _, exists := doc[field]; !exists {
			return map[string]interface{}{
				"valid":   false,
				"message": fmt.Sprintf("Missing required field: %s", field),
				"details": map[string]interface{}{
					"yaml": yamlContent,
				},
			}, nil
		}
	}

	// Check measures structure
	if measures, ok := doc["measures"].([]interface{}); ok {
		for i, measure := range measures {
			if measureMap, ok := measure.(map[string]interface{}); ok {
				measureFields := []string{"name", "display_name", "expression"}
				for _, field := range measureFields {
					if _, exists := measureMap[field]; !exists {
						return map[string]interface{}{
							"valid":   false,
							"message": fmt.Sprintf("Measure %d missing required field: %s", i, field),
							"details": map[string]interface{}{
								"yaml": yamlContent,
							},
						}, nil
					}
				}
			}
		}
	}

	return map[string]interface{}{
		"valid":   true,
		"message": "MetricsView YAML is valid",
		"details": map[string]interface{}{
			"yaml": yamlContent,
		},
	}, nil
}

// Helper function to create a MetricsView YAML template (based on existing code)
func (m *MetricsViewAgent) createMetricsViewTemplate() string {
	// Using the same template structure as metricsViewYAMLSystemPrompt() in generate_metrics_view.go
	template := map[string]interface{}{
		"display_name": "<human-friendly display name based on the table name and column names>",
		"measures": []map[string]interface{}{
			{
				"name":         "<unique name for the metric in snake case, such as average_sales>",
				"display_name": "<short descriptive display name for the metric>",
				"expression":   "<SQL expression to calculate the KPI in the requested SQL dialect>",
				"description":  "<short description of the metric>",
			},
		},
	}

	out, err := yaml.Marshal(template)
	if err != nil {
		return "Error creating template"
	}

	return string(out)
}

// getSystemPrompt returns the system prompt based on the existing MetricsView generation logic
func (m *MetricsViewAgent) getSystemPrompt() string {
	template := m.createMetricsViewTemplate()
	
	return fmt.Sprintf(`You are an agent whose only task is to suggest relevant business metrics (KPIs) based on a table schema.
The metrics should be valid SQL aggregation expressions that use only the COUNT, SUM, MIN, MAX, and AVG functions.
Do not use any complex aggregations and do not use WHERE or FILTER in the metrics expressions.
Your output should only consist of valid YAML in the format below:

%s`, template)
}

// getUserPrompt creates the user prompt based on the existing logic
func (m *MetricsViewAgent) getUserPrompt(dialect, modelName string, schema *runtimev1.StructType) string {
	prompt := fmt.Sprintf(`Give me up to 10 suggested metrics using the %q SQL dialect based on the model named %q, which has the following schema:
`, dialect, modelName)
	
	if schema != nil {
		for _, field := range schema.Fields {
			prompt += fmt.Sprintf("- column=%s, type=%s\n", field.Name, field.Type.Code.String())
		}
	}
	
	return prompt
}