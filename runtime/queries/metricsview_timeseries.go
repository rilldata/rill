package queries

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/pbutil"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type MetricsViewTimeSeries struct {
	MetricsViewName string                       `json:"metrics_view_name,omitempty"`
	MeasureNames    []string                     `json:"measure_names,omitempty"`
	InlineMeasures  []*runtimev1.InlineMeasure   `json:"inline_measures,omitempty"`
	TimeStart       *timestamppb.Timestamp       `json:"time_start,omitempty"`
	TimeEnd         *timestamppb.Timestamp       `json:"time_end,omitempty"`
	Limit           int64                        `json:"limit,omitempty"`
	Offset          int64                        `json:"offset,omitempty"`
	Sort            []*runtimev1.MetricsViewSort `json:"sort,omitempty"`
	Filter          *runtimev1.MetricsViewFilter `json:"filter,omitempty"`
	TimeGranularity runtimev1.TimeGrain          `json:"time_granularity,omitempty"`

	Result *runtimev1.MetricsViewTimeSeriesResponse `json:"-"`
}

var _ runtime.Query = &MetricsViewTimeSeries{}

func (q *MetricsViewTimeSeries) Key() string {
	r, err := json.Marshal(q)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("MetricsViewTimeSeries:%s", string(r))
}

func (q *MetricsViewTimeSeries) Deps() []string {
	return []string{q.MetricsViewName}
}

func (q *MetricsViewTimeSeries) MarshalResult() *runtime.QueryResult {
	return &runtime.QueryResult{
		Value: q.Result,
		Bytes: sizeProtoMessage(q.Result),
	}
}

func (q *MetricsViewTimeSeries) UnmarshalResult(v any) error {
	res, ok := v.(*runtimev1.MetricsViewTimeSeriesResponse)
	if !ok {
		return fmt.Errorf("MetricsViewTimeSeries: mismatched unmarshal input")
	}
	q.Result = res
	return nil
}

func (q *MetricsViewTimeSeries) Resolve(ctx context.Context, rt *runtime.Runtime, instanceID string, priority int) error {
	mv, err := lookupMetricsView(ctx, rt, instanceID, q.MetricsViewName)
	if err != nil {
		return err
	}

	if mv.TimeDimension == "" {
		return fmt.Errorf("metrics view '%s' does not have a time dimension", q.MetricsViewName)
	}

	olap, err := rt.OLAP(ctx, instanceID)
	if err != nil {
		return err
	}

	switch olap.Dialect() {
	case drivers.DialectDuckDB:
		return q.resolveDuckDB(ctx, rt, instanceID, mv, priority)
	case drivers.DialectDruid:
		return q.resolveDruid(ctx, olap, mv, priority)
	default:
		return fmt.Errorf("not available for dialect '%s'", olap.Dialect())
	}
}

func (q *MetricsViewTimeSeries) resolveDuckDB(ctx context.Context, rt *runtime.Runtime, instanceID string, mv *runtimev1.MetricsView, priority int) error {
	ms, err := resolveMeasures(mv, q.InlineMeasures, q.MeasureNames)
	if err != nil {
		return err
	}

	measures, err := toColumnTimeseriesMeasures(ms)
	if err != nil {
		return err
	}

	tsq := &ColumnTimeseries{
		TableName:           mv.Model,
		TimestampColumnName: mv.TimeDimension,
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Start:    q.TimeStart,
			End:      q.TimeEnd,
			Interval: q.TimeGranularity,
		},
		Measures:          measures,
		MetricsView:       mv,
		MetricsViewFilter: q.Filter,
	}
	err = rt.Query(ctx, instanceID, tsq, priority)
	if err != nil {
		return err
	}

	r := tsq.Result

	q.Result = &runtimev1.MetricsViewTimeSeriesResponse{
		Meta: r.Meta,
		Data: r.Results,
	}

	return nil
}

func toColumnTimeseriesMeasures(measures []*runtimev1.MetricsView_Measure) ([]*runtimev1.ColumnTimeSeriesRequest_BasicMeasure, error) {
	res := make([]*runtimev1.ColumnTimeSeriesRequest_BasicMeasure, len(measures))
	for i, m := range measures {
		res[i] = &runtimev1.ColumnTimeSeriesRequest_BasicMeasure{
			SqlName:    m.Name,
			Expression: m.Expression,
		}
	}
	return res, nil
}

func (q *MetricsViewTimeSeries) resolveDruid(ctx context.Context, olap drivers.OLAPStore, mv *runtimev1.MetricsView, priority int) error {
	sql, tsAlias, args, err := q.buildDruidMetricsTimeseriesSQL(mv)
	if err != nil {
		return fmt.Errorf("error building query: %w", err)
	}

	rows, err := olap.Execute(ctx, &drivers.Statement{
		Query:            sql,
		Args:             args,
		Priority:         priority,
		ExecutionTimeout: defaultExecutionTimeout,
	})
	if err != nil {
		return err
	}
	defer rows.Close()

	var data []*runtimev1.TimeSeriesValue
	for rows.Next() {
		rowMap := make(map[string]any)
		err := rows.MapScan(rowMap)
		if err != nil {
			return err
		}

		var t time.Time
		switch v := rowMap[tsAlias].(type) {
		case time.Time:
			t = v
		default:
			panic(fmt.Sprintf("unexpected type for timestamp column: %T", v))
		}

		delete(rowMap, tsAlias)
		records, err := pbutil.ToStruct(rowMap)
		if err != nil {
			return err
		}

		data = append(data, &runtimev1.TimeSeriesValue{
			Ts:      timestamppb.New(t),
			Records: records,
		})
	}

	meta := structTypeToMetricsViewColumn(rows.Schema)

	q.Result = &runtimev1.MetricsViewTimeSeriesResponse{
		Meta: meta,
		Data: data,
	}

	return nil
}

func (q *MetricsViewTimeSeries) buildDruidMetricsTimeseriesSQL(mv *runtimev1.MetricsView) (string, string, []any, error) {
	ms, err := resolveMeasures(mv, q.InlineMeasures, q.MeasureNames)
	if err != nil {
		return "", "", nil, err
	}

	selectCols := []string{}
	for _, m := range ms {
		expr := fmt.Sprintf(`%s as "%s"`, m.Expression, m.Name)
		selectCols = append(selectCols, expr)
	}

	whereClause := "1=1"
	args := []any{}
	if q.TimeStart != nil {
		whereClause += fmt.Sprintf(" AND %s >= ?", safeName(mv.TimeDimension))
		args = append(args, q.TimeStart.AsTime())
	}
	if q.TimeEnd != nil {
		whereClause += fmt.Sprintf(" AND %s < ?", safeName(mv.TimeDimension))
		args = append(args, q.TimeEnd.AsTime())
	}

	if q.Filter != nil {
		clause, clauseArgs, err := buildFilterClauseForMetricsViewFilter(mv, q.Filter, drivers.DialectDruid)
		if err != nil {
			return "", "", nil, err
		}
		whereClause += " " + clause
		args = append(args, clauseArgs...)
	}

	tsAlias := tempName("_ts_")
	tsSpecifier := convertToDateTruncSpecifier(q.TimeGranularity)

	sql := fmt.Sprintf(
		`SELECT date_trunc('%s', %s) AS %s, %s FROM %q WHERE %s GROUP BY 1 ORDER BY 1`,
		tsSpecifier,
		safeName(mv.TimeDimension),
		tsAlias,
		strings.Join(selectCols, ", "),
		mv.Model,
		whereClause,
	)

	return sql, tsAlias, args, nil
}
