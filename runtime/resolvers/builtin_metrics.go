package resolvers

import (
	"context"

	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/metricsview"
)

func init() {
	runtime.RegisterResolver("builtin_metrics", newBuiltinMetrics, analyzeBuiltinMetrics)
	runtime.RegisterBuiltinAPI(&runtime.BuiltinAPIOptions{
		Name:                 "metrics",
		Resolver:             "builtin_metrics",
		OpenAPISummary:       "The main API for querying metrics",
		OpenAPIRequestSchema: metricsview.QueryJSONSchema,
	})
}

// newBuiltinMetrics is a resolver for the built-in /metrics API.
// It executes a metrics query provided dynamically through the args.
// It errors if the user identified by the attributes does not have access to read metrics.
func newBuiltinMetrics(ctx context.Context, opts *runtime.ResolverOptions) (runtime.Resolver, error) {
	props, args := builtinMetricsProperties(opts.Args)

	// Rewrite to the metrics resolver
	return newMetrics(ctx, &runtime.ResolverOptions{
		Runtime:    opts.Runtime,
		InstanceID: opts.InstanceID,
		Properties: props,
		Args:       args,
		Claims:     opts.Claims,
		ForExport:  opts.ForExport,
	})
}

func builtinMetricsProperties(arguments map[string]any) (map[string]any, map[string]any) {
	// We translate the API args to props for the metrics resolver
	props := arguments

	// We need to separate out the values that the resolver considers as args
	args := map[string]any{}
	if priority, ok := arguments["priority"]; ok {
		args["priority"] = priority
	}
	if executionTime, ok := arguments["execution_time"]; ok {
		args["execution_time"] = executionTime
	}

	return props, args
}

func analyzeBuiltinMetrics(ctx context.Context, rt *runtime.Runtime, opts *runtime.ResolverAnalysisOptions) (*runtime.ResolverAnalysis, error) {
	child := *opts
	child.Properties, child.Args = builtinMetricsProperties(opts.Args)
	return rt.AnalyzeResolver(ctx, "metrics", &child)
}
