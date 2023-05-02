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
	"google.golang.org/protobuf/types/known/timestamppb"
)

type MetricsViewComparisonToplist struct {
	MetricsViewName     string                                 `json:"metrics_view_name,omitempty"`
	DimensionName       string                                 `json:"dimension_name,omitempty"`
	MeasureNames        []string                               `json:"measure_names,omitempty"`
	BaseTimeStart       *timestamppb.Timestamp                 `json:"base_time_start,omitempty"`
	BaseTimeEnd         *timestamppb.Timestamp                 `json:"base_time_end,omitempty"`
	ComparisonTimeStart *timestamppb.Timestamp                 `json:"comparison_time_start,omitempty"`
	ComparisonTimeEnd   *timestamppb.Timestamp                 `json:"comparison_time_end,omitempty"`
	Limit               int64                                  `json:"limit,omitempty"`
	Offset              int64                                  `json:"offset,omitempty"`
	Sort                []*runtimev1.MetricsViewComparisonSort `json:"sort,omitempty"`
	Filter              *runtimev1.MetricsViewFilter           `json:"filter,omitempty"`

	Result *runtimev1.MetricsViewCompareToplistResponse `json:"-"`
}

var _ runtime.Query = &MetricsViewComparisonToplist{}

func (q *MetricsViewComparisonToplist) Key() string {
	r, err := json.Marshal(q)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("MetricsViewCompareToplist:%s", string(r))
}

func (q *MetricsViewComparisonToplist) Deps() []string {
	return []string{q.MetricsViewName}
}

func (q *MetricsViewComparisonToplist) MarshalResult() any {
	return q.Result
}

func (q *MetricsViewComparisonToplist) UnmarshalResult(v any) error {
	res, ok := v.(*runtimev1.MetricsViewCompareToplistResponse)
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

	if olap.Dialect() != drivers.DialectDuckDB {
		return fmt.Errorf("not available for dialect '%s'", olap.Dialect())
	}

	mv, err := lookupMetricsView(ctx, rt, instanceID, q.MetricsViewName)
	if err != nil {
		return err
	}

	if mv.TimeDimension == "" {
		return fmt.Errorf("metrics view '%s' does not have a time dimension", q.MetricsViewName)
	}
	if q.BaseTimeStart == nil || q.BaseTimeEnd == nil || q.ComparisonTimeStart == nil || q.ComparisonTimeEnd == nil {
		return fmt.Errorf("undefined time range for comparison on '%s' metrics view ", q.MetricsViewName)
	}

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
		data = append(data, &runtimev1.MetricsViewComparisonRow{
			DimensionName:  q.DimensionName,
			DimensionValue: dv,
			MeasureValues:  measureValues,
		})
	}

	q.Result = &runtimev1.MetricsViewCompareToplistResponse{
		Data: data,
	}

	return nil
}

func timeRangeClause(start *timestamppb.Timestamp, end *timestamppb.Timestamp, td string, args *[]any) string {
	var clause string
	clause += fmt.Sprintf(" AND %s >= ?", td)
	*args = append(*args, start.AsTime())

	clause += fmt.Sprintf(" AND %s < ?", td)
	*args = append(*args, end.AsTime())

	return clause
}

func (q *MetricsViewComparisonToplist) buildMetricsTopListSQL(mv *runtimev1.MetricsView, dialect drivers.Dialect) (string, []any, error) {
	dimName := safeName(q.DimensionName)
	selectCols := []string{dimName}
	finalSelectCols := []string{}
	measureMap := make(map[string]int)
	for i, n := range q.MeasureNames {
		measureMap[n] = i
		found := false
		for _, m := range mv.Measures {
			if m.Name == n {
				expr := fmt.Sprintf(`%s as "%s"`, m.Expression, m.Name)
				selectCols = append(selectCols, expr)
				finalSelectCols = append(
					finalSelectCols,
					fmt.Sprintf(
						"base.%[1]s, comparison.%[1]s, comparison.%[1]s - base.%[1]s, (comparison.%[1]s - base.%[1]s)/base.%[1]s::DOUBLE",
						m.Name,
					),
				)
				found = true
				break
			}
		}
		if !found {
			return "", nil, fmt.Errorf("measure does not exist: '%s'", n)
		}
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

	baseWhereClause += timeRangeClause(q.BaseTimeStart, q.BaseTimeEnd, td, &args)
	if q.Filter != nil {
		clause, clauseArgs, err := buildFilterClauseForMetricsViewFilter(q.Filter, dialect)
		if err != nil {
			return "", nil, err
		}
		baseWhereClause += " " + clause

		args = append(args, clauseArgs...)
	}

	comparisonWhereClause += timeRangeClause(q.ComparisonTimeStart, q.ComparisonTimeEnd, td, &args)
	if q.Filter != nil {
		clause, clauseArgs, err := buildFilterClauseForMetricsViewFilter(q.Filter, dialect)
		if err != nil {
			return "", nil, err
		}
		comparisonWhereClause += " " + clause

		args = append(args, clauseArgs...)
	}

	orderClause := "true"
	for _, s := range q.Sort {
		i := measureMap[s.MeasureName]
		orderClause += ", "
		var pos int
		switch s.Type {
		case runtimev1.ComparisonSortType_COMPARISON_SORT_TYPE_BASE_VALUE:
			pos = 2 + i*4
		case runtimev1.ComparisonSortType_COMPARISON_SORT_TYPE_COMPARISON_VALUE:
			pos = 3 + i*4
		case runtimev1.ComparisonSortType_COMPARISON_SORT_TYPE_DELTA:
			pos = 4 + i*4
		default:
			return "", nil, fmt.Errorf("undefined sort type for measure %s", s.MeasureName)
		}
		orderClause += fmt.Sprint(pos)
		if !s.Ascending {
			orderClause += " DESC"
		}
		orderClause += " NULLS LAST"
	}

	if q.Limit == 0 {
		q.Limit = 100
	}

	sql := fmt.Sprintf(`
		SELECT COALESCE(base.%[2]s, comparison.%[2]s), %[8]s FROM 
			(
				SELECT %[1]s, %[2]s FROM %[3]q WHERE %[4]s GROUP BY %[2]s
			) base
		FULL JOIN
			(
				SELECT %[1]s, %[2]s FROM %[3]q WHERE %[5]s GROUP BY %[2]s
			) comparison
		ON
				base.%[2]s = comparison.%[2]s
		ORDER BY
			%[6]s
		LIMIT
			%[7]d
		`,
		subSelectClause,       // 1
		dimName,               // 2
		mv.Model,              // 3
		baseWhereClause,       // 4
		comparisonWhereClause, // 5
		orderClause,           // 6
		q.Limit,               // 7
		finalSelectClause,     // 8
	)
	fmt.Println("sql " + sql)

	return sql, args, nil
}
