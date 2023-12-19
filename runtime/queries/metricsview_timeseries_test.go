package queries_test

import (
	"context"
	// "fmt"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/expressionpb"
	"github.com/rilldata/rill/runtime/queries"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestMetricsViewsTimeseries_month_grain(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "timeseries")

	ctrl, err := rt.Controller(context.Background(), instanceID)
	require.NoError(t, err)
	r, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "timeseries_year"}, false)
	require.NoError(t, err)
	mv := r.GetMetricsView()

	q := &queries.MetricsViewTimeSeries{
		MeasureNames:    []string{"max_clicks"},
		MetricsViewName: "timeseries_year",
		MetricsView:     mv.Spec,
		TimeStart:       parseTime(t, "2023-01-01T00:00:00Z"),
		TimeEnd:         parseTime(t, "2024-01-01T00:00:00Z"),
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_MONTH,
		Limit:           250,
	}

	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	rows := q.Result.Data
	require.Len(t, rows, 12)
	i := 0
	require.Equal(t, parseTime(t, "2023-01-01T00:00:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-02-01T00:00:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-03-01T00:00:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-04-01T00:00:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-05-01T00:00:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-06-01T00:00:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-07-01T00:00:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-08-01T00:00:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-09-01T00:00:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-10-01T00:00:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-11-01T00:00:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-12-01T00:00:00Z").AsTime(), rows[i].Ts.AsTime())
}

func TestMetricsViewsTimeseries_month_grain_IST(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "timeseries")

	ctrl, err := rt.Controller(context.Background(), instanceID)
	require.NoError(t, err)
	r, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "timeseries_year"}, false)
	require.NoError(t, err)
	mv := r.GetMetricsView()

	q := &queries.MetricsViewTimeSeries{
		MeasureNames:    []string{"max_clicks"},
		MetricsViewName: "timeseries_year",
		MetricsView:     mv.Spec,
		TimeStart:       parseTime(t, "2022-12-31T18:30:00Z"),
		TimeEnd:         parseTime(t, "2024-01-31T18:30:00Z"),
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_MONTH,
		TimeZone:        "Asia/Kolkata",
		Limit:           250,
	}

	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	rows := q.Result.Data
	require.Len(t, rows, 13)
	i := 0
	require.Equal(t, parseTime(t, "2022-12-31T18:30:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-01-31T18:30:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-02-28T18:30:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-03-31T18:30:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-04-30T18:30:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-05-31T18:30:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-06-30T18:30:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-07-31T18:30:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-08-31T18:30:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-09-30T18:30:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-10-31T18:30:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-11-30T18:30:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-12-31T18:30:00Z").AsTime(), rows[i].Ts.AsTime())
}

func TestMetricsViewsTimeseries_quarter_grain_IST(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "timeseries")

	ctrl, err := rt.Controller(context.Background(), instanceID)
	require.NoError(t, err)
	r, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "timeseries_year"}, false)
	require.NoError(t, err)
	mv := r.GetMetricsView()

	q := &queries.MetricsViewTimeSeries{
		MeasureNames:    []string{"max_clicks"},
		MetricsViewName: "timeseries_year",
		MetricsView:     mv.Spec,
		TimeStart:       parseTime(t, "2022-12-31T18:30:00Z"),
		TimeEnd:         parseTime(t, "2024-01-31T18:30:00Z"),
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_QUARTER,
		TimeZone:        "Asia/Kolkata",
		Limit:           250,
	}

	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	rows := q.Result.Data
	require.Len(t, rows, 6)
	i := 0
	require.Equal(t, parseTime(t, "2022-10-31T18:30:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2022-12-31T18:30:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-03-31T18:30:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-06-30T18:30:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-09-30T18:30:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-12-31T18:30:00Z").AsTime(), rows[i].Ts.AsTime())
}

