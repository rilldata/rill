package server

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/marcboeker/go-duckdb"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
	structpb "google.golang.org/protobuf/types/known/structpb"
)

func CreateSimpleTimeseriesTable(server *Server, instanceId string, t *testing.T, tableName string) *drivers.Result {
	result, err := server.query(context.Background(), instanceId, &drivers.Statement{
		Query: "create table " + quoteName(tableName) + " (clicks double, time timestamp, device varchar)",
	})
	require.NoError(t, err)
	result.Close()
	result, err = server.query(context.Background(), instanceId, &drivers.Statement{
		Query: "insert into " + quoteName(tableName) + " values (1.0, '2019-01-01 00:00:00', 'android'), (1.0, '2019-01-02 00:00:00', 'iphone')",
	})
	require.NoError(t, err)
	result.Close()
	result, err = server.query(context.Background(), instanceId, &drivers.Statement{
		Query: "select count(*) from " + quoteName(tableName),
	})
	require.NoError(t, err)
	return result
}

func CreateTimeseriesTable(server *Server, instanceId string, t *testing.T, tableName string) *drivers.Result {
	result, err := server.query(context.Background(), instanceId, &drivers.Statement{
		Query: "create table " + quoteName(tableName) + " (clicks double, time timestamp, device varchar)",
	})
	require.NoError(t, err)
	result.Close()
	result, _ = server.query(context.Background(), instanceId, &drivers.Statement{
		Query: "insert into " + quoteName(tableName) + ` values 
		(1.0, '2019-01-01 00:00:00', 'android'), 
		(1.0, '2019-01-02 00:00:00', 'iphone'),
		(1.0, '2019-01-02 00:00:00', 'android'), 
		(1.0, '2019-01-03 00:00:00', 'iphone'),
		(1.0, '2019-01-04 00:00:00', 'android'), 

		(1.0, '2019-01-05 00:00:00', 'iphone'),
		(1.0, '2019-01-06 00:00:00', 'android'), 
		(1.0, '2019-01-07 00:00:00', 'iphone'),
		(1.0, '2019-01-07 00:00:00', 'android'), 
		(1.0, '2019-01-08 00:00:00', 'iphone'),
		`,
	})
	require.NoError(t, err)
	result.Close()
	result, err = server.query(context.Background(), instanceId, &drivers.Statement{
		Query: "select count(*) from " + quoteName(tableName),
	})
	require.NoError(t, err)
	return result
}

func TestServer_Timeseries(t *testing.T) {
	server, instanceId := getTestServer(t)

	result := CreateSimpleTimeseriesTable(server, instanceId, t, "timeseries")
	require.Equal(t, 2, getSingleValue(t, result.Rows))

	mx := "max"
	response, err := server.GenerateTimeSeries(context.Background(), &runtimev1.GenerateTimeSeriesRequest{
		InstanceId: instanceId,
		TableName:  "timeseries",
		Measures: &runtimev1.GenerateTimeSeriesRequest_BasicMeasures{
			BasicMeasures: []*runtimev1.BasicMeasureDefinition{
				{
					Expression: "max(clicks)",
					SqlName:    &mx,
				},
			},
		},
		TimestampColumnName: "time",
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Start:    parseTime(t, "2019-01-01T00:00:00Z"),
			End:      parseTime(t, "2019-12-01T00:00:00Z"),
			Interval: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		},
		Filters: &runtimev1.MetricsViewRequestFilter{
			Include: []*runtimev1.MetricsViewDimensionValue{
				{
					Name: "device",
					In:   []*structpb.Value{structpb.NewStringValue("android"), structpb.NewStringValue("iphone")},
				},
			},
		},
	})

	require.NoError(t, err)
	results := response.GetRollup().Results
	// printResults(results)
	require.Equal(t, 1, len(results))
	require.Equal(t, 1.0, results[0].Records["max"])
}

