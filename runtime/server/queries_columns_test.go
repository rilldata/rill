package server

import (
	"context"
	"fmt"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	_ "github.com/rilldata/rill/runtime/drivers/duckdb"
	"github.com/rilldata/rill/runtime/queries"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

func TestServer_GetTopK_HugeInt(t *testing.T) {
	server, instanceId := getColumnTestServerWithModel(t, "select 170141183460469231731687303715884105727::hugeint as metric, 'a' as dim", 1)

	res, err := server.ColumnTopK(
		testCtx(),
		&runtimev1.ColumnTopKRequest{
			InstanceId: instanceId,
			TableName:  "test",
			ColumnName: "dim",
			Agg:        "sum(metric)",
		},
	)
	require.NoError(t, err)
	require.NotEmpty(t, res)
	topk := res.CategoricalSummary.GetTopK()
	require.Equal(t, 1, len(topk.Entries))
	require.Equal(t, "a", topk.Entries[0].Value.GetStringValue())
	require.True(t, topk.Entries[0].Count >= 170141183460469231731687303715884105727.0)
}

func TestServer_GetTopK_1dim_HugeInt(t *testing.T) {
	server, instanceId := getColumnTestServerWithModel(t, "select 170141183460469231731687303715884105727::hugeint as metric", 1)

	res, err := server.ColumnTopK(
		testCtx(),
		&runtimev1.ColumnTopKRequest{
			InstanceId: instanceId,
			TableName:  "test",
			ColumnName: "metric",
		},
	)
	require.NoError(t, err)
	require.NotEmpty(t, res)
	topk := res.CategoricalSummary.GetTopK()
	require.Equal(t, 1, len(topk.Entries))
	require.Equal(t, 170141183460469231731687303715884105727.0, topk.Entries[0].Value.GetNumberValue())
	require.Equal(t, 1.0, topk.Entries[0].Count)
}

func TestServer_GetTopK(t *testing.T) {
	server, instanceId := getColumnTestServerWithModel(
		t,
		`
		SELECT 'abc' AS col, 1 AS val, TIMESTAMP '2022-11-01 00:00:00' AS times, DATE '2007-04-01' AS dates
		UNION ALL 
		SELECT 'def' AS col, 5 AS val, TIMESTAMP '2022-11-02 00:00:00' AS times, DATE '2009-06-01' AS dates
		UNION ALL 
		SELECT 'abc' AS col, 3 AS val, TIMESTAMP '2022-11-03 00:00:00' AS times, DATE '2010-04-11' AS dates
		UNION ALL 
		SELECT null AS col, 1 AS val, TIMESTAMP '2022-11-03 00:00:00' AS times, DATE '2010-11-21' AS dates
		UNION ALL 
		SELECT 12 AS col, 1 AS val, TIMESTAMP '2022-11-03 00:00:00' AS times, DATE '2011-06-30' AS dates
		`,
		5,
	)

	res, err := server.ColumnTopK(
		testCtx(),
		&runtimev1.ColumnTopKRequest{
			InstanceId: instanceId,
			TableName:  "test",
			ColumnName: "col",
		},
	)
	require.NoError(t, err)
	require.NotEmpty(t, res)
	topk := res.CategoricalSummary.GetTopK()
	require.Equal(t, 4, len(topk.Entries))
	require.Equal(t, "abc", topk.Entries[0].Value.GetStringValue())
	require.Equal(t, 2, int(topk.Entries[0].Count))
	require.Equal(t, structpb.NewNullValue(), topk.Entries[1].Value)
	require.Equal(t, 1, int(topk.Entries[1].Count))
	require.Equal(t, "12", topk.Entries[2].Value.GetStringValue())
	require.Equal(t, 1, int(topk.Entries[2].Count))
	require.Equal(t, "def", topk.Entries[3].Value.GetStringValue())
	require.Equal(t, 1, int(topk.Entries[3].Count))

	agg := "sum(val)"
	res, err = server.ColumnTopK(testCtx(), &runtimev1.ColumnTopKRequest{InstanceId: instanceId, TableName: "test", ColumnName: "col", Agg: agg})
	require.NoError(t, err)
	require.NotEmpty(t, res)
	require.Equal(t, 4, len(res.CategoricalSummary.GetTopK().Entries))
	require.Equal(t, "def", res.CategoricalSummary.GetTopK().Entries[0].Value.GetStringValue())
	require.Equal(t, 5, int(res.CategoricalSummary.GetTopK().Entries[0].Count))
	require.Equal(t, "abc", res.CategoricalSummary.GetTopK().Entries[1].Value.GetStringValue())
	require.Equal(t, 4, int(res.CategoricalSummary.GetTopK().Entries[1].Count))
	require.Equal(t, structpb.NewNullValue(), res.CategoricalSummary.GetTopK().Entries[2].Value)
	require.Equal(t, 1, int(res.CategoricalSummary.GetTopK().Entries[2].Count))
	require.Equal(t, "12", res.CategoricalSummary.GetTopK().Entries[3].Value.GetStringValue())
	require.Equal(t, 1, int(res.CategoricalSummary.GetTopK().Entries[3].Count))

	k := int32(1)
	res, err = server.ColumnTopK(testCtx(), &runtimev1.ColumnTopKRequest{InstanceId: instanceId, TableName: "test", ColumnName: "col", K: k})
	require.NoError(t, err)
	require.NotEmpty(t, res)
	require.Equal(t, 1, len(res.CategoricalSummary.GetTopK().Entries))
	require.Equal(t, "abc", res.CategoricalSummary.GetTopK().Entries[0].Value.GetStringValue())
	require.Equal(t, 2, int(res.CategoricalSummary.GetTopK().Entries[0].Count))
}

func TestServer_ColumnNullCount(t *testing.T) {
	server, instanceId := getColumnTestServer(t)

	res, err := server.ColumnNullCount(testCtx(), &runtimev1.ColumnNullCountRequest{InstanceId: instanceId, TableName: "test", ColumnName: "col"})
	require.NoError(t, err)
	require.NotEmpty(t, res)
	require.Equal(t, 1.0, res.Count)

	res, err = server.ColumnNullCount(testCtx(), &runtimev1.ColumnNullCountRequest{InstanceId: instanceId, TableName: "test", ColumnName: "times"})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, 0.0, res.Count)
}

