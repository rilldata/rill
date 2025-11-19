package queries

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/metricsview/executor"
	"github.com/rilldata/rill/runtime/pkg/pbutil"

	// Load IANA time zone data
	_ "time/tzdata"
)

type MetricsViewComparison struct {
	MetricsViewName     string                                         `json:"metrics_view_name,omitempty"`
	DimensionName       string                                         `json:"dimension_name,omitempty"`
	Measures            []*runtimev1.MetricsViewAggregationMeasure     `json:"measures,omitempty"`
	ComparisonMeasures  []string                                       `json:"comparison_measures,omitempty"`
	TimeRange           *runtimev1.TimeRange                           `json:"base_time_range,omitempty"`
	ComparisonTimeRange *runtimev1.TimeRange                           `json:"comparison_time_range,omitempty"`
	Limit               int64                                          `json:"limit,omitempty"`
	Offset              int64                                          `json:"offset,omitempty"`
	Sort                []*runtimev1.MetricsViewComparisonSort         `json:"sort,omitempty"`
	Where               *runtimev1.Expression                          `json:"where,omitempty"`
	WhereSQL            string                                         `json:"where_sql,omitempty"`
	Having              *runtimev1.Expression                          `json:"having,omitempty"`
	HavingSQL           string                                         `json:"having_sql,omitempty"`
	Filter              *runtimev1.MetricsViewFilter                   `json:"filter"` // Backwards compatibility
	Aliases             []*runtimev1.MetricsViewComparisonMeasureAlias `json:"aliases,omitempty"`
	Exact               bool                                           `json:"exact"`
	SecurityClaims      *runtime.SecurityClaims                        `json:"security_claims,omitempty"`
	ExecutionTime       *time.Time                                     `json:"execution_time,omitempty"`

	Result       *runtimev1.MetricsViewComparisonResponse `json:"-"`
	measuresMeta map[string]metricsViewMeasureMeta        `json:"-"`
}

type metricsViewMeasureMeta struct {
	baseSubqueryIndex int  // relative position of the measure in the inner query, 1 based
	outerIndex        int  // relative position of the measure in the outer query, this different from innerIndex as there may be derived measures like comparison, delta etc in the outer query after each base measure, 1 based
	expand            bool // whether the measure has derived measures like comparison, delta etc
}

var _ runtime.Query = &MetricsViewComparison{}

func (q *MetricsViewComparison) Key() string {
	r, err := json.Marshal(q)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("MetricsViewComparison:%s", r)
}

func (q *MetricsViewComparison) Deps() []*runtimev1.ResourceName {
	return []*runtimev1.ResourceName{
		{Kind: runtime.ResourceKindMetricsView, Name: q.MetricsViewName},
	}
}

func (q *MetricsViewComparison) MarshalResult() *runtime.QueryResult {
	return &runtime.QueryResult{
		Value: q.Result,
		Bytes: sizeProtoMessage(q.Result),
	}
}

func (q *MetricsViewComparison) UnmarshalResult(v any) error {
	res, ok := v.(*runtimev1.MetricsViewComparisonResponse)
	if !ok {
		return fmt.Errorf("MetricsViewComparison: mismatched unmarshal input")
	}
	q.Result = res
	return nil
}

func (q *MetricsViewComparison) Resolve(ctx context.Context, rt *runtime.Runtime, instanceID string, priority int) error {
	// Resolve metrics view
	mv, security, err := resolveMVAndSecurityFromAttributes(ctx, rt, instanceID, q.MetricsViewName, q.SecurityClaims)
	if err != nil {
		return err
	}

	err = q.calculateMeasuresMeta()
	if err != nil {
		return err
	}

	qry, err := q.rewriteToMetricsViewQuery(false)
	if err != nil {
		return fmt.Errorf("error rewriting to metrics query: %w", err)
	}

	e, err := executor.New(ctx, rt, instanceID, mv.ValidSpec, mv.Streaming, security, priority)
	if err != nil {
		return err
	}
	defer e.Close()

	res, err := e.Query(ctx, qry, q.ExecutionTime)
	if err != nil {
		return err
	}
	defer res.Close()

	data, err := rowsToData(res)
	if err != nil {
		return err
	}

	var rows []*runtimev1.MetricsViewComparisonRow
	for _, val := range data {
		val := val.AsMap()

		dv, err := pbutil.ToValue(val[q.DimensionName], safeFieldTypeName(res.Schema, q.DimensionName))
		if err != nil {
			return err
		}

		out := &runtimev1.MetricsViewComparisonRow{DimensionValue: dv}

		for _, m := range q.Measures {
			mv := &runtimev1.MetricsViewComparisonValue{MeasureName: m.Name}

			bv := val[m.Name]
			mv.BaseValue, err = pbutil.ToValue(bv, safeFieldTypeName(res.Schema, m.Name))
			if err != nil {
				return err
			}

			cv, ok := val[m.Name+"__previous"]
			if ok {
				mv.ComparisonValue, err = pbutil.ToValue(cv, safeFieldTypeName(res.Schema, m.Name+"__previous"))
				if err != nil {
					return err
				}
			}

			da, ok := val[m.Name+"__delta_abs"]
			if ok {
				mv.DeltaAbs, err = pbutil.ToValue(da, safeFieldTypeName(res.Schema, m.Name+"__delta_abs"))
				if err != nil {
					return err
				}
			}

			dr, ok := val[m.Name+"__delta_rel"]
			if ok {
				mv.DeltaRel, err = pbutil.ToValue(dr, safeFieldTypeName(res.Schema, m.Name+"__delta_rel"))
				if err != nil {
					return err
				}
			}

			out.MeasureValues = append(out.MeasureValues, mv)
		}

		rows = append(rows, out)
	}

	q.Result = &runtimev1.MetricsViewComparisonResponse{
		Rows: rows,
	}

	return nil
}

