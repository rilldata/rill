package queries_test

import (
	"context"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers/clickhouse/testclickhouse"
	"github.com/rilldata/rill/runtime/pkg/expressionpb"
	"github.com/rilldata/rill/runtime/queries"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/rilldata/rill/runtime/testruntime/testmode"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestMetricsViewsTimeseriesAgainstClickHouse(t *testing.T) {
	testmode.Expensive(t)
	// Create a test ClickHouse cluster
	dsn := testclickhouse.Start(t)
	t.Setenv("RILL_RUNTIME_TEST_OLAP_DRIVER", "clickhouse")
	t.Setenv("RILL_RUNTIME_TEST_OLAP_DSN", dsn)
	rt, instanceID := testruntime.NewInstanceForProject(t, "timeseries_clickhouse")
	t.Run("testMetricsViewsTimeseries_month_grain", func(t *testing.T) { testMetricsViewsTimeseries_month_grain(t, rt, instanceID) })
	t.Run("testMetricsViewsTimeseries_month_grain_IST", func(t *testing.T) { testMetricsViewsTimeseries_month_grain_IST(t, rt, instanceID) })
	t.Run("testMetricsViewsTimeseries_quarter_grain_IST", func(t *testing.T) { testMetricsViewsTimeseries_quarter_grain_IST(t, rt, instanceID) })
	t.Run("testMetricsViewsTimeseries_year_grain_IST", func(t *testing.T) { testMetricsViewsTimeseries_year_grain_IST(t, rt, instanceID) })
	t.Run("testMetricsViewTimeSeries_DayLightSavingsBackwards_Continuous_Weekly", func(t *testing.T) {
		testMetricsViewTimeSeries_DayLightSavingsBackwards_Continuous_Weekly(t, rt, instanceID)
	})
	t.Run("testMetricsViewTimeSeries_DayLightSavingsBackwards_Continuous_WeeklyOnSaturday", func(t *testing.T) {
		testMetricsViewTimeSeries_DayLightSavingsBackwards_Continuous_WeeklyOnSaturday(t, rt, instanceID)
	})
	t.Run("testMetricsViewTimeSeries_DayLightSavingsBackwards_Continuous_Daily", func(t *testing.T) {
		testMetricsViewTimeSeries_DayLightSavingsBackwards_Continuous_Daily(t, rt, instanceID)
	})
	t.Run("testMetricsViewTimeSeries_DayLightSavingsBackwards_Sparse_Daily", func(t *testing.T) { testMetricsViewTimeSeries_DayLightSavingsBackwards_Sparse_Daily(t, rt, instanceID) })
	t.Run("testMetricsViewTimeSeries_DayLightSavingsBackwards_Continuous_Second", func(t *testing.T) {
		testMetricsViewTimeSeries_DayLightSavingsBackwards_Continuous_Second(t, true, rt, instanceID)
	})
	t.Run("testMetricsViewTimeSeries_DayLightSavingsBackwards_Continuous_Minute", func(t *testing.T) {
		testMetricsViewTimeSeries_DayLightSavingsBackwards_Continuous_Minute(t, true, rt, instanceID)
	})
	t.Run("testMetricsViewTimeSeries_DayLightSavingsBackwards_Continuous_Hourly", func(t *testing.T) {
		testMetricsViewTimeSeries_DayLightSavingsBackwards_Continuous_Hourly(t, true, rt, instanceID)
	})
	t.Run("testMetricsViewTimeSeries_DayLightSavingsBackwards_Sparse_Hourly", func(t *testing.T) {
		testMetricsViewTimeSeries_DayLightSavingsBackwards_Sparse_Hourly(t, true, rt, instanceID)
	})
	t.Run("testMetricsViewTimeSeries_DayLightSavingsForwards_Continuous_Weekly", func(t *testing.T) {
		testMetricsViewTimeSeries_DayLightSavingsForwards_Continuous_Weekly(t, rt, instanceID)
	})
	t.Run("testMetricsViewTimeSeries_DayLightSavingsForwards_Continuous_Daily", func(t *testing.T) {
		testMetricsViewTimeSeries_DayLightSavingsForwards_Continuous_Daily(t, rt, instanceID)
	})
	t.Run("testMetricsViewTimeSeries_DayLightSavingsForwards_Sparse_Daily", func(t *testing.T) { testMetricsViewTimeSeries_DayLightSavingsForwards_Sparse_Daily(t, rt, instanceID) })
	t.Run("testMetricsViewTimeSeries_DayLightSavingsForwards_Continuous_Hourly", func(t *testing.T) {
		testMetricsViewTimeSeries_DayLightSavingsForwards_Continuous_Hourly(t, true, rt, instanceID)
	})
	testMetricsViewTimeSeries_DayLightSavingsForwards_Continuous_Hourly(t, true, rt, instanceID)
	t.Run("testMetricsViewTimeSeries_DayLightSavingsForwards_Sparse_Hourly", func(t *testing.T) {
		testMetricsViewTimeSeries_DayLightSavingsForwards_Sparse_Hourly(t, true, rt, instanceID)
	})
	t.Run("testMetricsViewTimeSeries_having_clause", func(t *testing.T) { testMetricsViewTimeSeries_having_clause(t, rt, instanceID) })
}

