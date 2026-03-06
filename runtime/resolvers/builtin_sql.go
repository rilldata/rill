package resolvers

import (
	"context"
	"errors"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func init() {
	runtime.RegisterResolverInitializer("builtin_sql", newBuiltinSQL)
	runtime.RegisterBuiltinAPI(&runtime.BuiltinAPIOptions{
		Name:               "sql",
		Resolver:           "builtin_sql",
		ResolverProperties: nil,
		OpenAPISummary:     "Execute a raw SQL query. Access is restricted to admins.",
		OpenAPIRequestSchema: `{
			"type":"object",
			"properties": {
				"sql": {"type":"string"},
				"connector": {"type":"string"}
			},
			"required":["sql"]
		}`,
	})
}

type builtinSQLArgs struct {
	Connector string `mapstructure:"connector"`
	SQL       string `mapstructure:"sql"`
	Priority  int    `mapstructure:"priority"`
}

// newBuiltinSQL is the resolver for the built-in /sql API.
// It executes a SQL query provided dynamically through the args.
// It errors if the user identified by the attributes is not an admin.
func newBuiltinSQL(ctx context.Context, opts *runtime.ResolverOptions) (runtime.Resolver, error) {
	// Only admins can run arbitrary SQL queries.
	if !opts.Claims.SkipChecks && !opts.Claims.Admin() {
		return nil, errors.New("must be an admin to run arbitrary SQL queries")
	}

	// Decode the args
	args := &builtinSQLArgs{}
	if err := mapstructure.Decode(opts.Args, args); err != nil {
		return nil, err
	}

	// Set the span attributes
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		span.SetAttributes(
			attribute.String("sql", args.SQL),
			attribute.String("connector", args.Connector),
		)
	}

	// Rewrite to the regular SQL resolver
	return newSQL(ctx, &runtime.ResolverOptions{
		Runtime:    opts.Runtime,
		InstanceID: opts.InstanceID,
		Properties: map[string]any{
			"connector": args.Connector,
			"sql":       args.SQL,
		},
		Args: map[string]any{
			"priority": args.Priority,
		},
		Claims:                 opts.Claims,
		ForExport:              opts.ForExport,
		SkipPropertyValidation: true,
	})
}
