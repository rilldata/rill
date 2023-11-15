package queries_test

import (
	"context"
	"fmt"
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
	assertTimeSeriesResponse(t, q.Result, []string{
		"2023-01-01T00:00:00Z",
		"2023-02-01T00:00:00Z",
		"2023-03-01T00:00:00Z",
		"2023-04-01T00:00:00Z",
		"2023-05-01T00:00:00Z",
		"2023-06-01T00:00:00Z",
		"2023-07-01T00:00:00Z",
		"2023-08-01T00:00:00Z",
		"2023-09-01T00:00:00Z",
		"2023-10-01T00:00:00Z",
		"2023-11-01T00:00:00Z",
		"2023-12-01T00:00:00Z",
	})
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
	assertTimeSeriesResponse(t, q.Result, []string{
		"2022-12-31T18:30:00Z",
		"2023-01-31T18:30:00Z",
		"2023-02-28T18:30:00Z",
		"2023-03-31T18:30:00Z",
		"2023-04-30T18:30:00Z",
		"2023-05-31T18:30:00Z",
		"2023-06-30T18:30:00Z",
		"2023-07-31T18:30:00Z",
		"2023-08-31T18:30:00Z",
		"2023-09-30T18:30:00Z",
		"2023-10-31T18:30:00Z",
		"2023-11-30T18:30:00Z",
		"2023-12-31T18:30:00Z",
	})
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
	assertTimeSeriesResponse(t, q.Result, []string{
		"2022-10-31T18:30:00Z",
		"2022-12-31T18:30:00Z",
		"2023-03-31T18:30:00Z",
		"2023-06-30T18:30:00Z",
		"2023-09-30T18:30:00Z",
		"2023-12-31T18:30:00Z",
	})
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
	assertTimeSeriesResponse(t, q.Result, []string{"2022-12-31T18:30:00Z", "2023-12-31T18:30:00Z"})
}

func TestMetricsViewTimeSeries_DayLightSavings(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "timeseries")

	ctrl, err := rt.Controller(context.Background(), instanceID)
	require.NoError(t, err)
	r, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "timeseries_year"}, false)
	require.NoError(t, err)
	mv := r.GetMetricsView()

	q := &queries.MetricsViewTimeSeries{
		MeasureNames:    []string{"total_records"},
		MetricsViewName: "timeseries_dst",
		MetricsView:     mv.Spec,
		TimeStart:       parseTime(t, "2023-11-03T04:00:00.000Z"),
		TimeEnd:         parseTime(t, "2023-11-07T05:00:00.000Z"),
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_DAY,
		TimeZone:        "America/New_York",
		Limit:           250,
	}
	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	assertTimeSeriesResponse(t, q.Result, []string{
		"2023-11-03T04:00:00Z",
		"2023-11-04T04:00:00Z",
		"2023-11-05T04:00:00Z",
		"2023-11-06T05:00:00Z",
	})

	q = &queries.MetricsViewTimeSeries{
		MeasureNames:    []string{"total_records"},
		MetricsViewName: "timeseries_dst",
		MetricsView:     mv.Spec,
		TimeStart:       parseTime(t, "2023-11-05T03:00:00.000Z"),
		TimeEnd:         parseTime(t, "2023-11-05T08:00:00.000Z"),
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_HOUR,
		TimeZone:        "America/New_York",
		Limit:           250,
	}
	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	for _, r := range q.Result.Data {
		fmt.Println(r.Ts.AsTime().String(), r.Records.AsMap())
	}
	assertTimeSeriesResponse(t, q.Result, []string{
		"2023-11-05T03:00:00.000Z",
		"2023-11-05T04:00:00.000Z",
		"2023-11-05T06:00:00.000Z",
		"2023-11-05T07:00:00.000Z",
	})
}

func assertTimeSeriesResponse(t *testing.T, res *runtimev1.MetricsViewTimeSeriesResponse, dates []string) {
	require.NotEmpty(t, res)
	require.Len(t, res.Data, len(dates))
	for i, r := range res.Data {
		require.Equal(t, parseTime(t, dates[i]).AsTime(), r.Ts.AsTime())
	}
}