func TestServer_ColumnDescriptiveStatistics(t *testing.T) {
	server, instanceId := getColumnTestServer(t)

	_, err := server.ColumnDescriptiveStatistics(testCtx(), &runtimev1.ColumnDescriptiveStatisticsRequest{InstanceId: instanceId, TableName: "test", ColumnName: "col"})
	if err != nil {
		// "col" is a varchar column, so this should fail
		require.ErrorContains(t, err, "No function matches the given name and argument types 'approx_quantile(VARCHAR, DECIMAL(3,2))'")
	}

	res, err := server.ColumnDescriptiveStatistics(testCtx(), &runtimev1.ColumnDescriptiveStatisticsRequest{InstanceId: instanceId, TableName: "test", ColumnName: "val"})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, 1.0, res.NumericSummary.GetNumericStatistics().Min)
	require.Equal(t, 5.0, res.NumericSummary.GetNumericStatistics().Max)
	require.Equal(t, 2.2, res.NumericSummary.GetNumericStatistics().Mean)
	require.Equal(t, 1.0, res.NumericSummary.GetNumericStatistics().Q25)
	require.Equal(t, 1.0, res.NumericSummary.GetNumericStatistics().Q50)
	require.Equal(t, 4.0, res.NumericSummary.GetNumericStatistics().Q75)
	require.Equal(t, 1.6, res.NumericSummary.GetNumericStatistics().Sd)
}

func TestServer_ColumnDescriptiveStatistics_EmptyModel(t *testing.T) {
	server, instanceId := getColumnTestServerWithEmptyModel(t)

	res, err := server.ColumnDescriptiveStatistics(testCtx(), &runtimev1.ColumnDescriptiveStatisticsRequest{InstanceId: instanceId, TableName: "test", ColumnName: "val"})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Nil(t, res.NumericSummary.GetNumericStatistics())
}

