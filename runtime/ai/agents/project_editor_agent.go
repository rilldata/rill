package agents

import (
	"context"
	"fmt"

	"github.com/pontus-devoteam/agent-sdk-go/pkg/agent"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/runner"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/ai/tools"
)

// NewProjectEditorAgent creates a new ProjectEditorAgent with handoff capabilities
func NewProjectEditorAgent(ctx context.Context, instanceID, modelName string, r *runtime.Runtime, runner *runner.Runner) (*agent.Agent, error) {
	a := agent.NewAgent("ProjectEditorAgent")
	a.WithModel(modelName)
	a.SetSystemInstructions(`You are a ProjectEditor Agent manages Rill projects
- A rill project is a collection of resources like models, dashboards, and metrics views.
- A model is a resource that defines how data is ingested in the system. A model can be built on top of other model as well.
- A metrics view is a resource that defines how data is aggregated and visualized in the system. It is built on top of models and can be used to create dashboards.
- You can be asked to either create new resources or edit existing ones based on user input
- Must not ask for user input. Try to create resources based on the context provided in the user input as much as feasible

Your primary responsibilities:
- Ingest sample data as models using "create_model" tool
- Create metrics views using "create_metrics_view" tool. You must always pass the model name to the tool.
- Edits existing models using "edit_model" tool
- You can get existing resources using "list_resources" tool
- Try to infer the resource from the context provided in the user input
DECISION LOGIC:
- If user mentions "synthetic data", "generate data", or "sample data" -> Use "create_model" tool
- If user wants to edit existing models -> Use "model_editor" tool with full context
- If user wants to create metrics views -> Use "create_metrics_view" tool with full context
_ If a user wants to fix a model use "model_editor" tool
- A user can also ask to create multiple resources at once
- The ask can also be implicit (e.g., "create a rill project for analysing sales data implies create a model that ingests sample sales data and a metrics view for visualizing it")
`)

	// create model agent
	dataAgent, err := NewSyntheticDataAgent(ctx, instanceID, modelName, r)
	if err != nil {
		return nil, fmt.Errorf("failed to create SyntheticDataAgent: %w", err)
	}
	tool, err := tools.RunAgent(dataAgent, runner, "create_model", "Create a `model` resource that ingest data based on user input")
	if err != nil {
		return nil, fmt.Errorf("failed to run SyntheticDataAgent: %w", err)
	}
	a.WithTools(tool)

	editModelAgent, err := NewModelEditorAgent(ctx, instanceID, modelName, r)
	if err != nil {
		return nil, fmt.Errorf("failed to create ModelEditorAgent: %w", err)
	}
	tool, err = tools.RunAgent(editModelAgent, runner, "edit_model", "Edit an existing model based on user input")
	if err != nil {
		return nil, fmt.Errorf("failed to run ModelEditorAgent: %w", err)
	}
	a.WithTools(tool)

	// metrics view agent
	metricsViewAgent, err := NewMetricsViewAgent(ctx, instanceID, modelName, r)
	if err != nil {
		return nil, fmt.Errorf("failed to create MetricsViewAgent: %w", err)
	}
	tool, err = tools.RunAgent(metricsViewAgent, runner, "create_metrics_view", "Create a metrics view based on the generated model")
	if err != nil {
		return nil, fmt.Errorf("failed to run MetricsViewAgent: %w", err)
	}
	a.WithTools(tool)

	// existing resources tool
	a.WithTools(tools.ListResources(instanceID, r))
	return a, nil
}
