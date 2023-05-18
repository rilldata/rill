package queries

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/pbutil"
)

type MetricsViewComparisonToplist struct {
	MetricsViewName     string                                 `json:"metrics_view_name,omitempty"`
	DimensionName       string                                 `json:"dimension_name,omitempty"`
	MeasureNames        []string                               `json:"measure_names,omitempty"`
	InlineMeasures      []*runtimev1.InlineMeasure             `json:"inline_measures,omitempty"`
	BaseTimeRange       *runtimev1.TimeRange                   `json:"base_time_range,omitempty"`
	ComparisonTimeRange *runtimev1.TimeRange                   `json:"comparison_time_range,omitempty"`
	Limit               int64                                  `json:"limit,omitempty"`
	Offset              int64                                  `json:"offset,omitempty"`
	Sort                []*runtimev1.MetricsViewComparisonSort `json:"sort,omitempty"`
	Filter              *runtimev1.MetricsViewFilter           `json:"filter,omitempty"`

	Result *runtimev1.MetricsViewComparisonToplistResponse `json:"-"`
}

var _ runtime.Query = &MetricsViewComparisonToplist{}

func (q *MetricsViewComparisonToplist) Key() string {
	r, err := json.Marshal(q)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("MetricsViewComparisonToplist:%s", string(r))
}

func (q *MetricsViewComparisonToplist) Deps() []string {
	return []string{q.MetricsViewName}
}

func (q *MetricsViewComparisonToplist) MarshalResult() any {
	return q.Result
}

func (q *MetricsViewComparisonToplist) UnmarshalResult(v any) error {
	res, ok := v.(*runtimev1.MetricsViewComparisonToplistResponse)
	if !ok {
		return fmt.Errorf("MetricsViewComparisonToplist: mismatched unmarshal input")
	}
	q.Result = res
	return nil
}

func (q *MetricsViewComparisonToplist) Resolve(ctx context.Context, rt *runtime.Runtime, instanceID string, priority int) error {
	olap, err := rt.OLAP(ctx, instanceID)
	if err != nil {
		return err
	}

	if olap.Dialect() != drivers.DialectDuckDB && olap.Dialect() != drivers.DialectDruid {
		return fmt.Errorf("not available for dialect '%s'", olap.Dialect())
	}

	mv, err := lookupMetricsView(ctx, rt, instanceID, q.MetricsViewName)
	if err != nil {
		return err
	}

	if mv.TimeDimension == "" && (q.BaseTimeRange != nil || q.ComparisonTimeRange != nil) {
		return fmt.Errorf("metrics view '%s' does not have a time dimension", q.MetricsViewName)
	}

	if q.ComparisonTimeRange != nil {
		return q.executeComparisonToplist(ctx, olap, mv, priority)
	}

	return q.executeToplist(ctx, olap, mv, priority)
}

func (q *MetricsViewComparisonToplist) executeToplist(ctx context.Context, olap drivers.OLAPStore, mv *runtimev1.MetricsView, priority int) error {
	sql, args, err := q.buildMetricsTopListSQL(mv, olap.Dialect())
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
		measureValues := make([]*runtimev1.MetricsViewComparisonValue, 0, len(q.MeasureNames))

		for i, name := range q.MeasureNames {
			v, err := pbutil.ToValue(values[1+i])
			if err != nil {
				return err
			}

			measureValues = append(measureValues, &runtimev1.MetricsViewComparisonValue{
				MeasureName: name,
				BaseValue:   v,
			})
		}

		dv, err := pbutil.ToValue(values[0])
		if err != nil {
			return err
		}

		data = append(data, &runtimev1.MetricsViewComparisonRow{
			DimensionValue: dv,
			MeasureValues:  measureValues,
		})
	}

	q.Result = &runtimev1.MetricsViewComparisonToplistResponse{
		Rows: data,
	}

	return nil
}

