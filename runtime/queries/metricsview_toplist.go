package queries

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type MetricsViewToplist struct {
	MetricsViewName string                       `json:"metrics_view_name,omitempty"`
	DimensionName   string                       `json:"dimension_name,omitempty"`
	MeasureNames    []string                     `json:"measure_names,omitempty"`
	TimeStart       *timestamppb.Timestamp       `json:"time_start,omitempty"`
	TimeEnd         *timestamppb.Timestamp       `json:"time_end,omitempty"`
	Limit           int64                        `json:"limit,omitempty"`
	Offset          int64                        `json:"offset,omitempty"`
	Sort            []*runtimev1.MetricsViewSort `json:"sort,omitempty"`
	Filter          *runtimev1.MetricsViewFilter `json:"filter,omitempty"`

	Result *runtimev1.MetricsViewToplistResponse `json:"-"`
}

var _ runtime.Query = &MetricsViewToplist{}

func (q *MetricsViewToplist) Key() string {
	r, err := json.Marshal(q)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("MetricsViewToplist:%s", string(r))
}

func (q *MetricsViewToplist) Deps() []string {
	return []string{q.MetricsViewName}
}

func (q *MetricsViewToplist) MarshalResult() any {
	return q.Result
}

func (q *MetricsViewToplist) UnmarshalResult(v any) error {
	res, ok := v.(*runtimev1.MetricsViewToplistResponse)
	if !ok {
		return fmt.Errorf("MetricsViewToplist: mismatched unmarshal input")
	}
	q.Result = res
	return nil
}

func (q *MetricsViewToplist) Resolve(ctx context.Context, rt *runtime.Runtime, instanceID string, priority int) error {
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

	// Build query
	sql, args, err := q.buildMetricsTopListSQL(mv)
	if err != nil {
		return fmt.Errorf("error building query: %w", err)
	}

	// Execute
	meta, data, err := metricsQuery(ctx, olap, priority, sql, args)
	if err != nil {
		return err
	}

	q.Result = &runtimev1.MetricsViewToplistResponse{
		Meta: meta,
		Data: data,
	}

	return nil
}

func (q *MetricsViewToplist) buildMetricsTopListSQL(mv *runtimev1.MetricsView) (string, []any, error) {
	dimName := safeName(q.DimensionName)
	selectCols := []string{dimName}
	for _, n := range q.MeasureNames {
		found := false
		for _, m := range mv.Measures {
			if m.Name == n {
				expr := fmt.Sprintf(`%s as "%s"`, m.Expression, m.Name)
				selectCols = append(selectCols, expr)
				found = true
				break
			}
		}
		if !found {
			return "", nil, fmt.Errorf("measure does not exist: '%s'", n)
		}
	}

	args := []any{}
	whereClause := "1=1"
	timestampColumnName := safeName(mv.TimeDimension)
	if mv.TimeDimension != "" {
		if q.TimeStart != nil {
			whereClause += fmt.Sprintf(" AND %s >= ?", timestampColumnName)
			args = append(args, q.TimeStart.AsTime())
		}
		if q.TimeEnd != nil {
			whereClause += fmt.Sprintf(" AND %s < ?", timestampColumnName)
			args = append(args, q.TimeEnd.AsTime())
		}
	}

	if q.Filter != nil {
		clause, clauseArgs, err := buildFilterClauseForMetricsViewFilter(q.Filter)
		if err != nil {
			return "", nil, err
		}
		whereClause += " " + clause
		args = append(args, clauseArgs...)
	}

	orderClause := "true"
	for _, s := range q.Sort {
		orderClause += ", "
		orderClause += safeName(s.Name)
		if !s.Ascending {
			orderClause += " DESC"
		}
		orderClause += " NULLS LAST"
	}

	if q.Limit == 0 {
		q.Limit = 100
	}

	sql := fmt.Sprintf("SELECT %s FROM %s WHERE %s GROUP BY %s ORDER BY %s LIMIT %d",
		strings.Join(selectCols, ", "),
		mv.Model,
		whereClause,
		dimName,
		orderClause,
		q.Limit,
	)

	return sql, args, nil
}