func TestMetricsViewsTimeseries_year_grain_IST(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "timeseries")

	ctrl, err := rt.Controller(context.Background(), instanceID)
	require.NoError(t, err)
	r, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "timeseries_year"}, false)
	require.NoError(t, err)
	mv := r.GetMetricsView()

	q := &queries.MetricsViewTimeSeries{
		MeasureNames:    []string{"max_clicks"},
		MetricsViewName: "timeseries_year",
		MetricsView:     mv.Spec,
		TimeStart:       parseTime(t, "2022-12-31T18:30:00Z"),
		TimeEnd:         parseTime(t, "2024-12-31T00:00:00Z"),
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		TimeZone:        "Asia/Kolkata",
		Limit:           250,
	}

	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	rows := q.Result.Data
	require.Len(t, rows, 2)
	i := 0
	require.Equal(t, parseTime(t, "2022-12-31T18:30:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-12-31T18:30:00Z").AsTime(), rows[i].Ts.AsTime())
}

func TestMetricsViewTimeSeries_DayLightSavingsBackwards_Continuous_Weekly(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "timeseries")

	ctrl, err := rt.Controller(context.Background(), instanceID)
	require.NoError(t, err)
	r, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "timeseries_dst_backwards"}, false)
	require.NoError(t, err)
	mv := r.GetMetricsView()

	q := &queries.MetricsViewTimeSeries{
		MeasureNames:    []string{"total_records"},
		MetricsViewName: "timeseries_dst_backwards",
		MetricsView:     mv.Spec,
		TimeStart:       parseTime(t, "2023-10-28T04:00:00.000Z"),
		TimeEnd:         parseTime(t, "2023-11-19T05:00:00.000Z"),
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_WEEK,
		TimeZone:        "America/New_York",
		Limit:           250,
	}
	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	rows := q.Result.Data
	require.Len(t, rows, 4)
	i := 0
	require.Equal(t, parseTime(t, "2023-10-22T04:00:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-10-29T04:00:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-11-05T04:00:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-11-12T05:00:00Z").AsTime(), rows[i].Ts.AsTime())
}

func TestMetricsViewTimeSeries_DayLightSavingsBackwards_Continuous_WeeklyOnSaturday(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "timeseries")

	ctrl, err := rt.Controller(context.Background(), instanceID)
	require.NoError(t, err)
	r, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "timeseries_dst_backwards"}, false)
	require.NoError(t, err)
	mv := r.GetMetricsView()

	mv.GetSpec().FirstDayOfWeek = 6
	q := &queries.MetricsViewTimeSeries{
		MeasureNames:    []string{"total_records"},
		MetricsViewName: "timeseries_dst_backwards",
		MetricsView:     mv.Spec,
		TimeStart:       parseTime(t, "2023-10-28T04:00:00.000Z"),
		TimeEnd:         parseTime(t, "2023-11-19T05:00:00.000Z"),
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_WEEK,
		TimeZone:        "America/New_York",
		Limit:           250,
	}
	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	rows := q.Result.Data
	require.Len(t, rows, 4)
	i := 0
	require.Equal(t, parseTime(t, "2023-10-28T04:00:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-11-04T04:00:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-11-11T05:00:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-11-18T05:00:00Z").AsTime(), rows[i].Ts.AsTime())
}

func TestMetricsViewTimeSeries_DayLightSavingsBackwards_Continuous_Daily(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "timeseries")

	ctrl, err := rt.Controller(context.Background(), instanceID)
	require.NoError(t, err)
	r, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "timeseries_dst_backwards"}, false)
	require.NoError(t, err)
	mv := r.GetMetricsView()

	q := &queries.MetricsViewTimeSeries{
		MeasureNames:    []string{"total_records"},
		MetricsViewName: "timeseries_dst_backwards",
		MetricsView:     mv.Spec,
		TimeStart:       parseTime(t, "2023-11-03T04:00:00.000Z"),
		TimeEnd:         parseTime(t, "2023-11-07T05:00:00.000Z"),
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_DAY,
		TimeZone:        "America/New_York",
		Limit:           250,
	}
	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	rows := q.Result.Data
	require.Len(t, rows, 4)
	i := 0
	require.Equal(t, parseTime(t, "2023-11-03T04:00:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-11-04T04:00:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-11-05T04:00:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-11-06T05:00:00Z").AsTime(), rows[i].Ts.AsTime())
}