func (q *MetricsViewComparison) Export(ctx context.Context, rt *runtime.Runtime, instanceID string, w io.Writer, opts *runtime.ExportOptions) error {
	// Resolve metrics view
	mv, security, err := resolveMVAndSecurityFromAttributes(ctx, rt, instanceID, q.MetricsViewName, q.SecurityClaims)
	if err != nil {
		return err
	}

	err = q.calculateMeasuresMeta()
	if err != nil {
		return err
	}

	qry, err := q.rewriteToMetricsViewQuery(true)
	if err != nil {
		return fmt.Errorf("error rewriting to metrics query: %w", err)
	}

	e, err := executor.New(ctx, rt, instanceID, mv.ValidSpec, mv.Streaming, security, opts.Priority)
	if err != nil {
		return err
	}
	defer e.Close()

	var format drivers.FileFormat
	switch opts.Format {
	case runtimev1.ExportFormat_EXPORT_FORMAT_CSV:
		format = drivers.FileFormatCSV
	case runtimev1.ExportFormat_EXPORT_FORMAT_XLSX:
		format = drivers.FileFormatXLSX
	case runtimev1.ExportFormat_EXPORT_FORMAT_PARQUET:
		format = drivers.FileFormatParquet
	default:
		return fmt.Errorf("unsupported format: %s", opts.Format.String())
	}

	path, err := e.Export(ctx, qry, q.ExecutionTime, format, nil)
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(path) }()

	filename := q.generateFilename()
	err = opts.PreWriteHook(filename)
	if err != nil {
		return err
	}

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(w, f)
	if err != nil {
		return err
	}

	return nil
}

func (q *MetricsViewComparison) generateFilename() string {
	filename := strings.ReplaceAll(q.MetricsViewName, `"`, `_`)
	filename += "_" + q.DimensionName
	if q.Where != nil || q.Having != nil {
		filename += "_filtered"
	}
	return filename
}

