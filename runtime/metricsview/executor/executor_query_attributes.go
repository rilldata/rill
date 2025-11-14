package executor

import (
	"context"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/parser"
)

// resolveQueryAttributes resolves the query attributes defined in metrics views
func resolveQueryAttributes(ctx context.Context, rt *runtime.Runtime, instanceID string, mv *runtimev1.MetricsViewSpec, userAttrs map[string]any) (map[string]string, error) {
	if len(mv.QueryAttributes) == 0 {
		return nil, nil
	}

	// Get instance for template data
	inst, err := rt.Instance(ctx, instanceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get instance for query attributes: %w", err)
	}

	td := parser.TemplateData{
		Environment: inst.Environment,
		Variables:   inst.ResolveVariables(false),
		User:        userAttrs,
	}

	// Resolve templates
	resolved := make(map[string]string)
	for key, template := range mv.QueryAttributes {
		val, err := parser.ResolveTemplate(template, td, false)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve query attribute %q: %w", key, err)
		}
		resolved[key] = val
	}

	return resolved, nil
}
