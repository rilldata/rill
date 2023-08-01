package server

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/queries"
	"github.com/stretchr/testify/require"
	"github.com/xitongsys/parquet-go-source/buffer"

	"github.com/xitongsys/parquet-go/parquet"
	"github.com/xitongsys/parquet-go/reader"
	"github.com/xuri/excelize/v2"
)

func TestServer_MetricsViewRows(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewRows(testCtx(), &runtimev1.MetricsViewRowsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
	})
	require.NoError(t, err)
	require.Equal(t, 2, len(tr.Data))
	require.Equal(t, 11, len(tr.Meta))

}

func TestServer_MetricsViewRows_Granularity(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewRows(testCtx(), &runtimev1.MetricsViewRowsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_DAY,
	})
	require.NoError(t, err)
	require.Equal(t, 2, len(tr.Data))
	require.Equal(t, 12, len(tr.Meta))
	require.Equal(t, "timestamp__day", tr.Meta[0].Name)
}

/*
|id |timestamp               |publisher|domain   |bid_price|volume|impressions|ad words|clicks|device|
|---|------------------------|---------|---------|---------|------|-----------|--------|------|------|
|0  |2022-01-01T14:49:50.459Z|         |msn.com  |2        |4     |2          |cars    |      |iphone|
|1  |2022-01-02T11:58:12.475Z|Yahoo    |yahoo.com|2        |4     |1          |cars    |1     |      |
*/
func TestServer_MetricsViewRows_Granularity_Kathmandu(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewRows(testCtx(), &runtimev1.MetricsViewRowsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_HOUR,
		TimeStart:       parseTimeToProtoTimeStamps(t, "2022-01-01T14:15:00Z"),
		TimeEnd:         parseTimeToProtoTimeStamps(t, "2022-01-01T15:15:00Z"),
		TimeZone:        "Asia/Kathmandu",
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data))
	require.Equal(t, "timestamp__hour", tr.Meta[0].Name)
	require.Equal(t, "2022-01-01T14:15:00Z", tr.Data[0].Fields["timestamp__hour"].GetStringValue())
}

func TestServer_MetricsViewRows_export(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	ctx := testCtx()
	q := &queries.MetricsViewRows{
		MetricsViewName: "ad_bids_metrics",
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_DAY,
	}

	var buf bytes.Buffer

	err := q.Export(ctx, server.runtime, instanceId, &buf, &runtime.ExportOptions{
		Format: runtimev1.ExportFormat_EXPORT_FORMAT_XLSX,
	})
	require.NoError(t, err)

	file, err := excelize.OpenReader(&buf)
	rows, err := file.GetRows("Sheet1")
	require.NoError(t, err)

	require.Equal(t, 3, len(rows))
	require.Equal(t, []string{"timestamp", "publisher", "domain", "bid_price", "volume", "impressions", "ad words", "clicks", "numeric_dim", "device"}, rows[0][2:])
	require.Equal(t, []string{"2022-01-01T14:49:50.459Z", "", "msn.com", "2", "4", "2", "cars", "", "1", "iphone"}, rows[1][2:])
	require.Equal(t, []string{"2022-01-02T11:58:12.475Z", "Yahoo", "yahoo.com", "2", "4", "1", "cars", "1", "1"}, rows[2][2:])
}

func TestServer_MetricsViewRows_parquet_export(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	ctx := testCtx()
	q := &queries.MetricsViewRows{
		MetricsViewName: "ad_bids_metrics_parquet",
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_DAY,
	}

	var buf bytes.Buffer

	err := q.Export(ctx, server.runtime, instanceId, &buf, &runtime.ExportOptions{
		Format: runtimev1.ExportFormat_EXPORT_FORMAT_PARQUET,
	})
	require.NoError(t, err)

	fw := buffer.NewBufferFileFromBytes(buf.Bytes())

	reader, err := reader.NewParquetReader(fw, nil, 1)

	require.NoError(t, err)

	values := make(map[string]interface{})
	for k, columnBuffer := range reader.ColumnBuffers {
		table, _ := columnBuffer.ReadRows(1)
		v := table.Values[0]
		splits := strings.Split(k, "\x01")
		k = strings.ToLower(splits[len(splits)-1])
		values[k] = v
		fmt.Printf("%s %v\n", k, v)
		require.NotNil(t, v)
	}

	require.Equal(t, true, values["tbool"])
	require.Equal(t, int32(-1), values["tint1"])
	require.Equal(t, int32(-2), values["tint2"])
	require.Equal(t, int32(-4), values["tint4"])
	require.Equal(t, int64(-8), values["tint8"])
	require.Equal(t, int32(1), values["tuint1"])
	require.Equal(t, int32(2), values["tuint2"])
	require.Equal(t, int32(4), values["tuint4"])
	require.Equal(t, int64(8), values["tuint8"])
	require.Equal(t, float64(1.0), values["thugeint"])
	require.Equal(t, float32(4.0), values["tfloat4"])
	require.Equal(t, float64(8.0), values["tfloat8"])
	require.Equal(t, float64(1.0), values["tdecimal"])
	require.Equal(t, "2023-01-01T00:00:00Z", values["timestamp"])
	require.Equal(t, `["a", "b"]`, values["tlist"])
	require.Equal(t, `{"f1":1,"f2":2}`, values["tmap"])
	require.Equal(t, `{"f1":1,"f2":{"f3":3}}`, values["tstruct"])

	schema := reader.Footer.Schema
	meta := make(map[string]*parquet.SchemaElement)
	for _, elem := range schema {
		meta[strings.ToLower(elem.GetName())] = elem
	}

	require.Equal(t, "BOOLEAN", meta["tbool"].Type.String())
	require.Equal(t, "IntType({BitWidth:8 IsSigned:true})", meta["tint1"].GetLogicalType().GetINTEGER().String())
	require.Equal(t, "IntType({BitWidth:16 IsSigned:true})", meta["tint2"].GetLogicalType().GetINTEGER().String())
	require.Equal(t, "IntType({BitWidth:32 IsSigned:true})", meta["tint4"].GetLogicalType().GetINTEGER().String())
	require.Equal(t, "IntType({BitWidth:64 IsSigned:true})", meta["tint8"].GetLogicalType().GetINTEGER().String())
	require.Equal(t, "IntType({BitWidth:8 IsSigned:false})", meta["tuint1"].GetLogicalType().GetINTEGER().String())
	require.Equal(t, "IntType({BitWidth:16 IsSigned:false})", meta["tuint2"].GetLogicalType().GetINTEGER().String())
	require.Equal(t, "IntType({BitWidth:32 IsSigned:false})", meta["tuint4"].GetLogicalType().GetINTEGER().String())
	require.Equal(t, "IntType({BitWidth:64 IsSigned:false})", meta["tuint8"].GetLogicalType().GetINTEGER().String())
	require.Equal(t, "DOUBLE", meta["thugeint"].Type.String())
	require.Equal(t, "FLOAT", meta["tfloat4"].Type.String())
	require.Equal(t, "DOUBLE", meta["tfloat8"].Type.String())
	require.Equal(t, "DOUBLE", meta["tdecimal"].Type.String())
	require.Equal(t, "BYTE_ARRAY", meta["timestamp"].Type.String())
	require.Equal(t, "BYTE_ARRAY", meta["tlist"].Type.String())
	require.Equal(t, "BYTE_ARRAY", meta["tmap"].Type.String())
	require.Equal(t, "BYTE_ARRAY", meta["tstruct"].Type.String())
	require.Equal(t, "BYTE_ARRAY", meta["tuuid"].Type.String())

	reader.ReadStop()
}