func TestServer_Timeseries_2measures(t *testing.T) {
	server, instanceId := getTestServer(t)

	result := CreateSimpleTimeseriesTable(server, instanceId, t, "timeseries")
	require.Equal(t, 2, getSingleValue(t, result.Rows))

	mx := "max"
	sm := "sum"
	response, err := server.GenerateTimeSeries(context.Background(), &runtimev1.GenerateTimeSeriesRequest{
		InstanceId: instanceId,
		TableName:  "timeseries",
		Measures: &runtimev1.GenerateTimeSeriesRequest_BasicMeasures{
			BasicMeasures: []*runtimev1.BasicMeasureDefinition{
				{
					Expression: "max(clicks)",
					SqlName:    &mx,
				},
				{
					Expression: "sum(clicks)",
					SqlName:    &sm,
				},
			},
		},
		TimestampColumnName: "time",
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Start:    parseTime(t, "2019-01-01T00:00:00Z"),
			End:      parseTime(t, "2019-12-01T00:00:00Z"),
			Interval: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		},
		Filters: &runtimev1.MetricsViewRequestFilter{
			Include: []*runtimev1.MetricsViewDimensionValue{
				{
					Name: "device",
					In:   []*structpb.Value{structpb.NewStringValue("android"), structpb.NewStringValue("iphone")},
				},
			},
		},
	})

	require.NoError(t, err)
	results := response.GetRollup().Results
	// printResults(results)
	require.Equal(t, 1, len(results))
	require.Equal(t, 1.0, results[0].Records["max"])
	require.Equal(t, 2.0, results[0].Records["sum"])
}

func TestServer_Timeseries_1dim(t *testing.T) {
	server, instanceId := getTestServer(t)

	result := CreateSimpleTimeseriesTable(server, instanceId, t, "timeseries")
	require.Equal(t, 2, getSingleValue(t, result.Rows))

	sm := "sum"
	response, err := server.GenerateTimeSeries(context.Background(), &runtimev1.GenerateTimeSeriesRequest{
		InstanceId: instanceId,
		TableName:  "timeseries",
		Measures: &runtimev1.GenerateTimeSeriesRequest_BasicMeasures{
			BasicMeasures: []*runtimev1.BasicMeasureDefinition{
				{
					Expression: "sum(clicks)",
					SqlName:    &sm,
				},
			},
		},
		TimestampColumnName: "time",
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Start:    parseTime(t, "2019-01-01T00:00:00Z"),
			End:      parseTime(t, "2019-12-01T00:00:00Z"),
			Interval: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		},
		Filters: &runtimev1.MetricsViewRequestFilter{
			Include: []*runtimev1.MetricsViewDimensionValue{
				{
					Name: "device",
					In:   []*structpb.Value{structpb.NewStringValue("android")},
				},
			},
		},
	})

	require.NoError(t, err)
	results := response.GetRollup().Results
	// printResults(results)
	require.Equal(t, 1, len(results))
	require.Equal(t, 1.0, results[0].Records["sum"])
}

func printResults(results []*runtimev1.TimeSeriesValue) {
	for _, result := range results {
		fmt.Printf("%v ", result.Ts)
		for k, value := range result.Records {
			fmt.Printf("%v:%v ", k, value)
		}
		fmt.Println()
	}
}

func TestServer_Timeseries_1day(t *testing.T) {
	server, instanceId := getTestServer(t)

	result := CreateSimpleTimeseriesTable(server, instanceId, t, "timeseries")
	require.Equal(t, 2, getSingleValue(t, result.Rows))

	mx := "max"
	response, err := server.GenerateTimeSeries(context.Background(), &runtimev1.GenerateTimeSeriesRequest{
		InstanceId: instanceId,
		TableName:  "timeseries",
		Measures: &runtimev1.GenerateTimeSeriesRequest_BasicMeasures{
			BasicMeasures: []*runtimev1.BasicMeasureDefinition{
				{
					Expression: "max(clicks)",
					SqlName:    &mx,
				},
			},
		},
		TimestampColumnName: "time",
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Start:    parseTime(t, "2019-01-01T00:00:00Z"),
			End:      parseTime(t, "2019-01-02T00:00:00Z"),
			Interval: runtimev1.TimeGrain_TIME_GRAIN_DAY,
		},
		Filters: &runtimev1.MetricsViewRequestFilter{
			Include: []*runtimev1.MetricsViewDimensionValue{
				{
					Name: "device",
					In:   []*structpb.Value{structpb.NewStringValue("android"), structpb.NewStringValue("iphone")},
				},
			},
		},
	})

	require.NoError(t, err)
	results := response.GetRollup().Results
	require.Equal(t, 2, len(results))
}

