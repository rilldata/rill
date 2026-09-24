package resolvers

import (
	"context"
	"errors"
	"fmt"

	"github.com/mitchellh/mapstructure"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/metricsview/executor"
	"github.com/rilldata/rill/runtime/metricsview/metricssql"
	"github.com/rilldata/rill/runtime/pkg/mapstructureutil"
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
	// AdditionalWhereByMetricsView is a map of metrics view names to filters to apply to the metrics SQL.
	AdditionalWhereByMetricsView map[string]*metricsview.Expression `mapstructure:"additional_where_by_metrics_view"`
	// AdditionalTimeRange is a time range filter to apply to the metrics SQL.
	AdditionalTimeRange *metricsview.TimeRange `mapstructure:"additional_time_range"`
	// AdditionalTimeGrain buckets the time dimensions selected by the metrics SQL at this grain.
	// Dimensions that already declare a grain, i.e. an explicit date_trunc in the SQL, are left alone.
	AdditionalTimeGrain metricsview.TimeGrain `mapstructure:"additional_time_grain"`
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
	if err := mapstructureutil.WeakDecode(opts.Properties, props); err != nil {
		return nil, err
	}
	if props.SQL == "" {
		return nil, errors.New(`metrics SQL: missing required property "sql"`)
	}
	if !props.AdditionalTimeGrain.Valid() {
		return nil, fmt.Errorf("metrics SQL: invalid time grain %q", props.AdditionalTimeGrain)
	}

	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		span.SetAttributes(
			attribute.String("metrics_sql", props.SQL),
			attribute.String("time_zone", props.TimeZone),
			attribute.Bool("has_additional_where", props.AdditionalWhere != nil || len(props.AdditionalWhereByMetricsView) > 0),
			attribute.Bool("has_additional_time_range", props.AdditionalTimeRange != nil),
			attribute.String("additional_time_grain", string(props.AdditionalTimeGrain)),
		)
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

	getMetricsView := func(ctx context.Context, name string) (*runtimev1.Resource, error) {
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
	}

	// Create a metrics SQL parser
	compiler := metricssql.New(&metricssql.CompilerOptions{
		GetMetricsView: getMetricsView,
		GetTimestamps: func(ctx context.Context, mv *runtimev1.Resource, timeDim string) (metricsview.TimestampsResult, error) {
			sec, err := opts.Runtime.ResolveSecurity(ctx, ctrl.InstanceID, opts.Claims, mv)
			if err != nil {
				return metricsview.TimestampsResult{}, err
			}
			var userAttrs map[string]any
			if opts.Claims != nil {
				userAttrs = opts.Claims.UserAttributes
			}
			e, err := executor.New(ctx, opts.Runtime, opts.InstanceID, mv.GetMetricsView().State.ValidSpec, false, sec, sqlArgs.Priority, userAttrs)
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
	if where, ok := props.AdditionalWhereByMetricsView[query.MetricsView]; ok {
		query.Where = applyAdditionalWhere(query.Where, where)
	}

	// Inject the additional time range if provided
	query.TimeRange = applyAdditionalTimeRange(query.TimeRange, props.AdditionalTimeRange)

	// Bucket the query's time dimensions if a grain was provided
	if props.AdditionalTimeGrain != metricsview.TimeGrainUnspecified {
		mv, err := getMetricsView(ctx, query.MetricsView)
		if err != nil {
			return nil, err
		}
		applyAdditionalTimeGrain(query, mv.GetMetricsView().State.ValidSpec, props.AdditionalTimeGrain)
	}

	// Set the additional timezone if provided
	if props.TimeZone != "" {
		query.TimeZone = props.TimeZone
	}

	// Build the options for the metrics resolver
	metricProps, err := query.AsMap()
	if err != nil {
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

// applyAdditionalTimeGrain floors the query's time dimensions at the given grain.
// A dimension that already computes a time floor came from an explicit date_trunc in the metrics SQL,
// so it keeps the bucket its author asked for.
// The dimension's name is left untouched, so the result columns are the same as without a grain.
func applyAdditionalTimeGrain(query *metricsview.Query, spec *runtimev1.MetricsViewSpec, grain metricsview.TimeGrain) {
	if spec == nil {
		return
	}

	for i, dim := range query.Dimensions {
		if dim.Compute != nil {
			continue
		}

		// The primary time dimension is not necessarily declared in the dimensions list, so check it separately.
		isTime := dim.Name == spec.TimeDimension
		for _, d := range spec.Dimensions {
			if d.Name == dim.Name {
				isTime = d.Type == runtimev1.MetricsViewSpec_DIMENSION_TYPE_TIME
				break
			}
		}
		if !isTime {
			continue
		}

		query.Dimensions[i].Compute = &metricsview.DimensionCompute{
			TimeFloor: &metricsview.DimensionComputeTimeFloor{
				Dimension: dim.Name,
				Grain:     grain,
			},
		}
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
