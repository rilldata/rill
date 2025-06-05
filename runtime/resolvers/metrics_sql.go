package resolvers

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime"
	metricssqlparser "github.com/rilldata/rill/runtime/pkg/metricssql"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func init() {
	runtime.RegisterResolverInitializer("metrics_sql", newMetricsSQL)
}

type metricsSQLProps struct {
	SQL string `mapstructure:"sql"`
}

type metricsSQLArgs struct {
	Priority int `mapstructure:"priority"`
	// NOTE: Not exhaustive. Any other args are passed to the "args" property of sqlResolverOpts.
}

// newMetricsSQL creates a resolver for evaluating metrics SQL.
// It wraps the regular SQL resolver and compiles the metrics SQL to a regular SQL query first.
// The compiler preserves templating in the SQL, allowing the regular SQL resolver to handle SQL templating rules.
func newMetricsSQL(ctx context.Context, opts *runtime.ResolverOptions) (runtime.Resolver, error) {
	props := &metricsSQLProps{}
	if err := mapstructure.Decode(opts.Properties, props); err != nil {
		return nil, err
	}
	if props.SQL == "" {
		return nil, errors.New(`metrics SQL: missing required property "sql"`)
	}

	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		span.SetAttributes(attribute.String("metrics_sql", props.SQL))
	}

	instance, err := opts.Runtime.Instance(ctx, opts.InstanceID)
	if err != nil {
		return nil, err
	}

	// todo handle refs
	props.SQL, _, err = resolveTemplate(props.SQL, opts.Args, instance, opts.Claims.UserAttributes, opts.ForExport)
	if err != nil {
		return nil, err
	}

	ctrl, err := opts.Runtime.Controller(ctx, opts.InstanceID)
	if err != nil {
		return nil, err
	}

	sqlArgs := &metricsSQLArgs{}
	if err := mapstructure.Decode(opts.Args, sqlArgs); err != nil {
		return nil, err
	}

	compiler := metricssqlparser.New(ctrl, opts.InstanceID, opts.Claims, sqlArgs.Priority)
	query, err := compiler.Rewrite(ctx, props.SQL)
	if err != nil {
		return nil, err
	}

	// Build the options for the metrics resolver
	metricProps := map[string]any{}
	if err := mapstructure.WeakDecode(query, &metricProps); err != nil {
		return nil, err
	}
	resolverOpts := &runtime.ResolverOptions{
		Runtime:    opts.Runtime,
		InstanceID: opts.InstanceID,
		Properties: metricProps,
		Args:       opts.Args,
		Claims:     opts.Claims,
		ForExport:  opts.ForExport,
	}
	res, err := newMetrics(ctx, resolverOpts)
	if err != nil {
		return nil, err
	}

	// If the resolver is a metricsResolver, wrap it to include meta in the result
	if mr, ok := res.(*metricsResolver); ok {
		return &metaResolver{
			Resolver: res,
			meta:     mr.Meta(),
		}, nil
	}

	return res, nil
}

// metaResolver wraps a runtime.Resolver and injects meta into the ResolverResult's JSON output.
type metaResolver struct {
	runtime.Resolver
	meta any
}

func (r *metaResolver) ResolveInteractive(ctx context.Context) (runtime.ResolverResult, error) {
	res, err := r.Resolver.ResolveInteractive(ctx)
	if err != nil {
		return nil, err
	}
	return &metaResolverResult{
		ResolverResult: res,
		meta:           r.meta,
	}, nil
}

type metaResolverResult struct {
	runtime.ResolverResult
	meta any
}

func (r *metaResolverResult) MarshalJSON() ([]byte, error) {
	data, err := r.ResolverResult.MarshalJSON()
	if err != nil {
		return nil, err
	}
	var dataVal any
	if err := json.Unmarshal(data, &dataVal); err != nil {
		return nil, err
	}
	return json.Marshal(map[string]any{
		"data": dataVal,
		"meta": r.meta,
	})
}
