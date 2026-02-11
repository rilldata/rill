package resolvers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/formatter"
	"github.com/rilldata/rill/runtime/pkg/mapstructureutil"
	"github.com/rilldata/rill/runtime/queries"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
)

func init() {
	runtime.RegisterResolverInitializer("legacy_metrics", newLegacyMetrics)
}

type legacyMetricsResolver struct {
	runtime         *runtime.Runtime
	instanceID      string
	query           runtime.Query
	args            *legacyMetricsResolverArgs
	metricsViewName string
	logger          *zap.Logger
}

type legacyMetricsResolverProps struct {
	QueryName     string `mapstructure:"query_name"`
	QueryArgsJSON string `mapstructure:"query_args_json"`
}

type legacyMetricsResolverArgs struct {
	Priority      int        `mapstructure:"priority"`
	ExecutionTime *time.Time `mapstructure:"execution_time"`
	Limit         int        `mapstructure:"limit"`
	Format        bool       `mapstructure:"format"`
}

func newLegacyMetrics(ctx context.Context, opts *runtime.ResolverOptions) (runtime.Resolver, error) {
	props := &legacyMetricsResolverProps{}
	if err := mapstructureutil.WeakDecode(opts.Properties, props); err != nil {
		return nil, err
	}

	args := &legacyMetricsResolverArgs{}
	if err := mapstructureutil.WeakDecode(opts.Args, args); err != nil {
		return nil, err
	}

	// Build query proto
	qpb, err := queries.ProtoFromJSON(props.QueryName, props.QueryArgsJSON, args.ExecutionTime)
	if err != nil {
		return nil, fmt.Errorf("failed to parse query: %w", err)
	}

	metricsViewName, err := queries.MetricsViewFromQuery(props.QueryName, props.QueryArgsJSON)
	if err != nil {
		return nil, fmt.Errorf("failed extract metrics view name from query: %w", err)
	}

	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		span.SetAttributes(attribute.String("metrics_view", metricsViewName))
	}

	q, err := queries.ProtoToQuery(qpb, opts.Claims, args.ExecutionTime)
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	return &legacyMetricsResolver{
		runtime:         opts.Runtime,
		instanceID:      opts.InstanceID,
		query:           q,
		args:            args,
		metricsViewName: metricsViewName,
		logger:          opts.Runtime.Logger,
	}, nil
}

func (r *legacyMetricsResolver) Close() error {
	return nil
}

func (r *legacyMetricsResolver) CacheKey(ctx context.Context) ([]byte, bool, error) {
	return nil, false, nil
}

func (r *legacyMetricsResolver) Refs() []*runtimev1.ResourceName {
	return []*runtimev1.ResourceName{{Kind: runtime.ResourceKindMetricsView, Name: r.metricsViewName}}
}

func (r *legacyMetricsResolver) Validate(ctx context.Context) error {
	return nil
}

func (r *legacyMetricsResolver) ResolveInteractive(ctx context.Context) (runtime.ResolverResult, error) {
	ctrl, err := r.runtime.Controller(ctx, r.instanceID)
	if err != nil {
		return nil, err
	}

	metricsView, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: r.metricsViewName}, false)
	if err != nil {
		return nil, err
	}

	spec := metricsView.GetMetricsView().State.ValidSpec
	if spec == nil {
		return nil, fmt.Errorf("metrics view spec is not valid")
	}

	err = r.runtime.Query(ctx, r.instanceID, r.query, r.args.Priority)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	var out []map[string]any
	var schema *runtimev1.StructType

	switch q := r.query.(type) {
	case *queries.MetricsViewAggregation:
		schema = q.Result.Schema
		if q.Result != nil {
			for i, row := range q.Result.Data {
				if r.args.Limit > 0 && i >= r.args.Limit {
					break
				}
				if r.args.Format {
					out = append(out, r.formatMetricsViewAggregationResult(row.AsMap(), q, spec.Measures))
					continue
				}
				out = append(out, row.AsMap())
			}
		}
	case *queries.MetricsViewComparison:
		if q.Result != nil {
			for i, row := range q.Result.Rows {
				if r.args.Limit > 0 && i >= r.args.Limit {
					break
				}
				if r.args.Format {
					out = append(out, r.formatMetricsViewComparisonResult(row, q, spec.Measures))
					continue
				}
				r := make(map[string]any)
				r[q.DimensionName] = row.DimensionValue
				for _, v := range row.MeasureValues {
					r[v.MeasureName] = v.BaseValue.AsInterface()
					if v.ComparisonValue != nil {
						r[v.MeasureName+" (prev)"] = v.ComparisonValue.AsInterface()
					}
					if v.DeltaAbs != nil {
						r[v.MeasureName+" (Δ)"] = v.DeltaAbs.AsInterface()
					}
					if v.DeltaRel != nil {
						r[v.MeasureName+" (Δ%)"] = v.DeltaRel.AsInterface()
					}
				}
				out = append(out, r)
			}
		}
	default:
		return nil, fmt.Errorf("query type %T not supported", q)
	}

	data, err := json.Marshal(out)
	if err != nil {
		return nil, err
	}

	return &legacyResolverResult{
		data:   data,
		schema: schema,
	}, nil
}

