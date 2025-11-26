package resolvers

import (
	"context"
	"errors"

	"github.com/mitchellh/mapstructure"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/metricsview/executor"
	"github.com/rilldata/rill/runtime/metricsview/metricssql"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func init() {
	runtime.RegisterResolverInitializer("metrics_sql", newMetricsSQL)
}

type metricsSQLProps struct {
	// SQL is the metrics SQL to evaluate.
	SQL string `mapstructure:"sql"`
	// TimeZone is a timezone to apply to the metrics SQL.
	TimeZone string `mapstructure:"time_zone"`
	// AdditionalWhere is a filter to apply to the metrics SQL. (additional WHERE clause)
	AdditionalWhere *metricsview.Expression `mapstructure:"additional_where"`
	// AdditionalTimeRange is a time range filter to apply to the metrics SQL.
	AdditionalTimeRange *metricsview.TimeRange `mapstructure:"additional_time_range"`
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
		span.SetAttributes(attribute.Bool("has_additional_time_range", props.AdditionalTimeRange != nil))
		span.SetAttributes(attribute.Bool("has_additional_time_zone", props.TimeZone != ""))
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

	// Create a metrics SQL parser
	compiler := metricssql.New(&metricssql.CompilerOptions{
		GetMetricsView: func(ctx context.Context, name string) (*runtimev1.Resource, error) {
			mv, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: name}, false)
			if err != nil {
				return nil, err
			}
			sec, err := opts.Runtime.ResolveSecurity(ctx, ctrl.InstanceID, opts.Claims, mv)
			if err != nil {
				return nil, err
			}
			if !sec.CanAccess() {
				return nil, runtime.ErrForbidden
			}
			return mv, nil
		},
		GetTimestamps: func(ctx context.Context, mv *runtimev1.Resource, timeDim string) (metricsview.TimestampsResult, error) {
			sec, err := opts.Runtime.ResolveSecurity(ctx, ctrl.InstanceID, opts.Claims, mv)
			if err != nil {
				return metricsview.TimestampsResult{}, err
			}
			e, err := executor.New(ctx, opts.Runtime, opts.InstanceID, mv.GetMetricsView().State.ValidSpec, false, sec, sqlArgs.Priority)
			if err != nil {
				return metricsview.TimestampsResult{}, err
			}
			return e.Timestamps(ctx, timeDim)
		},
	})

	// Parse the metrics SQL query
	query, err := compiler.Parse(ctx, props.SQL)
	if err != nil {
		return nil, err
	}

	// Inject the additional where clause if provided
	query.Where = applyAdditionalWhere(query.Where, props.AdditionalWhere)

	// Inject the additional time range if provided
	query.TimeRange = applyAdditionalTimeRange(query.TimeRange, props.AdditionalTimeRange)

	// Set the additional timezone if provided
	if props.TimeZone != "" {
		query.TimeZone = props.TimeZone
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

// applyAdditionalWhere combines the existing where clause with the additional where clause
func applyAdditionalWhere(current, additional *metricsview.Expression) *metricsview.Expression {
	if current == nil {
		return additional
	}
	if additional == nil {
		return current
	}

	// Combine the existing where clause with the additional where clause
	return &metricsview.Expression{
		Condition: &metricsview.Condition{
			Operator: metricsview.OperatorAnd,
			Expressions: []*metricsview.Expression{
				current,
				additional,
			},
		},
	}
}

// applyAdditionalTimeRange merges the existing time range with the additional time range
func applyAdditionalTimeRange(current, additional *metricsview.TimeRange) *metricsview.TimeRange {
	if current == nil {
		return additional
	}
	if additional == nil {
		return current
	}

	timeRange := &metricsview.TimeRange{
		Start:         current.Start,
		End:           current.End,
		Expression:    current.Expression,
		IsoDuration:   current.IsoDuration,
		IsoOffset:     current.IsoOffset,
		RoundToGrain:  current.RoundToGrain,
		TimeDimension: current.TimeDimension,
	}

	if !additional.Start.IsZero() && (timeRange.Start.IsZero() || additional.Start.After(timeRange.Start)) {
		timeRange.Start = additional.Start
	}
	if !additional.End.IsZero() && (timeRange.End.IsZero() || additional.End.Before(timeRange.End)) {
		timeRange.End = additional.End
	}
	if additional.Expression != "" && timeRange.Expression == "" {
		timeRange.Expression = additional.Expression
	}
	if additional.IsoDuration != "" && timeRange.IsoDuration == "" {
		timeRange.IsoDuration = additional.IsoDuration
	}
	if additional.IsoOffset != "" && timeRange.IsoOffset == "" {
		timeRange.IsoOffset = additional.IsoOffset
	}
	if additional.RoundToGrain != metricsview.TimeGrainUnspecified && timeRange.RoundToGrain == metricsview.TimeGrainUnspecified {
		timeRange.RoundToGrain = additional.RoundToGrain
	}
	if additional.TimeDimension != "" && timeRange.TimeDimension == "" {
		timeRange.TimeDimension = additional.TimeDimension
	}

	return timeRange
}
