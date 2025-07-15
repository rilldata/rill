package tools

import (
	"context"
	"fmt"

	"github.com/pontus-devoteam/agent-sdk-go/pkg/tool"
	"github.com/rilldata/rill/runtime"
)

// ListModels is a tool that lists all models in a Rill project
// This is separate from ListResources to focus specifically on models and return model specific fields
func ListModels(instanceID string, r *runtime.Runtime) *tool.FunctionTool {
	t := tool.NewFunctionTool(
		"list_models",
		"List all models in the Rill project",
		func(ctx context.Context, _ map[string]any) (any, error) {
			ctrl, err := r.Controller(ctx, instanceID)
			if err != nil {
				return nil, fmt.Errorf("failed to get controller: %w", err)
			}
			models, err := ctrl.List(ctx, runtime.ResourceKindModel, "", false)
			if err != nil {
				return nil, fmt.Errorf("failed to list models: %w", err)
			}

			modelsMap := make([]map[string]any, len(models))
			for i, model := range models {
				m := model.GetModel()
				// TODO : Add support for YAML models
				props := m.Spec.InputProperties.AsMap()
				modelsMap[i] = map[string]any{
					"name":            model.Meta.Name,
					"reconcile_error": model.Meta.ReconcileError,
					"sql":             props["sql"],
				}
			}
			return map[string]any{"models": modelsMap}, nil
		},
	)
	return t
}