func (r *legacyMetricsResolver) ResolveExport(ctx context.Context, w io.Writer, opts *runtime.ResolverExportOptions) error {
	return errors.New("not implemented")
}

func (r *legacyMetricsResolver) InferRequiredSecurityRules() ([]*runtimev1.SecurityRule, error) {
	// Extract fields and row filter from the query using the queries.SecurityFromRuntimeQuery helper
	rowFilter, fields, err := queries.SecurityFromRuntimeQuery(r.query)
	if err != nil {
		return nil, fmt.Errorf("failed to extract accessible fields: %w", err)
	}

	var rules []*runtimev1.SecurityRule

	if rowFilter != "" {
		expr := &runtimev1.Expression{}
		err := protojson.Unmarshal([]byte(rowFilter), expr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse row filter expression: %w", err)
		}

		rules = append(rules, &runtimev1.SecurityRule{
			Rule: &runtimev1.SecurityRule_RowFilter{
				RowFilter: &runtimev1.SecurityRuleRowFilter{
					ConditionResources: []*runtimev1.ResourceName{{Kind: runtime.ResourceKindMetricsView, Name: r.metricsViewName}},
					Expression:         expr,
				},
			},
		})
	}

	if len(fields) > 0 {
		rules = append(rules, &runtimev1.SecurityRule{
			Rule: &runtimev1.SecurityRule_FieldAccess{
				FieldAccess: &runtimev1.SecurityRuleFieldAccess{
					ConditionResources: []*runtimev1.ResourceName{{Kind: runtime.ResourceKindMetricsView, Name: r.metricsViewName}},
					Fields:             fields,
					Allow:              true,
					Exclusive:          true,
				},
			},
		})
	}

	return rules, nil
}

func (r *legacyMetricsResolver) formatMetricsViewAggregationResult(row map[string]interface{}, q *queries.MetricsViewAggregation, measures []*runtimev1.MetricsViewSpec_Measure) map[string]any {
	res := make(map[string]any)
	for k, v := range row {
		measureLabel, f := r.getComparisonMeasureLabelAndFormatter(k, q.Measures, measures)
		res[measureLabel] = r.formatValue(f, v)
	}
	return res
}

func (r *legacyMetricsResolver) formatMetricsViewComparisonResult(row *runtimev1.MetricsViewComparisonRow, q *queries.MetricsViewComparison, measures []*runtimev1.MetricsViewSpec_Measure) map[string]any {
	res := make(map[string]any)
	res[q.DimensionName] = row.DimensionValue
	for _, v := range row.MeasureValues {
		measureLabel, f := r.getMeasureLabelAndFormatter(v.MeasureName, measures)
		res[measureLabel] = r.formatValue(f, v.BaseValue.AsInterface())
		if v.ComparisonValue != nil {
			res[measureLabel+" (prev)"] = r.formatValue(f, v.ComparisonValue.AsInterface())
		}
		if v.DeltaAbs != nil {
			res[measureLabel+" (Δ)"] = r.formatValue(f, v.DeltaAbs.AsInterface())
		}
		if v.DeltaRel != nil {
			fp, err := formatter.NewPresetFormatter("percentage", false)
			if err != nil {
				r.logger.Warn("Failed to get formatter, using no formatter", zap.Error(err))
				fp = nil
			}
			res[measureLabel+" (Δ%)"] = r.formatValue(fp, v.DeltaRel.AsInterface())
		}
	}
	return res
}

