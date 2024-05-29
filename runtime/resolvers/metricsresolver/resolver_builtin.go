package metricsresolver

import (
	"context"

	"github.com/rilldata/rill/runtime"
)

// NewBuiltin is a resolver for the built-in /metrics API.
// It executes a metrics query provided dynamically through the args.
// It errors if the user identified by the attributes does not have access to read metrics.
func NewBuiltin(ctx context.Context, opts *runtime.ResolverOptions) (runtime.Resolver, error) {
	// We translate the API args to props for the metrics resolver
	props := opts.Args

	// We need to separate out the values that the resolver considers as args
	args := map[string]any{}
	if priority, ok := opts.Args["priority"]; ok {
		args["priority"] = priority
	}
	if executionTime, ok := opts.Args["execution_time"]; ok {
		args["execution_time"] = executionTime
	}

	// Rewrite to the metrics resolver
	return New(ctx, &runtime.ResolverOptions{
		Runtime:        opts.Runtime,
		InstanceID:     opts.InstanceID,
		Properties:     props,
		Args:           args,
		UserAttributes: opts.UserAttributes,
		ForExport:      opts.ForExport,
	})
}
