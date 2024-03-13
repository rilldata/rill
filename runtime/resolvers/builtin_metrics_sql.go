package resolvers

import (
	"context"
	"errors"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime"
)

func init() {
	runtime.RegisterResolverInitializer("builtin_metrics_sql", newBuiltinMetricsSQL)
	runtime.RegisterBuiltinAPI("metrics-sql", "builtin_metrics_sql", nil)
}

type builtinMetricsSQLArgs struct {
	SQL      string `mapstructure:"sql"`
	Priority int    `mapstructure:"priority"`
}

// newBuiltinMetricsSQL is the resolver for the built-in /metrics-sql API.
// It executes a metrics SQL query provided dynamically through the args.
// It errors if the user identified by the attributes is not an admin.
func newBuiltinMetricsSQL(ctx context.Context, opts *runtime.ResolverOptions) (runtime.Resolver, error) {
	// Only admins and non-users (i.e. local users and service accounts) can run arbitrary SQL queries.
	if len(opts.UserAttributes) > 0 {
		admin, ok := opts.UserAttributes["admin"].(bool)
		if !ok || !admin {
			return nil, errors.New("must be an admin to run arbitrary SQL queries")
		}
	}

	// Decode the args
	args := &builtinMetricsSQLArgs{}
	if err := mapstructure.Decode(opts.Args, args); err != nil {
		return nil, err
	}

	// Build the options for the metrics SQL resolver
	metricsSQLResolverOpts := &runtime.ResolverOptions{
		Runtime:    opts.Runtime,
		InstanceID: opts.InstanceID,
		Properties: map[string]any{
			"sql": args.SQL,
		},
		Args: map[string]any{
			"priority": args.Priority,
		},
		UserAttributes: opts.UserAttributes,
		ForExport:      opts.ForExport,
	}

	return newMetricsSQL(ctx, metricsSQLResolverOpts)
}
