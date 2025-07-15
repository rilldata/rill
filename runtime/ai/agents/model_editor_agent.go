package agents

import (
	"context"

	"github.com/pontus-devoteam/agent-sdk-go/pkg/agent"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/ai/tools"
)

func NewModelEditorAgent(ctx context.Context, instanceID, modelName string, r *runtime.Runtime) (*agent.Agent, error) {
	a := agent.NewAgent("ModelEditorAgent")
	a.WithModel(modelName)
	a.SetSystemInstructions(`You are a ModelEditor Agent that edits Rill models.
- A model is a resource that defines how data is ingested in the system. A model can be built on top of other models as well.
- You can be asked to fix a model based on user input or context. When fixing a model only fix SQL errors without changing model semantics.
- You can also be asked to add new functionality to an existing model.
- You can get details of existing models using "list_models" tool.
- Must not ask for user input. Try to create or edit models based on the context provided in the user input as much as feasible.
`)
	a.WithTools(tools.ListModels(instanceID, r))

	return a, nil
}