func TestServer_ColumnTimeGrain(t *testing.T) {
	server, instanceId := getColumnTestServer(t)

	_, err := server.ColumnTimeGrain(testCtx(), &runtimev1.ColumnTimeGrainRequest{InstanceId: instanceId, TableName: "test", ColumnName: "val"})
	if err != nil {
		// "val" is a numeric column, so this should fail
		require.ErrorContains(t, err, "Binder Error: No function matches the given name and argument types 'date_part(VARCHAR, INTEGER)'")
	}
	res, err := server.ColumnTimeGrain(testCtx(), &runtimev1.ColumnTimeGrainRequest{InstanceId: instanceId, TableName: "test", ColumnName: "times"})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, "TIME_GRAIN_DAY", res.TimeGrain.String())
}

func TestServer_ColumnTimeGrain_EmptyModel(t *testing.T) {
	server, instanceId := getColumnTestServerWithEmptyModel(t)

	_, err := server.ColumnTimeGrain(testCtx(), &runtimev1.ColumnTimeGrainRequest{InstanceId: instanceId, TableName: "test", ColumnName: "val"})
	if err != nil {
		// "val" is a numeric column, so this should fail
		require.ErrorContains(t, err, "Binder Error: No function matches the given name and argument types 'date_part(VARCHAR, INTEGER)'")
	}
	res, err := server.ColumnTimeGrain(testCtx(), &runtimev1.ColumnTimeGrainRequest{InstanceId: instanceId, TableName: "test", ColumnName: "times"})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, "TIME_GRAIN_UNSPECIFIED", res.TimeGrain.String())
}

func TestServer_GetNumericHistogram_FD(t *testing.T) {
	server, instanceId := getColumnTestServer(t)

	res, err := server.ColumnNumericHistogram(
		testCtx(),
		&runtimev1.ColumnNumericHistogramRequest{
			InstanceId:      instanceId,
			TableName:       "test",
			ColumnName:      "val",
			HistogramMethod: runtimev1.HistogramMethod_HISTOGRAM_METHOD_FD,
		},
	)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, 3, len(res.NumericSummary.GetNumericHistogramBins().Bins))
	require.Equal(t, 0, int(res.NumericSummary.GetNumericHistogramBins().Bins[0].Bucket))
	require.Equal(t, 1.0, res.NumericSummary.GetNumericHistogramBins().Bins[0].Low)
	require.Equal(t, 2.333333333333333, res.NumericSummary.GetNumericHistogramBins().Bins[0].High)
	require.Equal(t, 3.0, res.NumericSummary.GetNumericHistogramBins().Bins[0].Count)
}

func TestServer_GetNumericHistogram_Diagnostic(t *testing.T) {
	server, instanceId := getColumnTestServer(t)

	start, _, gap := queries.NiceAndStep(1, 5, 5)

	res, err := server.ColumnNumericHistogram(
		testCtx(),
		&runtimev1.ColumnNumericHistogramRequest{
			InstanceId:      instanceId,
			TableName:       "test",
			ColumnName:      "val",
			HistogramMethod: runtimev1.HistogramMethod_HISTOGRAM_METHOD_DIAGNOSTIC,
		})
	require.NoError(t, err)
	require.NotNil(t, res)
	bins := res.NumericSummary.GetNumericHistogramBins().Bins
	for i := 0; i < len(bins); i++ {
		fmt.Printf("%d %f %f %f\n", bins[i].Bucket, bins[i].Low, bins[i].High, bins[i].Count)
	}
	require.Equal(t, 5, len(bins))

	require.Equal(t, int32(0), bins[0].Bucket)
	require.Equal(t, start, bins[0].Low)
	require.Equal(t, start+gap, bins[0].High)
	require.Equal(t, 3.0, bins[0].Count)

	require.Equal(t, int32(1), bins[1].Bucket)
	require.Equal(t, start+gap, bins[1].Low)
	require.Equal(t, start+gap*2, bins[1].High)
	require.Equal(t, 0.0, bins[1].Count)

	require.Equal(t, 1.0, bins[2].Count)

	require.Equal(t, 0.0, bins[3].Count)
}

