package executor

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/runtime/parser"
)

// resolveQueryAttributes resolves the query attributes defined in metrics views
func (e *Executor) resolveQueryAttributes(ctx context.Context) (map[string]string, error) {
	if len(e.metricsView.QueryAttributes) == 0 {
		return nil, nil
	}

	// Get instance for template data
	inst, err := e.rt.Instance(ctx, e.instanceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get instance for query attributes: %w", err)
	}

	td := parser.TemplateData{
		Environment: inst.Environment,
		Variables:   inst.ResolveVariables(false),
	}

	// Resolve templates
	resolved := make(map[string]string)
	for key, template := range e.metricsView.QueryAttributes {
		val, err := parser.ResolveTemplate(template, td, false)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve query attribute %q: %w", key, err)
		}
		resolved[key] = val
	}

	return resolved, nil
}