func TestMetricsViewsTimeseriesAgainstDuckdb(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "timeseries")
	t.Run("testMetricsViewsTimeseries_month_grain", func(t *testing.T) { testMetricsViewsTimeseries_month_grain(t, rt, instanceID) })
	t.Run("testMetricsViewsTimeseries_month_grain_IST", func(t *testing.T) { testMetricsViewsTimeseries_month_grain_IST(t, rt, instanceID) })
	t.Run("testMetricsViewsTimeseries_quarter_grain_IST", func(t *testing.T) { testMetricsViewsTimeseries_quarter_grain_IST(t, rt, instanceID) })
	t.Run("testMetricsViewsTimeseries_year_grain_IST", func(t *testing.T) { testMetricsViewsTimeseries_year_grain_IST(t, rt, instanceID) })
	t.Run("testMetricsViewTimeSeries_DayLightSavingsBackwards_Continuous_Weekly", func(t *testing.T) {
		testMetricsViewTimeSeries_DayLightSavingsBackwards_Continuous_Weekly(t, rt, instanceID)
	})
	t.Run("testMetricsViewTimeSeries_DayLightSavingsBackwards_Continuous_WeeklyOnSaturday", func(t *testing.T) {
		testMetricsViewTimeSeries_DayLightSavingsBackwards_Continuous_WeeklyOnSaturday(t, rt, instanceID)
	})
	t.Run("testMetricsViewTimeSeries_DayLightSavingsBackwards_Continuous_Daily", func(t *testing.T) {
		testMetricsViewTimeSeries_DayLightSavingsBackwards_Continuous_Daily(t, rt, instanceID)
	})
	t.Run("testMetricsViewTimeSeries_DayLightSavingsBackwards_Sparse_Daily", func(t *testing.T) { testMetricsViewTimeSeries_DayLightSavingsBackwards_Sparse_Daily(t, rt, instanceID) })
	t.Run("testMetricsViewTimeSeries_DayLightSavingsBackwards_Continuous_Second", func(t *testing.T) {
		testMetricsViewTimeSeries_DayLightSavingsBackwards_Continuous_Second(t, false, rt, instanceID)
	})
	t.Run("testMetricsViewTimeSeries_DayLightSavingsBackwards_Continuous_Minute", func(t *testing.T) {
		testMetricsViewTimeSeries_DayLightSavingsBackwards_Continuous_Minute(t, false, rt, instanceID)
	})
	t.Run("testMetricsViewTimeSeries_DayLightSavingsBackwards_Continuous_Hourly", func(t *testing.T) {
		testMetricsViewTimeSeries_DayLightSavingsBackwards_Continuous_Hourly(t, false, rt, instanceID)
	})
	t.Run("testMetricsViewTimeSeries_DayLightSavingsBackwards_Sparse_Hourly", func(t *testing.T) {
		testMetricsViewTimeSeries_DayLightSavingsBackwards_Sparse_Hourly(t, false, rt, instanceID)
	})
	t.Run("testMetricsViewTimeSeries_DayLightSavingsForwards_Continuous_Weekly", func(t *testing.T) {
		testMetricsViewTimeSeries_DayLightSavingsForwards_Continuous_Weekly(t, rt, instanceID)
	})
	t.Run("testMetricsViewTimeSeries_DayLightSavingsForwards_Continuous_Daily", func(t *testing.T) {
		testMetricsViewTimeSeries_DayLightSavingsForwards_Continuous_Daily(t, rt, instanceID)
	})
	t.Run("testMetricsViewTimeSeries_DayLightSavingsForwards_Sparse_Daily", func(t *testing.T) { testMetricsViewTimeSeries_DayLightSavingsForwards_Sparse_Daily(t, rt, instanceID) })
	t.Run("testMetricsViewTimeSeries_DayLightSavingsForwards_Continuous_Hourly", func(t *testing.T) {
		testMetricsViewTimeSeries_DayLightSavingsForwards_Continuous_Hourly(t, false, rt, instanceID)
	})
	t.Run("testMetricsViewTimeSeries_DayLightSavingsForwards_Sparse_Hourly", func(t *testing.T) {
		testMetricsViewTimeSeries_DayLightSavingsForwards_Sparse_Hourly(t, false, rt, instanceID)
	})
	t.Run("testMetricsViewTimeSeries_having_clause", func(t *testing.T) { testMetricsViewTimeSeries_having_clause(t, rt, instanceID) })
}

