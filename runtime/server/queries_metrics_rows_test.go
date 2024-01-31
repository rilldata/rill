package server_test

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/apache/arrow/go/v14/arrow"
	"github.com/apache/arrow/go/v14/arrow/array"
	"github.com/apache/arrow/go/v14/arrow/memory"
	"github.com/apache/arrow/go/v14/parquet"

	"github.com/apache/arrow/go/v14/parquet/file"
	"github.com/apache/arrow/go/v14/parquet/pqarrow"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/queries"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"

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

func TestServer_MetricsViewRows_export_xlsx(t *testing.T) {
	t.Parallel()
	rt, instanceId := testruntime.NewInstanceForProject(t, "ad_bids_2rows")

	ctx := testCtx()
	mvName := "ad_bids_metrics"
	mv, security := resolveMVAndSecurity(t, rt, instanceId, mvName)

	q := &queries.MetricsViewRows{
		MetricsViewName:    mvName,
		TimeGranularity:    runtimev1.TimeGrain_TIME_GRAIN_DAY,
		MetricsView:        mv,
		ResolvedMVSecurity: security,
	}

	var buf bytes.Buffer

	err := q.Export(ctx, rt, instanceId, &buf, &runtime.ExportOptions{
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

func getColumnChunk(tbl arrow.Table, col int) arrow.Array {
	return tbl.Column(col).Data().Chunk(0)
}

func TestServer_MetricsViewRows_parquet_export(t *testing.T) {
	t.Parallel()
	rt, instanceId := testruntime.NewInstanceForProject(t, "ad_bids_2rows")

	ctx := testCtx()
	mvName := "ad_bids_metrics_parquet"
	mv, security := resolveMVAndSecurity(t, rt, instanceId, mvName)

	q := &queries.MetricsViewRows{
		MetricsViewName:    mvName,
		TimeGranularity:    runtimev1.TimeGrain_TIME_GRAIN_DAY,
		MetricsView:        mv,
		ResolvedMVSecurity: security,
	}

	var buf bytes.Buffer

	err := q.Export(ctx, rt, instanceId, &buf, &runtime.ExportOptions{
		Format: runtimev1.ExportFormat_EXPORT_FORMAT_PARQUET,
	})
	require.NoError(t, err)

	f, err := os.CreateTemp("", "test-export-*")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	defer f.Close()

	f.Write(buf.Bytes())
	mem := memory.NewCheckedAllocator(memory.DefaultAllocator)
	rdr, err := file.OpenParquetFile(f.Name(), false, file.WithReadProps(parquet.NewReaderProperties(mem)))
	require.NoError(t, err)
	defer rdr.Close()

	arrowRdr, err := pqarrow.NewFileReader(rdr, pqarrow.ArrowReadProperties{}, mem)
	require.NoError(t, err)
	tbl, err := arrowRdr.ReadTable(ctx)
	require.NoError(t, err)

	/*
	   -1::INT1 as tint1,
	   -2::INT2 as tint2,
	   -4::INT4 as tint4,
	   -8::INT8 as tint8,
	   1::UTINYINT as tuint1,
	   2::USMALLINT as tuint2,
	   4::UINTEGER as tuint4,
	   8::UBIGINT as tuint8,
	   1::HUGEINT as thugeint,
	   4::FLOAT4 as tfloat4,
	   8::FLOAT8 as tfloat8,
	   1::DECIMAL(18,3) as tdecimal,
	   TRUE as tbool,
	   ['a','b'] as tlist,
	   map {'f1' : 1, 'f2': 2} as tmap,
	   {'f1' : 1, 'f2': { 'f3': 3 }} as tstruct,
	   TIMESTAMP '2023-01-01' as timestamp,
	   uuid() as tuuid
	*/
	index := 0
	flds := arrowRdr.Manifest.Fields
	require.Equal(t, "timestamp__day", flds[index].Field.Name)
	require.Equal(t, arrow.TIMESTAMP, flds[index].Field.Type.ID())
	td := getColumnChunk(tbl, index).(*array.Timestamp)
	require.Equal(t, "2023-01-01T00:00:00Z", td.Value(0).ToTime(arrow.Microsecond).Format(time.RFC3339))
	index++

	require.Equal(t, "tint1", flds[index].Field.Name)
	require.Equal(t, arrow.INT8, flds[index].Field.Type.ID())
	tint1 := getColumnChunk(tbl, index).(*array.Int8)
	require.Equal(t, int8(-1), tint1.Value(0))
	index++

	require.Equal(t, "tint2", flds[index].Field.Name)
	require.Equal(t, arrow.INT16, flds[index].Field.Type.ID())
	tint2 := getColumnChunk(tbl, index).(*array.Int16)
	require.Equal(t, int16(-2), tint2.Value(0))
	index++

	require.Equal(t, "tint4", flds[index].Field.Name)
	require.Equal(t, arrow.INT32, flds[index].Field.Type.ID())
	tint4 := getColumnChunk(tbl, index).(*array.Int32)
	require.Equal(t, int32(-4), tint4.Value(0))
	index++

	require.Equal(t, "tint8", flds[index].Field.Name)
	require.Equal(t, arrow.INT64, flds[index].Field.Type.ID())
	tint8 := getColumnChunk(tbl, index).(*array.Int64)
	require.Equal(t, int64(-8), tint8.Value(0))
	index++

	require.Equal(t, "tuint1", flds[index].Field.Name)
	require.Equal(t, arrow.UINT8, flds[index].Field.Type.ID())
	tuint1 := getColumnChunk(tbl, index).(*array.Uint8)
	require.Equal(t, uint8(1), tuint1.Value(0))
	index++

	require.Equal(t, "tuint2", flds[index].Field.Name)
	require.Equal(t, arrow.UINT16, flds[index].Field.Type.ID())
	tuint2 := getColumnChunk(tbl, index).(*array.Uint16)
	require.Equal(t, uint16(2), tuint2.Value(0))
	index++

	require.Equal(t, "tuint4", flds[index].Field.Name)
	require.Equal(t, arrow.UINT32, flds[index].Field.Type.ID())
	tuint4 := getColumnChunk(tbl, index).(*array.Uint32)
	require.Equal(t, uint32(4), tuint4.Value(0))
	index++

	require.Equal(t, "tuint8", flds[index].Field.Name)
	require.Equal(t, arrow.UINT64, flds[index].Field.Type.ID())
	tuint8 := getColumnChunk(tbl, index).(*array.Uint64)
	require.Equal(t, uint64(8), tuint8.Value(0))
	index++

	require.Equal(t, "thugeint", flds[index].Field.Name)
	require.Equal(t, arrow.FLOAT64, flds[index].Field.Type.ID())
	thugeint := getColumnChunk(tbl, index).(*array.Float64)
	require.Equal(t, float64(1), thugeint.Value(0))
	index++

	require.Equal(t, "tfloat4", flds[index].Field.Name)
	require.Equal(t, arrow.FLOAT32, flds[index].Field.Type.ID())
	tfloat4 := getColumnChunk(tbl, index).(*array.Float32)
	require.Equal(t, float32(4), tfloat4.Value(0))
	index++

	require.Equal(t, "tfloat8", flds[index].Field.Name)
	require.Equal(t, arrow.FLOAT64, flds[index].Field.Type.ID())
	tfloat8 := getColumnChunk(tbl, index).(*array.Float64)
	require.Equal(t, float64(8), tfloat8.Value(0))
	index++

	require.Equal(t, "tdecimal", flds[index].Field.Name)
	require.Equal(t, arrow.DECIMAL128, flds[index].Field.Type.ID())
	tdecimal := getColumnChunk(tbl, index).(*array.Decimal128)
	require.Equal(t, float32(1), tdecimal.Value(0).ToFloat32(3))
	index++

	require.Equal(t, "tbool", flds[index].Field.Name)
	require.Equal(t, arrow.BOOL, flds[index].Field.Type.ID())
	tbool := getColumnChunk(tbl, index).(*array.Boolean)
	require.Equal(t, true, tbool.Value(0))
	index++

	require.Equal(t, "tlist", flds[index].Field.Name)
	require.Equal(t, arrow.LIST, flds[index].Field.Type.ID())
	tlist := getColumnChunk(tbl, index).(*array.List)
	strList := tlist.ListValues().(*array.String)
	require.Equal(t, "a", strList.Value(0))
	require.Equal(t, "b", strList.Value(1))
	index++

	require.Equal(t, "tmap", flds[index].Field.Name)
	require.Equal(t, arrow.MAP, flds[index].Field.Type.ID())
	tmap := getColumnChunk(tbl, index).(*array.Map)
	keys := tmap.Keys().(*array.String)
	values := tmap.Items().(*array.Int32)
	require.Equal(t, "f1", keys.Value(0))
	require.Equal(t, "f2", keys.Value(1))
	require.Equal(t, int32(1), values.Value(0))
	require.Equal(t, int32(2), values.Value(1))
	index++

	require.Equal(t, "tstruct", flds[index].Field.Name)
	require.Equal(t, arrow.STRUCT, flds[index].Field.Type.ID())
	tstruct := getColumnChunk(tbl, index).(*array.Struct)
	fields := tstruct.Field(0).(*array.Int32)
	require.Equal(t, int32(1), fields.Value(0))
	substruct := tstruct.Field(1).(*array.Struct)
	subfield := substruct.Field(0).(*array.Int32)
	require.Equal(t, int32(3), subfield.Value(0))
	index++

	require.Equal(t, "timestamp", flds[index].Field.Name)
	require.Equal(t, arrow.TIMESTAMP, flds[index].Field.Type.ID())
	ttimestamp := getColumnChunk(tbl, index).(*array.Timestamp)
	require.Equal(t, "2023-01-01T00:00:00Z", ttimestamp.Value(0).ToTime(arrow.Microsecond).Format(time.RFC3339))
	index++

	require.Equal(t, "ttime", flds[index].Field.Name)
	require.Equal(t, arrow.TIME64, flds[index].Field.Type.ID())
	ttime := getColumnChunk(tbl, index).(*array.Time64)
	require.Equal(t, "12:00:00", ttime.Value(0).ToTime(arrow.Microsecond).Format(time.TimeOnly))
	index++

	require.Equal(t, "tdate", flds[index].Field.Name)
	require.Equal(t, arrow.DATE32, flds[index].Field.Type.ID())
	tdate := getColumnChunk(tbl, index).(*array.Date32)
	require.Equal(t, "2023-01-02", tdate.Value(0).FormattedString())
	index++

	require.Equal(t, "tuuid", flds[index].Field.Name)
	require.Equal(t, arrow.FIXED_SIZE_BINARY, flds[index].Field.Type.ID())
	tuuid := getColumnChunk(tbl, index).(*array.FixedSizeBinary)
	require.True(t, len(tuuid.Value(0)) > 0)
}

func TestServer_MetricsViewRows_export_csv(t *testing.T) {
	t.Parallel()
	rt, instanceId := testruntime.NewInstanceForProject(t, "ad_bids_2rows")

	ctx := testCtx()
	mvName := "ad_bids_metrics"
	mv, security := resolveMVAndSecurity(t, rt, instanceId, mvName)

	q := &queries.MetricsViewRows{
		MetricsViewName:    mvName,
		TimeGranularity:    runtimev1.TimeGrain_TIME_GRAIN_DAY,
		MetricsView:        mv,
		ResolvedMVSecurity: security,
	}

	var buf bytes.Buffer

	err := q.Export(ctx, rt, instanceId, &buf, &runtime.ExportOptions{
		Format: runtimev1.ExportFormat_EXPORT_FORMAT_CSV,
	})
	require.NoError(t, err)

	require.True(t, len(buf.Bytes()) > 0)
}