func TestMetricsViewTimeSeries_DayLightSavingsBackwards_Sparse_Daily(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "timeseries")

	ctrl, err := rt.Controller(context.Background(), instanceID)
	require.NoError(t, err)
	r, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "timeseries_dst_backwards"}, false)
	require.NoError(t, err)
	mv := r.GetMetricsView()

	q := &queries.MetricsViewTimeSeries{
		MeasureNames: []string{"total_records"},
		Where: expressionpb.In(
			expressionpb.Identifier("label"),
			[]*runtimev1.Expression{expressionpb.Value(toStructpbValue(t, "sparse_day"))},
		),
		MetricsViewName: "timeseries_dst_backwards",
		MetricsView:     mv.Spec,
		TimeStart:       parseTime(t, "2023-11-03T04:00:00.000Z"),
		TimeEnd:         parseTime(t, "2023-11-07T05:00:00.000Z"),
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_DAY,
		TimeZone:        "America/New_York",
		Limit:           250,
	}
	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	rows := q.Result.Data
	require.Len(t, rows, 4)
	i := 0
	require.Equal(t, parseTime(t, "2023-11-03T04:00:00Z").AsTime(), rows[i].Ts.AsTime())
	require.NotNil(t, q.Result.Data[i].Records.AsMap()["total_records"])
	i++
	require.Equal(t, parseTime(t, "2023-11-04T04:00:00Z").AsTime(), rows[i].Ts.AsTime())
	require.Nil(t, q.Result.Data[i].Records.AsMap()["total_records"])
	i++
	require.Equal(t, parseTime(t, "2023-11-05T04:00:00Z").AsTime(), rows[i].Ts.AsTime())
	require.NotNil(t, q.Result.Data[i].Records.AsMap()["total_records"])
	i++
	require.Equal(t, parseTime(t, "2023-11-06T05:00:00Z").AsTime(), rows[i].Ts.AsTime())
	require.Nil(t, q.Result.Data[i].Records.AsMap()["total_records"])
}

func TestMetricsViewTimeSeries_DayLightSavingsBackwards_Continuous_Second(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "timeseries")

	ctrl, err := rt.Controller(context.Background(), instanceID)
	require.NoError(t, err)
	r, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "timeseries_dst_backwards"}, false)
	require.NoError(t, err)
	mv := r.GetMetricsView()

	q := &queries.MetricsViewTimeSeries{
		MeasureNames:    []string{"total_records"},
		MetricsViewName: "timeseries_dst_backwards",
		MetricsView:     mv.Spec,
		TimeStart:       parseTime(t, "2023-11-05T05:00:00.000Z"),
		TimeEnd:         parseTime(t, "2023-11-05T05:00:01.000Z"),
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_SECOND,
		TimeZone:        "America/New_York",
		Limit:           250,
	}
	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	rows := q.Result.Data
	require.Len(t, rows, 1)
	i := 0
	require.Equal(t, parseTime(t, "2023-11-05T05:00:00Z").AsTime(), rows[i].Ts.AsTime())

	q = &queries.MetricsViewTimeSeries{
		MeasureNames:    []string{"total_records"},
		MetricsViewName: "timeseries_dst_backwards",
		MetricsView:     mv.Spec,
		TimeStart:       parseTime(t, "2023-11-05T06:00:00.000Z"),
		TimeEnd:         parseTime(t, "2023-11-05T06:00:01.000Z"),
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_SECOND,
		TimeZone:        "America/New_York",
		Limit:           250,
	}
	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	rows = q.Result.Data
	require.Len(t, rows, 1)
	i = 0
	require.Equal(t, parseTime(t, "2023-11-05T06:00:00Z").AsTime(), rows[i].Ts.AsTime())
}

