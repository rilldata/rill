package tools

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/tool"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
)

type createResourceInputProps struct {
	ResourceName string `mapstructure:"resource_name"`
	ResourceType string `mapstructure:"resource_type"`
	Contents     string `mapstructure:"contents"`
	Path         string `mapstructure:"path"`
}

func newCreateResourceInput(in map[string]any) (*createResourceInputProps, error) {
	var input createResourceInputProps
	if err := mapstructure.Decode(in, &input); err != nil {
		return nil, fmt.Errorf("failed to decode input: %w", err)
	}
	if err := input.validate(); err != nil {
		return nil, fmt.Errorf("input validation failed: %w", err)
	}
	return &input, nil
}

func (i *createResourceInputProps) validate() error {
	if i.ResourceName == "" {
		return fmt.Errorf("`resource_name` parameter is required and must be a string")
	}
	if i.ResourceType == "" {
		return fmt.Errorf("`resource_type` parameter is required and must be a string")
	}
	if i.Contents == "" {
		return fmt.Errorf("`contents` parameter is required and must be a string")
	}
	if i.Path == "" {
		return fmt.Errorf("`path` parameter is required and must be a string")
	}
	return nil
}

func CreateAndReconcileResource(instanceID string, rt *runtime.Runtime) *tool.FunctionTool {
	tool := tool.NewFunctionTool(
		"create_and_reconcile_resource",
		"Creates and reconciles a resource in the Rill runtime",
		func(ctx context.Context, params map[string]any) (res any, resErr error) {
			input, err := newCreateResourceInput(params)
			if err != nil {
				return nil, err
			}

			res, err = putResourceAndWaitForReconcile(ctx, rt, instanceID, input.Path, input.Contents, &runtimev1.ResourceName{
				Kind: runtime.ResourceKindFromShorthand(input.ResourceType),
				Name: input.ResourceName,
			})
			if err != nil {
				return map[string]any{
					"error": fmt.Sprintf("Encountered error while creating or reconciling resource '%s' of type '%s': %s", input.ResourceName, input.ResourceType, err.Error()),
				}, nil
			}
			return res, nil
		},
	)

	tool.WithSchema(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"resource_name": map[string]any{
				"type":        "string",
				"description": "The name of the resource to create or reconcile",
			},
			"resource_type": map[string]any{
				"type":        "string",
				"description": "The type of the resource (e.g., 'model', 'dashboard')",
			},
			"contents": map[string]any{
				"type":        "string",
				"description": "The contents of the resource, typically in YAML or SQL format",
			},
			"path": map[string]any{
				"type":        "string",
				"description": "The path where the resource should be stored in the repository",
			},
		},
		"required": []string{"resource_name", "resource_type", "contents", "path"},
	})
	return tool
}

func putResourceAndWaitForReconcile(ctx context.Context, rt *runtime.Runtime, instanceID, path, contents string, resource *runtimev1.ResourceName) (res map[string]any, resErr error) {
	repo, release, err := rt.Repo(ctx, instanceID)
	if err != nil {
		return nil, err
	}
	defer release()

	// ensure directory exists
	switch resource.Kind {
	case runtime.ResourceKindModel:
		err = repo.MkdirAll(ctx, "models")
		if err != nil {
			return nil, err
		}
	case runtime.ResourceKindMetricsView:
		err = repo.MkdirAll(ctx, "metrics_views")
		if err != nil {
			return nil, err
		}
	}

	// Create the resource in the repository
	err = repo.Put(ctx, path, strings.NewReader(contents))
	if err != nil {
		return nil, fmt.Errorf("failed to put resource at path '%s': %w", path, err)
	}
	defer func() {
		if resErr != nil {
			// Clean up if there was an error
			_ = repo.Delete(ctx, path, true)
		}
	}()

	ctrl, err := rt.Controller(ctx, instanceID)
	if err != nil {
		return nil, err
	}

	// TODO : Find a better way to handle this than just waiting for full reconciliation
	// wait for 5 seconds for reconciler to pick up the changes
	// Find a better way to wait for reconciliation for a specific resource
	// May be just subscribe to resource events and wait for the specific resource to be reconciled
	time.Sleep(5 * time.Second)
	err = ctrl.WaitUntilIdle(ctx, true)
	if err != nil {
		return nil, err
	}

	// get the resource
	r, err := ctrl.Get(ctx, resource, false)
	if err != nil {
		return nil, err
	}
	if r.Meta.ReconcileError == "" {
		return newToolResult(fmt.Sprintf("Resource '%s' of type '%s' created and reconciled successfully", resource.Name, resource.Kind), nil), nil
	}
	return newToolResult(fmt.Sprintf("Reconilation of resource '%s' of type '%s' failed: %s", resource.Name, resource.Kind, r.Meta.ReconcileError), nil), nil
}