func testMetricsViewsTimeseries_month_grain(t *testing.T, rt *runtime.Runtime, instanceID string) {
	q := &queries.MetricsViewTimeSeries{
		MeasureNames:    []string{"max_clicks"},
		MetricsViewName: "timeseries_year",
		TimeStart:       parseTime(t, "2023-01-01T00:00:00Z"),
		TimeEnd:         parseTime(t, "2024-01-01T00:00:00Z"),
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_MONTH,
		Limit:           250,
		SecurityClaims:  testClaims(),
	}

	err := q.Resolve(context.Background(), rt, instanceID, 0)
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

func testMetricsViewsTimeseries_month_grain_IST(t *testing.T, rt *runtime.Runtime, instanceID string) {
	q := &queries.MetricsViewTimeSeries{
		MeasureNames:    []string{"max_clicks"},
		MetricsViewName: "timeseries_year",
		TimeStart:       parseTime(t, "2022-12-31T18:30:00Z"),
		TimeEnd:         parseTime(t, "2024-01-31T18:30:00Z"),
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_MONTH,
		TimeZone:        "Asia/Kolkata",
		Limit:           250,
		SecurityClaims:  testClaims(),
	}

	err := q.Resolve(context.Background(), rt, instanceID, 0)
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

func testMetricsViewsTimeseries_quarter_grain_IST(t *testing.T, rt *runtime.Runtime, instanceID string) {
	q := &queries.MetricsViewTimeSeries{
		MeasureNames:    []string{"max_clicks"},
		MetricsViewName: "timeseries_year",
		TimeStart:       parseTime(t, "2022-12-31T18:30:00Z"),
		TimeEnd:         parseTime(t, "2024-01-31T18:30:00Z"),
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_QUARTER,
		TimeZone:        "Asia/Kolkata",
		Limit:           250,
		SecurityClaims:  testClaims(),
	}

	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	rows := q.Result.Data
	require.Len(t, rows, 5)
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

func testMetricsViewsTimeseries_year_grain_IST(t *testing.T, rt *runtime.Runtime, instanceID string) {
	q := &queries.MetricsViewTimeSeries{
		MeasureNames:    []string{"max_clicks"},
		MetricsViewName: "timeseries_year",
		TimeStart:       parseTime(t, "2022-12-31T18:30:00Z"),
		TimeEnd:         parseTime(t, "2024-12-31T00:00:00Z"),
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		TimeZone:        "Asia/Kolkata",
		Limit:           250,
		SecurityClaims:  testClaims(),
	}

	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	rows := q.Result.Data
	require.Len(t, rows, 2)
	i := 0
	require.Equal(t, parseTime(t, "2022-12-31T18:30:00Z").AsTime(), rows[i].Ts.AsTime())
	i++
	require.Equal(t, parseTime(t, "2023-12-31T18:30:00Z").AsTime(), rows[i].Ts.AsTime())
}

func testMetricsViewTimeSeries_DayLightSavingsBackwards_Continuous_Weekly(t *testing.T, rt *runtime.Runtime, instanceID string) {
	q := &queries.MetricsViewTimeSeries{
		MeasureNames:    []string{"total_records"},
		MetricsViewName: "timeseries_dst_backwards",
		TimeStart:       parseTime(t, "2023-10-28T04:00:00.000Z"),
		TimeEnd:         parseTime(t, "2023-11-19T05:00:00.000Z"),
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_WEEK,
		TimeZone:        "America/New_York",
		Limit:           250,
		SecurityClaims:  testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
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

func testMetricsViewTimeSeries_DayLightSavingsBackwards_Continuous_WeeklyOnSaturday(t *testing.T, rt *runtime.Runtime, instanceID string) {
	q := &queries.MetricsViewTimeSeries{
		MeasureNames:    []string{"total_records"},
		MetricsViewName: "timeseries_dst_backwards_fdow6",
		TimeStart:       parseTime(t, "2023-10-28T04:00:00.000Z"),
		TimeEnd:         parseTime(t, "2023-11-19T05:00:00.000Z"),
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_WEEK,
		TimeZone:        "America/New_York",
		Limit:           250,
		SecurityClaims:  testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
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

func testMetricsViewTimeSeries_DayLightSavingsBackwards_Continuous_Daily(t *testing.T, rt *runtime.Runtime, instanceID string) {
	q := &queries.MetricsViewTimeSeries{
		MeasureNames:    []string{"total_records"},
		MetricsViewName: "timeseries_dst_backwards",
		TimeStart:       parseTime(t, "2023-11-03T04:00:00.000Z"),
		TimeEnd:         parseTime(t, "2023-11-07T05:00:00.000Z"),
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_DAY,
		TimeZone:        "America/New_York",
		Limit:           250,
		SecurityClaims:  testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
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

func testMetricsViewTimeSeries_DayLightSavingsBackwards_Sparse_Daily(t *testing.T, rt *runtime.Runtime, instanceID string) {
	q := &queries.MetricsViewTimeSeries{
		MeasureNames: []string{"total_records"},
		Where: expressionpb.In(
			expressionpb.Identifier("label"),
			[]*runtimev1.Expression{expressionpb.Value(toStructpbValue(t, "sparse_day"))},
		),
		MetricsViewName: "timeseries_dst_backwards",
		TimeStart:       parseTime(t, "2023-11-03T04:00:00.000Z"),
		TimeEnd:         parseTime(t, "2023-11-07T05:00:00.000Z"),
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_DAY,
		TimeZone:        "America/New_York",
		Limit:           250,
		SecurityClaims:  testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
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

func testMetricsViewTimeSeries_DayLightSavingsBackwards_Continuous_Second(t *testing.T, isClickHouse bool, rt *runtime.Runtime, instanceID string) {
	q := &queries.MetricsViewTimeSeries{
		MeasureNames:    []string{"total_records"},
		MetricsViewName: "timeseries_dst_backwards",
		TimeStart:       parseTime(t, "2023-11-05T05:00:00.000Z"),
		TimeEnd:         parseTime(t, "2023-11-05T05:00:01.000Z"),
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_SECOND,
		TimeZone:        "America/New_York",
		Limit:           250,
		SecurityClaims:  testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	rows := q.Result.Data
	require.Len(t, rows, 1)
	i := 0
	// ClickHouse returns 05:00 UTC for this range; DuckDB returns 06:00 UTC (DST difference).
	if isClickHouse {
		require.Equal(t, parseTime(t, "2023-11-05T05:00:00Z").AsTime(), rows[i].Ts.AsTime())
	} else {
		require.Equal(t, parseTime(t, "2023-11-05T06:00:00Z").AsTime(), rows[i].Ts.AsTime())
	}

	q = &queries.MetricsViewTimeSeries{
		MeasureNames:    []string{"total_records"},
		MetricsViewName: "timeseries_dst_backwards",
		TimeStart:       parseTime(t, "2023-11-05T06:00:00.000Z"),
		TimeEnd:         parseTime(t, "2023-11-05T06:00:01.000Z"),
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_SECOND,
		TimeZone:        "America/New_York",
		Limit:           250,
		SecurityClaims:  testClaims(),
	}
	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	rows = q.Result.Data
	require.Len(t, rows, 1)
	i = 0
	require.Equal(t, parseTime(t, "2023-11-05T06:00:00Z").AsTime(), rows[i].Ts.AsTime())
}

func testMetricsViewTimeSeries_DayLightSavingsBackwards_Continuous_Minute(t *testing.T, isClickHouse bool, rt *runtime.Runtime, instanceID string) {
	q := &queries.MetricsViewTimeSeries{
		MeasureNames:    []string{"total_records"},
		MetricsViewName: "timeseries_dst_backwards",
		TimeStart:       parseTime(t, "2023-11-05T05:00:00.000Z"),
		TimeEnd:         parseTime(t, "2023-11-05T05:01:00.000Z"),
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_MINUTE,
		TimeZone:        "America/New_York",
		Limit:           250,
		SecurityClaims:  testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	rows := q.Result.Data
	require.Len(t, rows, 1)
	i := 0
	// ClickHouse returns 05:00 UTC for this range; DuckDB returns 06:00 UTC (DST difference).
	if isClickHouse {
		require.Equal(t, parseTime(t, "2023-11-05T05:00:00Z").AsTime(), rows[i].Ts.AsTime())
	} else {
		require.Equal(t, parseTime(t, "2023-11-05T06:00:00Z").AsTime(), rows[i].Ts.AsTime())
	}

	q = &queries.MetricsViewTimeSeries{
		MeasureNames:    []string{"total_records"},
		MetricsViewName: "timeseries_dst_backwards",
		TimeStart:       parseTime(t, "2023-11-05T06:00:00.000Z"),
		TimeEnd:         parseTime(t, "2023-11-05T06:01:00.000Z"),
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_MINUTE,
		TimeZone:        "America/New_York",
		Limit:           250,
		SecurityClaims:  testClaims(),
	}
	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	rows = q.Result.Data
	require.Len(t, rows, 1)
	i = 0
	require.Equal(t, parseTime(t, "2023-11-05T06:00:00Z").AsTime(), rows[i].Ts.AsTime())
}

func testMetricsViewTimeSeries_DayLightSavingsBackwards_Continuous_Hourly(t *testing.T, isClickHouse bool, rt *runtime.Runtime, instanceID string) {
	q := &queries.MetricsViewTimeSeries{
		MeasureNames:    []string{"total_records"},
		MetricsViewName: "timeseries_dst_backwards",
		TimeStart:       parseTime(t, "2023-11-05T03:00:00.000Z"),
		TimeEnd:         parseTime(t, "2023-11-05T08:00:00.000Z"),
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_HOUR,
		TimeZone:        "America/New_York",
		Limit:           250,
		SecurityClaims:  testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	rows := q.Result.Data
	if isClickHouse {
		// ClickHouse includes 05:00 UTC (5 bins); DuckDB skips it (4 bins) due to DST fall back.
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
	} else {
		require.Len(t, rows, 4)
		i := 0
		require.Equal(t, parseTime(t, "2023-11-05T03:00:00Z").AsTime(), rows[i].Ts.AsTime())
		i++
		require.Equal(t, parseTime(t, "2023-11-05T04:00:00Z").AsTime(), rows[i].Ts.AsTime())
		i++
		// no 05:00 hour since 04:00 to 05:00 UTC are same because of DST fall back
		require.Equal(t, parseTime(t, "2023-11-05T06:00:00Z").AsTime(), rows[i].Ts.AsTime())
		i++
		require.Equal(t, parseTime(t, "2023-11-05T07:00:00Z").AsTime(), rows[i].Ts.AsTime())
	}
}

func testMetricsViewTimeSeries_DayLightSavingsBackwards_Sparse_Hourly(t *testing.T, isClickHouse bool, rt *runtime.Runtime, instanceID string) {
	q := &queries.MetricsViewTimeSeries{
		MeasureNames: []string{"total_records"},
		Where: expressionpb.In(
			expressionpb.Identifier("label"),
			[]*runtimev1.Expression{expressionpb.Value(toStructpbValue(t, "sparse_hour"))},
		),
		MetricsViewName: "timeseries_dst_backwards",
		TimeStart:       parseTime(t, "2023-11-05T03:00:00.000Z"),
		TimeEnd:         parseTime(t, "2023-11-05T08:00:00.000Z"),
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_HOUR,
		TimeZone:        "America/New_York",
		Limit:           250,
		SecurityClaims:  testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	rows := q.Result.Data
	if isClickHouse {
		// ClickHouse includes 05:00 UTC (5 bins); DuckDB skips it (4 bins) due to DST fall back.
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
	} else {
		require.Len(t, rows, 4)
		i := 0
		require.Equal(t, parseTime(t, "2023-11-05T03:00:00Z").AsTime(), rows[i].Ts.AsTime())
		require.NotNil(t, q.Result.Data[i].Records.AsMap()["total_records"])
		i++
		require.Equal(t, parseTime(t, "2023-11-05T04:00:00Z").AsTime(), rows[i].Ts.AsTime())
		require.Nil(t, q.Result.Data[i].Records.AsMap()["total_records"])
		i++
		// no 05:00 hour since 04:00 to 05:00 UTC are same because of DST fall back
		require.Equal(t, parseTime(t, "2023-11-05T06:00:00Z").AsTime(), rows[i].Ts.AsTime())
		require.Nil(t, q.Result.Data[i].Records.AsMap()["total_records"])
		i++
		require.Equal(t, parseTime(t, "2023-11-05T07:00:00Z").AsTime(), rows[i].Ts.AsTime())
		require.NotNil(t, q.Result.Data[i].Records.AsMap()["total_records"])
	}
}

func testMetricsViewTimeSeries_DayLightSavingsForwards_Continuous_Weekly(t *testing.T, rt *runtime.Runtime, instanceID string) {
	q := &queries.MetricsViewTimeSeries{
		MeasureNames:    []string{"total_records"},
		MetricsViewName: "timeseries_dst_forwards",
		TimeStart:       parseTime(t, "2023-02-26T05:00:00.000Z"),
		TimeEnd:         parseTime(t, "2023-03-26T04:00:00.000Z"),
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_WEEK,
		TimeZone:        "America/New_York",
		Limit:           250,
		SecurityClaims:  testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
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

func testMetricsViewTimeSeries_DayLightSavingsForwards_Continuous_Daily(t *testing.T, rt *runtime.Runtime, instanceID string) {
	q := &queries.MetricsViewTimeSeries{
		MeasureNames:    []string{"total_records"},
		MetricsViewName: "timeseries_dst_forwards",
		TimeStart:       parseTime(t, "2023-03-10T05:00:00.000Z"),
		TimeEnd:         parseTime(t, "2023-03-14T04:00:00.000Z"),
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_DAY,
		TimeZone:        "America/New_York",
		Limit:           250,
		SecurityClaims:  testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
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

func testMetricsViewTimeSeries_DayLightSavingsForwards_Sparse_Daily(t *testing.T, rt *runtime.Runtime, instanceID string) {
	q := &queries.MetricsViewTimeSeries{
		MeasureNames: []string{"total_records"},
		Where: expressionpb.In(
			expressionpb.Identifier("label"),
			[]*runtimev1.Expression{expressionpb.Value(toStructpbValue(t, "sparse_day"))},
		),
		MetricsViewName: "timeseries_dst_forwards",
		TimeStart:       parseTime(t, "2023-03-10T05:00:00.000Z"),
		TimeEnd:         parseTime(t, "2023-03-14T04:00:00.000Z"),
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_DAY,
		TimeZone:        "America/New_York",
		Limit:           250,
		SecurityClaims:  testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
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

func testMetricsViewTimeSeries_DayLightSavingsForwards_Continuous_Hourly(t *testing.T, isClickHouse bool, rt *runtime.Runtime, instanceID string) {
	q := &queries.MetricsViewTimeSeries{
		MeasureNames:    []string{"total_records"},
		MetricsViewName: "timeseries_dst_forwards",
		TimeStart:       parseTime(t, "2023-03-12T04:00:00.000Z"),
		TimeEnd:         parseTime(t, "2023-03-12T09:00:00.000Z"),
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_HOUR,
		TimeZone:        "America/New_York",
		Limit:           250,
		SecurityClaims:  testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	rows := q.Result.Data
	if isClickHouse {
		// ClickHouse has 5 bins (no duplicate 07:00 for DST spring forward); DuckDB has 6 with two 07:00.
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
	} else {
		require.Len(t, rows, 6)
		i := 0
		require.Equal(t, parseTime(t, "2023-03-12T04:00:00Z").AsTime(), rows[i].Ts.AsTime())
		i++
		require.Equal(t, parseTime(t, "2023-03-12T05:00:00Z").AsTime(), rows[i].Ts.AsTime())
		i++
		require.Equal(t, parseTime(t, "2023-03-12T06:00:00Z").AsTime(), rows[i].Ts.AsTime())
		i++
		require.Equal(t, parseTime(t, "2023-03-12T07:00:00Z").AsTime(), rows[i].Ts.AsTime())
		i++
		require.Equal(t, parseTime(t, "2023-03-12T07:00:00Z").AsTime(), rows[i].Ts.AsTime())
		i++
		require.Equal(t, parseTime(t, "2023-03-12T08:00:00Z").AsTime(), rows[i].Ts.AsTime())
	}
}

func testMetricsViewTimeSeries_DayLightSavingsForwards_Sparse_Hourly(t *testing.T, isClickHouse bool, rt *runtime.Runtime, instanceID string) {
	q := &queries.MetricsViewTimeSeries{
		MeasureNames: []string{"total_records"},
		Where: expressionpb.In(
			expressionpb.Identifier("label"),
			[]*runtimev1.Expression{expressionpb.Value(toStructpbValue(t, "sparse_hour"))},
		),
		MetricsViewName: "timeseries_dst_forwards",
		TimeStart:       parseTime(t, "2023-03-12T04:00:00.000Z"),
		TimeEnd:         parseTime(t, "2023-03-12T09:00:00.000Z"),
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_HOUR,
		TimeZone:        "America/New_York",
		Limit:           250,
		SecurityClaims:  testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	rows := q.Result.Data
	if isClickHouse {
		// ClickHouse has 5 bins (no duplicate 07:00 for DST spring forward); DuckDB has 6 with two 07:00.
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
	} else {
		require.Len(t, rows, 6)
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
		require.Equal(t, parseTime(t, "2023-03-12T07:00:00Z").AsTime(), rows[i].Ts.AsTime())
		require.NotNil(t, q.Result.Data[i].Records.AsMap()["total_records"])
		i++
		require.Equal(t, parseTime(t, "2023-03-12T08:00:00Z").AsTime(), rows[i].Ts.AsTime())
		require.Nil(t, q.Result.Data[i].Records.AsMap()["total_records"])
	}
}

func testMetricsViewTimeSeries_having_clause(t *testing.T, rt *runtime.Runtime, instanceID string) {
	q := &queries.MetricsViewTimeSeries{
		MeasureNames:    []string{"sum_imps"},
		MetricsViewName: "timeseries_gaps",
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
		Limit:          250,
		SecurityClaims: testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	rows := q.Result.Data
	require.Len(t, rows, 2)
	i := 0
	require.Equal(t, parseTime(t, "2019-01-01T00:00:00Z").AsTime(), rows[i].Ts.AsTime())
	require.NotNil(t, q.Result.Data[i].Records.AsMap()["sum_imps"])
	i++
	require.Equal(t, parseTime(t, "2019-01-06T00:00:00Z").AsTime(), rows[i].Ts.AsTime())
	require.NotNil(t, q.Result.Data[i].Records.AsMap()["sum_imps"])
}

func TestMetricsTimeseries_measure_filters_same_name(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

	ctr := &queries.ColumnTimeRange{
		TableName:  "ad_bids",
		ColumnName: "timestamp",
	}
	err := ctr.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)

	lmt := int64(25)
	q := &queries.MetricsViewTimeSeries{
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"bid_price"},
		TimeStart:       timestamppb.New(ctr.Result.Min.AsTime().Truncate(time.Hour)),
		TimeEnd:         ctr.Result.Max,
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_DAY,
		Having: &runtimev1.Expression{
			Expression: &runtimev1.Expression_Cond{
				Cond: &runtimev1.Condition{
					Op: runtimev1.Operation_OPERATION_LTE,
					Exprs: []*runtimev1.Expression{
						{
							Expression: &runtimev1.Expression_Ident{
								Ident: "bid_price",
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
		Limit:          lmt,
		SecurityClaims: testClaims(),
	}

	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	rows := q.Result.Data
	require.Len(t, rows, 25)
	i := 0
	require.Equal(t, parseTime(t, "2022-01-03T00:00:00Z").AsTime(), rows[i].Ts.AsTime())
	require.NotNil(t, q.Result.Data[i].Records.AsMap()["bid_price"])
	require.LessOrEqual(t, q.Result.Data[i].Records.AsMap()["bid_price"], 3.0)
	i++
	require.Equal(t, parseTime(t, "2022-01-04T00:00:00Z").AsTime(), rows[i].Ts.AsTime())
	require.NotNil(t, q.Result.Data[i].Records.AsMap()["bid_price"])
	require.LessOrEqual(t, q.Result.Data[i].Records.AsMap()["bid_price"], 3.0)
	i++
	require.Equal(t, parseTime(t, "2022-01-06T00:00:00Z").AsTime(), rows[i].Ts.AsTime())
	require.NotNil(t, q.Result.Data[i].Records.AsMap()["bid_price"])
	require.LessOrEqual(t, q.Result.Data[i].Records.AsMap()["bid_price"], 3.0)
	i++
	require.Equal(t, parseTime(t, "2022-01-07T00:00:00Z").AsTime(), rows[i].Ts.AsTime())
	require.NotNil(t, q.Result.Data[i].Records.AsMap()["bid_price"])
	require.LessOrEqual(t, q.Result.Data[i].Records.AsMap()["bid_price"], 3.0)
	i++
	require.Equal(t, parseTime(t, "2022-01-08T00:00:00Z").AsTime(), rows[i].Ts.AsTime())
	require.NotNil(t, q.Result.Data[i].Records.AsMap()["bid_price"])
	require.LessOrEqual(t, q.Result.Data[i].Records.AsMap()["bid_price"], 3.0)
	i++
	require.Equal(t, parseTime(t, "2022-01-09T00:00:00Z").AsTime(), rows[i].Ts.AsTime())
	require.NotNil(t, q.Result.Data[i].Records.AsMap()["bid_price"])
	require.LessOrEqual(t, q.Result.Data[i].Records.AsMap()["bid_price"], 3.0)
	i++
	require.Equal(t, parseTime(t, "2022-01-11T00:00:00Z").AsTime(), rows[i].Ts.AsTime())
	require.NotNil(t, q.Result.Data[i].Records.AsMap()["bid_price"])
	require.LessOrEqual(t, q.Result.Data[i].Records.AsMap()["bid_price"], 3.0)
	i++
	require.Equal(t, parseTime(t, "2022-01-12T00:00:00Z").AsTime(), rows[i].Ts.AsTime())
	require.NotNil(t, q.Result.Data[i].Records.AsMap()["bid_price"])
	require.LessOrEqual(t, q.Result.Data[i].Records.AsMap()["bid_price"], 3.0)
	i++
	require.Equal(t, parseTime(t, "2022-01-13T00:00:00Z").AsTime(), rows[i].Ts.AsTime())
	require.NotNil(t, q.Result.Data[i].Records.AsMap()["bid_price"])
	require.LessOrEqual(t, q.Result.Data[i].Records.AsMap()["bid_price"], 3.0)
	i++
	require.Equal(t, parseTime(t, "2022-01-15T00:00:00Z").AsTime(), rows[i].Ts.AsTime())
	require.NotNil(t, q.Result.Data[i].Records.AsMap()["bid_price"])
	require.LessOrEqual(t, q.Result.Data[i].Records.AsMap()["bid_price"], 3.0)
	i++
	require.Equal(t, parseTime(t, "2022-01-18T00:00:00Z").AsTime(), rows[i].Ts.AsTime())
	require.NotNil(t, q.Result.Data[i].Records.AsMap()["bid_price"])
	require.LessOrEqual(t, q.Result.Data[i].Records.AsMap()["bid_price"], 3.0)
	i++
	require.Equal(t, parseTime(t, "2022-01-21T00:00:00Z").AsTime(), rows[i].Ts.AsTime())
	require.NotNil(t, q.Result.Data[i].Records.AsMap()["bid_price"])
	require.LessOrEqual(t, q.Result.Data[i].Records.AsMap()["bid_price"], 3.0)
	i++
	require.Equal(t, parseTime(t, "2022-01-23T00:00:00Z").AsTime(), rows[i].Ts.AsTime())
	require.NotNil(t, q.Result.Data[i].Records.AsMap()["bid_price"])
	require.LessOrEqual(t, q.Result.Data[i].Records.AsMap()["bid_price"], 3.0)
	// check rest of the rows have bid_price < 3
	for ; i < len(rows); i++ {
		require.NotNil(t, q.Result.Data[i].Records.AsMap()["bid_price"])
		require.LessOrEqual(t, q.Result.Data[i].Records.AsMap()["bid_price"], 3.0)
	}
}

func toStructpbValue(t *testing.T, v any) *structpb.Value {
	sv, err := structpb.NewValue(v)
	require.NoError(t, err)
	return sv
}