func (q *MetricsViewComparisonToplist) executeComparisonToplist(ctx context.Context, olap drivers.OLAPStore, mv *runtimev1.MetricsView, priority int) error {
	sql, args, err := q.buildMetricsComparisonTopListSQL(mv, olap.Dialect())
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

		for i, name := range q.MeasureNames {
			bv, err := pbutil.ToValue(values[1+i*4])
			if err != nil {
				return err
			}

			cv, err := pbutil.ToValue(values[2+i*4])
			if err != nil {
				return err
			}

			da, err := pbutil.ToValue(values[3+i*4])
			if err != nil {
				return err
			}

			dr, err := pbutil.ToValue(values[4+i*4])
			if err != nil {
				return err
			}

			measureValues = append(measureValues, &runtimev1.MetricsViewComparisonValue{
				MeasureName:     name,
				BaseValue:       bv,
				ComparisonValue: cv,
				DeltaAbs:        da,
				DeltaRel:        dr,
			})
		}

		dv, err := pbutil.ToValue(values[0])
		if err != nil {
			return err
		}

		data = append(data, &runtimev1.MetricsViewComparisonRow{
			DimensionValue: dv,
			MeasureValues:  measureValues,
		})
	}

	q.Result = &runtimev1.MetricsViewComparisonToplistResponse{
		Rows: data,
	}

	return nil
}

func timeRangeClause(timeRange *runtimev1.TimeRange, td string, args *[]any) string {
	var clause string
	if timeRange == nil {
		return clause
	}

	if timeRange.Start != nil {
		clause += fmt.Sprintf(" AND %s >= ?", td)
		*args = append(*args, timeRange.Start.AsTime())
	}

	if timeRange.End != nil {
		clause += fmt.Sprintf(" AND %s < ?", td)
		*args = append(*args, timeRange.End.AsTime())
	}

	return clause
}

func (q *MetricsViewComparisonToplist) buildMetricsTopListSQL(mv *runtimev1.MetricsView, dialect drivers.Dialect) (string, []any, error) {
	ms, err := resolveMeasures(mv, q.InlineMeasures, q.MeasureNames)
	if err != nil {
		return "", nil, err
	}

	dimName := safeName(q.DimensionName)
	selectCols := []string{dimName}
	for _, m := range ms {
		expr := fmt.Sprintf(`%s as %s`, m.Expression, safeName(m.Name))
		selectCols = append(selectCols, expr)
	}

	selectClause := strings.Join(selectCols, ", ")
	baseWhereClause := "1=1"

	args := []any{}
	td := safeName(mv.TimeDimension)

	baseWhereClause += timeRangeClause(q.BaseTimeRange, td, &args)
	if q.Filter != nil {
		clause, clauseArgs, err := buildFilterClauseForMetricsViewFilter(q.Filter, dialect)
		if err != nil {
			return "", nil, err
		}
		baseWhereClause += " " + clause

		args = append(args, clauseArgs...)
	}

	orderClause := "true"
	for _, s := range q.Sort {
		orderClause += ", "
		orderClause += safeName(s.MeasureName)
		if !s.Ascending {
			orderClause += " DESC"
		}
		if dialect == drivers.DialectDuckDB {
			orderClause += " NULLS LAST"
		}
	}

	if q.Limit == 0 {
		q.Limit = 100
	}

	sql := fmt.Sprintf(
		`SELECT %[1]s FROM %[3]q WHERE %[4]s GROUP BY %[2]s ORDER BY %[5]s LIMIT %[6]d OFFSET %[7]d`,
		selectClause,    // 1
		dimName,         // 2
		mv.Model,        // 3
		baseWhereClause, // 4
		orderClause,     // 5
		q.Limit,         // 6
		q.Offset,        // 7
	)

	return sql, args, nil
}

