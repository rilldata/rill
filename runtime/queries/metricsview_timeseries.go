package queries

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type MetricsViewTimeSeries struct {
	MetricsViewName string                       `json:"metrics_view_name,omitempty"`
	MeasureNames    []string                     `json:"measure_names,omitempty"`
	TimeStart       *timestamppb.Timestamp       `json:"time_start,omitempty"`
	TimeEnd         *timestamppb.Timestamp       `json:"time_end,omitempty"`
	Limit           int64                        `json:"limit,omitempty"`
	Offset          int64                        `json:"offset,omitempty"`
	Sort            []*runtimev1.MetricsViewSort `json:"sort,omitempty"`
	Filter          *runtimev1.MetricsViewFilter `json:"filter,omitempty"`
	TimeGranularity string                       `json:"time_granularity,omitempty"`

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

func (q *MetricsViewTimeSeries) MarshalResult() any {
	return q.Result
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

	measures, err := toMeasures(mv.Measures, q.MeasureNames)
	tsq := &ColumnTimeseries{
		TableName:           mv.Model,
		TimestampColumnName: mv.TimeDimension,
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Start:    q.TimeStart,
			End:      q.TimeEnd,
			Interval: runtimev1.TimeGrain(runtimev1.TimeGrain_value["TIME_GRAIN_"+strings.ToUpper(q.TimeGranularity)]),
		},
		Measures: measures,
		Filters:  toFilter(q.Filter),
	}
	err = rt.Query(ctx, instanceID, tsq, priority)
	if err != nil {
		return err
	}

	a := tsq.Result

	q.Result = &runtimev1.MetricsViewTimeSeriesResponse{
		Meta: a.Meta,
		Data: toData(a.Data.Results, mv),
	}

	return nil
}

func toData(data []*runtimev1.TimeSeriesValue, mv *runtimev1.MetricsView) []*structpb.Struct {
	var res []*structpb.Struct

	for _, d := range data {
		m := map[string]*structpb.Value{}
		m[mv.TimeDimension] = structpb.NewStringValue(d.Ts)
		for n, r := range d.Records {
			m[n] = structpb.NewNumberValue(r)
		}
		res = append(res, &structpb.Struct{
			Fields: m,
		})
	}
	return res
}

func toFilter(filter *runtimev1.MetricsViewFilter) *runtimev1.MetricsViewRequestFilter {
	res := &runtimev1.MetricsViewRequestFilter{}
	for _, f := range filter.Include {
		var likes []string
		for _, l := range f.Like {
			likes = append(likes, l.GetStringValue())
		}
		res.Include = append(res.Include, &runtimev1.MetricsViewDimensionValue{
			Name: f.Name,
			In:   f.In,
			Like: likes,
		})
	}
	for _, f := range filter.Exclude {
		var likes []string
		for _, l := range f.Like {
			likes = append(likes, l.GetStringValue())
		}
		res.Exclude = append(res.Exclude, &runtimev1.MetricsViewDimensionValue{
			Name: f.Name,
			In:   f.In,
			Like: likes,
		})
	}

	return res
}

func toMeasures(measures []*runtimev1.MetricsView_Measure, measureNames []string) ([]*runtimev1.GenerateTimeSeriesRequest_BasicMeasure, error) {
	var res []*runtimev1.GenerateTimeSeriesRequest_BasicMeasure
	for _, n := range measureNames {
		found := false
		for _, m := range measures {
			if m.Name == n {
				res = append(res, &runtimev1.GenerateTimeSeriesRequest_BasicMeasure{
					SqlName:    m.Name,
					Expression: m.Expression,
				})
				found = true
			}
		}
		if !found {
			return nil, fmt.Errorf("measure does not exist: '%s'", n)
		}
	}
	return res, nil
}