// getComparisonMeasureLabelAndFormatter gets the measure label and formatter by a measure name and adds a suffix if it was compared measure.
// for relative change comparison it uses percent formatter, uses defined preset for everything else
// if a measure is not found in the request list, it returns the measure name as the label and no formatter.
// if the measure is not found in the metrics view measures, it returns the measure name as the label and no formatter.
// if the formatter fails to load, it logs the error and returns the measure name as the label and no formatter.
func (r *legacyMetricsResolver) getComparisonMeasureLabelAndFormatter(measureName string, reqMeasures []*runtimev1.MetricsViewAggregationMeasure, measures []*runtimev1.MetricsViewSpec_Measure) (string, formatter.Formatter) {
	var reqMeasure *runtimev1.MetricsViewAggregationMeasure
	effectiveMeasure := measureName
	for _, m := range reqMeasures {
		if measureName == m.Name {
			reqMeasure = m
			// get the actual measure comparison is based on
			switch v := m.Compute.(type) {
			case *runtimev1.MetricsViewAggregationMeasure_ComparisonValue:
				effectiveMeasure = v.ComparisonValue.Measure
			case *runtimev1.MetricsViewAggregationMeasure_ComparisonDelta:
				effectiveMeasure = v.ComparisonDelta.Measure
			case *runtimev1.MetricsViewAggregationMeasure_ComparisonRatio:
				effectiveMeasure = v.ComparisonRatio.Measure
			case *runtimev1.MetricsViewAggregationMeasure_PercentOfTotal:
				effectiveMeasure = v.PercentOfTotal.Measure
			}
			break
		}
	}
	if reqMeasure == nil {
		return measureName, nil
	}

	var measure *runtimev1.MetricsViewSpec_Measure
	for _, m := range measures {
		if effectiveMeasure == m.Name {
			measure = m
			break
		}
	}

	if measure == nil {
		return effectiveMeasure, nil
	}

	measureLabel := measure.DisplayName
	if measureLabel == "" {
		measureLabel = measureName
	}
	formatPreset := measure.FormatPreset
	if effectiveMeasure != measureName {
		// comparison measure, add a suffix based on type
		switch reqMeasure.Compute.(type) {
		case *runtimev1.MetricsViewAggregationMeasure_ComparisonValue:
			measureLabel += " (prev)"
		case *runtimev1.MetricsViewAggregationMeasure_ComparisonDelta:
			measureLabel += " (Δ)"
		case *runtimev1.MetricsViewAggregationMeasure_ComparisonRatio:
			measureLabel += " (Δ%)"
			formatPreset = "percentage"
		case *runtimev1.MetricsViewAggregationMeasure_PercentOfTotal:
			measureLabel += " (Σ%)"
			formatPreset = "percentage"
		}
	}

	// D3 formatting isn't implemented yet so using the format preset only for now
	f, err := formatter.NewPresetFormatter(formatPreset, false)
	if err != nil {
		r.logger.Warn("Failed to get formatter, using no formatter", zap.Error(err))
		return measureLabel, nil
	}

	return measureLabel, f
}

// getMeasureLabelAndFormatter gets the measure label and formatter by a measure name.
// if the measure is not found, it returns the measure name as the label and no formatter.
// if the formatter fails to load, it logs the error and returns the measure name as the label and no formatter.
func (r *legacyMetricsResolver) getMeasureLabelAndFormatter(measureName string, measures []*runtimev1.MetricsViewSpec_Measure) (string, formatter.Formatter) {
	var measure *runtimev1.MetricsViewSpec_Measure
	for _, m := range measures {
		if measureName == m.Name {
			measure = m
			break
		}
	}

	if measure == nil {
		return measureName, nil
	}

	measureLabel := measure.DisplayName
	if measureLabel == "" {
		measureLabel = measureName
	}

	// D3 formatting isn't implemented yet so using the format preset only for now
	f, err := formatter.NewPresetFormatter(measure.FormatPreset, false)
	if err != nil {
		r.logger.Warn("Failed to get formatter, using no formatter", zap.Error(err))
		return measureLabel, nil
	}

	return measureLabel, f
}

// formatValue formats a measure value using the provided formatter.
// If the formatter is nil, or value is nil, or an error occurred, it will log a warning and return the value as is.
func (r *legacyMetricsResolver) formatValue(f formatter.Formatter, v any) any {
	if f == nil || v == nil {
		return v
	}
	if s, err := f.StringFormat(v); err == nil {
		return s
	}
	r.logger.Warn("Failed to format measure value", zap.Any("value", v))
	return fmt.Sprintf("%v", v)
}

type legacyResolverResult struct {
	data   []byte
	schema *runtimev1.StructType

	rows []map[string]any
	idx  int
}

func (r *legacyResolverResult) Close() error {
	return nil
}

// Meta implements runtime.ResolverResult.
func (r *legacyResolverResult) Meta() map[string]any {
	return nil
}

func (r *legacyResolverResult) Schema() *runtimev1.StructType {
	return r.schema
}

func (r *legacyResolverResult) Next() (map[string]any, error) {
	if r.rows == nil {
		if err := json.Unmarshal(r.data, &r.rows); err != nil {
			return nil, err
		}
	}
	if r.idx >= len(r.rows) {
		return nil, io.EOF
	}
	row := r.rows[r.idx]
	r.idx++
	return row, nil
}

func (r *legacyResolverResult) MarshalJSON() ([]byte, error) {
	return r.data, nil
}
