package tools

import (
	"context"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/tool"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
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

// GenerateMetricsViewYAML creates a tool that generates a metrics view YAML based on the provided model name.
// To keep things simple it also generates a dashboard YAML for the metrics view.
func GenerateMetricsViewYAML(instanceID string, rt *runtime.Runtime, s ServerTools) *tool.FunctionTool {
	tool := tool.NewFunctionTool(
		"generate_metrics_view_yaml",
		"Generates a YAML configuration for a metrics ,view based on the provided model name",
		func(ctx context.Context, params map[string]any) (any, error) {
			input, err := newGenerateMetricsViewInput(params)
			if err != nil {
				return nil, err
			}

			req := &runtimev1.GenerateMetricsViewFileRequest{
				InstanceId: instanceID,
				Model:      input.ModelName,
				Path:       fmt.Sprintf("metrics_views/%s.yaml", input.MetricsViewName),
				UseAi:      true,
			}
			_, err = s.GenerateMetricsViewFile(ctx, req)
			if err != nil {
				// try without AI
				req.UseAi = false
				_, err = s.GenerateMetricsViewFile(ctx, req)
				if err != nil {
					return newToolResult("", fmt.Errorf("failed to generate metrics view YAML: %w", err)), nil
				}
			}

			err = waitForReconcile(ctx, rt, instanceID, &runtimev1.ResourceName{
				Kind: runtime.ResourceKindMetricsView,
				Name: input.MetricsViewName,
			})
			if err != nil {
				return nil, err
			}

			// also create a dashboard for the metrics view
			dashboardYAML := fmt.Sprintf(dashboardYAML, input.MetricsViewName, input.MetricsViewName)
			err = putResourceAndWaitForReconcile(ctx, rt, instanceID, fmt.Sprintf("dashboards/%s.yaml", input.DashboardName), dashboardYAML, &runtimev1.ResourceName{
				Kind: runtime.ResourceKindExplore,
				Name: input.DashboardName,
			})
			if err != nil {
				return nil, err
			}
			return fmt.Sprintf("Metrics view %q and dashboard %q created successfully", input.MetricsViewName, input.DashboardName), nil
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

var dashboardYAML = `
# Explore YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/explore-dashboards

type: explore

display_name: "%s dashboard"
metrics_view: %s

dimensions: '*'
measures: '*'
`