func TestServer_Timeseries_1day_Count(t *testing.T) {
	server, instanceId := getTestServer(t)

	result := CreateSimpleTimeseriesTable(server, instanceId, t, "timeseries")
	require.Equal(t, 2, getSingleValue(t, result.Rows))

	cnt := "count"
	response, err := server.GenerateTimeSeries(context.Background(), &runtimev1.GenerateTimeSeriesRequest{
		InstanceId: instanceId,
		TableName:  "timeseries",
		Measures: &runtimev1.GenerateTimeSeriesRequest_BasicMeasures{
			BasicMeasures: []*runtimev1.BasicMeasureDefinition{
				{
					Expression: "count(*)",
					SqlName:    &cnt,
				},
			},
		},
		TimestampColumnName: "time",
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Start:    parseTime(t, "2019-01-01T00:00:00Z"),
			End:      parseTime(t, "2019-01-02T00:00:00Z"),
			Interval: runtimev1.TimeGrain_TIME_GRAIN_DAY,
		},
		Filters: &runtimev1.MetricsViewRequestFilter{
			Include: []*runtimev1.MetricsViewDimensionValue{
				{
					Name: "device",
					In:   []*structpb.Value{structpb.NewStringValue("android"), structpb.NewStringValue("iphone")},
				},
			},
		},
	})

	require.NoError(t, err)
	results := response.GetRollup().Results
	require.Equal(t, 2, len(results))
	require.Equal(t, 1.0, results[0].Records["count"])
}

func TestServer_RangeSanity(t *testing.T) {
	server, instanceId := getTestServer(t)

	result := CreateSimpleTimeseriesTable(server, instanceId, t, "timeseries")
	result.Close()
	result, err := server.query(context.Background(), instanceId, &drivers.Statement{
		Query: "select min(time) min, max(time) max, max(time)-min(time) as r from timeseries",
	})
	require.NoError(t, err)
	var min, max time.Time
	var r duckdb.Interval
	result.Next()
	err = result.Scan(&min, &max, &r)
	require.NoError(t, err)
	require.Equal(t, time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC), min)
	require.Equal(t, int32(1), r.Days)
}

func TestServer_normaliseTimeRange(t *testing.T) {
	server, instanceId := getTestServer(t)

	result := CreateSimpleTimeseriesTable(server, instanceId, t, "timeseries")
	require.Equal(t, 2, getSingleValue(t, result.Rows))
	r := &runtimev1.TimeSeriesTimeRange{
		Interval: runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED,
	}
	r, err := server.normaliseTimeRange(context.Background(), &runtimev1.GenerateTimeSeriesRequest{
		InstanceId:          instanceId,
		TimeRange:           r,
		TableName:           "timeseries",
		TimestampColumnName: "time",
	})
	require.NoError(t, err)
	require.Equal(t, parseTime(t, "2019-01-01T00:00:00.000Z"), r.Start)
	require.Equal(t, parseTime(t, "2019-01-02T00:00:00.000Z"), r.End)
	require.Equal(t, runtimev1.TimeGrain_TIME_GRAIN_HOUR, r.Interval)
}

func TestServer_normaliseTimeRange_NoEnd(t *testing.T) {
	server, instanceId := getTestServer(t)

	result := CreateSimpleTimeseriesTable(server, instanceId, t, "timeseries")
	require.Equal(t, 2, getSingleValue(t, result.Rows))
	r := &runtimev1.TimeSeriesTimeRange{
		Interval: runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED,
		Start:    parseTime(t, "2018-01-01T00:00:00Z"),
	}
	r, err := server.normaliseTimeRange(context.Background(), &runtimev1.GenerateTimeSeriesRequest{
		InstanceId:          instanceId,
		TimeRange:           r,
		TableName:           "timeseries",
		TimestampColumnName: "time",
	})
	require.NoError(t, err)
	require.Equal(t, parseTime(t, "2018-01-01T00:00:00Z"), r.Start)
	require.Equal(t, parseTime(t, "2019-01-02T00:00:00.000Z"), r.End)
	require.Equal(t, runtimev1.TimeGrain_TIME_GRAIN_HOUR, r.Interval)
}

