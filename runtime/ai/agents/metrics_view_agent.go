package agents

import (
	"context"

	"github.com/pontus-devoteam/agent-sdk-go/pkg/agent"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/ai/tools"
)

// NewMetricsViewAgent creates a new MetricsViewAgent
func NewMetricsViewAgent(ctx context.Context, instanceID, modelName string, r *runtime.Runtime, s tools.ServerTools) (*agent.Agent, error) {
	a := agent.NewAgent("MetricsViewAgent")
	a.WithModel(modelName)

	a.SetSystemInstructions(`
	Extract the model name from the input. You mustn't ask for user input. Fail if unable to extract the model name.
	Infer a metrics view name and a dashboard name from the model name. Pass all these to the "generate_metrics_view_yaml" tool to generate MetricsView YAML definitions.
	`)

	a.WithTools(
		tools.GenerateMetricsViewYAML(instanceID, r, s),
	)
	return a, nil
}
