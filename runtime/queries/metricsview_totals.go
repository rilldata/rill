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

type MetricsViewTotals struct {
	MetricsViewName string                       `json:"metrics_view_name,omitempty"`
	MeasureNames    []string                     `json:"measure_names,omitempty"`
	InlineMeasures  []*runtimev1.InlineMeasure   `json:"inline_measures,omitempty"`
	TimeStart       *timestamppb.Timestamp       `json:"time_start,omitempty"`
	TimeEnd         *timestamppb.Timestamp       `json:"time_end,omitempty"`
	Filter          *runtimev1.MetricsViewFilter `json:"filter,omitempty"`

	Result *runtimev1.MetricsViewTotalsResponse `json:"-"`
}

var _ runtime.Query = &MetricsViewTotals{}

func (q *MetricsViewTotals) Key() string {
	r, err := json.Marshal(q)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("MetricsViewTotals:%s", string(r))
}

func (q *MetricsViewTotals) Deps() []string {
	return []string{q.MetricsViewName}
}

func (q *MetricsViewTotals) MarshalResult() *runtime.QueryResult {
	return &runtime.QueryResult{
		Value: q.Result,
		Bytes: sizeProtoMessage(q.Result),
	}
}

func (q *MetricsViewTotals) UnmarshalResult(v any) error {
	res, ok := v.(*runtimev1.MetricsViewTotalsResponse)
	if !ok {
		return fmt.Errorf("MetricsViewTotals: mismatched unmarshal input")
	}
	q.Result = res
	return nil
}

func (q *MetricsViewTotals) Resolve(ctx context.Context, rt *runtime.Runtime, instanceID string, priority int) error {
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

	if mv.TimeDimension == "" && (q.TimeStart != nil || q.TimeEnd != nil) {
		return fmt.Errorf("metrics view '%s' does not have a time dimension", q.MetricsViewName)
	}

	ql, args, err := q.buildMetricsTotalsSQL(mv, olap.Dialect())
	if err != nil {
		return fmt.Errorf("error building query: %w", err)
	}

	meta, data, err := metricsQuery(ctx, olap, priority, ql, args)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		return fmt.Errorf("no rows received from totals query")
	}

	q.Result = &runtimev1.MetricsViewTotalsResponse{
		Meta: meta,
		Data: data[0],
	}

	return nil
}

func (q *MetricsViewTotals) buildMetricsTotalsSQL(mv *runtimev1.MetricsView, dialect drivers.Dialect) (string, []any, error) {
	ms, err := resolveMeasures(mv, q.InlineMeasures, q.MeasureNames)
	if err != nil {
		return "", nil, err
	}

	selectCols := []string{}
	for _, m := range ms {
		expr := fmt.Sprintf(`%s as "%s"`, m.Expression, m.Name)
		selectCols = append(selectCols, expr)
	}

	whereClause := "1=1"
	args := []any{}
	if mv.TimeDimension != "" {
		if q.TimeStart != nil {
			whereClause += fmt.Sprintf(" AND %s >= ?", safeName(mv.TimeDimension))
			args = append(args, q.TimeStart.AsTime())
		}
		if q.TimeEnd != nil {
			whereClause += fmt.Sprintf(" AND %s < ?", safeName(mv.TimeDimension))
			args = append(args, q.TimeEnd.AsTime())
		}
	}

	if q.Filter != nil {
		clause, clauseArgs, err := buildFilterClauseForMetricsViewFilter(q.Filter, metricsViewDimensionNameMap(mv), dialect)
		if err != nil {
			return "", nil, err
		}
		whereClause += " " + clause
		args = append(args, clauseArgs...)
	}

	sql := fmt.Sprintf(
		"SELECT %s FROM %q WHERE %s",
		strings.Join(selectCols, ", "),
		mv.Model,
		whereClause,
	)
	return sql, args, nil
}
