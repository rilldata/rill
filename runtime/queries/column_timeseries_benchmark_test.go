package queries_test

import (
	"context"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	_ "github.com/rilldata/rill/runtime/drivers/duckdb"
	"github.com/rilldata/rill/runtime/queries"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func BenchmarkTimeSeries_hourly(b *testing.B) {
	rt, instanceID := testruntime.NewInstanceForProject(b, "ad_bids")
	q := &queries.ColumnTimeseries{
		TableName:           "ad_bids",
		TimestampColumnName: "timestamp",
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Interval: runtimev1.TimeGrain_TIME_GRAIN_HOUR,
		},
		TimeZone: "Asia/Kathmandu",
		Measures: []*runtimev1.ColumnTimeSeriesRequest_BasicMeasure{
			{
				SqlName:    "avg_price",
				Expression: "avg(bid_price)",
			},
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := q.Resolve(context.Background(), rt, instanceID, 0)
		require.NoError(b, err)
		require.NotEmpty(b, q.Result)
	}
}

func BenchmarkTimeSeries_daily(b *testing.B) {
	rt, instanceID := testruntime.NewInstanceForProject(b, "ad_bids")
	q := &queries.ColumnTimeseries{
		TableName:           "ad_bids",
		TimestampColumnName: "timestamp",
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Interval: runtimev1.TimeGrain_TIME_GRAIN_DAY,
		},
		TimeZone: "Asia/Kathmandu",
		Measures: []*runtimev1.ColumnTimeSeriesRequest_BasicMeasure{
			{
				SqlName:    "avg_price",
				Expression: "avg(bid_price)",
			},
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := q.Resolve(context.Background(), rt, instanceID, 0)
		require.NoError(b, err)
		require.NotEmpty(b, q.Result)
	}
}

func BenchmarkTimeSeries_weekly(b *testing.B) {
	rt, instanceID := testruntime.NewInstanceForProject(b, "ad_bids")
	q := &queries.ColumnTimeseries{
		TableName:           "ad_bids",
		TimestampColumnName: "timestamp",
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Interval: runtimev1.TimeGrain_TIME_GRAIN_WEEK,
		},
		TimeZone: "Asia/Kathmandu",
		Measures: []*runtimev1.ColumnTimeSeriesRequest_BasicMeasure{
			{
				SqlName:    "avg_price",
				Expression: "avg(bid_price)",
			},
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := q.Resolve(context.Background(), rt, instanceID, 0)
		require.NoError(b, err)
		require.NotEmpty(b, q.Result)
	}
}

func BenchmarkTimeSeries_weekly_first_day_of_week_Monday(b *testing.B) {
	rt, instanceID := testruntime.NewInstanceForProject(b, "ad_bids")

	q := &queries.ColumnTimeseries{
		TableName:           "ad_bids",
		TimestampColumnName: "timestamp",
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Interval: runtimev1.TimeGrain_TIME_GRAIN_WEEK,
		},
		Measures: []*runtimev1.ColumnTimeSeriesRequest_BasicMeasure{
			{
				SqlName:    "avg_price",
				Expression: "avg(bid_price)",
			},
		},
		TimeZone: "Asia/Kathmandu",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := q.Resolve(context.Background(), rt, instanceID, 0)
		require.NoError(b, err)
		require.NotEmpty(b, q.Result)
	}
}

func BenchmarkTimeSeries_weekly_first_day_of_week_Sunday(b *testing.B) {
	rt, instanceID := testruntime.NewInstanceForProject(b, "ad_bids")

	q := &queries.ColumnTimeseries{
		TableName:           "ad_bids",
		TimestampColumnName: "timestamp",
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Interval: runtimev1.TimeGrain_TIME_GRAIN_WEEK,
		},
		Measures: []*runtimev1.ColumnTimeSeriesRequest_BasicMeasure{
			{
				SqlName:    "avg_price",
				Expression: "avg(bid_price)",
			},
		},
		FirstDayOfWeek: 7,
		TimeZone:       "Asia/Kathmandu",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := q.Resolve(context.Background(), rt, instanceID, 0)
		require.NoError(b, err)
		require.NotEmpty(b, q.Result)
	}
}
