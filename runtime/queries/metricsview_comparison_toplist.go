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
	"github.com/rilldata/rill/runtime/pkg/pbutil"
	"google.golang.org/protobuf/types/known/structpb"

	// Load IANA time zone data
	_ "time/tzdata"
)

type MetricsViewComparison struct {
	MetricsViewName     string                                     `json:"metrics_view_name,omitempty"`
	DimensionName       string                                     `json:"dimension_name,omitempty"`
	Measures            []*runtimev1.MetricsViewAggregationMeasure `json:"measures,omitempty"`
	TimeRange           *runtimev1.TimeRange                       `json:"base_time_range,omitempty"`
	ComparisonTimeRange *runtimev1.TimeRange                       `json:"comparison_time_range,omitempty"`
	Limit               int64                                      `json:"limit,omitempty"`
	Offset              int64                                      `json:"offset,omitempty"`
	Sort                []*runtimev1.MetricsViewComparisonSort     `json:"sort,omitempty"`
	Filter              *runtimev1.MetricsViewFilter               `json:"filter,omitempty"`
	MetricsView         *runtimev1.MetricsViewSpec                 `json:"-"`
	ResolvedMVSecurity  *runtime.ResolvedMetricsViewSecurity       `json:"security"`
	Exact               bool                                       `json:"exact"`

	Result *runtimev1.MetricsViewComparisonResponse `json:"-"`
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
	olap, release, err := rt.OLAP(ctx, instanceID)
	if err != nil {
		return err
	}
	defer release()

	if olap.Dialect() != drivers.DialectDuckDB && olap.Dialect() != drivers.DialectDruid {
		return fmt.Errorf("not available for dialect '%s'", olap.Dialect())
	}

	if q.MetricsView.TimeDimension == "" && (!isTimeRangeNil(q.TimeRange) || !isTimeRangeNil(q.ComparisonTimeRange)) {
		return fmt.Errorf("metrics view '%s' does not have a time dimension", q.MetricsViewName)
	}

	err = validateSort(q.Sort)
	if err != nil {
		return err
	}

	if !isTimeRangeNil(q.ComparisonTimeRange) {
		return q.executeComparisonToplist(ctx, olap, q.MetricsView, priority, q.ResolvedMVSecurity)
	}

	return q.executeToplist(ctx, olap, q.MetricsView, priority, q.ResolvedMVSecurity)
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
		measureValues := []*runtimev1.MetricsViewComparisonValue{}

		for i, m := range q.Measures {
			bv, err := pbutil.ToValue(values[1+i*4], safeFieldType(rows.Schema, 1+i*4))
			if err != nil {
				return err
			}

			cv, err := pbutil.ToValue(values[2+i*4], safeFieldType(rows.Schema, 2+i*4))
			if err != nil {
				return err
			}

			da, err := pbutil.ToValue(values[3+i*4], safeFieldType(rows.Schema, 3+i*4))
			if err != nil {
				return err
			}

			dr, err := pbutil.ToValue(values[4+i*4], safeFieldType(rows.Schema, 4+i*4))
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
	rawColName := metricsViewDimensionColumn(dim)
	colName := safeName(rawColName)
	unnestColName := safeName(tempName(fmt.Sprintf("%s_%s_", "unnested", rawColName)))

	labelMap := make(map[string]string, len(mv.Measures))
	for _, m := range mv.Measures {
		labelMap[m.Name] = m.Name
		if m.Label != "" {
			labelMap[m.Name] = m.Label
		}
	}

	var labelCols []string
	var selectCols []string
	unnestClause := ""
	dimLabel := colName
	if dim.Label != "" {
		dimLabel = safeName(dim.Label)
	}
	if dim.Unnest && dialect != drivers.DialectDruid {
		// select "unnested_colName" as "colName" ... FROM "mv_table", LATERAL UNNEST("mv_table"."colName") tbl("unnested_colName") ...
		selectCols = append(selectCols, fmt.Sprintf(`%s as %s`, unnestColName, colName))
		unnestClause = fmt.Sprintf(`, LATERAL UNNEST(%s.%s) tbl(%s)`, safeName(mv.Table), colName, unnestColName)
	} else {
		selectCols = append(selectCols, colName)
	}
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

	trc, err := timeRangeClause(q.TimeRange, mv, dialect, td, &args)
	if err != nil {
		return "", nil, err
	}
	baseWhereClause += trc

	filterClause, filterClauseArgs, err := buildFilterClauseForMetricsViewFilter(mv, q.Filter, dialect, policy)
	if err != nil {
		return "", nil, err
	}
	if filterClause != "" {
		baseWhereClause += " " + filterClause
		args = append(args, filterClauseArgs...)
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

	groupByCol := colName
	if dim.Unnest && dialect != drivers.DialectDruid {
		groupByCol = unnestColName
	}

	var sql string
	if export {
		labelSelectClause := strings.Join(labelCols, ", ")
		sql = fmt.Sprintf(
			`SELECT %[9]s FROM (SELECT %[1]s FROM %[3]s %[8]s WHERE %[4]s GROUP BY %[2]s %[5]s %[6]s OFFSET %[7]d)`,
			selectClause,       // 1
			groupByCol,         // 2
			safeName(mv.Table), // 3
			baseWhereClause,    // 4
			orderByClause,      // 5
			limitClause,        // 6
			q.Offset,           // 7
			unnestClause,       // 8
			labelSelectClause,  // 9
		)
	} else {
		sql = fmt.Sprintf(
			`SELECT %[1]s FROM %[3]s %[8]s WHERE %[4]s GROUP BY %[2]s %[5]s %[6]s OFFSET %[7]d`,
			selectClause,       // 1
			groupByCol,         // 2
			safeName(mv.Table), // 3
			baseWhereClause,    // 4
			orderByClause,      // 5
			limitClause,        // 6
			q.Offset,           // 7
			unnestClause,       // 8
		)
	}

	return sql, args, nil
}

func (q *MetricsViewComparison) buildMetricsComparisonTopListSQL(mv *runtimev1.MetricsViewSpec, dialect drivers.Dialect, policy *runtime.ResolvedMetricsViewSecurity, export bool) (string, []any, error) {
	dim, err := metricsViewDimension(mv, q.DimensionName)
	if err != nil {
		return "", nil, err
	}

	rawColName := metricsViewDimensionColumn(dim)
	colName := safeName(rawColName)
	unnestColName := safeName(tempName(fmt.Sprintf("%s_%s_", "unnested", rawColName)))

	labelMap := make(map[string]string, len(mv.Measures))
	for _, m := range mv.Measures {
		labelMap[m.Name] = m.Name
		if m.Label != "" {
			labelMap[m.Name] = m.Label
		}
	}

	var selectCols []string
	unnestClause := ""
	if dim.Unnest && dialect != drivers.DialectDruid {
		// select "unnested_colName" as "colName" ... FROM "mv_table", LATERAL UNNEST("mv_table"."colName") tbl("unnested_colName") ...
		selectCols = append(selectCols, fmt.Sprintf(`%s as %s`, unnestColName, colName))
		unnestClause = fmt.Sprintf(`, LATERAL UNNEST(%s.%s) tbl(%s)`, safeName(mv.Table), colName, unnestColName)
	} else {
		selectCols = append(selectCols, colName)
	}

	for _, m := range q.Measures {
		switch m.BuiltinMeasure {
		case runtimev1.BuiltinMeasure_BUILTIN_MEASURE_UNSPECIFIED:
			expr, err := metricsViewMeasureExpression(mv, m.Name)
			if err != nil {
				return "", nil, err
			}
			selectCols = append(selectCols, fmt.Sprintf("%s as %s", expr, safeName(m.Name)))
		case runtimev1.BuiltinMeasure_BUILTIN_MEASURE_COUNT:
			selectCols = append(selectCols, fmt.Sprintf("COUNT(*) as %s", safeName(m.Name)))
		case runtimev1.BuiltinMeasure_BUILTIN_MEASURE_COUNT_DISTINCT:
			if len(m.BuiltinMeasureArgs) != 1 {
				return "", nil, fmt.Errorf("builtin measure '%s' expects 1 argument", m.BuiltinMeasure.String())
			}
			arg := m.BuiltinMeasureArgs[0].GetStringValue()
			if arg == "" {
				return "", nil, fmt.Errorf("builtin measure '%s' expects non-empty string argument, got '%v'", m.BuiltinMeasure.String(), m.BuiltinMeasureArgs[0])
			}
			selectCols = append(selectCols, fmt.Sprintf("COUNT(DISTINCT %s) as %s", safeName(arg), safeName(m.Name)))
		default:
			return "", nil, fmt.Errorf("unknown builtin measure '%d'", m.BuiltinMeasure)
		}
	}

	finalSelectCols := []string{}
	labelCols := []string{}
	measureMap := make(map[string]int)
	for i, m := range q.Measures {
		measureMap[m.Name] = i
		var columnsTuple string
		var labelTuple string
		if dialect != drivers.DialectDruid {
			columnsTuple = fmt.Sprintf(
				"base.%[1]s, comparison.%[1]s AS %[2]s, base.%[1]s - comparison.%[1]s AS %[3]s, (base.%[1]s - comparison.%[1]s)/comparison.%[1]s::DOUBLE AS %[4]s",
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
			columnsTuple = fmt.Sprintf(
				"ANY_VALUE(base.%[1]s), ANY_VALUE(comparison.%[1]s), ANY_VALUE(base.%[1]s - comparison.%[1]s), ANY_VALUE(SAFE_DIVIDE(base.%[1]s - comparison.%[1]s, CAST(comparison.%[1]s AS DOUBLE)))",
				safeName(m.Name),
			)
			labelTuple = fmt.Sprintf(
				"ANY_VALUE(base.%[1]s) AS %[2]s, ANY_VALUE(comparison.%[1]s) AS %[3]s, ANY_VALUE(base.%[1]s - comparison.%[1]s) AS %[4]s, ANY_VALUE(SAFE_DIVIDE(base.%[1]s - comparison.%[1]s, CAST(comparison.%[1]s AS DOUBLE))) AS %[5]s",
				safeName(m.Name),
				safeName(labelMap[m.Name]),
				safeName(labelMap[m.Name]+" (prev)"),
				safeName(labelMap[m.Name]+" (Δ)"),
				safeName(labelMap[m.Name]+" (Δ%)"),
			)
		}
		finalSelectCols = append(
			finalSelectCols,
			columnsTuple,
		)
		labelCols = append(labelCols, labelTuple)
	}

	subSelectClause := strings.Join(selectCols, ", ")
	finalSelectClause := strings.Join(finalSelectCols, ", ")
	labelSelectClause := strings.Join(labelCols, ", ")
	if export {
		finalSelectClause = labelSelectClause
	}

	baseWhereClause := "1=1"
	comparisonWhereClause := "1=1"

	args := []any{}
	if mv.TimeDimension == "" {
		return "", nil, fmt.Errorf("metrics view '%s' doesn't have time dimension", q.MetricsViewName)
	}

	td := safeName(mv.TimeDimension)

	trc, err := timeRangeClause(q.TimeRange, mv, dialect, td, &args)
	if err != nil {
		return "", nil, err
	}
	baseWhereClause += trc

	trc, err = timeRangeClause(q.ComparisonTimeRange, mv, dialect, td, &args)
	if err != nil {
		return "", nil, err
	}
	comparisonWhereClause += trc

	filterClause, filterClauseArgs, err := buildFilterClauseForMetricsViewFilter(mv, q.Filter, dialect, policy)
	if err != nil {
		return "", nil, err
	}
	if filterClause != "" {
		baseWhereClause += " " + filterClause
		comparisonWhereClause += " " + filterClause
		args = append(args, filterClauseArgs...)
		args = append(args, filterClauseArgs...)
	}

	err = validateSort(q.Sort)
	if err != nil {
		return "", nil, err
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
		i, ok := measureMap[s.Name]
		if !ok {
			return "", nil, fmt.Errorf("metrics view '%s' doesn't contain '%s' sort column", q.MetricsViewName, s.Name)
		}

		var pos int
		switch s.Type {
		case runtimev1.MetricsViewComparisonSortType_METRICS_VIEW_COMPARISON_SORT_TYPE_BASE_VALUE:
			pos = 2 + i*4
		case runtimev1.MetricsViewComparisonSortType_METRICS_VIEW_COMPARISON_SORT_TYPE_COMPARISON_VALUE:
			pos = 3 + i*4
		case runtimev1.MetricsViewComparisonSortType_METRICS_VIEW_COMPARISON_SORT_TYPE_ABS_DELTA:
			pos = 4 + i*4
		case runtimev1.MetricsViewComparisonSortType_METRICS_VIEW_COMPARISON_SORT_TYPE_REL_DELTA:
			pos = 5 + i*4
		default:
			return "", nil, fmt.Errorf("undefined sort type for measure %s", s.Name)
		}
		orderClause := fmt.Sprint(pos)
		subQueryOrderClause := fmt.Sprint(i + 2) // 1-based + skip the first dim column
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
	twiceTheLimitClause := ""
	if q.Limit > 0 {
		limitClause = fmt.Sprintf(" LIMIT %d", q.Limit)
		twiceTheLimitClause = fmt.Sprintf(" LIMIT %d", q.Limit*2)
	} else if q.Limit == 0 {
		twiceTheLimitClause = fmt.Sprintf(" LIMIT %d", 100_000) // use Druid limit
	}

	baseLimitClause := ""
	comparisonLimitClause := ""

	joinType := "FULL"
	if !q.Exact {
		deltaComparison := q.Sort[0].Type == runtimev1.MetricsViewComparisonSortType_METRICS_VIEW_COMPARISON_SORT_TYPE_ABS_DELTA ||
			q.Sort[0].Type == runtimev1.MetricsViewComparisonSortType_METRICS_VIEW_COMPARISON_SORT_TYPE_REL_DELTA

		approximationLimit := q.Limit
		if q.Limit != 0 && q.Limit < 100 && deltaComparison {
			approximationLimit = 100
		}

		if q.Sort[0].Type == runtimev1.MetricsViewComparisonSortType_METRICS_VIEW_COMPARISON_SORT_TYPE_BASE_VALUE || deltaComparison {
			joinType = "LEFT OUTER"
			baseLimitClause = subQueryOrderByClause
			if approximationLimit > 0 {
				baseLimitClause += fmt.Sprintf(" LIMIT %d", approximationLimit)
			}
		} else if q.Sort[0].Type == runtimev1.MetricsViewComparisonSortType_METRICS_VIEW_COMPARISON_SORT_TYPE_COMPARISON_VALUE {
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
	groupByCol := colName
	if dim.Unnest && dialect != drivers.DialectDruid {
		groupByCol = unnestColName
	}

	finalDimName := safeName(q.DimensionName)
	if export && dim.Label != "" {
		finalDimName = safeName(dim.Label)
	}
	var sql string
	if dialect != drivers.DialectDruid {
		sql = fmt.Sprintf(`
		SELECT COALESCE(base.%[2]s, comparison.%[2]s) AS %[10]s, %[9]s FROM 
			(
				SELECT %[1]s FROM %[3]s %[14]s WHERE %[4]s GROUP BY %[15]s %[12]s 
			) base
		%[11]s JOIN
			(
				SELECT %[1]s FROM %[3]s %[14]s WHERE %[5]s GROUP BY %[15]s %[13]s 
			) comparison
		ON
				base.%[2]s = comparison.%[2]s OR (base.%[2]s is null and comparison.%[2]s is null)
		%[6]s
		%[7]s
		OFFSET
			%[8]d
		`,
			subSelectClause,       // 1
			colName,               // 2
			safeName(mv.Table),    // 3
			baseWhereClause,       // 4
			comparisonWhereClause, // 5
			orderByClause,         // 6
			limitClause,           // 7
			q.Offset,              // 8
			finalSelectClause,     // 9
			finalDimName,          // 10
			joinType,              // 11
			baseLimitClause,       // 12
			comparisonLimitClause, // 13
			unnestClause,          // 14
			groupByCol,            // 15
		)
	} else {
		/*
			Example of the SQL query:

			WITH base AS (
				SELECT "user",
					sum(added) measure
				FROM "wikipedia"
				WHERE 1=1
				GROUP BY "user"
				ORDER BY 2
				LIMIT 10000
				OFFSET 100
			),
			comparison AS (
				SELECT "user",
					sum(added) measure
				FROM "wikipedia"
				WHERE 1=1
					AND "user" IN (SELECT "user" FROM base)
				GROUP BY "user"
				ORDER BY 2
				LIMIT 10000
			)
			SELECT
				base."user",
				ANY_VALUE(base."measure"),
				ANY_VALUE(comparison."measure"),
				ANY_VALUE(base."measure" - comparison."measure"),
				ANY_VALUE(SAFE_DIVIDE(base."measure" - comparison."measure", CAST(comparison."measure" AS DOUBLE)))
			FROM base LEFT JOIN comparison
			ON base."user" = comparison."user"
			GROUP BY 1 -- Druid doesn't support ORDER BY non-time columns without GROUP BY
			ORDER BY 2
			LIMIT 100
			OFFSET 100

			Apache Druid requires that one part of the JOIN fits in memory, that can be achieved by pushing down the limit clause to a subquery (works only if the sorting is based entirely on a single subquery result)
		*/
		leftSubQueryAlias := "base"
		rightSubQueryAlias := "comparison"
		leftWhereClause := baseWhereClause
		rightWhereClause := comparisonWhereClause

		if q.Sort[0].Type == runtimev1.MetricsViewComparisonSortType_METRICS_VIEW_COMPARISON_SORT_TYPE_COMPARISON_VALUE {
			leftSubQueryAlias = "comparison"
			rightSubQueryAlias = "base"
			leftWhereClause = comparisonWhereClause
			rightWhereClause = baseWhereClause
		}

		sql = fmt.Sprintf(`
				WITH %[11]s AS (
					SELECT %[1]s FROM %[3]s WHERE %[4]s GROUP BY %[2]s %[13]s %[10]s OFFSET %[8]d
				), %[12]s AS (
					SELECT %[1]s FROM %[3]s WHERE %[5]s AND %[2]s IN (SELECT %[2]s FROM %[11]s) GROUP BY %[2]s %[10]s
				)
				SELECT %[11]s.%[2]s AS %[14]s, %[9]s FROM %[11]s LEFT JOIN %[12]s ON base.%[2]s = comparison.%[2]s
				GROUP BY 1
				%[6]s
				%[7]s
				OFFSET %[8]d
			`,
			subSelectClause,       // 1
			colName,               // 2
			safeName(mv.Table),    // 3
			leftWhereClause,       // 4
			rightWhereClause,      // 5
			orderByClause,         // 6
			limitClause,           // 7
			q.Offset,              // 8
			finalSelectClause,     // 9
			twiceTheLimitClause,   // 10
			leftSubQueryAlias,     // 11
			rightSubQueryAlias,    // 12
			subQueryOrderByClause, // 13
			finalDimName,          // 14
		)
	}

	return sql, args, nil
}

func (q *MetricsViewComparison) Export(ctx context.Context, rt *runtime.Runtime, instanceID string, w io.Writer, opts *runtime.ExportOptions) error {
	olap, release, err := rt.OLAP(ctx, instanceID)
	if err != nil {
		return err
	}
	defer release()

	switch olap.Dialect() {
	case drivers.DialectDuckDB:
		if opts.Format == runtimev1.ExportFormat_EXPORT_FORMAT_CSV || opts.Format == runtimev1.ExportFormat_EXPORT_FORMAT_PARQUET {
			var sql string
			var args []any
			if !isTimeRangeNil(q.ComparisonTimeRange) {
				sql, args, err = q.buildMetricsComparisonTopListSQL(q.MetricsView, olap.Dialect(), q.ResolvedMVSecurity, true)
				if err != nil {
					return fmt.Errorf("error building query: %w", err)
				}
			} else {
				sql, args, err = q.buildMetricsTopListSQL(q.MetricsView, olap.Dialect(), q.ResolvedMVSecurity, true)
				if err != nil {
					return fmt.Errorf("error building query: %w", err)
				}
			}

			filename := q.generateFilename()
			if err := duckDBCopyExport(ctx, w, opts, sql, args, filename, olap, opts.Format); err != nil {
				return err
			}
		} else {
			if err := q.generalExport(ctx, rt, instanceID, w, opts, q.MetricsView); err != nil {
				return err
			}
		}
	case drivers.DialectDruid:
		if err := q.generalExport(ctx, rt, instanceID, w, opts, q.MetricsView); err != nil {
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
	comparison := !isTimeRangeNil(q.ComparisonTimeRange)
	if comparison {
		metaLen = len(q.Result.Rows[0].MeasureValues) * 4
	} else {
		metaLen = len(q.Result.Rows[0].MeasureValues)
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
	}
	if comparison {
		for i, m := range q.Result.Rows[0].MeasureValues {
			meta[1+i*4] = &runtimev1.MetricsViewColumn{
				Name: labelMap[m.MeasureName],
			}
			meta[2+i*4] = &runtimev1.MetricsViewColumn{
				Name: fmt.Sprintf("%s (prev)", labelMap[m.MeasureName]),
			}
			meta[3+i*4] = &runtimev1.MetricsViewColumn{
				Name: fmt.Sprintf("%s (Δ)", labelMap[m.MeasureName]),
			}
			meta[4+i*4] = &runtimev1.MetricsViewColumn{
				Name: fmt.Sprintf("%s (Δ%%)", labelMap[m.MeasureName]),
			}
		}
	} else {
		for i, m := range q.Result.Rows[0].MeasureValues {
			meta[1+i] = &runtimev1.MetricsViewColumn{
				Name: labelMap[m.MeasureName],
			}
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
			if comparison {
				data[i].Fields[labelMap[m.MeasureName]] = &structpb.Value{
					Kind: &structpb.Value_NumberValue{
						NumberValue: m.BaseValue.GetNumberValue(),
					},
				}
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
			} else {
				data[i].Fields[labelMap[m.MeasureName]] = &structpb.Value{
					Kind: &structpb.Value_NumberValue{
						NumberValue: m.BaseValue.GetNumberValue(),
					},
				}
			}
		}
	}

	switch opts.Format {
	case runtimev1.ExportFormat_EXPORT_FORMAT_UNSPECIFIED:
		return fmt.Errorf("unspecified format")
	case runtimev1.ExportFormat_EXPORT_FORMAT_CSV:
		return writeCSV(meta, data, w)
	case runtimev1.ExportFormat_EXPORT_FORMAT_XLSX:
		return writeXLSX(meta, data, w)
	case runtimev1.ExportFormat_EXPORT_FORMAT_PARQUET:
		return writeParquet(meta, data, w)
	}

	return nil
}

func (q *MetricsViewComparison) generateFilename() string {
	filename := strings.ReplaceAll(q.MetricsViewName, `"`, `_`)
	filename += "_" + q.DimensionName
	if q.Filter != nil && (len(q.Filter.Include) > 0 || len(q.Filter.Exclude) > 0) {
		filename += "_filtered"
	}
	return filename
}

// TODO: a) Ensure correct time zone handling, b) Implement support for tr.RoundToGrain
// (Maybe consider pushing all this logic into the SQL instead?)
func timeRangeClause(tr *runtimev1.TimeRange, mv *runtimev1.MetricsViewSpec, dialect drivers.Dialect, timeCol string, args *[]any) (string, error) {
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

func validateSort(sorts []*runtimev1.MetricsViewComparisonSort) error {
	if len(sorts) == 0 {
		return fmt.Errorf("sorting is required")
	}
	firstSort := sorts[0].Type

	for _, s := range sorts {
		if firstSort != s.Type {
			return fmt.Errorf("diffirent sort types are not supported in a single query")
		}
	}
	return nil
}

func isTimeRangeNil(tr *runtimev1.TimeRange) bool {
	return tr == nil || (tr.Start == nil && tr.End == nil)
}