func TestMetricsViewTimeSeries_DayLightSavingsBackwards_Continuous_Minute(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "timeseries")

	ctrl, err := rt.Controller(context.Background(), instanceID)
	require.NoError(t, err)
	r, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "timeseries_dst_backwards"}, false)
	require.NoError(t, err)
	mv := r.GetMetricsView()

	q := &queries.MetricsViewTimeSeries{
		MeasureNames:    []string{"total_records"},
		MetricsViewName: "timeseries_dst_backwards",
		MetricsView:     mv.Spec,
		TimeStart:       parseTime(t, "2023-11-05T05:00:00.000Z"),
		TimeEnd:         parseTime(t, "2023-11-05T05:01:00.000Z"),
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_MINUTE,
		TimeZone:        "America/New_York",
		Limit:           250,
	}
	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	rows := q.Result.Data
	require.Len(t, rows, 1)
	i := 0
	require.Equal(t, parseTime(t, "2023-11-05T05:00:00Z").AsTime(), rows[i].Ts.AsTime())

	q = &queries.MetricsViewTimeSeries{
		MeasureNames:    []string{"total_records"},
		MetricsViewName: "timeseries_dst_backwards",
		MetricsView:     mv.Spec,
		TimeStart:       parseTime(t, "2023-11-05T06:00:00.000Z"),
		TimeEnd:         parseTime(t, "2023-11-05T06:01:00.000Z"),
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_MINUTE,
		TimeZone:        "America/New_York",
		Limit:           250,
	}
	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	rows = q.Result.Data
	require.Len(t, rows, 1)
	i = 0
	require.Equal(t, parseTime(t, "2023-11-05T06:00:00Z").AsTime(), rows[i].Ts.AsTime())
}

func TestMetricsViewTimeSeries_DayLightSavingsBackwards_Continuous_Hourly(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "timeseries")

	ctrl, err := rt.Controller(context.Background(), instanceID)
	require.NoError(t, err)
	r, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "timeseries_dst_backwards"}, false)
	require.NoError(t, err)
	mv := r.GetMetricsView()

	q := &queries.MetricsViewTimeSeries{
		MeasureNames:    []string{"total_records"},
		MetricsViewName: "timeseries_dst_backwards",
		MetricsView:     mv.Spec,
		TimeStart:       parseTime(t, "2023-11-05T03:00:00.000Z"),
		TimeEnd:         parseTime(t, "2023-11-05T08:00:00.000Z"),
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_HOUR,
		TimeZone:        "America/New_York",
		Limit:           250,
	}
	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	rows := q.Result.Data
	require.Len(t, rows, 5)
	i := 0
	require.Equal(t, parseTime(t, "2023-11-05T03:00:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-11-05T04:00:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-11-05T05:00:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-11-05T06:00:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-11-05T07:00:00Z").AsTime(), rows[i].Ts.AsTime())
}

