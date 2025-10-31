package queries_test

import (
	"context"
	"fmt"
	"math"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/queries"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/rilldata/rill/runtime/testruntime/testmode"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/clickhouse"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgainstClickHouse(t *testing.T) {
	testmode.Expensive(t)

	ctx := context.Background()
	clickHouseContainer, err := clickhouse.RunContainer(ctx,
		testcontainers.WithImage("clickhouse/clickhouse-server:latest"),
		clickhouse.WithUsername("clickhouse"),
		clickhouse.WithPassword("clickhouse"),
		clickhouse.WithConfigFile("../testruntime/testdata/clickhouse-config.xml"),
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		err := clickHouseContainer.Terminate(ctx)
		require.NoError(t, err)
	})

	host, err := clickHouseContainer.Host(ctx)
	require.NoError(t, err)
	port, err := clickHouseContainer.MappedPort(ctx, "9000/tcp")
	require.NoError(t, err)

	t.Setenv("RILL_RUNTIME_TEST_OLAP_DRIVER", "clickhouse")
	t.Setenv("RILL_RUNTIME_TEST_OLAP_DSN", fmt.Sprintf("clickhouse://clickhouse:clickhouse@%v:%v", host, port.Port()))
	t.Run("TestTimeseries_normaliseTimeRange", func(t *testing.T) { TestTimeseries_normaliseTimeRange(t) })
	t.Run("TestTimeseries_normaliseTimeRange_NoEnd", func(t *testing.T) { TestTimeseries_normaliseTimeRange_NoEnd(t) })
	t.Run("TestTimeseries_normaliseTimeRange_Specified", func(t *testing.T) { TestTimeseries_normaliseTimeRange_Specified(t) })
	t.Run("TestTimeseries_SparkOnly", func(t *testing.T) { TestTimeseries_SparkOnly(t) })
	t.Run("TestTimeseries_FirstDayOfWeek_Monday", func(t *testing.T) { TestTimeseries_FirstDayOfWeek_Monday(t) })
	t.Run("TestTimeseries_FirstDayOfWeek_Sunday", func(t *testing.T) { TestTimeseries_FirstDayOfWeek_Sunday(t) })
	// t.Run("TestTimeseries_FirstDayOfWeek_Sunday_OnSunday", func(t *testing.T) { TestTimeseries_FirstDayOfWeek_Sunday_OnSunday(t) })
	t.Run("TestTimeseries_FirstDayOfWeek_Saturday", func(t *testing.T) { TestTimeseries_FirstDayOfWeek_Saturday(t) })
	t.Run("TestTimeseries_FirstMonthOfYear_January", func(t *testing.T) { TestTimeseries_FirstMonthOfYear_January(t) })
	// t.Run("TestTimeseries_FirstMonthOfYear_March", func(t *testing.T) { TestTimeseries_FirstMonthOfYear_March(t) })
	t.Run("TestTimeseries_FirstMonthOfYear_December", func(t *testing.T) { TestTimeseries_FirstMonthOfYear_December(t) })
	// t.Run("TestTimeseries_FirstMonthOfYear_December_InDecember", func(t *testing.T) { TestTimeseries_FirstMonthOfYear_December_InDecember(t) })
}

func instanceWith2RowsModel(t *testing.T) (*runtime.Runtime, string) {
	rt, instanceID := testruntime.NewInstanceWithModel(t, "test", `
		SELECT 1.0 AS clicks, TIMESTAMP '2019-01-01 00:00:00' AS time, DATE '2019-01-01' as day, 'android' AS device, 'Google' AS publisher, 'google.com' AS domain
		UNION ALL
		SELECT 1.0 AS clicks, TIMESTAMP '2019-01-02 00:00:00' AS time, DATE '2019-01-02' as day, 'iphone' AS device, null AS publisher, 'msn.com' AS domain
	`)
	return rt, instanceID
}