func TestServer_normaliseTimeRange_Specified(t *testing.T) {
	server, instanceId := getTestServer(t)

	result := CreateSimpleTimeseriesTable(server, instanceId, t, "timeseries")
	require.Equal(t, 2, getSingleValue(t, result.Rows))
	r := &runtimev1.TimeSeriesTimeRange{
		Interval: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		Start:    parseTime(t, "2018-01-01T00:00:00Z"),
	}
	r, err := server.normaliseTimeRange(context.Background(), &runtimev1.GenerateTimeSeriesRequest{
		InstanceId:          instanceId,
		TimeRange:           r,
		TableName:           "timeseries",
		TimestampColumnName: "time",
	})
	require.NoError(t, err)
	require.Equal(t, parseTime(t, "2018-01-01T00:00:00Z"), r.Start)
	require.Equal(t, parseTime(t, "2019-01-02T00:00:00.000Z"), r.End)
	require.Equal(t, runtimev1.TimeGrain_TIME_GRAIN_YEAR, r.Interval)
}

func CreateAggregatedTableForSpark(server *Server, instanceId string, t *testing.T, tableName string) *drivers.Result {
	result, err := server.query(context.Background(), instanceId, &drivers.Statement{
		Query: "create table " + quoteName(tableName) + " (clicks double, time timestamp, device varchar)", // todo device is redundant - ie remove
	})
	require.NoError(t, err)
	result.Close()
	result, err = server.query(context.Background(), instanceId, &drivers.Statement{
		Query: "insert into " + quoteName(tableName) + ` values 
		(2.0, '2019-01-01T00:00:00Z', 'android'), 
		(3.0, '2019-01-02T00:00:00Z', 'iphone'),
		(1.0, '2019-01-03T00:00:00Z', 'iphone'),
		(2.0, '2019-01-04T00:00:00Z', 'android'), 

		(2.0, '2019-01-05T00:00:00Z', 'iphone'),
		(1.0, '2019-01-06T00:00:00Z', 'android'), 
		(4.0, '2019-01-07T00:00:00Z', 'android'), 
		(3, '2019-01-08T00:00:00Z', 'iphone'),

		(1.0, '2019-01-09T00:00:00Z', 'iphone'),
		`,
	})
	require.NoError(t, err)
	result.Close()
	result, err = server.query(context.Background(), instanceId, &drivers.Statement{
		Query: "select count(*) from " + quoteName(tableName),
	})
	require.NoError(t, err)
	return result
}

func TestServer_SparkOnly(t *testing.T) {
	time.Local = time.UTC
	server, instanceId := getTestServer(t)

	result := CreateAggregatedTableForSpark(server, instanceId, t, "timeseries")
	require.Equal(t, 9, getSingleValue(t, result.Rows))
	values, err := server.createTimestampRollupReduction(context.Background(), instanceId, "timeseries", "time", "clicks", 2.0)
	require.NoError(t, err)

	require.Equal(t, "2019-01-01T00:00:00.000Z", values[0].Ts)
	require.Equal(t, "2019-01-02T00:00:00.000Z", values[1].Ts)
	require.Equal(t, "2019-01-03T00:00:00.000Z", values[2].Ts)
	require.Equal(t, "2019-01-04T00:00:00.000Z", values[3].Ts)
	require.Equal(t, "2019-01-05T00:00:00.000Z", values[4].Ts)
	require.Equal(t, "2019-01-06T00:00:00.000Z", values[5].Ts)
	require.Equal(t, "2019-01-07T00:00:00.000Z", values[6].Ts)
	require.Equal(t, "2019-01-08T00:00:00.000Z", values[7].Ts)

	require.Equal(t, 0.0, *values[0].Bin)
	require.Equal(t, 0.0, *values[1].Bin)
	require.Equal(t, 0.0, *values[2].Bin)
	require.Equal(t, 0.0, *values[3].Bin)
	require.Equal(t, 1.0, *values[4].Bin)
	require.Equal(t, 1.0, *values[5].Bin)
	require.Equal(t, 1.0, *values[6].Bin)
	require.Equal(t, 1.0, *values[7].Bin)
	require.Equal(t, 2.0, *values[8].Bin)

	require.Equal(t, 2.0, values[0].Records["count"])
	require.Equal(t, 3.0, values[1].Records["count"])
	require.Equal(t, 1.0, values[2].Records["count"])
	require.Equal(t, 2.0, values[3].Records["count"])
	require.Equal(t, 2.0, values[4].Records["count"])
	require.Equal(t, 1.0, values[5].Records["count"])
	require.Equal(t, 4.0, values[6].Records["count"])
	require.Equal(t, 3.0, values[7].Records["count"])
	require.Equal(t, 1.0, values[8].Records["count"])
}
