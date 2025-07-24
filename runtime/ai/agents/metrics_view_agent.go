package agents

import (
	"context"

	"github.com/pontus-devoteam/agent-sdk-go/pkg/agent"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/ai/tools"
)

func init() {
	runtime.RegisterAgentInitializer("metrics_view_agent", func(ctx context.Context, opts *runtime.AgentInitializerOptions) (*agent.Agent, error) {
		return NewMetricsViewAgent(ctx, opts.InstanceID, "gpt-4o", opts.Runtime, opts.ServerTools)
	})
}

// NewMetricsViewAgent creates a new MetricsViewAgent
func NewMetricsViewAgent(ctx context.Context, instanceID, modelName string, r *runtime.Runtime, s runtime.ServerTools) (*agent.Agent, error) {
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