func (q *MetricsViewComparison) rewriteToMetricsViewQuery(export bool) (*metricsview.Query, error) {
	qry := &metricsview.Query{MetricsView: q.MetricsViewName}

	qry.Dimensions = append(qry.Dimensions, metricsview.Dimension{Name: q.DimensionName})

	var measureFilter *runtimev1.Expression

	for _, m := range q.Measures {
		res := metricsview.Measure{Name: m.Name}
		switch m.BuiltinMeasure {
		case runtimev1.BuiltinMeasure_BUILTIN_MEASURE_COUNT:
			res.Compute = &metricsview.MeasureCompute{Count: true}
		case runtimev1.BuiltinMeasure_BUILTIN_MEASURE_COUNT_DISTINCT:
			res.Compute = &metricsview.MeasureCompute{CountDistinct: &metricsview.MeasureComputeCountDistinct{
				Dimension: m.BuiltinMeasureArgs[0].GetStringValue(),
			}}
		}

		if m.Filter != nil {
			if len(q.Measures) > 1 {
				return nil, fmt.Errorf("measure-level filter is not supported when multiple measures are present")
			}
			measureFilter = m.Filter
		}

		if m.Compute != nil {
			switch c := m.Compute.(type) {
			case *runtimev1.MetricsViewAggregationMeasure_ComparisonValue:
				res.Compute = &metricsview.MeasureCompute{ComparisonValue: &metricsview.MeasureComputeComparisonValue{
					Measure: c.ComparisonValue.Measure,
				}}
			case *runtimev1.MetricsViewAggregationMeasure_ComparisonDelta:
				res.Compute = &metricsview.MeasureCompute{ComparisonDelta: &metricsview.MeasureComputeComparisonDelta{
					Measure: c.ComparisonDelta.Measure,
				}}
			case *runtimev1.MetricsViewAggregationMeasure_ComparisonRatio:
				res.Compute = &metricsview.MeasureCompute{ComparisonRatio: &metricsview.MeasureComputeComparisonRatio{
					Measure: c.ComparisonRatio.Measure,
				}}
			}
		}

		qry.Measures = append(qry.Measures, res)
	}

	for _, m := range q.ComparisonMeasures {
		qry.Measures = append(qry.Measures, metricsview.Measure{
			Name: q.aliasForMeasure(m, runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_COMPARISON_VALUE),
			Compute: &metricsview.MeasureCompute{
				ComparisonValue: &metricsview.MeasureComputeComparisonValue{
					Measure: m,
				},
			},
		}, metricsview.Measure{
			Name: q.aliasForMeasure(m, runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_ABS_DELTA),
			Compute: &metricsview.MeasureCompute{
				ComparisonDelta: &metricsview.MeasureComputeComparisonDelta{
					Measure: m,
				},
			},
		}, metricsview.Measure{
			Name: q.aliasForMeasure(m, runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_REL_DELTA),
			Compute: &metricsview.MeasureCompute{
				ComparisonRatio: &metricsview.MeasureComputeComparisonRatio{
					Measure: m,
				},
			},
		})
	}

	if q.TimeRange != nil {
		res := &metricsview.TimeRange{}
		if q.TimeRange.Start != nil {
			res.Start = q.TimeRange.Start.AsTime()
		}
		if q.TimeRange.End != nil {
			res.End = q.TimeRange.End.AsTime()
		}
		res.IsoDuration = q.TimeRange.IsoDuration
		res.IsoOffset = q.TimeRange.IsoOffset
		res.RoundToGrain = metricsview.TimeGrainFromProto(q.TimeRange.RoundToGrain)
		res.TimeDimension = q.TimeRange.TimeDimension
		qry.TimeRange = res
		qry.TimeZone = q.TimeRange.TimeZone
	}

	if q.ComparisonTimeRange != nil {
		res := &metricsview.TimeRange{}
		if q.ComparisonTimeRange.Start != nil {
			res.Start = q.ComparisonTimeRange.Start.AsTime()
		}
		if q.ComparisonTimeRange.End != nil {
			res.End = q.ComparisonTimeRange.End.AsTime()
		}
		res.IsoDuration = q.ComparisonTimeRange.IsoDuration
		res.IsoOffset = q.ComparisonTimeRange.IsoOffset
		res.RoundToGrain = metricsview.TimeGrainFromProto(q.ComparisonTimeRange.RoundToGrain)
		res.TimeDimension = q.ComparisonTimeRange.TimeDimension
		qry.ComparisonTimeRange = res
	}

	if q.Limit != 0 {
		qry.Limit = &q.Limit
	}

	if q.Offset != 0 {
		qry.Offset = &q.Offset
	}

	for _, s := range q.Sort {
		qry.Sort = append(qry.Sort, metricsview.Sort{
			Name: q.aliasForMeasure(s.Name, s.SortType),
			Desc: s.Desc,
		})
	}

	if q.Filter != nil { // Backwards compatibility
		if q.Where != nil {
			return nil, fmt.Errorf("both filter and where is provided")
		}
		q.Where = convertFilterToExpression(q.Filter)
	}

	var err error
	qry.Where, err = metricViewExpression(q.Where, q.WhereSQL)
	if err != nil {
		return nil, fmt.Errorf("error converting where clause: %w", err)
	}

	// If a measure-level filter is present, we set qry.Where as the spine, and use (qry.Where AND measuresFilter) as the new where clause
	if measureFilter != nil {
		qry.Spine = &metricsview.Spine{Where: &metricsview.WhereSpine{Expression: qry.Where}}

		if qry.Where == nil {
			qry.Where = metricsview.NewExpressionFromProto(measureFilter)
		} else {
			qry.Where = &metricsview.Expression{
				Condition: &metricsview.Condition{
					Operator: metricsview.OperatorAnd,
					Expressions: []*metricsview.Expression{
						qry.Where,
						metricsview.NewExpressionFromProto(measureFilter),
					},
				},
			}
		}
	}

	qry.Having, err = metricViewExpression(q.Having, q.HavingSQL)
	if err != nil {
		return nil, fmt.Errorf("error converting having clause: %w", err)
	}

	qry.UseDisplayNames = export

	return qry, nil
}

func (q *MetricsViewComparison) aliasForMeasure(name string, t runtimev1.MetricsViewComparisonMeasureType) string {
	for _, a := range q.Aliases {
		if a.Name == name && a.Type == t {
			return a.Alias
		}
	}

	switch t {
	case runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_COMPARISON_VALUE:
		return name + "__previous"
	case runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_ABS_DELTA:
		return name + "__delta_abs"
	case runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_REL_DELTA:
		return name + "__delta_rel"
	}

	return name
}