func TestServer_Model_Nulls(t *testing.T) {
	sql := `SELECT null as val`
	server, instanceId := getColumnTestServerWithModel(t, sql, 1)
	require.NotNil(t, server)
	require.NotEmpty(t, instanceId)
}

func TestServer_GetNumericHistogram_2rows_all_nulls(t *testing.T) {
	sql := `
		SELECT null as val
		UNION ALL
		SELECT null as val
	`
	server, instanceId := getColumnTestServerWithModel(t, sql, 2)

	res, err := server.ColumnNumericHistogram(
		testCtx(),
		&runtimev1.ColumnNumericHistogramRequest{
			InstanceId:      instanceId,
			TableName:       "test",
			ColumnName:      "val",
			HistogramMethod: runtimev1.HistogramMethod_HISTOGRAM_METHOD_FD,
		},
	)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, 0, len(res.NumericSummary.GetNumericHistogramBins().Bins))
}

func TestServer_GetNumericHistogram_2rows_single_null(t *testing.T) {
	sql := `
		SELECT null as val
		UNION ALL
		SELECT 2 as val
	`
	server, instanceId := getColumnTestServerWithModel(t, sql, 2)

	res, err := server.ColumnNumericHistogram(
		testCtx(),
		&runtimev1.ColumnNumericHistogramRequest{
			InstanceId:      instanceId,
			TableName:       "test",
			ColumnName:      "val",
			HistogramMethod: runtimev1.HistogramMethod_HISTOGRAM_METHOD_FD,
		},
	)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, 0, len(res.NumericSummary.GetNumericHistogramBins().Bins))
}

func TestServer_GetNumericHistogram_2rows(t *testing.T) {
	sql := `
		SELECT NULL as val
		UNION ALL
		SELECT 2 as val
		UNION ALL
		SELECT 4 as val
	`
	server, instanceId := getColumnTestServerWithModel(t, sql, 3)

	res, err := server.ColumnNumericHistogram(
		testCtx(),
		&runtimev1.ColumnNumericHistogramRequest{
			InstanceId:      instanceId,
			TableName:       "test",
			ColumnName:      "val",
			HistogramMethod: runtimev1.HistogramMethod_HISTOGRAM_METHOD_FD,
		},
	)
	require.NoError(t, err)
	require.NotNil(t, res)
	bins := res.NumericSummary.GetNumericHistogramBins().Bins
	require.Equal(t, 2, len(bins))
	require.Equal(t, int32(0), bins[0].Bucket)
	require.Equal(t, 2.0, bins[0].Low)
	require.Equal(t, 3.0, bins[0].High)
	require.Equal(t, 1.0, bins[0].Count)

	require.Equal(t, int32(1), bins[1].Bucket)
	require.Equal(t, 3.0, bins[1].Low)
	require.Equal(t, 4.0, bins[1].High)
	require.Equal(t, 1.0, bins[1].Count)
}

func TestServer_GetNumericHistogram_EmptyModel(t *testing.T) {
	server, instanceId := getColumnTestServerWithEmptyModel(t)

	res, err := server.ColumnNumericHistogram(
		testCtx(),
		&runtimev1.ColumnNumericHistogramRequest{
			InstanceId:      instanceId,
			TableName:       "test",
			ColumnName:      "val",
			HistogramMethod: runtimev1.HistogramMethod_HISTOGRAM_METHOD_FD,
		},
	)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Nil(t, res.NumericSummary.GetNumericHistogramBins().Bins)
}

func TestServer_GetRugHistogram(t *testing.T) {
	server, instanceId := getColumnTestServer(t)

	res, err := server.ColumnRugHistogram(testCtx(), &runtimev1.ColumnRugHistogramRequest{InstanceId: instanceId, TableName: "test", ColumnName: "val"})
	require.NoError(t, err)
	require.NotNil(t, res)
	outliers := res.NumericSummary.GetNumericOutliers().Outliers
	require.Equal(t, 3, len(outliers))
	require.Equal(t, 0, int(outliers[0].Bucket))
	require.Equal(t, 1.0, outliers[0].Low)
	require.Equal(t, 1.008, outliers[0].High)
	require.Equal(t, true, outliers[0].Present)
	require.True(t, outliers[0].Count > 0)

	// works only with numeric columns
	_, err = server.ColumnRugHistogram(testCtx(), &runtimev1.ColumnRugHistogramRequest{InstanceId: instanceId, TableName: "test", ColumnName: "times"})
	require.ErrorContains(t, err, "Conversion Error: Unimplemented type for cast (TIMESTAMP -> DOUBLE)")
}