func instanceWith1RowModel(t *testing.T) (*runtime.Runtime, string) {
	rt, instanceID := testruntime.NewInstanceWithModel(t, "test", `
		SELECT 1.0 AS clicks, TIMESTAMP '2023-10-03 00:00:00' AS time, DATE '2019-01-01' as day, 'android' AS device, 'Google' AS publisher, 'google.com' AS domain
	`)
	return rt, instanceID
}

func instanceWith1RowModelWithTime(t *testing.T, tm string) (*runtime.Runtime, string) {
	rt, instanceID := testruntime.NewInstanceWithModel(t, "test", `
		SELECT 1.0 AS clicks, TIMESTAMP '`+tm+`' AS time, DATE '2019-01-01' as day, 'android' AS device, 'Google' AS publisher, 'google.com' AS domain
	`)
	return rt, instanceID
}

func instanceWithSparkModel(t *testing.T) (*runtime.Runtime, string) {
	rt, instanceID := testruntime.NewInstanceWithModel(t, "test", `
		SELECT 1.0 AS clicks, TIMESTAMP '2019-01-01 00:00:00' AS time, 'android' AS device
		UNION ALL
		SELECT 2.0 AS clicks, TIMESTAMP '2019-01-02 00:00:00' AS time, 'iphone' AS device
		UNION ALL
		SELECT 3.0 AS clicks, TIMESTAMP '2019-01-03 00:00:00' AS time, 'iphone' AS device
		UNION ALL
		SELECT 4.0 AS clicks, TIMESTAMP '2019-01-04 00:00:00' AS time, 'android' AS device
		UNION ALL
		SELECT 5.0 AS clicks, TIMESTAMP '2019-01-05 00:00:00' AS time, 'iphone' AS device
		UNION ALL
		SELECT 4.5 AS clicks, TIMESTAMP '2019-01-06 00:00:00' AS time, 'android' AS device
		UNION ALL
		SELECT 3.5 AS clicks, TIMESTAMP '2019-01-07 00:00:00' AS time, 'android' AS device
		UNION ALL
		SELECT 2.5 AS clicks, TIMESTAMP '2019-01-08 00:00:00' AS time, 'iphone' AS device
		UNION ALL
		SELECT 1.5 AS clicks, TIMESTAMP '2019-01-09 00:00:00' AS time, 'iphone' AS device
	`)
	return rt, instanceID
}

func instanceWithSparkSameTimestampModel(t *testing.T) (*runtime.Runtime, string) {
	rt, instanceID := testruntime.NewInstanceWithModel(t, "test", `
		SELECT 1.0 AS clicks, TIMESTAMP '2019-01-01T00:00:00Z' AS time, 'android' AS device
		UNION ALL
		SELECT 1.0 AS clicks, TIMESTAMP '2019-01-01T00:00:00Z' AS time, 'android' AS device
		UNION ALL
		SELECT 1.0 AS clicks, TIMESTAMP '2019-01-01T00:00:00Z' AS time, 'android' AS device
		UNION ALL
		SELECT 1.0 AS clicks, TIMESTAMP '2019-01-01T00:00:00Z' AS time, 'android' AS device
		UNION ALL
		SELECT 1.0 AS clicks, TIMESTAMP '2019-01-01T00:00:00Z' AS time, 'android' AS device
		UNION ALL
		SELECT 1.0 AS clicks, TIMESTAMP '2019-01-01T00:00:00Z' AS time, 'android' AS device
		UNION ALL
		SELECT 1.0 AS clicks, TIMESTAMP '2019-01-01T00:00:00Z' AS time, 'android' AS device
		UNION ALL
		SELECT 1.0 AS clicks, TIMESTAMP '2019-01-01T00:00:00Z' AS time, 'android' AS device
	`)
	return rt, instanceID
}

func TestTimeseries_normaliseTimeRange(t *testing.T) {
	rt, instanceID := instanceWith2RowsModel(t)

	q := &queries.ColumnTimeseries{
		TableName:           "test",
		TimestampColumnName: "time",
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Interval: runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED,
		},
	}
	tr, err := q.ResolveNormaliseTimeRange(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.Equal(t, parseTime(t, "2019-01-01T00:00:00.000Z").AsTime(), tr.Start.AsTime())
	require.Equal(t, parseTime(t, "2019-01-02T01:00:00.000Z").AsTime(), tr.End.AsTime())
	require.Equal(t, runtimev1.TimeGrain_TIME_GRAIN_HOUR, tr.Interval)
}

