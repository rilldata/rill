package resolvers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/pkg/jsonschemautil"
)

func init() {
	expressionDefs := jsonschemautil.MustExtractReferencedDefs(metricsview.QueryJSONSchema, "Expression")
	expressionDefsJSON, err := json.Marshal(expressionDefs)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal expression defs: %v", err))
	}

	runtime.RegisterResolver("builtin_metrics_sql", newBuiltinMetricsSQL, analyzeBuiltinMetricsSQL)
	runtime.RegisterBuiltinAPI(&runtime.BuiltinAPIOptions{
		Name:               "metrics-sql",
		Resolver:           "builtin_metrics_sql",
		ResolverProperties: nil,
		OpenAPISummary:     "Query metrics with SQL",
		OpenAPIRequestSchema: fmt.Sprintf(`{
			"type":"object",
			"properties": {
				"sql": {"type":"string"},
				"additional_where": {"$ref": "#/$defs/Expression"}
			},
			"required":["sql"],
			"$defs": %s
		}`, expressionDefsJSON),
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

	props, args, err := builtinMetricsSQLProperties(opts.Args)
	if err != nil {
		return nil, err
	}

	// Rewrite to the metrics SQL resolver
	return newMetricsSQL(ctx, &runtime.ResolverOptions{
		Runtime:    opts.Runtime,
		InstanceID: opts.InstanceID,
		Properties: props,
		Args:       args,
		Claims:     opts.Claims,
		ForExport:  opts.ForExport,
	})
}

func builtinMetricsSQLProperties(arguments map[string]any) (map[string]any, map[string]any, error) {
	args := &builtinMetricsSQLArgs{}
	if err := mapstructure.Decode(arguments, args); err != nil {
		return nil, nil, err
	}
	return map[string]any{"sql": args.SQL}, map[string]any{"priority": args.Priority}, nil
}

func analyzeBuiltinMetricsSQL(ctx context.Context, rt *runtime.Runtime, opts *runtime.ResolverAnalysisOptions) (*runtime.ResolverAnalysis, error) {
	props, args, err := builtinMetricsSQLProperties(opts.Args)
	if err != nil {
		return nil, err
	}
	child := *opts
	child.Properties = props
	child.Args = args
	return rt.AnalyzeResolver(ctx, "metrics_sql", &child)
}
