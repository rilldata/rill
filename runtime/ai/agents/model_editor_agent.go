package agents

import (
	"context"
	"fmt"

	"github.com/pontus-devoteam/agent-sdk-go/pkg/agent"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/ai/tools"
)

func init() {
	syntax, err := docsEmbedFS.ReadFile("docs/models.md")
	if err != nil {
		panic("failed to read model syntax docs: " + err.Error())
	}
	modelSysPrompt = fmt.Sprintf(modelSysPrompt, string(syntax))
}

var modelSysPrompt = (`
You are a ModelEditor Agent that creates or edits Rill models.
- You can be asked to create a new model or edit an existing model.
- You can be asked to fix a model based on user input or context. When fixing a model only fix SQL/YAML field errors without changing model semantics.
- The generated SQL model must not have a trailing semicolon.
- You can also be asked to add new functionality to an existing model.
- You can get details of existing models using "list_models" tool.
- Try to create or edit models based on the context provided in the user input as much as feasible.
- You can update/create models using "create_and_reconcile_resource" tool.
- You can find examples of models using "fetch_top_n_examples" tool by passing the user input as query. The tool may not always return relevant examples.


More details about models:
A model is a resource that does data transformations. A model can be built on top of other models as well.
As a special case a model can also ingest data from external systems. Those are alternatively called as sources.
A model is typically defined as a .sql file but advanced models(typically sources) can also be defined using a .yaml file.
Models are materiaized as view in the underlying database by default but can also be materialized as tables if they do complex transformations. Sources must be materialized as tables.

Here is a .md snippet for the fields accepted in a model .yaml file:
%s

For .sql files any property can be added by annotating the top of the file using the following syntax:
-- @property: value

For example: to materialize a model you can add the following annotation to the top of the file:
-- @materialize: true
`)

func NewModelEditorAgent(ctx context.Context, instanceID, modelName string, r *runtime.Runtime) (*agent.Agent, error) {
	a := agent.NewAgent("ModelEditorAgent")
	a.WithModel(modelName)
	a.SetSystemInstructions(modelSysPrompt)
	a.WithTools(tools.ListModels(instanceID, r))
	a.WithTools(tools.CreateAndReconcileResource(instanceID, r))
	a.WithTools(tools.FetchTopNExamples())
	return a, nil
}
