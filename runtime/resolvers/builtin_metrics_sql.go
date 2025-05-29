package resolvers

import (
	"context"
	"errors"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime"
)

func init() {
	runtime.RegisterResolverInitializer("builtin_metrics_sql", newBuiltinMetricsSQL)
	runtime.RegisterBuiltinAPI(&runtime.BuiltinAPIOptions{
		Name:               "metrics-sql",
		Resolver:           "builtin_metrics_sql",
		ResolverProperties: nil,
		OpenAPISummary:     "Query metrics with SQL",
		OpenAPIRequestSchema: `{
			"type":"object",
			"properties": {
				"sql": {"type":"string"}
			},
			"required":["sql"]
		}`,
	})
}

type builtinMetricsSQLArgs struct {
	SQL      string `mapstructure:"sql"`
	Priority int    `mapstructure:"priority"`
}

// newBuiltinMetricsSQL is the resolver for the built-in /metrics-sql API.
// It executes a metrics SQL query provided dynamically through the args.
// It errors if the user identified by the attributes is not an admin.
func newBuiltinMetricsSQL(ctx context.Context, opts *runtime.ResolverOptions) (runtime.Resolver, error) {
	// Only admins can run arbitrary SQL queries.
	if !opts.Claims.SkipChecks && !opts.Claims.Admin() {
		return nil, errors.New("must be an admin to run arbitrary SQL queries")
	}

	// Decode the args
	args := &builtinMetricsSQLArgs{}
	if err := mapstructure.Decode(opts.Args, args); err != nil {
		return nil, err
	}

	// Rewrite to the metrics SQL resolver
	return newMetricsSQL(ctx, &runtime.ResolverOptions{
		Runtime:    opts.Runtime,
		InstanceID: opts.InstanceID,
		Properties: map[string]any{
			"sql": args.SQL,
		},
		Args: map[string]any{
			"priority": args.Priority,
		},
		Claims:    opts.Claims,
		ForExport: opts.ForExport,
	})
}
