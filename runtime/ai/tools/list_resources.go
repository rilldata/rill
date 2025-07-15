package tools

import (
	"context"
	"fmt"

	"github.com/pontus-devoteam/agent-sdk-go/pkg/tool"
	"github.com/rilldata/rill/runtime"
)

func ListResources(instanceID string, r *runtime.Runtime) *tool.FunctionTool {
	return tool.NewFunctionTool(
		"list_resources",
		"List all resources in the Rill project",
		func(ctx context.Context, _ map[string]any) (any, error) {
			ctrl, err := r.Controller(ctx, instanceID)
			if err != nil {
				return nil, fmt.Errorf("failed to get controller: %w", err)
			}
			resources, err := ctrl.List(ctx, "", "", false)
			if err != nil {
				return nil, fmt.Errorf("failed to list resources: %w", err)
			}

			resourcesMap := make([]map[string]any, len(resources))
			for i, res := range resources {
				if res.Meta.Hidden {
					continue // Skip hidden resources
				}
				resourcesMap[i] = map[string]any{
					"name":            res.Meta.Name.Name,
					"type":            res.Meta.Name.Kind,
					"reconcile_error": res.Meta.ReconcileError,
				}
			}
			return map[string]any{"resources": resourcesMap}, nil
		},
	)
}