func TestMetricsViewTimeSeries_DayLightSavingsBackwards_Sparse_Hourly(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "timeseries")

	ctrl, err := rt.Controller(context.Background(), instanceID)
	require.NoError(t, err)
	r, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "timeseries_dst_backwards"}, false)
	require.NoError(t, err)
	mv := r.GetMetricsView()

	q := &queries.MetricsViewTimeSeries{
		MeasureNames: []string{"total_records"},
		Where: expressionpb.In(
			expressionpb.Identifier("label"),
			[]*runtimev1.Expression{expressionpb.Value(toStructpbValue(t, "sparse_hour"))},
		),
		MetricsViewName: "timeseries_dst_backwards",
		MetricsView:     mv.Spec,
		TimeStart:       parseTime(t, "2023-11-05T03:00:00.000Z"),
		TimeEnd:         parseTime(t, "2023-11-05T08:00:00.000Z"),
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_HOUR,
		TimeZone:        "America/New_York",
		Limit:           250,
	}
	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	rows := q.Result.Data
	require.Len(t, rows, 5)
	i := 0
	require.Equal(t, parseTime(t, "2023-11-05T03:00:00Z").AsTime(), rows[i].Ts.AsTime())
	require.NotNil(t, q.Result.Data[i].Records.AsMap()["total_records"])
	i++
	require.Equal(t, parseTime(t, "2023-11-05T04:00:00Z").AsTime(), rows[i].Ts.AsTime())
	require.Nil(t, q.Result.Data[i].Records.AsMap()["total_records"])
	i++
	require.Equal(t, parseTime(t, "2023-11-05T05:00:00Z").AsTime(), rows[i].Ts.AsTime())
	require.NotNil(t, q.Result.Data[i].Records.AsMap()["total_records"])
	i++
	require.Equal(t, parseTime(t, "2023-11-05T06:00:00Z").AsTime(), rows[i].Ts.AsTime())
	require.Nil(t, q.Result.Data[i].Records.AsMap()["total_records"])
	i++
	require.Equal(t, parseTime(t, "2023-11-05T07:00:00Z").AsTime(), rows[i].Ts.AsTime())
	require.NotNil(t, q.Result.Data[i].Records.AsMap()["total_records"])
}

func TestMetricsViewTimeSeries_DayLightSavingsForwards_Continuous_Weekly(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "timeseries")

	ctrl, err := rt.Controller(context.Background(), instanceID)
	require.NoError(t, err)
	r, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "timeseries_dst_forwards"}, false)
	require.NoError(t, err)
	mv := r.GetMetricsView()

	q := &queries.MetricsViewTimeSeries{
		MeasureNames:    []string{"total_records"},
		MetricsViewName: "timeseries_dst_forwards",
		MetricsView:     mv.Spec,
		TimeStart:       parseTime(t, "2023-02-26T05:00:00.000Z"),
		TimeEnd:         parseTime(t, "2023-03-26T04:00:00.000Z"),
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_WEEK,
		TimeZone:        "America/New_York",
		Limit:           250,
	}
	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	rows := q.Result.Data
	require.Len(t, rows, 4)
	i := 0
	require.Equal(t, parseTime(t, "2023-02-26T05:00:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-03-05T05:00:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-03-12T05:00:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-03-19T04:00:00Z").AsTime(), rows[i].Ts.AsTime())
}

func TestMetricsViewTimeSeries_DayLightSavingsForwards_Continuous_Daily(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "timeseries")

	ctrl, err := rt.Controller(context.Background(), instanceID)
	require.NoError(t, err)
	r, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "timeseries_dst_forwards"}, false)
	require.NoError(t, err)
	mv := r.GetMetricsView()

	q := &queries.MetricsViewTimeSeries{
		MeasureNames:    []string{"total_records"},
		MetricsViewName: "timeseries_dst_forwards",
		MetricsView:     mv.Spec,
		TimeStart:       parseTime(t, "2023-03-10T05:00:00.000Z"),
		TimeEnd:         parseTime(t, "2023-03-14T04:00:00.000Z"),
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_DAY,
		TimeZone:        "America/New_York",
		Limit:           250,
	}
	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	rows := q.Result.Data
	require.Len(t, rows, 4)
	i := 0
	require.Equal(t, parseTime(t, "2023-03-10T05:00:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-03-11T05:00:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-03-12T05:00:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-03-13T04:00:00Z").AsTime(), rows[i].Ts.AsTime())
}

