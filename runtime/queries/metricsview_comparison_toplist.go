package queries

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/expressionpb"
	"github.com/rilldata/rill/runtime/pkg/pbutil"
	"google.golang.org/protobuf/types/known/structpb"

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
	Having              *runtimev1.Expression                          `json:"having,omitempty"`
	Filter              *runtimev1.MetricsViewFilter                   `json:"filter"` // Backwards compatibility
	Aliases             []*runtimev1.MetricsViewComparisonMeasureAlias `json:"aliases,omitempty"`
	Exact               bool                                           `json:"exact"`
	SecurityAttributes  map[string]any                                 `json:"security_attributes,omitempty"`

	Result *runtimev1.MetricsViewComparisonResponse `json:"-"`

	// Internal
	measuresMeta map[string]metricsViewMeasureMeta `json:"-"`
}

type metricsViewMeasureMeta struct {
	baseSubqueryIndex       int // relative position of the measure in the inner query, 1 based
	outerIndex              int // relative position of the measure in the outer query, this different from innerIndex as there may be derived measures like comparison, delta etc in the outer query after each base measure, 1 based
	comparisonSubqueryIndex int
	expand                  bool // whether the measure has derived measures like comparison, delta etc
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
	mv, security, err := resolveMVAndSecurityFromAttributes(ctx, rt, instanceID, q.MetricsViewName, q.SecurityAttributes, []*runtimev1.MetricsViewAggregationDimension{{Name: q.DimensionName}}, q.Measures)
	if err != nil {
		return err
	}

	olap, release, err := rt.OLAP(ctx, instanceID, mv.Connector)
	if err != nil {
		return err
	}
	defer release()

	if olap.Dialect() != drivers.DialectDuckDB && olap.Dialect() != drivers.DialectDruid && olap.Dialect() != drivers.DialectClickHouse {
		return fmt.Errorf("not available for dialect '%s'", olap.Dialect())
	}

	if mv.TimeDimension == "" && (!isTimeRangeNil(q.TimeRange) || !isTimeRangeNil(q.ComparisonTimeRange)) {
		return fmt.Errorf("metrics view '%s' does not have a time dimension", q.MetricsViewName)
	}

	// backwards compatibility
	if q.Filter != nil {
		if q.Where != nil {
			return fmt.Errorf("both filter and where is provided")
		}
		q.Where = convertFilterToExpression(q.Filter)
	}

	err = q.calculateMeasuresMeta()
	if err != nil {
		return err
	}

	// comparison toplist
	if !isTimeRangeNil(q.ComparisonTimeRange) {
		// execute toplist for base and get dim list
		// create and add filter and execute comprison toplist
		// remove strict limits in comp toplist sql
		if drivers.DialectDruid != olap.Dialect() || q.Exact {
			return q.executeComparisonToplist(ctx, olap, mv, priority, security)
		}

		return q.executeDruidApproximateComparison(ctx, olap, mv, priority, security)
	}

	// general toplist
	if drivers.DialectDruid != olap.Dialect() || q.Exact {
		return q.executeToplist(ctx, olap, mv, priority, security)
	}

	return q.executeDruidApproximateToplist(ctx, olap, mv, priority, security)
}

// Druid-based `exactify` approach:
// 1. The first query fetch topN dimensions.
// 2. The second query fetches topN filtered by the collected dimensions.
// The dimension filter contrains topN table avoiding approximation in measures (due to mearging multiple topN Druid results from different nodes).
func (q *MetricsViewComparison) executeDruidApproximateComparison(ctx context.Context, olap drivers.OLAPStore, mv *runtimev1.MetricsViewSpec, priority int, security *runtime.ResolvedMetricsViewSecurity) error {
	originalMeasures := q.removeNoSortMeasures()

	if q.isBase() || q.isDeltaComparison() {
		err := q.executeToplist(ctx, olap, mv, priority, security)
		if err != nil {
			return err
		}
	} else {
		ttr := q.TimeRange
		q.TimeRange = q.ComparisonTimeRange
		err := q.executeToplist(ctx, olap, mv, priority, security)
		if err != nil {
			return err
		}

		q.TimeRange = ttr
	}

	q.addDimsAsFilter()
	q.Measures = originalMeasures
	return q.executeComparisonToplist(ctx, olap, mv, priority, security)
}

// Druid-based `exactify` approach (see comments above)
// Optimizations:
// * the first query fetches only sorted dimensions
// * the second query isn't run if the topN already less than the limit
func (q *MetricsViewComparison) executeDruidApproximateToplist(ctx context.Context, olap drivers.OLAPStore, mv *runtimev1.MetricsViewSpec, priority int, security *runtime.ResolvedMetricsViewSecurity) error {
	originalMeasures := q.Measures
	if len(q.Measures) >= 5 {
		originalMeasures = q.removeNoSortMeasures()
	}

	err := q.executeToplist(ctx, olap, mv, priority, security)
	if err != nil {
		return err
	}

	if len(q.Result.Rows) < int(q.Limit) && len(q.Measures) == len(originalMeasures) {
		return nil
	}

	q.addDimsAsFilter()

	q.Measures = originalMeasures

	// remove limit since we have already added filter with only toplist values and order clause will be present
	q.Limit = 0

	return q.executeToplist(ctx, olap, mv, priority, security)
}

func (q *MetricsViewComparison) removeNoSortMeasures() []*runtimev1.MetricsViewAggregationMeasure {
	measures := q.Measures
	sortMeasures := make([]*runtimev1.MetricsViewAggregationMeasure, 0, len(q.Sort))
	for _, m := range q.Measures {
		for _, s := range q.Sort {
			if s.Name == m.Name {
				sortMeasures = append(sortMeasures, m)
			}
		}
	}
	q.Measures = sortMeasures
	return measures
}