func TestServer_GetRugHistogram_all_nulls(t *testing.T) {
	sql := `
		SELECT NULL as val
		UNION ALL
		SELECT NULL as val
	`
	server, instanceId := getColumnTestServerWithModel(t, sql, 2)

	res, err := server.ColumnRugHistogram(testCtx(), &runtimev1.ColumnRugHistogramRequest{InstanceId: instanceId, TableName: "test", ColumnName: "val"})
	require.NoError(t, err)
	require.NotNil(t, res)
	outliers := res.NumericSummary.GetNumericOutliers().Outliers
	require.Equal(t, 0, len(outliers))
}

func TestServer_GetRugHistogram_2rows_null(t *testing.T) {
	sql := `
		SELECT NULL as val
		UNION ALL
		SELECT 2 as val
	`
	server, instanceId := getColumnTestServerWithModel(t, sql, 2)

	res, err := server.ColumnRugHistogram(testCtx(), &runtimev1.ColumnRugHistogramRequest{InstanceId: instanceId, TableName: "test", ColumnName: "val"})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, 500, len(res.NumericSummary.GetNumericOutliers().Outliers))
}

func TestServer_GetRugHistogram_3rows_null(t *testing.T) {
	sql := `
		SELECT NULL as val
		UNION ALL
		SELECT 2 as val
		UNION ALL
		SELECT 4 as val
	`
	server, instanceId := getColumnTestServerWithModel(t, sql, 3)

	res, err := server.ColumnRugHistogram(testCtx(), &runtimev1.ColumnRugHistogramRequest{InstanceId: instanceId, TableName: "test", ColumnName: "val"})
	require.NoError(t, err)
	require.NotNil(t, res)
	outliers := res.NumericSummary.GetNumericOutliers().Outliers
	require.Equal(t, 2, len(outliers))
}

func TestServer_GetCategoricalHistogram_EmptyModel(t *testing.T) {
	server, instanceId := getColumnTestServerWithEmptyModel(t)

	res, err := server.ColumnRugHistogram(testCtx(), &runtimev1.ColumnRugHistogramRequest{InstanceId: instanceId, TableName: "test", ColumnName: "val"})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, 0, len(res.NumericSummary.GetNumericOutliers().Outliers))
}

func TestServer_GetTimeRangeSummary(t *testing.T) {
	server, instanceId := getColumnTestServer(t)

	// Get Time Range Summary works with timestamp columns
	res, err := server.ColumnTimeRange(testCtx(), &runtimev1.ColumnTimeRangeRequest{InstanceId: instanceId, TableName: "test", ColumnName: "times"})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, parseTime(t, "2022-11-01T00:00:00Z"), res.TimeRangeSummary.Min)
	require.Equal(t, parseTime(t, "2022-11-03T00:00:00Z"), res.TimeRangeSummary.Max)
	require.Equal(t, int32(0), res.TimeRangeSummary.Interval.Months)
	require.Equal(t, int32(2), res.TimeRangeSummary.Interval.Days)
	require.Equal(t, int64(0), res.TimeRangeSummary.Interval.Micros)
}

func TestServer_GetTimeRangeSummary_EmptyModel(t *testing.T) {
	server, instanceId := getColumnTestServerWithEmptyModel(t)

	// Get Time Range Summary works with timestamp columns
	res, err := server.ColumnTimeRange(testCtx(), &runtimev1.ColumnTimeRangeRequest{InstanceId: instanceId, TableName: "test", ColumnName: "times"})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Nil(t, res.TimeRangeSummary.Max)
	require.Nil(t, res.TimeRangeSummary.Min)
	require.Nil(t, res.TimeRangeSummary.Interval)
}