func TestMetricsViewTimeSeries_DayLightSavingsForwards_Sparse_Daily(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "timeseries")

	ctrl, err := rt.Controller(context.Background(), instanceID)
	require.NoError(t, err)
	r, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "timeseries_dst_forwards"}, false)
	require.NoError(t, err)
	mv := r.GetMetricsView()

	q := &queries.MetricsViewTimeSeries{
		MeasureNames: []string{"total_records"},
		Where: expressionpb.In(
			expressionpb.Identifier("label"),
			[]*runtimev1.Expression{expressionpb.Value(toStructpbValue(t, "sparse_day"))},
		),
		MetricsViewName: "timeseries_dst_forwards",
		MetricsView:     mv.Spec,
		TimeStart:       parseTime(t, "2023-03-10T05:00:00.000Z"),
		TimeEnd:         parseTime(t, "2023-03-14T04:00:00.000Z"),
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_DAY,
		TimeZone:        "America/New_York",
		Limit:           250,
	}
	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	rows := q.Result.Data
	require.Len(t, rows, 4)
	i := 0
	require.Equal(t, parseTime(t, "2023-03-10T05:00:00Z").AsTime(), rows[i].Ts.AsTime())
	require.NotNil(t, q.Result.Data[i].Records.AsMap()["total_records"])
	i++
	require.Equal(t, parseTime(t, "2023-03-11T05:00:00Z").AsTime(), rows[i].Ts.AsTime())
	require.Nil(t, q.Result.Data[i].Records.AsMap()["total_records"])
	i++
	require.Equal(t, parseTime(t, "2023-03-12T05:00:00Z").AsTime(), rows[i].Ts.AsTime())
	require.NotNil(t, q.Result.Data[i].Records.AsMap()["total_records"])
	i++
	require.Equal(t, parseTime(t, "2023-03-13T04:00:00Z").AsTime(), rows[i].Ts.AsTime())
	require.Nil(t, q.Result.Data[i].Records.AsMap()["total_records"])
}

func TestMetricsViewTimeSeries_DayLightSavingsForwards_Continuous_Hourly(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "timeseries")

	ctrl, err := rt.Controller(context.Background(), instanceID)
	require.NoError(t, err)
	r, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "timeseries_dst_forwards"}, false)
	require.NoError(t, err)
	mv := r.GetMetricsView()

	q := &queries.MetricsViewTimeSeries{
		MeasureNames:    []string{"total_records"},
		MetricsViewName: "timeseries_dst_forwards",
		MetricsView:     mv.Spec,
		TimeStart:       parseTime(t, "2023-03-12T04:00:00.000Z"),
		TimeEnd:         parseTime(t, "2023-03-12T09:00:00.000Z"),
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_HOUR,
		TimeZone:        "America/New_York",
		Limit:           250,
	}
	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	rows := q.Result.Data
	require.Len(t, rows, 5)
	i := 0
	require.Equal(t, parseTime(t, "2023-03-12T04:00:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-03-12T05:00:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-03-12T06:00:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-03-12T07:00:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-03-12T08:00:00Z").AsTime(), rows[i].Ts.AsTime())
}

func TestMetricsViewTimeSeries_DayLightSavingsForwards_Sparse_Hourly(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "timeseries")

	ctrl, err := rt.Controller(context.Background(), instanceID)
	require.NoError(t, err)
	r, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "timeseries_dst_forwards"}, false)
	require.NoError(t, err)
	mv := r.GetMetricsView()

	q := &queries.MetricsViewTimeSeries{
		MeasureNames: []string{"total_records"},
		Where: expressionpb.In(
			expressionpb.Identifier("label"),
			[]*runtimev1.Expression{expressionpb.Value(toStructpbValue(t, "sparse_hour"))},
		),
		MetricsViewName: "timeseries_dst_forwards",
		MetricsView:     mv.Spec,
		TimeStart:       parseTime(t, "2023-03-12T04:00:00.000Z"),
		TimeEnd:         parseTime(t, "2023-03-12T09:00:00.000Z"),
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_HOUR,
		TimeZone:        "America/New_York",
		Limit:           250,
	}
	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	rows := q.Result.Data
	require.Len(t, rows, 5)
	i := 0
	require.Equal(t, parseTime(t, "2023-03-12T04:00:00Z").AsTime(), rows[i].Ts.AsTime())
	require.Nil(t, q.Result.Data[i].Records.AsMap()["total_records"])
	i++
	require.Equal(t, parseTime(t, "2023-03-12T05:00:00Z").AsTime(), rows[i].Ts.AsTime())
	require.NotNil(t, q.Result.Data[i].Records.AsMap()["total_records"])
	i++
	require.Equal(t, parseTime(t, "2023-03-12T06:00:00Z").AsTime(), rows[i].Ts.AsTime())
	require.Nil(t, q.Result.Data[i].Records.AsMap()["total_records"])
	i++
	require.Equal(t, parseTime(t, "2023-03-12T07:00:00Z").AsTime(), rows[i].Ts.AsTime())
	require.NotNil(t, q.Result.Data[i].Records.AsMap()["total_records"])
	i++
	require.Equal(t, parseTime(t, "2023-03-12T08:00:00Z").AsTime(), rows[i].Ts.AsTime())
	require.Nil(t, q.Result.Data[i].Records.AsMap()["total_records"])
}