func (q *MetricsViewComparison) addDimsAsFilter() {
	inExpressions := make([]*runtimev1.Expression, 0, len(q.Result.Rows))
	for _, r := range q.Result.Rows {
		inExpressions = append(inExpressions, expressionpb.Value(r.DimensionValue))
	}
	if q.Where != nil {
		q.Where = expressionpb.And([]*runtimev1.Expression{q.Where, expressionpb.In(expressionpb.Identifier(q.DimensionName), inExpressions)})
	} else {
		q.Where = expressionpb.In(expressionpb.Identifier(q.DimensionName), inExpressions)
	}
}

func (q *MetricsViewComparison) calculateMeasuresMeta() error {
	compare := !isTimeRangeNil(q.ComparisonTimeRange)

	if !compare && len(q.ComparisonMeasures) > 0 {
		return fmt.Errorf("comparison measures are provided but comparison time range is not")
	}

	if len(q.ComparisonMeasures) == 0 && compare {
		// backwards compatibility
		q.ComparisonMeasures = make([]string, len(q.Measures))
		for i, m := range q.Measures {
			q.ComparisonMeasures[i] = m.Name
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

func (q *MetricsViewComparison) executeToplist(ctx context.Context, olap drivers.OLAPStore, mv *runtimev1.MetricsViewSpec, priority int, policy *runtime.ResolvedMetricsViewSecurity) error {
	sql, args, err := q.buildMetricsTopListSQL(mv, olap.Dialect(), policy, false)
	if err != nil {
		return fmt.Errorf("error building query: %w", err)
	}

	rows, err := olap.Execute(ctx, &drivers.Statement{
		Query:    sql,
		Args:     args,
		Priority: priority,
	})
	if err != nil {
		return err
	}
	defer rows.Close()

	var data []*runtimev1.MetricsViewComparisonRow
	for rows.Next() {
		values, err := rows.SliceScan()
		if err != nil {
			return err
		}
		measureValues := make([]*runtimev1.MetricsViewComparisonValue, 0, len(q.Measures))

		for i, m := range q.Measures {
			v, err := pbutil.ToValue(values[1+i], safeFieldType(rows.Schema, 1+i))
			if err != nil {
				return err
			}

			measureValues = append(measureValues, &runtimev1.MetricsViewComparisonValue{
				MeasureName: m.Name,
				BaseValue:   v,
			})
		}

		dv, err := pbutil.ToValue(values[0], safeFieldType(rows.Schema, 0))
		if err != nil {
			return err
		}

		data = append(data, &runtimev1.MetricsViewComparisonRow{
			DimensionValue: dv,
			MeasureValues:  measureValues,
		})
	}

	q.Result = &runtimev1.MetricsViewComparisonResponse{
		Rows: data,
	}

	return nil
}

func (q *MetricsViewComparison) executeComparisonToplist(ctx context.Context, olap drivers.OLAPStore, mv *runtimev1.MetricsViewSpec, priority int, policy *runtime.ResolvedMetricsViewSecurity) error {
	sql, args, err := q.buildMetricsComparisonTopListSQL(mv, olap.Dialect(), policy, false)
	if err != nil {
		return fmt.Errorf("error building query: %w", err)
	}

	rows, err := olap.Execute(ctx, &drivers.Statement{
		Query:    sql,
		Args:     args,
		Priority: priority,
	})
	if err != nil {
		return err
	}
	defer rows.Close()

	var data []*runtimev1.MetricsViewComparisonRow
	for rows.Next() {
		values, err := rows.SliceScan()
		if err != nil {
			return err
		}
		var measureValues []*runtimev1.MetricsViewComparisonValue

		for _, m := range q.Measures {
			measureMeta := q.measuresMeta[m.Name]
			index := measureMeta.outerIndex
			if measureMeta.expand {
				bv, err := pbutil.ToValue(values[index], safeFieldType(rows.Schema, index))
				if err != nil {
					return err
				}

				cv, err := pbutil.ToValue(values[1+index], safeFieldType(rows.Schema, 1+index))
				if err != nil {
					return err
				}

				da, err := pbutil.ToValue(values[2+index], safeFieldType(rows.Schema, 2+index))
				if err != nil {
					return err
				}

				dr, err := pbutil.ToValue(values[3+index], safeFieldType(rows.Schema, 3+index))
				if err != nil {
					return err
				}

				measureValues = append(measureValues, &runtimev1.MetricsViewComparisonValue{
					MeasureName:     m.Name,
					BaseValue:       bv,
					ComparisonValue: cv,
					DeltaAbs:        da,
					DeltaRel:        dr,
				})
			} else {
				v, err := pbutil.ToValue(values[index], safeFieldType(rows.Schema, index))
				if err != nil {
					return err
				}

				measureValues = append(measureValues, &runtimev1.MetricsViewComparisonValue{
					MeasureName: m.Name,
					BaseValue:   v,
				})
			}
		}

		dv, err := pbutil.ToValue(values[0], safeFieldType(rows.Schema, 0))
		if err != nil {
			return err
		}

		data = append(data, &runtimev1.MetricsViewComparisonRow{
			DimensionValue: dv,
			MeasureValues:  measureValues,
		})
	}

	q.Result = &runtimev1.MetricsViewComparisonResponse{
		Rows: data,
	}

	return nil
}

func (q *MetricsViewComparison) buildMetricsTopListSQL(mv *runtimev1.MetricsViewSpec, dialect drivers.Dialect, policy *runtime.ResolvedMetricsViewSecurity, export bool) (string, []any, error) {
	dim, err := metricsViewDimension(mv, q.DimensionName)
	if err != nil {
		return "", nil, err
	}
	colName := safeName(dim.Name)

	labelMap := make(map[string]string, len(mv.Measures))
	for _, m := range mv.Measures {
		labelMap[m.Name] = m.Name
		if m.Label != "" {
			labelMap[m.Name] = m.Label
		}
	}

	var labelCols []string
	var selectCols []string
	dimLabel := colName
	if dim.Label != "" {
		dimLabel = safeName(dim.Label)
	}
	dimSel, unnestClause := dialect.DimensionSelect(mv.Database, mv.DatabaseSchema, mv.Table, dim)
	selectCols = append(selectCols, dimSel)
	labelCols = []string{fmt.Sprintf("%s as %s", safeName(dim.Name), dimLabel)}

	for _, m := range q.Measures {
		switch m.BuiltinMeasure {
		case runtimev1.BuiltinMeasure_BUILTIN_MEASURE_UNSPECIFIED:
			expr, err := metricsViewMeasureExpression(mv, m.Name)
			if err != nil {
				return "", nil, err
			}
			selectCols = append(selectCols, fmt.Sprintf("%s as %s", expr, safeName(m.Name)))
			labelCols = append(labelCols, fmt.Sprintf("%s as %s", safeName(m.Name), safeName(labelMap[m.Name])))
		case runtimev1.BuiltinMeasure_BUILTIN_MEASURE_COUNT:
			selectCols = append(selectCols, fmt.Sprintf("COUNT(*) as %s", safeName(m.Name)))
			labelCols = append(labelCols, fmt.Sprintf("%s as %s", safeName(m.Name), safeName(labelMap[m.Name])))
		case runtimev1.BuiltinMeasure_BUILTIN_MEASURE_COUNT_DISTINCT:
			if len(m.BuiltinMeasureArgs) != 1 {
				return "", nil, fmt.Errorf("builtin measure '%s' expects 1 argument", m.BuiltinMeasure.String())
			}
			arg := m.BuiltinMeasureArgs[0].GetStringValue()
			if arg == "" {
				return "", nil, fmt.Errorf("builtin measure '%s' expects non-empty string argument, got '%v'", m.BuiltinMeasure.String(), m.BuiltinMeasureArgs[0])
			}
			selectCols = append(selectCols, fmt.Sprintf("COUNT(DISTINCT %s) as %s", safeName(arg), safeName(m.Name)))
			labelCols = append(labelCols, fmt.Sprintf("%s as %s", safeName(m.Name), safeName(labelMap[m.Name])))
		default:
			return "", nil, fmt.Errorf("unknown builtin measure '%d'", m.BuiltinMeasure)
		}
	}

	selectClause := strings.Join(selectCols, ", ")
	baseWhereClause := "1=1"

	args := []any{}
	td := safeName(mv.TimeDimension)
	if dialect == drivers.DialectDuckDB {
		td = fmt.Sprintf("%s::TIMESTAMP", td)
	}

	trc, err := timeRangeClause(q.TimeRange, mv, td, &args)
	if err != nil {
		return "", nil, err
	}
	baseWhereClause += trc

	if q.Where != nil {
		clause, clauseArgs, err := buildExpression(mv, q.Where, nil, dialect)
		if err != nil {
			return "", nil, err
		}
		if strings.TrimSpace(clause) != "" {
			baseWhereClause += fmt.Sprintf(" AND (%s)", clause)
		}

		args = append(args, clauseArgs...)
	}

	if policy != nil && policy.RowFilter != "" {
		baseWhereClause += fmt.Sprintf(" AND (%s)", policy.RowFilter)
	}

	havingClause := ""
	if q.Having != nil {
		var havingClauseArgs []any
		havingClause, havingClauseArgs, err = buildExpression(mv, q.Having, q.Aliases, dialect)
		if err != nil {
			return "", nil, err
		}
		if strings.TrimSpace(havingClause) != "" {
			havingClause = "HAVING " + havingClause
		}
		args = append(args, havingClauseArgs...)
	}

	var orderClauses []string
	for _, s := range q.Sort {
		if s.Name == q.DimensionName {
			clause := "1"
			if s.Desc {
				clause += " DESC"
			}
			if dialect == drivers.DialectDuckDB {
				clause += " NULLS LAST"
			}
			orderClauses = append(orderClauses, clause)
			break
		}
		clause := safeName(s.Name)
		if s.Desc {
			clause += " DESC"
		}
		if dialect == drivers.DialectDuckDB {
			clause += " NULLS LAST"
		}
		orderClauses = append(orderClauses, clause)
	}

	orderByClause := ""
	if len(orderClauses) > 0 {
		orderByClause = "ORDER BY " + strings.Join(orderClauses, ", ")
	}

	limitClause := ""
	if q.Limit > 0 {
		limitClause = fmt.Sprintf(" LIMIT %d", q.Limit)
	}

	var sql string
	if export {
		labelSelectClause := strings.Join(labelCols, ", ")
		sql = fmt.Sprintf(
			`SELECT %[8]s FROM (SELECT %[1]s FROM %[2]s %[7]s WHERE %[3]s GROUP BY 1 %[9]s %[4]s %[5]s OFFSET %[6]d)`,
			selectClause,                        // 1
			escapeMetricsViewTable(dialect, mv), // 2
			baseWhereClause,                     // 3
			orderByClause,                       // 4
			limitClause,                         // 5
			q.Offset,                            // 6
			unnestClause,                        // 7
			labelSelectClause,                   // 8
			havingClause,                        // 9
		)
	} else {
		sql = fmt.Sprintf(
			`SELECT %[1]s FROM %[2]s %[7]s WHERE %[3]s GROUP BY 1 %[8]s %[4]s %[5]s OFFSET %[6]d`,
			selectClause,                        // 1
			escapeMetricsViewTable(dialect, mv), // 2
			baseWhereClause,                     // 3
			orderByClause,                       // 4
			limitClause,                         // 5
			q.Offset,                            // 6
			unnestClause,                        // 7
			havingClause,                        // 8
		)
	}

	return sql, args, nil
}

func (q *MetricsViewComparison) buildMetricsComparisonTopListSQL(mv *runtimev1.MetricsViewSpec, dialect drivers.Dialect, policy *runtime.ResolvedMetricsViewSecurity, export bool) (string, []any, error) {
	dim, err := metricsViewDimension(mv, q.DimensionName)
	if err != nil {
		return "", nil, err
	}

	colName := safeName(dim.Name)

	labelMap := make(map[string]string, len(mv.Measures))
	for _, m := range mv.Measures {
		labelMap[m.Name] = m.Name
		if m.Label != "" {
			labelMap[m.Name] = m.Label
		}
	}

	var selectCols []string
	var comparisonSelectCols []string
	dimSel, unnestClause := dialect.DimensionSelect(mv.Database, mv.DatabaseSchema, mv.Table, dim)
	selectCols = append(selectCols, dimSel)
	comparisonSelectCols = append(comparisonSelectCols, dimSel)

	for _, m := range q.Measures {
		switch m.BuiltinMeasure {
		case runtimev1.BuiltinMeasure_BUILTIN_MEASURE_UNSPECIFIED:
			expr, err := metricsViewMeasureExpression(mv, m.Name)
			if err != nil {
				return "", nil, err
			}
			selectCols = append(selectCols, fmt.Sprintf("%s as %s", expr, safeName(m.Name)))
			if q.measuresMeta[m.Name].expand {
				comparisonSelectCols = append(comparisonSelectCols, fmt.Sprintf("%s as %s", expr, safeName(m.Name)))
			}
		case runtimev1.BuiltinMeasure_BUILTIN_MEASURE_COUNT:
			selectCols = append(selectCols, fmt.Sprintf("COUNT(*) as %s", safeName(m.Name)))
			if q.measuresMeta[m.Name].expand {
				comparisonSelectCols = append(comparisonSelectCols, fmt.Sprintf("COUNT(*) as %s", safeName(m.Name)))
			}
		case runtimev1.BuiltinMeasure_BUILTIN_MEASURE_COUNT_DISTINCT:
			if len(m.BuiltinMeasureArgs) != 1 {
				return "", nil, fmt.Errorf("builtin measure '%s' expects 1 argument", m.BuiltinMeasure.String())
			}
			arg := m.BuiltinMeasureArgs[0].GetStringValue()
			if arg == "" {
				return "", nil, fmt.Errorf("builtin measure '%s' expects non-empty string argument, got '%v'", m.BuiltinMeasure.String(), m.BuiltinMeasureArgs[0])
			}
			selectCols = append(selectCols, fmt.Sprintf("COUNT(DISTINCT %s) as %s", safeName(arg), safeName(m.Name)))
			if q.measuresMeta[m.Name].expand {
				comparisonSelectCols = append(comparisonSelectCols, fmt.Sprintf("COUNT(DISTINCT %s) as %s", safeName(arg), safeName(m.Name)))
			}
		default:
			return "", nil, fmt.Errorf("unknown builtin measure '%d'", m.BuiltinMeasure)
		}
	}

	var finalSelectCols []string
	var labelCols []string
	for _, m := range q.Measures {
		var columnsTuple string
		var labelTuple string
		if dialect != drivers.DialectDruid {
			if q.measuresMeta[m.Name].expand {
				columnsTuple = fmt.Sprintf(
					"base.%[1]s AS %[1]s, comparison.%[1]s AS %[2]s, base.%[1]s - comparison.%[1]s AS %[3]s, (base.%[1]s - comparison.%[1]s)/comparison.%[1]s::DOUBLE AS %[4]s",
					safeName(m.Name),
					safeName(m.Name+"__previous"),
					safeName(m.Name+"__delta_abs"),
					safeName(m.Name+"__delta_rel"),
				)
				labelTuple = fmt.Sprintf(
					"base.%[1]s AS %[5]s, comparison.%[1]s AS %[2]s, base.%[1]s - comparison.%[1]s AS %[3]s, (base.%[1]s - comparison.%[1]s)/comparison.%[1]s::DOUBLE AS %[4]s",
					safeName(m.Name),
					safeName(labelMap[m.Name]+" (prev)"),
					safeName(labelMap[m.Name]+" (Δ)"),
					safeName(labelMap[m.Name]+" (Δ%)"),
					safeName(labelMap[m.Name]),
				)
			} else {
				columnsTuple = fmt.Sprintf("base.%[1]s AS %[1]s", safeName(m.Name))
				labelTuple = fmt.Sprintf("base.%[1]s AS %[2]s", safeName(m.Name), safeName(labelMap[m.Name]))
			}
		} else {
			if q.measuresMeta[m.Name].expand {
				columnsTuple = fmt.Sprintf(
					"ANY_VALUE(base.%[1]s) AS %[1]s, ANY_VALUE(comparison.%[1]s) AS %[2]s, ANY_VALUE(base.%[1]s - comparison.%[1]s) AS %[3]s, ANY_VALUE(SAFE_DIVIDE(base.%[1]s - comparison.%[1]s, CAST(comparison.%[1]s AS DOUBLE))) AS %[4]s",
					safeName(m.Name),
					safeName(m.Name+"__previous"),
					safeName(m.Name+"__delta_abs"),
					safeName(m.Name+"__delta_rel"),
				)
				labelTuple = fmt.Sprintf(
					"ANY_VALUE(base.%[1]s) AS %[2]s, ANY_VALUE(comparison.%[1]s) AS %[3]s, ANY_VALUE(base.%[1]s - comparison.%[1]s) AS %[4]s, ANY_VALUE(SAFE_DIVIDE(base.%[1]s - comparison.%[1]s, CAST(comparison.%[1]s AS DOUBLE))) AS %[5]s",
					safeName(m.Name),
					safeName(labelMap[m.Name]),
					safeName(labelMap[m.Name]+" (prev)"),
					safeName(labelMap[m.Name]+" (Δ)"),
					safeName(labelMap[m.Name]+" (Δ%)"),
				)
			} else {
				columnsTuple = fmt.Sprintf("ANY_VALUE(base.%[1]s) AS %[1]s", safeName(m.Name))
				labelTuple = fmt.Sprintf("ANY_VALUE(base.%[1]s) AS %[2]s", safeName(m.Name), safeName(labelMap[m.Name]))
			}
		}
		finalSelectCols = append(
			finalSelectCols,
			columnsTuple,
		)
		labelCols = append(labelCols, labelTuple)
	}

	subSelectClause := strings.Join(selectCols, ", ")
	subComparisonSelectClause := strings.Join(comparisonSelectCols, ", ")
	finalSelectClause := strings.Join(finalSelectCols, ", ")
	labelSelectClause := strings.Join(labelCols, ", ")
	if export {
		finalSelectClause = labelSelectClause
	}

	baseWhereClause := "1=1"
	comparisonWhereClause := "1=1"

	var args []any
	if mv.TimeDimension == "" {
		return "", nil, fmt.Errorf("metrics view '%s' doesn't have time dimension", q.MetricsViewName)
	}

	td := safeName(mv.TimeDimension)
	if dialect == drivers.DialectDuckDB {
		td = fmt.Sprintf("%s::TIMESTAMP", td)
	}

	whereClause, whereClauseArgs, err := buildExpression(mv, q.Where, nil, dialect)
	if err != nil {
		return "", nil, err
	}

	trc, err := timeRangeClause(q.TimeRange, mv, td, &args)
	if err != nil {
		return "", nil, err
	}
	baseWhereClause += trc

	if whereClause != "" {
		baseWhereClause += fmt.Sprintf(" AND (%s)", whereClause)
		args = append(args, whereClauseArgs...)
	}

	trc, err = timeRangeClause(q.ComparisonTimeRange, mv, td, &args)
	if err != nil {
		return "", nil, err
	}
	comparisonWhereClause += trc

	if whereClause != "" {
		comparisonWhereClause += fmt.Sprintf(" AND (%s)", whereClause)
		args = append(args, whereClauseArgs...)
	}

	if policy != nil && policy.RowFilter != "" {
		baseWhereClause += fmt.Sprintf(" AND (%s)", policy.RowFilter)
		comparisonWhereClause += fmt.Sprintf(" AND (%s)", policy.RowFilter)
	}

	havingClause := "1=1"
	if q.Having != nil {
		var havingClauseArgs []any
		havingClause, havingClauseArgs, err = buildExpression(mv, q.Having, q.Aliases, dialect)
		if err != nil {
			return "", nil, err
		}
		args = append(args, havingClauseArgs...)
	}

	var orderClauses []string
	var subQueryOrderClauses []string
	for _, s := range q.Sort {
		if s.Name == q.DimensionName {
			clause := "1"
			subQueryClause := "1"
			var ending string
			if s.Desc {
				ending += " DESC"
			}
			if dialect == drivers.DialectDuckDB {
				ending += " NULLS LAST"
			}
			clause += ending
			subQueryClause += ending
			orderClauses = append(orderClauses, clause)
			subQueryOrderClauses = append(subQueryOrderClauses, subQueryClause)
			break
		}
		measureMeta, ok := q.measuresMeta[s.Name]
		if !ok {
			return "", nil, fmt.Errorf("metrics view '%s' doesn't contain '%s' sort column", q.MetricsViewName, s.Name)
		}

		var pos int
		switch s.SortType {
		case runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_BASE_VALUE:
			pos = 1 + measureMeta.outerIndex
		case runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_COMPARISON_VALUE:
			pos = 2 + measureMeta.outerIndex
		case runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_ABS_DELTA:
			pos = 3 + measureMeta.outerIndex
		case runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_REL_DELTA:
			pos = 4 + measureMeta.outerIndex
		default:
			return "", nil, fmt.Errorf("undefined sort type for measure %s", s.Name)
		}
		orderClause := fmt.Sprint(pos)
		subQueryOrderClause := fmt.Sprint(measureMeta.baseSubqueryIndex + 1) // 1-based + skip the first dim column
		ending := ""
		if s.Desc {
			ending += " DESC"
		}
		if dialect == drivers.DialectDuckDB {
			ending += " NULLS LAST"
		}
		orderClause += ending
		subQueryOrderClause += ending
		orderClauses = append(orderClauses, orderClause)
		subQueryOrderClauses = append(subQueryOrderClauses, subQueryOrderClause)
	}

	orderByClause := ""
	subQueryOrderByClause := ""
	if len(orderClauses) > 0 {
		orderByClause = "ORDER BY " + strings.Join(orderClauses, ", ")
		subQueryOrderByClause = "ORDER BY " + strings.Join(subQueryOrderClauses, ", ")
	}

	limitClause := ""
	if q.Limit > 0 {
		limitClause = fmt.Sprintf(" LIMIT %d", q.Limit)
	}

	baseLimitClause := ""
	comparisonLimitClause := ""

	joinType := "FULL"
	if !q.Exact {
		deltaComparison := q.Sort[0].SortType == runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_ABS_DELTA ||
			q.Sort[0].SortType == runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_REL_DELTA

		approximationLimit := int(q.Limit)
		if q.Limit != 0 && q.Limit < 100 && deltaComparison {
			approximationLimit = 100
		}

		if q.Sort[0].SortType == runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_BASE_VALUE || deltaComparison {
			joinType = "LEFT OUTER"
			baseLimitClause = subQueryOrderByClause
			if approximationLimit > 0 {
				baseLimitClause += fmt.Sprintf(" LIMIT %d", approximationLimit)
			}
		} else if q.Sort[0].SortType == runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_COMPARISON_VALUE {
			joinType = "RIGHT OUTER"
			comparisonLimitClause = subQueryOrderByClause
			if approximationLimit > 0 {
				comparisonLimitClause += fmt.Sprintf(" LIMIT %d", approximationLimit)
			}
		}
	}

	/*
		Example of the SQL:

		SELECT COALESCE(base."domain", comparison."domain") AS "dom", base."measure_1", comparison."measure_1" AS "measure_1__previous", base."measure_1" - comparison."measure_1" AS "measure_1__delta_abs", (base."measure_1" - comparison."measure_1")/comparison."measure_1"::DOUBLE AS "measure_1__delta_rel" FROM
			(
				SELECT "domain", avg(bid_price) as "measure_1" FROM "ad_bids" WHERE 1=1 AND "timestamp" >= ? AND "timestamp" < ? GROUP BY "domain" ORDER BY true, 1 NULLS LAST LIMIT 100
			) base
		LEFT OUTER JOIN
			(
				SELECT "domain", avg(bid_price) as "measure_1" FROM "ad_bids" WHERE 1=1 AND "timestamp" >= ? AND "timestamp" < ? GROUP BY "domain"
			) comparison
		ON
				base."domain" = comparison."domain" OR (base."domain" is null and comparison."domain" is null)
		ORDER BY
			true, 1 NULLS LAST
		LIMIT 10
		OFFSET 0
	*/

	finalDimName := safeName(q.DimensionName)
	if export && dim.Label != "" {
		finalDimName = safeName(dim.Label)
	}
	var sql string
	if dialect != drivers.DialectDruid {
		var joinOnClause string
		if dialect == drivers.DialectClickHouse {
			joinOnClause = fmt.Sprintf("isNotDistinctFrom(base.%[1]s, comparison.%[1]s)", colName)
		} else {
			joinOnClause = fmt.Sprintf("base.%[1]s = comparison.%[1]s OR (base.%[1]s is null and comparison.%[1]s is null)", colName)
		}
		if havingClause != "" {
			// measure filter could include the base measure name.
			// this leads to ambiguity whether it applies to the base.measure ot comparison.measure.
			// to keep the clause builder consistent we add an outer query here.
			sql = fmt.Sprintf(`
				SELECT * from (
					SELECT COALESCE(base.%[2]s, comparison.%[2]s) AS %[10]s, %[9]s FROM 
						(
							SELECT %[1]s FROM %[3]s %[14]s WHERE %[4]s GROUP BY 1 %[12]s 
						) base
					%[11]s JOIN
						(
							SELECT %[16]s FROM %[3]s %[14]s WHERE %[5]s GROUP BY 1 %[13]s 
						) comparison
					ON
							%[17]s
					%[6]s
					%[7]s
					OFFSET
						%[8]d
				) WHERE %[15]s 
			`,
				subSelectClause,                     // 1
				colName,                             // 2
				escapeMetricsViewTable(dialect, mv), // 3
				baseWhereClause,                     // 4
				comparisonWhereClause,               // 5
				orderByClause,                       // 6
				limitClause,                         // 7
				q.Offset,                            // 8
				finalSelectClause,                   // 9
				finalDimName,                        // 10
				joinType,                            // 11
				baseLimitClause,                     // 12
				comparisonLimitClause,               // 13
				unnestClause,                        // 14
				havingClause,                        // 15
				subComparisonSelectClause,           // 16
				joinOnClause,                        // 17
			)
		} else {
			sql = fmt.Sprintf(`
		SELECT COALESCE(base.%[2]s, comparison.%[2]s) AS %[10]s, %[9]s FROM 
			(
				SELECT %[1]s FROM %[3]s %[14]s WHERE %[4]s GROUP BY 1 %[12]s 
			) base
		%[11]s JOIN
			(
				SELECT %[15]s FROM %[3]s %[14]s WHERE %[5]s GROUP BY 1 %[13]s 
			) comparison
		ON
				%[16]s)
		%[6]s
		%[7]s
		OFFSET
			%[8]d
		`,
				subSelectClause,                     // 1
				colName,                             // 2
				escapeMetricsViewTable(dialect, mv), // 3
				baseWhereClause,                     // 4
				comparisonWhereClause,               // 5
				orderByClause,                       // 6
				limitClause,                         // 7
				q.Offset,                            // 8
				finalSelectClause,                   // 9
				finalDimName,                        // 10
				joinType,                            // 11
				baseLimitClause,                     // 12
				comparisonLimitClause,               // 13
				unnestClause,                        // 14
				subComparisonSelectClause,           // 15
				joinOnClause,                        // 16
			)
		}
	} else {
		/*
			Example of the SQL query with expression based dimension:

				WITH base AS (
				  SELECT (replace("channel", 'a', 'b')) as "b",
					count(*) as "total_records", sum("added") as "sum"
					FROM "wikipedia"
					WHERE 1=1 AND "__time" >= '2016-06-27T02:00:00.000Z' AND "__time" < '2016-06-27T03:00:00.000Z'
					GROUP BY 1 -- Druid does not support group by aliases
					ORDER BY 2 DESC
					LIMIT 500 OFFSET 0
				), comparison AS (
				  SELECT (replace("channel", 'a', 'b')) as "c",
					count(*) as "total_records"
					FROM "wikipedia"
					WHERE 1=1 AND "__time" >= '2016-06-27T01:00:00.000Z' AND "__time" < '2016-06-27T02:00:00.000Z'
					AND replace("channel", 'a', 'b') IN (SELECT "b" FROM base)
					GROUP BY 1 -- Druid does not support group by aliases
					LIMIT 500
				)
				SELECT base."b" AS "channel",
					ANY_VALUE(base."total_records") AS "total_records",
					ANY_VALUE(comparison."total_records") AS "total_records__previous",
					ANY_VALUE(base."total_records" - comparison."total_records") AS "total_records__delta_abs",
					ANY_VALUE(SAFE_DIVIDE(base."total_records" - comparison."total_records", CAST(comparison."total_records" AS DOUBLE))) AS "total_records__delta_rel",
					ANY_VALUE(base."sum") AS "sum",
				FROM base LEFT JOIN comparison ON base."b" = comparison."c"
				GROUP BY 1 -- Druid does not support group by aliases
				HAVING 1=1
				ORDER BY 2 DESC -- order by without group by is not supported by Druid
				 LIMIT 250
				OFFSET 0

			Apache Druid requires that one part of the JOIN fits in memory, that can be achieved by pushing down the limit clause to a subquery (works only if the sorting is based entirely on a single subquery result)
		*/
		leftSubQueryAlias := "base"
		rightSubQueryAlias := "comparison"
		leftWhereClause := baseWhereClause
		rightWhereClause := comparisonWhereClause

		if q.Sort[0].SortType == runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_COMPARISON_VALUE {
			leftSubQueryAlias = "comparison"
			rightSubQueryAlias = "base"
			leftWhereClause = comparisonWhereClause
			rightWhereClause = baseWhereClause
		}

		twiceTheLimitClause := ""
		if q.Exact {
			if q.Limit > 0 {
				twiceTheLimitClause = fmt.Sprintf(" LIMIT %d", q.Limit*2)
			} else if q.Limit == 0 {
				twiceTheLimitClause = fmt.Sprintf(" LIMIT %d", 100_000) // use Druid limit
			}
		}

		sql = fmt.Sprintf(`
				WITH %[11]s AS (
					SELECT %[1]s FROM %[3]s WHERE %[4]s GROUP BY 1 %[13]s %[10]s OFFSET %[8]d
				), %[12]s AS (
					SELECT %[17]s FROM %[3]s WHERE %[5]s AND %[16]s IN (SELECT %[2]s FROM %[11]s) GROUP BY 1 %[10]s
				)
				SELECT %[11]s.%[2]s AS %[14]s, %[9]s FROM %[11]s LEFT JOIN %[12]s ON base.%[2]s = comparison.%[2]s
				GROUP BY 1
				HAVING %[15]s
				%[6]s
				%[7]s
				OFFSET %[8]d
			`,
			subSelectClause,                     // 1
			colName,                             // 2
			escapeMetricsViewTable(dialect, mv), // 3
			leftWhereClause,                     // 4
			rightWhereClause,                    // 5
			orderByClause,                       // 6
			limitClause,                         // 7
			q.Offset,                            // 8
			finalSelectClause,                   // 9
			twiceTheLimitClause,                 // 10
			leftSubQueryAlias,                   // 11
			rightSubQueryAlias,                  // 12
			subQueryOrderByClause,               // 13
			finalDimName,                        // 14
			havingClause,                        // 15
			dialect.MetricsViewDimensionExpression(dim), // 16
			subComparisonSelectClause,                   // 17
		)
	}

	return sql, args, nil
}

func (q *MetricsViewComparison) Export(ctx context.Context, rt *runtime.Runtime, instanceID string, w io.Writer, opts *runtime.ExportOptions) error {
	// Resolve metrics view
	mv, security, err := resolveMVAndSecurityFromAttributes(ctx, rt, instanceID, q.MetricsViewName, q.SecurityAttributes, []*runtimev1.MetricsViewAggregationDimension{{Name: q.DimensionName}}, q.Measures)
	if err != nil {
		return err
	}

	olap, release, err := rt.OLAP(ctx, instanceID, mv.Connector)
	if err != nil {
		return err
	}
	defer release()

	switch olap.Dialect() {
	case drivers.DialectDuckDB:
		if opts.Format == runtimev1.ExportFormat_EXPORT_FORMAT_CSV || opts.Format == runtimev1.ExportFormat_EXPORT_FORMAT_PARQUET {
			// temporary backwards compatibility
			if q.Filter != nil {
				if q.Where != nil {
					return fmt.Errorf("both filter and where is provided")
				}
				q.Where = convertFilterToExpression(q.Filter)
			}

			err := q.calculateMeasuresMeta()
			if err != nil {
				return err
			}
			var sql string
			var args []any
			if !isTimeRangeNil(q.ComparisonTimeRange) {
				sql, args, err = q.buildMetricsComparisonTopListSQL(mv, olap.Dialect(), security, true)
				if err != nil {
					return fmt.Errorf("error building query: %w", err)
				}
			} else {
				sql, args, err = q.buildMetricsTopListSQL(mv, olap.Dialect(), security, true)
				if err != nil {
					return fmt.Errorf("error building query: %w", err)
				}
			}

			filename := q.generateFilename()
			if err := DuckDBCopyExport(ctx, w, opts, sql, args, filename, olap, opts.Format); err != nil {
				return err
			}
		} else {
			if err := q.generalExport(ctx, rt, instanceID, w, opts, mv); err != nil {
				return err
			}
		}
	case drivers.DialectDruid:
		if err := q.generalExport(ctx, rt, instanceID, w, opts, mv); err != nil {
			return err
		}
	case drivers.DialectClickHouse:
		if err := q.generalExport(ctx, rt, instanceID, w, opts, mv); err != nil {
			return err
		}
	default:
		return fmt.Errorf("not available for dialect '%s'", olap.Dialect())
	}

	return nil
}

func (q *MetricsViewComparison) generalExport(ctx context.Context, rt *runtime.Runtime, instanceID string, w io.Writer, opts *runtime.ExportOptions, mv *runtimev1.MetricsViewSpec) error {
	err := q.Resolve(ctx, rt, instanceID, opts.Priority)
	if err != nil {
		return err
	}

	if opts.PreWriteHook != nil {
		err = opts.PreWriteHook(q.generateFilename())
		if err != nil {
			return err
		}
	}

	labelMap := make(map[string]string, len(mv.Measures))
	for _, m := range mv.Measures {
		labelMap[m.Name] = m.Name
		if m.Label != "" {
			labelMap[m.Name] = m.Label
		}
	}
	var metaLen int
	for _, m := range q.Result.Rows[0].MeasureValues {
		if q.measuresMeta[m.MeasureName].expand {
			metaLen += 4
		} else {
			metaLen++
		}
	}

	meta := make([]*runtimev1.MetricsViewColumn, metaLen+1)
	dimName := q.DimensionName
	for _, d := range mv.Dimensions {
		if d.Name == q.DimensionName && d.Label != "" {
			dimName = d.Label
		}
	}
	meta[0] = &runtimev1.MetricsViewColumn{
		Name: dimName,
		Type: runtimev1.Type_CODE_STRING.String(),
	}

	i := 1
	for _, m := range q.Result.Rows[0].MeasureValues {
		meta[i] = &runtimev1.MetricsViewColumn{
			Name: labelMap[m.MeasureName],
			Type: runtimev1.Type_CODE_FLOAT64.String(),
		}
		i++
		if q.measuresMeta[m.MeasureName].expand {
			meta[i] = &runtimev1.MetricsViewColumn{
				Name: fmt.Sprintf("%s (prev)", labelMap[m.MeasureName]),
				Type: runtimev1.Type_CODE_FLOAT64.String(),
			}
			meta[i+1] = &runtimev1.MetricsViewColumn{
				Name: fmt.Sprintf("%s (Δ)", labelMap[m.MeasureName]),
				Type: runtimev1.Type_CODE_FLOAT64.String(),
			}
			meta[i+2] = &runtimev1.MetricsViewColumn{
				Name: fmt.Sprintf("%s (Δ%%)", labelMap[m.MeasureName]),
				Type: runtimev1.Type_CODE_FLOAT64.String(),
			}
			i += 3
		}
	}

	data := make([]*structpb.Struct, len(q.Result.Rows))
	for i, row := range q.Result.Rows {
		data[i] = &structpb.Struct{
			Fields: map[string]*structpb.Value{
				dimName: {
					Kind: &structpb.Value_StringValue{
						StringValue: row.DimensionValue.GetStringValue(),
					},
				},
			},
		}
		for _, m := range row.MeasureValues {
			data[i].Fields[labelMap[m.MeasureName]] = &structpb.Value{
				Kind: &structpb.Value_NumberValue{
					NumberValue: m.BaseValue.GetNumberValue(),
				},
			}
			if q.measuresMeta[m.MeasureName].expand {
				data[i].Fields[fmt.Sprintf("%s (prev)", labelMap[m.MeasureName])] = &structpb.Value{
					Kind: &structpb.Value_NumberValue{
						NumberValue: m.ComparisonValue.GetNumberValue(),
					},
				}
				data[i].Fields[fmt.Sprintf("%s (Δ)", labelMap[m.MeasureName])] = &structpb.Value{
					Kind: &structpb.Value_NumberValue{
						NumberValue: m.DeltaAbs.GetNumberValue(),
					},
				}
				data[i].Fields[fmt.Sprintf("%s (Δ%%)", labelMap[m.MeasureName])] = &structpb.Value{
					Kind: &structpb.Value_NumberValue{
						NumberValue: m.DeltaRel.GetNumberValue(),
					},
				}
			}
		}
	}

	switch opts.Format {
	case runtimev1.ExportFormat_EXPORT_FORMAT_UNSPECIFIED:
		return fmt.Errorf("unspecified format")
	case runtimev1.ExportFormat_EXPORT_FORMAT_CSV:
		return WriteCSV(meta, data, w)
	case runtimev1.ExportFormat_EXPORT_FORMAT_XLSX:
		return WriteXLSX(meta, data, w)
	case runtimev1.ExportFormat_EXPORT_FORMAT_PARQUET:
		return WriteParquet(meta, data, w)
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

// TODO: a) Ensure correct time zone handling, b) Implement support for tr.RoundToGrain
// (Maybe consider pushing all this logic into the SQL instead?)
func timeRangeClause(tr *runtimev1.TimeRange, mv *runtimev1.MetricsViewSpec, timeCol string, args *[]any) (string, error) {
	var clause string
	if isTimeRangeNil(tr) {
		return clause, nil
	}

	start, end, err := ResolveTimeRange(tr, mv)
	if err != nil {
		return "", err
	}

	if !start.IsZero() {
		clause += fmt.Sprintf(" AND %s >= ?", timeCol)
		*args = append(*args, start)
	}

	if !end.IsZero() {
		clause += fmt.Sprintf(" AND %s < ?", timeCol)
		*args = append(*args, end)
	}

	return clause, nil
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

func (q *MetricsViewComparison) isDeltaComparison() bool {
	return q.Sort[0].SortType == runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_ABS_DELTA ||
		q.Sort[0].SortType == runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_REL_DELTA
}

func (q *MetricsViewComparison) isBase() bool {
	return q.Sort[0].SortType == runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_BASE_VALUE
}