func TestTimeseries_normaliseTimeRange_NoEnd(t *testing.T) {
	rt, instanceID := instanceWith2RowsModel(t)

	q := &queries.ColumnTimeseries{
		TableName:           "test",
		TimestampColumnName: "time",
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Interval: runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED,
			Start:    parseTime(t, "2018-01-01T00:00:00Z"),
		},
	}

	r, err := q.ResolveNormaliseTimeRange(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.Equal(t, parseTime(t, "2018-01-01T00:00:00Z").AsTime(), r.Start.AsTime())
	require.Equal(t, parseTime(t, "2019-01-02T01:00:00.000Z").AsTime(), r.End.AsTime())
	require.Equal(t, runtimev1.TimeGrain_TIME_GRAIN_HOUR, r.Interval)
}

func TestTimeseries_normaliseTimeRange_Specified(t *testing.T) {
	rt, instanceID := instanceWith2RowsModel(t)

	q := &queries.ColumnTimeseries{
		TableName:           "test",
		TimestampColumnName: "time",
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Interval: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
			Start:    parseTime(t, "2018-01-01T00:00:00Z"),
		},
	}

	r, err := q.ResolveNormaliseTimeRange(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.Equal(t, parseTime(t, "2018-01-01T00:00:00Z").AsTime(), r.Start.AsTime())
	require.Equal(t, parseTime(t, "2020-01-01T00:00:00.000Z").AsTime(), r.End.AsTime())
	require.Equal(t, runtimev1.TimeGrain_TIME_GRAIN_YEAR, r.Interval)
}

func TestTimeseries_SparkOnly_same_timestamp(t *testing.T) {
	testmode.Expensive(t)

	time.Local = time.UTC

	rt, instanceID := instanceWithSparkSameTimestampModel(t)

	q := &queries.ColumnTimeseries{
		TableName:           "test",
		TimestampColumnName: "time",
		Pixels:              2.0,
	}
	ctx := context.Background()
	olap, release, err := rt.OLAP(ctx, instanceID, q.Connector)
	require.NoError(t, err)
	defer release()
	values, err := q.CreateTimestampRollupReduction(context.Background(), rt, olap, instanceID, 0, "test", "time", "clicks")
	require.NoError(t, err)
	require.True(t, math.IsNaN(values[0].Bin))
}

