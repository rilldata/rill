package agents

import (
	"context"
	"fmt"

	"github.com/pontus-devoteam/agent-sdk-go/pkg/agent"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/runner"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/ai/tools"
)

func init() {
	runtime.RegisterAgentInitializer("project_editor_agent", func(ctx context.Context, opts *runtime.AgentInitializerOptions) (*agent.Agent, error) {
		return NewProjectEditorAgent(ctx, opts.InstanceID, "gpt-4o", opts.Runtime, opts.Runner, opts.ServerTools)
	})
}

// NewProjectEditorAgent creates a new ProjectEditorAgent with handoff capabilities
func NewProjectEditorAgent(ctx context.Context, instanceID, modelName string, r *runtime.Runtime, runner *runner.Runner, s runtime.ServerTools) (*agent.Agent, error) {
	a := agent.NewAgent("ProjectEditorAgent")
	a.WithModel(modelName)
	a.SetSystemInstructions(`You are a ProjectEditor Agent that manages Rill projects. 
- A rill project is a collection of resources like models, dashboards, and metrics views.
- You can be asked to either create new resources or edit existing ones based on user input. 
- Try to infer if the user wants to create resource or edit existing ones. You can get details of existing resources using "list_resources" tool.
- Only ask for user input in case of ambiguity. Try to infer as much as feasible from the context provided in the user input.
- A user can also ask to create/edit multiple resources at once.
- The ask can also be implicit (e.g., "create a rill project for analysing sales data implies create a model that ingests sample sales data and a metrics view for visualizing it")
- Understand from the tool output if there was a problem and return tool's output


You have following tools that can help you create and edit resources:
1. "create_synthetic_data"
	- Creates a model that ingest sample data based on user input
2. "generate_metrics_view_yaml"
	- Creates a metrics view based on specified model
3. "model_editor_agent_tool"
	- Edits existing models based on user input.
	- Create a model that ingests data from a external system. 
	- Do not add additional context to the user input that might change the model semantics.
4. "list_resources"
	- Lists existing resources in the project
`)

	// create model agent
	dataAgent, err := NewSyntheticDataAgent(ctx, instanceID, modelName, r)
	if err != nil {
		return nil, fmt.Errorf("failed to create SyntheticDataAgent: %w", err)
	}
	tool, err := tools.RunAgent(dataAgent, runner, "create_synthetic_data", "An agent acting as a tool that can create a `model` resource that ingest sample data based on user input")
	if err != nil {
		return nil, fmt.Errorf("failed to run SyntheticDataAgent: %w", err)
	}
	a.WithTools(tool)

	modelEditorTool, err := NewModelEditorAgentTool(ctx, instanceID, "gpt-4o", r, runner)
	if err != nil {
		return nil, fmt.Errorf("failed to run ModelEditorAgent: %w", err)
	}
	a.WithTools(modelEditorTool)

	// metrics view agent
	metricsViewAgent, err := NewMetricsViewAgent(ctx, instanceID, modelName, r, s)
	if err != nil {
		return nil, fmt.Errorf("failed to create MetricsViewAgent: %w", err)
	}
	tool, err = tools.RunAgent(metricsViewAgent, runner, "create_metrics_view", "An agent acting as a tool that can create a metrics view based on specified model and user input")
	if err != nil {
		return nil, fmt.Errorf("failed to run MetricsViewAgent: %w", err)
	}
	a.WithTools(tool)

	// existing resources tool
	a.WithTools(tools.ListResources(instanceID, r))
	a.WithTools(tools.GenerateMetricsViewYAML(instanceID, r, s))
	return a, nil
}