func (q *MetricsViewComparison) calculateMeasuresMeta() error {
	compare := !isTimeRangeNil(q.ComparisonTimeRange)

	if !compare && len(q.ComparisonMeasures) > 0 {
		return fmt.Errorf("comparison measures are provided but comparison time range is not")
	}

	if len(q.ComparisonMeasures) == 0 && compare {
		// backwards compatibility
		q.ComparisonMeasures = make([]string, 0, len(q.Measures))
		for _, m := range q.Measures {
			if m.BuiltinMeasure != runtimev1.BuiltinMeasure_BUILTIN_MEASURE_UNSPECIFIED {
				continue
			}
			q.ComparisonMeasures = append(q.ComparisonMeasures, m.Name)
		}
	}

	q.measuresMeta = make(map[string]metricsViewMeasureMeta, len(q.Measures))

	inner := 1
	outer := 1
	for _, m := range q.Measures {
		expand := false
		for _, cm := range q.ComparisonMeasures {
			if m.Name == cm {
				expand = true
				break
			}
		}
		q.measuresMeta[m.Name] = metricsViewMeasureMeta{
			baseSubqueryIndex: inner,
			outerIndex:        outer,
			expand:            expand,
		}
		if expand {
			outer += 4
		} else {
			outer++
		}
		inner++
	}

	// check all comparison measures are present in the measures list
	for _, cm := range q.ComparisonMeasures {
		if _, ok := q.measuresMeta[cm]; !ok {
			return fmt.Errorf("comparison measure '%s' is not present in the measures list", cm)
		}
	}

	err := validateSort(q.Sort, q.measuresMeta, compare)
	if err != nil {
		return err
	}

	err = validateMeasureAliases(q.Aliases, q.measuresMeta, compare)
	if err != nil {
		return err
	}

	return nil
}

func validateSort(sorts []*runtimev1.MetricsViewComparisonSort, measuresMeta map[string]metricsViewMeasureMeta, hasComparison bool) error {
	if len(sorts) == 0 {
		return fmt.Errorf("sorting is required")
	}
	firstSort := sorts[0].Type

	for _, s := range sorts {
		if firstSort != s.Type {
			return fmt.Errorf("different sort types are not supported in a single query")
		}
		// Update sort to make sure it is backwards compatible
		if s.SortType == runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_UNSPECIFIED && s.Type != runtimev1.MetricsViewComparisonSortType_METRICS_VIEW_COMPARISON_SORT_TYPE_UNSPECIFIED {
			switch s.Type {
			case runtimev1.MetricsViewComparisonSortType_METRICS_VIEW_COMPARISON_SORT_TYPE_BASE_VALUE:
				s.SortType = runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_BASE_VALUE
			case runtimev1.MetricsViewComparisonSortType_METRICS_VIEW_COMPARISON_SORT_TYPE_COMPARISON_VALUE:
				s.SortType = runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_COMPARISON_VALUE
			case runtimev1.MetricsViewComparisonSortType_METRICS_VIEW_COMPARISON_SORT_TYPE_ABS_DELTA:
				s.SortType = runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_ABS_DELTA
			case runtimev1.MetricsViewComparisonSortType_METRICS_VIEW_COMPARISON_SORT_TYPE_REL_DELTA:
				s.SortType = runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_REL_DELTA
			}
		}

		if hasComparison {
			// check if sorting measure is a derived measure, if it is then it should be present in comparison measures list
			// don't do this check for non comparison query for backward compatibility, UI uses the old state while switching from compare to no comparison
			// in that case just sort the measure by base value
			if s.SortType == runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_COMPARISON_VALUE ||
				s.SortType == runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_ABS_DELTA ||
				s.SortType == runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_REL_DELTA {
				if !measuresMeta[s.Name].expand {
					return fmt.Errorf("comparison not enabled for sort measure '%s'", s.Name)
				}
			}
		}
	}
	return nil
}

func validateMeasureAliases(aliases []*runtimev1.MetricsViewComparisonMeasureAlias, measuresMeta map[string]metricsViewMeasureMeta, hasComparison bool) error {
	for _, alias := range aliases {
		switch alias.Type {
		case runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_COMPARISON_VALUE,
			runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_ABS_DELTA,
			runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_REL_DELTA:
			if !hasComparison || !measuresMeta[alias.Name].expand {
				return fmt.Errorf("comparison not enabled for alias %s", alias.Alias)
			}
		}
	}
	return nil
}

func isTimeRangeNil(tr *runtimev1.TimeRange) bool {
	return tr == nil || (tr.Start == nil && tr.End == nil)
}
