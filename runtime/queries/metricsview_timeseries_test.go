package queries_test

import (
	"context"
	"fmt"
	// "fmt"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
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
	assertTimeSeriesResponse(t, q.Result, []string{
		"2023-10-22T04:00:00Z",
		"2023-10-29T04:00:00Z",
		"2023-11-05T04:00:00Z",
		"2023-11-12T05:00:00Z",
	})
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
	assertTimeSeriesResponse(t, q.Result, []string{
		"2023-11-03T04:00:00Z",
		"2023-11-04T04:00:00Z",
		"2023-11-05T04:00:00Z",
		"2023-11-06T05:00:00Z",
	})
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
		Filter: &runtimev1.MetricsViewFilter{
			Include: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "label",
					In:   []*structpb.Value{toStructpbValue(t, "sparse_day")},
				},
			},
		},
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
	assertTimeSeriesResponse(t, q.Result, []string{
		"2023-11-03T04:00:00Z",
		"2023-11-04T04:00:00Z",
		"2023-11-05T04:00:00Z",
		"2023-11-06T05:00:00Z",
	})
	require.NotNil(t, q.Result.Data[0].Records.AsMap()["total_records"])
	require.Nil(t, q.Result.Data[1].Records.AsMap()["total_records"])
	require.NotNil(t, q.Result.Data[2].Records.AsMap()["total_records"])
	require.Nil(t, q.Result.Data[3].Records.AsMap()["total_records"])
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
	assertTimeSeriesResponse(t, q.Result, []string{
		"2023-11-05T03:00:00Z",
		"2023-11-05T04:00:00Z",
		"2023-11-05T05:00:00Z",
		"2023-11-05T06:00:00Z",
		"2023-11-05T07:00:00Z",
	})
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
		Filter: &runtimev1.MetricsViewFilter{
			Include: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "label",
					In:   []*structpb.Value{toStructpbValue(t, "sparse_hour")},
				},
			},
		},
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
	assertTimeSeriesResponse(t, q.Result, []string{
		"2023-11-05T03:00:00Z",
		"2023-11-05T04:00:00Z",
		"2023-11-05T05:00:00Z",
		"2023-11-05T06:00:00Z",
		"2023-11-05T07:00:00Z",
	})
	require.NotNil(t, q.Result.Data[0].Records.AsMap()["total_records"])
	require.Nil(t, q.Result.Data[1].Records.AsMap()["total_records"])
	require.NotNil(t, q.Result.Data[2].Records.AsMap()["total_records"])
	require.Nil(t, q.Result.Data[3].Records.AsMap()["total_records"])
	require.NotNil(t, q.Result.Data[4].Records.AsMap()["total_records"])
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
	assertTimeSeriesResponse(t, q.Result, []string{
		"2023-02-26T05:00:00Z",
		"2023-03-05T05:00:00Z",
		"2023-03-12T05:00:00Z",
		"2023-03-19T04:00:00Z",
	})
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
	assertTimeSeriesResponse(t, q.Result, []string{
		"2023-03-10T05:00:00Z",
		"2023-03-11T05:00:00Z",
		"2023-03-12T05:00:00Z",
		"2023-03-13T04:00:00Z",
	})
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
		Filter: &runtimev1.MetricsViewFilter{
			Include: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "label",
					In:   []*structpb.Value{toStructpbValue(t, "sparse_day")},
				},
			},
		},
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
	assertTimeSeriesResponse(t, q.Result, []string{
		"2023-03-10T05:00:00Z",
		"2023-03-11T05:00:00Z",
		"2023-03-12T05:00:00Z",
		"2023-03-13T04:00:00Z",
	})
	require.NotNil(t, q.Result.Data[0].Records.AsMap()["total_records"])
	require.Nil(t, q.Result.Data[1].Records.AsMap()["total_records"])
	require.NotNil(t, q.Result.Data[2].Records.AsMap()["total_records"])
	require.Nil(t, q.Result.Data[3].Records.AsMap()["total_records"])
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
	assertTimeSeriesResponse(t, q.Result, []string{
		"2023-03-12T04:00:00Z",
		"2023-03-12T05:00:00Z",
		"2023-03-12T06:00:00Z",
		"2023-03-12T07:00:00Z",
		"2023-03-12T08:00:00Z",
	})
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
		Filter: &runtimev1.MetricsViewFilter{
			Include: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "label",
					In:   []*structpb.Value{toStructpbValue(t, "sparse_hour")},
				},
			},
		},
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
	assertTimeSeriesResponse(t, q.Result, []string{
		"2023-03-12T04:00:00Z",
		"2023-03-12T05:00:00Z",
		"2023-03-12T06:00:00Z",
		"2023-03-12T07:00:00Z",
		"2023-03-12T08:00:00Z",
	})
	require.Nil(t, q.Result.Data[0].Records.AsMap()["total_records"])
	require.NotNil(t, q.Result.Data[1].Records.AsMap()["total_records"])
	require.Nil(t, q.Result.Data[2].Records.AsMap()["total_records"])
	require.NotNil(t, q.Result.Data[3].Records.AsMap()["total_records"])
	require.Nil(t, q.Result.Data[4].Records.AsMap()["total_records"])
}

func assertTimeSeriesResponse(t *testing.T, res *runtimev1.MetricsViewTimeSeriesResponse, expected []string) {
	require.NotEmpty(t, res)
	actual := make([]string, 0)
	for _, r := range res.Data {
		actual = append(actual, r.Ts.AsTime().Format(time.RFC3339))
		fmt.Println(r.Ts.AsTime().UTC(), r.Records.AsMap())
	}
	require.ElementsMatch(t, actual, expected)
}

func toStructpbValue(t *testing.T, v any) *structpb.Value {
	sv, err := structpb.NewValue(v)
	require.NoError(t, err)
	return sv
}
