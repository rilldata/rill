package agents

import (
	"context"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/agent"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/runner"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/tool"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
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
	return a, nil
}

func NewModelEditorAgentTool(ctx context.Context, instanceID, modelName string, r *runtime.Runtime, agentRunner *runner.Runner) (tool.Tool, error) {
	agent, err := NewModelEditorAgent(ctx, instanceID, modelName, r)
	if err != nil {
		return nil, fmt.Errorf("failed to create ModelEditorAgent: %w", err)
	}
	tool := tool.NewFunctionTool(
		"model_editor_agent_tool",
		"An agent that can create or edit Rill models",
		func(ctx context.Context, params map[string]any) (any, error) {
			parsed, err := newModelEditorInput(params)
			if err != nil {
				return nil, err
			}

			var input any
			// build a prompt
			if parsed.ModelName == "" {
				// If ask is to create a new model. No need to add more context
				input = parsed.Input
			} else {
				// Build the full context
				input = fmt.Sprintf("Edit model %q with the following user request: %s", parsed.ModelName, parsed.Input)
				// also add existing model details
				ctrl, err := r.Controller(ctx, instanceID)
				if err != nil {
					return nil, fmt.Errorf("failed to get controller: %w", err)
				}
				model, err := ctrl.Get(ctx, &runtimev1.ResourceName{
					Kind: runtime.ResourceKindModel,
					Name: parsed.ModelName,
				}, false)
				if err != nil {
					return nil, fmt.Errorf("failed to get model %q: %w", parsed.ModelName, err)
				}

				repo, release, err := r.Repo(ctx, instanceID)
				if err != nil {
					return nil, err
				}
				defer release()

				// get the model contents
				// TODO: handle multiple files
				contents, err := repo.Get(ctx, model.Meta.GetFilePaths()[0])
				if err != nil {
					return nil, fmt.Errorf("failed to get model contents for %q: %w", parsed.ModelName, err)
				}

				input = fmt.Sprintf("%s\n\nExisting model contents:\n%s", input, contents)
			}
			result, err := agentRunner.Run(ctx, agent, &runner.RunOptions{
				Input:    input,
				MaxTurns: 10,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to run ModelEditorAgent: %w", err)
			}
			return result, nil
		},
	).WithSchema(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"input": map[string]any{
				"type":        "string",
				"description": "Context to the tool",
			},
			"model_name": map[string]any{
				"type":        "string",
				"description": "Name of the model to edit if editing an existing model",
			},
		},
		"required": []string{"input"},
	})
	return tool, nil
}

type modelEditorInputProps struct {
	Input     string `mapstructure:"input"`
	ModelName string `mapstructure:"model_name"`
}

func (i *modelEditorInputProps) validate() error {
	if i.Input == "" {
		return fmt.Errorf("expected 'input' parameter to be a non empty string")
	}
	return nil
}

func newModelEditorInput(in map[string]any) (*modelEditorInputProps, error) {
	var input modelEditorInputProps
	if err := mapstructure.Decode(in, &input); err != nil {
		return nil, fmt.Errorf("failed to decode input: %w", err)
	}
	if err := input.validate(); err != nil {
		return nil, fmt.Errorf("input validation failed: %w", err)
	}
	return &input, nil
}
