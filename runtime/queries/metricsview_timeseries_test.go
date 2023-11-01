package queries_test

import (
	"context"
	// "fmt"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/queries"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
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
	i := 0
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
	i := 0
	require.Equal(t, parseTime(t, "2022-12-31T18:30:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-12-31T18:30:00Z").AsTime(), rows[i].Ts.AsTime())
}