func TestTimeseries_SparkOnly(t *testing.T) {
	testmode.Expensive(t)

	time.Local = time.UTC

	rt, instanceID := instanceWithSparkModel(t)

	q := &queries.ColumnTimeseries{
		TableName:           "test",
		TimestampColumnName: "time",
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Interval: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
			Start:    parseTime(t, "2018-01-01T00:00:00Z"),
		},
		Pixels: 2.0,
	}
	ctx := context.Background()
	olap, release, err := rt.OLAP(ctx, instanceID, q.Connector)
	require.NoError(t, err)
	defer release()
	values, err := q.CreateTimestampRollupReduction(context.Background(), rt, olap, instanceID, 0, "test", "time", "clicks")
	require.NoError(t, err)

	// for i := 0; i < 12; i++ {
	// 	fmt.Println(fmt.Sprintf("%s %.1f %.1f", values[i].Ts.AsTime(), values[i].Bin, values[i].Records.Fields["count"].GetNumberValue()))
	// }

	require.Equal(t, 12, len(values))
	require.Equal(t, parseTime(t, "2019-01-01T00:00:00Z"), values[0].Ts)
	require.Equal(t, parseTime(t, "2019-01-01T00:00:00Z"), values[1].Ts)
	require.Equal(t, parseTime(t, "2019-01-02T00:00:00Z"), values[2].Ts)
	require.Equal(t, parseTime(t, "2019-01-02T00:00:00Z"), values[3].Ts)
	require.Equal(t, parseTime(t, "2019-01-03T00:00:00Z"), values[4].Ts)
	require.Equal(t, parseTime(t, "2019-01-03T00:00:00Z"), values[5].Ts)
	require.Equal(t, parseTime(t, "2019-01-05T00:00:00Z"), values[6].Ts)
	require.Equal(t, parseTime(t, "2019-01-06T00:00:00Z"), values[7].Ts)
	require.Equal(t, parseTime(t, "2019-01-07T00:00:00Z"), values[8].Ts)
	require.Equal(t, parseTime(t, "2019-01-07T00:00:00Z"), values[9].Ts)
	require.Equal(t, parseTime(t, "2019-01-09T00:00:00Z"), values[10].Ts)
	require.Equal(t, parseTime(t, "2019-01-09T00:00:00Z"), values[11].Ts)

	require.Equal(t, 0.0, values[0].Bin)
	require.Equal(t, 0.0, values[1].Bin)
	require.Equal(t, 0.0, values[2].Bin)
	require.Equal(t, 0.0, values[3].Bin)
	require.Equal(t, 1.0, values[4].Bin)
	require.Equal(t, 1.0, values[5].Bin)
	require.Equal(t, 1.0, values[6].Bin)
	require.Equal(t, 1.0, values[7].Bin)
	require.Equal(t, 2.0, values[8].Bin)
	require.Equal(t, 2.0, values[9].Bin)
	require.Equal(t, 2.0, values[10].Bin)
	require.Equal(t, 2.0, values[11].Bin)

	require.Equal(t, 1.0, values[0].Records.Fields["count"].GetNumberValue())
	require.Equal(t, 1.0, values[1].Records.Fields["count"].GetNumberValue())
	require.Equal(t, 2.0, values[2].Records.Fields["count"].GetNumberValue())
	require.Equal(t, 2.0, values[3].Records.Fields["count"].GetNumberValue())
	require.Equal(t, 3.0, values[4].Records.Fields["count"].GetNumberValue())
	require.Equal(t, 3.0, values[5].Records.Fields["count"].GetNumberValue())
	require.Equal(t, 5.0, values[6].Records.Fields["count"].GetNumberValue())
	require.Equal(t, 4.5, values[7].Records.Fields["count"].GetNumberValue())
	require.Equal(t, 3.5, values[8].Records.Fields["count"].GetNumberValue())
	require.Equal(t, 3.5, values[9].Records.Fields["count"].GetNumberValue())
	require.Equal(t, 1.5, values[10].Records.Fields["count"].GetNumberValue())
	require.Equal(t, 1.5, values[11].Records.Fields["count"].GetNumberValue())
}

func TestTimeseries_Key(t *testing.T) {
	q := &queries.ColumnTimeseries{
		TableName:           "test",
		TimestampColumnName: "time",
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Interval: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
			Start:    parseTime(t, "2018-01-01T00:00:00Z"),
		},
		Pixels: 2.0,
	}

	k1 := q.Key()
	assert.NotEmpty(t, k1)

	q = &queries.ColumnTimeseries{
		TableName:           "test",
		TimestampColumnName: "time",
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Interval: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
			Start:    parseTime(t, "2018-01-02T00:00:00Z"),
		},
		Pixels: 2.0,
	}

	k2 := q.Key()
	assert.NotEmpty(t, k2)

	assert.NotEqual(t, k1, k2)
}

func TestTimeseries_FirstDayOfWeek_Monday(t *testing.T) {
	rt, instanceID := instanceWith1RowModel(t)

	q := &queries.ColumnTimeseries{
		TableName:           "test",
		TimestampColumnName: "time",
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Interval: runtimev1.TimeGrain_TIME_GRAIN_WEEK,
		},
		FirstDayOfWeek: 1,
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.Equal(t, parseTime(t, "2023-10-02T00:00:00.000Z").AsTime(), q.Result.Results[0].Ts.AsTime())
}

