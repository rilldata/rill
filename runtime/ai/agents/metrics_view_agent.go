package agents

import (
	"context"

	"github.com/pontus-devoteam/agent-sdk-go/pkg/agent"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/ai/tools"
)

// NewMetricsViewAgent creates a new MetricsViewAgent
func NewMetricsViewAgent(ctx context.Context, instanceID, modelName string, r *runtime.Runtime) (*agent.Agent, error) {
	a := agent.NewAgent("MetricsViewAgent")
	a.WithModel(modelName)

	a.SetSystemInstructions(`
	Extract the model name from the input and pass it to the generate_metrics_view_yaml tool to generate MetricsView YAML definitions.
	You mustn't ask for user input. Fail if unable to extract the model name.
	Pass the generated YAML to the create_and_reconcile_resource tool to create the MetricsView and Dashboard resource in the Rill project.
	The metrics view name should be in snake_case and inferred from the model name.
	`)

	a.WithTools(
		tools.GenerateMetricsViewYAML(instanceID, r),
	)
	return a, nil
}
