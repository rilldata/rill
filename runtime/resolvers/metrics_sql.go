package resolvers

import (
	"context"
	"errors"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/metricsview"
	metricssqlparser "github.com/rilldata/rill/runtime/pkg/metricssql"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func init() {
	runtime.RegisterResolverInitializer("metrics_sql", newMetricsSQL)
}

type metricsSQLProps struct {
	// SQL is the metrics SQL to evaluate.
	SQL string `mapstructure:"sql"`
	// AdditionalWhere is a filter to apply to the metrics SQL. (additional WHERE clause)
	AdditionalWhere *metricsview.Expression `mapstructure:"additional_where"`
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
		span.SetAttributes(attribute.Bool("has_additional_where", props.AdditionalWhere != nil))
	}

	instance, err := opts.Runtime.Instance(ctx, opts.InstanceID)
	if err != nil {
		return nil, err
	}

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

	// Inject the additional where clause if provided
	if props.AdditionalWhere != nil {
		expr := props.AdditionalWhere
		if query.Where != nil {
			query.Where = &metricsview.Expression{
				Condition: &metricsview.Condition{
					Operator:    metricsview.OperatorAnd,
					Expressions: []*metricsview.Expression{query.Where, expr},
				},
			}
		} else {
			query.Where = expr
		}
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
	return newMetrics(ctx, resolverOpts)
}