func TestTimeseries_FirstDayOfWeek_Sunday(t *testing.T) {
	rt, instanceID := instanceWith1RowModel(t)

	q := &queries.ColumnTimeseries{
		TableName:           "test",
		TimestampColumnName: "time",
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Interval: runtimev1.TimeGrain_TIME_GRAIN_WEEK,
		},
		FirstDayOfWeek: 7,
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.Equal(t, parseTime(t, "2023-10-01T00:00:00.000Z").AsTime(), q.Result.Results[0].Ts.AsTime())
}

func TestTimeseries_FirstDayOfWeek_Sunday_OnSunday(t *testing.T) {
	rt, instanceID := instanceWith1RowModelWithTime(t, "2023-10-01 00:00:00")

	q := &queries.ColumnTimeseries{
		TableName:           "test",
		TimestampColumnName: "time",
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Interval: runtimev1.TimeGrain_TIME_GRAIN_WEEK,
		},
		FirstDayOfWeek: 7,
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.Equal(t, 1, len(q.Result.Results))
	require.Equal(t, parseTime(t, "2023-10-01T00:00:00.000Z").AsTime(), q.Result.Results[0].Ts.AsTime())
}

func TestTimeseries_FirstDayOfWeek_Saturday(t *testing.T) {
	rt, instanceID := instanceWith1RowModel(t)

	q := &queries.ColumnTimeseries{
		TableName:           "test",
		TimestampColumnName: "time",
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Interval: runtimev1.TimeGrain_TIME_GRAIN_WEEK,
		},
		FirstDayOfWeek: 6,
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.Equal(t, parseTime(t, "2023-09-30T00:00:00.000Z").AsTime(), q.Result.Results[0].Ts.AsTime())
}

func TestTimeseries_FirstMonthOfYear_January(t *testing.T) {
	rt, instanceID := instanceWith1RowModel(t)

	q := &queries.ColumnTimeseries{
		TableName:           "test",
		TimestampColumnName: "time",
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Interval: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		},
		FirstMonthOfYear: 1,
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.Equal(t, parseTime(t, "2023-01-01T00:00:00.000Z").AsTime(), q.Result.Results[0].Ts.AsTime())
}

func TestTimeseries_FirstMonthOfYear_March(t *testing.T) {
	rt, instanceID := instanceWith1RowModel(t)

	q := &queries.ColumnTimeseries{
		TableName:           "test",
		TimestampColumnName: "time",
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Interval: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		},
		FirstMonthOfYear: 3,
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.Equal(t, 1, len(q.Result.Results))
	require.Equal(t, parseTime(t, "2023-03-01T00:00:00.000Z").AsTime(), q.Result.Results[0].Ts.AsTime())
}

func TestTimeseries_FirstMonthOfYear_December(t *testing.T) {
	rt, instanceID := instanceWith1RowModel(t)

	q := &queries.ColumnTimeseries{
		TableName:           "test",
		TimestampColumnName: "time",
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Interval: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		},
		FirstMonthOfYear: 12,
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.Equal(t, parseTime(t, "2022-12-01T00:00:00.000Z").AsTime(), q.Result.Results[0].Ts.AsTime())
}

func TestTimeseries_FirstMonthOfYear_December_InDecember(t *testing.T) {
	rt, instanceID := instanceWith1RowModelWithTime(t, "2023-12-04 00:00:00")

	q := &queries.ColumnTimeseries{
		TableName:           "test",
		TimestampColumnName: "time",
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Interval: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		},
		FirstMonthOfYear: 12,
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.Equal(t, 1, len(q.Result.Results))
	require.Equal(t, parseTime(t, "2023-12-01T00:00:00.000Z").AsTime(), q.Result.Results[0].Ts.AsTime())
}

func parseTime(tst *testing.T, t string) *timestamppb.Timestamp {
	ts, err := time.Parse(time.RFC3339, t)
	require.NoError(tst, err)
	return timestamppb.New(ts)
}

func parseTimeB(tst *testing.B, t string) *timestamppb.Timestamp {
	ts, err := time.Parse(time.RFC3339, t)
	require.NoError(tst, err)
	return timestamppb.New(ts)
}