func (q *MetricsViewComparisonToplist) buildMetricsComparisonTopListSQL(mv *runtimev1.MetricsView, dialect drivers.Dialect) (string, []any, error) {
	ms, err := resolveMeasures(mv, q.InlineMeasures, q.MeasureNames)
	if err != nil {
		return "", nil, err
	}

	dimName := safeName(q.DimensionName)
	selectCols := []string{dimName}
	finalSelectCols := []string{}
	measureMap := make(map[string]int)
	for i, m := range ms {
		measureMap[m.Name] = i
		expr := fmt.Sprintf(`%s as %s`, m.Expression, safeName(m.Name))
		selectCols = append(selectCols, expr)
		var columnsTuple string
		if dialect != drivers.DialectDruid {
			columnsTuple = fmt.Sprintf(
				"base.%[1]s, comparison.%[1]s, comparison.%[1]s - base.%[1]s, (comparison.%[1]s - base.%[1]s)/base.%[1]s::DOUBLE",
				safeName(m.Name),
			)
		} else {
			columnsTuple = fmt.Sprintf(
				"ANY_VALUE(base.%[1]s), ANY_VALUE(comparison.%[1]s), ANY_VALUE(comparison.%[1]s - base.%[1]s), ANY_VALUE(SAFE_DIVIDE(comparison.%[1]s - base.%[1]s, CAST(base.%[1]s AS FLOAT))",
				safeName(m.Name),
			)
		}
		finalSelectCols = append(
			finalSelectCols,
			columnsTuple,
		)
	}

	subSelectClause := strings.Join(selectCols, ", ")
	finalSelectClause := strings.Join(finalSelectCols, ", ")

	baseWhereClause := "1=1"
	comparisonWhereClause := "1=1"

	args := []any{}
	if mv.TimeDimension == "" {
		return "", nil, fmt.Errorf("Metrics view '%s' doesn't have time dimension", mv.Name)
	}

	td := safeName(mv.TimeDimension)

	baseWhereClause += timeRangeClause(q.BaseTimeRange, td, &args)
	if q.Filter != nil {
		clause, clauseArgs, err := buildFilterClauseForMetricsViewFilter(q.Filter, dialect)
		if err != nil {
			return "", nil, err
		}
		baseWhereClause += " " + clause

		args = append(args, clauseArgs...)
	}

	comparisonWhereClause += timeRangeClause(q.ComparisonTimeRange, td, &args)
	if q.Filter != nil {
		clause, clauseArgs, err := buildFilterClauseForMetricsViewFilter(q.Filter, dialect)
		if err != nil {
			return "", nil, err
		}
		comparisonWhereClause += " " + clause

		args = append(args, clauseArgs...)
	}

	validateSort(q.Sort)
	orderClause := "true"
	subQueryOrderClause := "true"
	for _, s := range q.Sort {
		i, ok := measureMap[s.MeasureName]
		if !ok {
			return "", nil, fmt.Errorf("Metrics view '%s' doesn't contain '%s' sort column", q.MetricsViewName, s.MeasureName)
		}
		orderClause += ", "
		subQueryOrderClause += ", "
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
			return "", nil, fmt.Errorf("undefined sort type for measure %s", s.MeasureName)
		}
		orderClause += fmt.Sprint(pos)
		subQueryOrderClause += fmt.Sprint(i + 2)
		if !s.Ascending {
			orderClause += " DESC"
		}
		if dialect == drivers.DialectDuckDB {
			orderClause += " NULLS LAST"
		}
	}

	if q.Limit == 0 {
		q.Limit = 100
	}

	var sql string
	if dialect != drivers.DialectDruid {
		sql = fmt.Sprintf(`
		SELECT COALESCE(base.%[2]s, comparison.%[2]s), %[9]s FROM 
			(
				SELECT %[1]s FROM %[3]q WHERE %[4]s GROUP BY %[2]s
			) base
		FULL JOIN
			(
				SELECT %[1]s FROM %[3]q WHERE %[5]s GROUP BY %[2]s
			) comparison
		ON
				base.%[2]s = comparison.%[2]s
		ORDER BY
			%[6]s
		LIMIT
			%[7]d
		OFFSET
			%[8]d
		`,
			subSelectClause,       // 1
			dimName,               // 2
			mv.Model,              // 3
			baseWhereClause,       // 4
			comparisonWhereClause, // 5
			orderClause,           // 6
			q.Limit,               // 7
			q.Offset,              // 8
			finalSelectClause,     // 9
		)
	} else {
		/*
			Example of the SQL query:

			SELECT COALESCE(a."user", b."user"),
			       ANY_VALUE(a."measure"),
			       ANY_VALUE(b."measure"),
			       ANY_VALUE(a."measure" - b."measure"),
			       ANY_VALUE(SAFE_DIVIDE(a."measure" - b."measure",a."measure"))
			FROM
			  (SELECT "user",
			          sum(added)
			   FROM "wikipedia"
			   WHERE 1=1
			   GROUP BY "user"
			   ORDER BY 2
			   LIMIT 10) b
			LEFT OUTER JOIN
			  (SELECT "user",
			          sum(added)
			   FROM "wikipedia"
			   WHERE 1=1
			   GROUP BY 2
			   ORDER BY "measure"
			   LIMIT 10) a ON a."user" = b."user"
			GROUP BY 1
			ORDER BY 2
			LIMIT 10
		*/
		// Apache Druid requires that one part of the JOIN fits in memory, that can be achieved by pushing down the limit clause to a subquery (works only if the sorting is based entirely on a single subquery result)
		if q.Sort[0].Type == runtimev1.MetricsViewComparisonSortType_METRICS_VIEW_COMPARISON_SORT_TYPE_COMPARISON_VALUE || q.Sort[0].Type == runtimev1.MetricsViewComparisonSortType_METRICS_VIEW_COMPARISON_SORT_TYPE_BASE_VALUE {
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
				SELECT COALESCE(base.%[2]s, comparison.%[2]s), %[9]s FROM 
					(
						SELECT %[1]s FROM %[3]q WHERE %[4]s GROUP BY %[2]s ORDER BY %[13]s LIMIT %[10]d OFFSET %[8]d 
					) %[11]s
				LEFT OUTER JOIN
					(
						SELECT %[1]s FROM %[3]q WHERE %[5]s GROUP BY %[2]s
					) %[12]s
				ON
						base.%[2]s = comparison.%[2]s
				ORDER BY
					%[6]s
				LIMIT
					%[7]d
				OFFSET
					%[8]d
				`,
				subSelectClause,     // 1
				dimName,             // 2
				mv.Model,            // 3
				leftWhereClause,     // 4
				rightWhereClause,    // 5
				orderClause,         // 6
				q.Limit,             // 7
				q.Offset,            // 8
				finalSelectClause,   // 9
				q.Limit*2,           // 10
				leftSubQueryAlias,   // 11
				rightSubQueryAlias,  // 12
				subQueryOrderClause, // 13
			)
		} else {
			sql = fmt.Sprintf(`
				SELECT COALESCE(base.%[2]s, comparison.%[2]s), %[9]s FROM 
					(
						SELECT %[1]s FROM %[3]q WHERE %[4]s GROUP BY %[2]s
					) base
				FULL JOIN
					(
						SELECT %[1]s FROM %[3]q WHERE %[5]s GROUP BY %[2]s
					) comparison
				ON
						base.%[2]s = comparison.%[2]s
				ORDER BY
					%[6]s
				LIMIT
					%[7]d
				OFFSET
					%[8]d
				`,
				subSelectClause,       // 1
				dimName,               // 2
				mv.Model,              // 3
				baseWhereClause,       // 4
				comparisonWhereClause, // 5
				orderClause,           // 6
				q.Limit,               // 7
				q.Offset,              // 8
				finalSelectClause,     // 9
			)
		}
	}

	return sql, args, nil
}

func validateSort(sorts []*runtimev1.MetricsViewComparisonSort) error {
	if len(sorts) == 0 {
		return fmt.Errorf("Sorting is required")
	}
	firstSort := sorts[0]

	for _, s := range sorts {
		if firstSort != s {
			return fmt.Errorf("Diffirent sort types are not supported in a single query")
		}
	}
	return nil
}