func TestServer_GetTimeRangeSummary_Date_Column(t *testing.T) {
	server, instanceId := getColumnTestServer(t)

	// Test Get Time Range Summary with Date type column
	res, err := server.ColumnTimeRange(testCtx(), &runtimev1.ColumnTimeRangeRequest{InstanceId: instanceId, TableName: "test", ColumnName: "dates"})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, parseTime(t, "2007-04-01T00:00:00Z"), res.TimeRangeSummary.Min)
	require.Equal(t, parseTime(t, "2011-06-30T00:00:00Z"), res.TimeRangeSummary.Max)
	require.Equal(t, int32(0), res.TimeRangeSummary.Interval.Months)
	require.Equal(t, int32(1551), res.TimeRangeSummary.Interval.Days)
	require.Equal(t, int64(0), res.TimeRangeSummary.Interval.Micros)
}

func parseTime(tst *testing.T, t string) *timestamppb.Timestamp {
	ts, err := time.Parse(time.RFC3339, t)
	require.NoError(tst, err)
	return timestamppb.New(ts)
}

func TestServer_GetCardinalityOfColumn(t *testing.T) {
	server, instanceId := getColumnTestServer(t)

	// Get Cardinality of Column works with all columns
	res, err := server.ColumnCardinality(testCtx(), &runtimev1.ColumnCardinalityRequest{InstanceId: instanceId, TableName: "test", ColumnName: "val"})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, 3.0, res.CategoricalSummary.GetCardinality())

	res, err = server.ColumnCardinality(testCtx(), &runtimev1.ColumnCardinalityRequest{InstanceId: instanceId, TableName: "test", ColumnName: "times"})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, 3.0, res.CategoricalSummary.GetCardinality())

	res, err = server.ColumnCardinality(testCtx(), &runtimev1.ColumnCardinalityRequest{InstanceId: instanceId, TableName: "test", ColumnName: "col"})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, 3.0, res.CategoricalSummary.GetCardinality())
}

func getColumnTestServer(t *testing.T) (*Server, string) {
	sql := `
		SELECT 'abc' AS col, 1 AS val, TIMESTAMP '2022-11-01 00:00:00' AS times, DATE '2007-04-01' AS dates
		UNION ALL 
		SELECT 'def' AS col, 5 AS val, TIMESTAMP '2022-11-02 00:00:00' AS times, DATE '2009-06-01' AS dates
		UNION ALL 
		SELECT 'abc' AS col, 3 AS val, TIMESTAMP '2022-11-03 00:00:00' AS times, DATE '2010-04-11' AS dates
		UNION ALL 
		SELECT null AS col, 1 AS val, TIMESTAMP '2022-11-03 00:00:00' AS times, DATE '2010-11-21' AS dates
		UNION ALL 
		SELECT 12 AS col, 1 AS val, TIMESTAMP '2022-11-03 00:00:00' AS times, DATE '2011-06-30' AS dates
	`

	return getColumnTestServerWithModel(t, sql, 5)
}

func getColumnTestServerWithEmptyModel(t *testing.T) (*Server, string) {
	sql := `
		SELECT 'abc' AS col, 1 AS val, TIMESTAMP '2022-11-01 00:00:00' AS times, DATE '2007-04-01' AS dates where 1<>1
	`
	return getColumnTestServerWithModel(t, sql, 0)
}

func getColumnTestServerWithModel(t *testing.T, sql string, expectation int) (*Server, string) {
	rt, instanceID := testruntime.NewInstanceWithModel(t, "test", sql)

	server, err := NewServer(&Options{}, rt, nil)
	require.NoError(t, err)

	olap, err := rt.OLAP(testCtx(), instanceID)
	require.NoError(t, err)

	res, err := olap.Execute(testCtx(), &drivers.Statement{Query: "SELECT count(*) FROM test"})
	require.NoError(t, err)
	defer res.Close()

	var n int
	for res.Next() {
		err := res.Scan(&n)
		require.NoError(t, err)
	}
	if expectation >= 0 {
		require.Equal(t, expectation, n)
	}

	return server, instanceID
}