func TestMetricsViewTimeSeries_having_clause(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "timeseries")

	ctrl, err := rt.Controller(context.Background(), instanceID)
	require.NoError(t, err)
	r, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "timeseries_gaps"}, false)
	require.NoError(t, err)
	mv := r.GetMetricsView()

	q := &queries.MetricsViewTimeSeries{
		MeasureNames:    []string{"sum_imps"},
		MetricsViewName: "timeseries_gaps",
		MetricsView:     mv.Spec,
		TimeStart:       parseTime(t, "2019-01-01T00:00:00Z"),
		TimeEnd:         parseTime(t, "2019-01-07T00:00:00Z"),
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_DAY,
		Having: &runtimev1.Expression{
			Expression: &runtimev1.Expression_Cond{
				Cond: &runtimev1.Condition{
					Op: runtimev1.Operation_OPERATION_LTE,
					Exprs: []*runtimev1.Expression{
						{
							Expression: &runtimev1.Expression_Ident{
								Ident: "sum_imps",
							},
						},
						{
							Expression: &runtimev1.Expression_Val{
								Val: structpb.NewNumberValue(3),
							},
						},
					},
				},
			},
		},
		Limit: 250,
	}
	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	rows := q.Result.Data
	require.Len(t, rows, 6)
	i := 0
	require.Equal(t, parseTime(t, "2019-01-01T00:00:00Z").AsTime(), rows[i].Ts.AsTime())
	require.NotNil(t, q.Result.Data[i].Records.AsMap()["sum_imps"])
	i++
	require.Equal(t, parseTime(t, "2019-01-02T00:00:00Z").AsTime(), rows[i].Ts.AsTime())
	require.Nil(t, q.Result.Data[i].Records.AsMap()["sum_imps"])
	i++
	require.Equal(t, parseTime(t, "2019-01-03T00:00:00Z").AsTime(), rows[i].Ts.AsTime())
	require.Nil(t, q.Result.Data[i].Records.AsMap()["sum_imps"])
	i++
	require.Equal(t, parseTime(t, "2019-01-04T00:00:00Z").AsTime(), rows[i].Ts.AsTime())
	require.Nil(t, q.Result.Data[i].Records.AsMap()["sum_imps"])
	i++
	require.Equal(t, parseTime(t, "2019-01-05T00:00:00Z").AsTime(), rows[i].Ts.AsTime())
	require.Nil(t, q.Result.Data[i].Records.AsMap()["sum_imps"])
	i++
	require.Equal(t, parseTime(t, "2019-01-06T00:00:00Z").AsTime(), rows[i].Ts.AsTime())
	require.NotNil(t, q.Result.Data[i].Records.AsMap()["sum_imps"])
}

func toStructpbValue(t *testing.T, v any) *structpb.Value {
	sv, err := structpb.NewValue(v)
	require.NoError(t, err)
	return sv
}
