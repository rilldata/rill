package druid

import (
	"context"
	"strings"
	"testing"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/pbutil"
	"github.com/rilldata/rill/runtime/storage"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

/*
 * Requires standalone Druid instance running on localhost:8888.
 */
func Ignore_TestDriver_types(t *testing.T) {
	driver := &driver{}
	handle, err := driver.Open("default", map[string]any{"pool_size": 2, "dsn": "http://localhost:8888/druid/v2/sql"}, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)

	olap, ok := handle.AsOLAP("")
	require.True(t, ok)

	res, err := olap.Query(context.Background(), &drivers.Statement{
		Query: `select 
		cast(1 as boolean) as bool1, 
		cast(1 as bigint) as bigint1, 
		timestamp '2021-01-01 00:00:00' as ts1,
		cast(1 as real) as double1,
		cast(1 as float) as float1,
		cast(1 as integer) as integer1,
		date '2023-01-01' as date1
		`,
	})
	require.NoError(t, err)
	require.True(t, len(res.Schema.Fields) > 0)

	data, err := rowsToData(res)

	require.NoError(t, err)

	require.True(t, data[0].Fields["bool1"].GetBoolValue())
	require.Equal(t, 1.0, data[0].Fields["bigint1"].GetNumberValue())
	require.Equal(t, "2021-01-01T00:00:00Z", data[0].Fields["ts1"].GetStringValue())
	require.Equal(t, 1.0, data[0].Fields["double1"].GetNumberValue())
	require.Equal(t, 1.0, data[0].Fields["float1"].GetNumberValue())
	require.Equal(t, 1.0, data[0].Fields["integer1"].GetNumberValue())
	require.Equal(t, "2023-01-01T00:00:00.000Z", data[0].Fields["date1"].GetStringValue())
}

func Ignore_TestDriver_array_type(t *testing.T) {
	driver := &driver{}
	handle, err := driver.Open("default", map[string]any{"pool_size": 2, "dsn": "http://localhost:8888/druid/v2/sql"}, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)

	olap, ok := handle.AsOLAP("")
	require.True(t, ok)

	res, err := olap.Query(context.Background(), &drivers.Statement{
		Query: `select 
		array [1,2] as array1
		`,
	})
	require.NoError(t, err)
	require.True(t, len(res.Schema.Fields) > 0)

	data, err := rowsToData(res)

	require.NoError(t, err)

	require.Equal(t, 1.0, data[0].Fields["array1"].GetListValue().Values[0].GetNumberValue())
	require.Equal(t, 2.0, data[0].Fields["array1"].GetListValue().Values[1].GetNumberValue())
}

func Ignore_TestDriver_json_type(t *testing.T) {
	driver := &driver{}
	handle, err := driver.Open("default", map[string]any{"pool_size": 2, "dsn": "http://localhost:8888/druid/v2/sql"}, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)

	olap, ok := handle.AsOLAP("")
	require.True(t, ok)

	res, err := olap.Query(context.Background(), &drivers.Statement{
		Query: `select 
			json_object('a':'b') as json1 
		`,
	})
	require.NoError(t, err)
	require.True(t, len(res.Schema.Fields) > 0)

	data, err := rowsToData(res)

	require.NoError(t, err)

	require.Equal(t, "b", data[0].Fields["json1"].GetStructValue().Fields["a"].GetStringValue())
}

func Ignore_TestDriver_multiple_rows(t *testing.T) {
	driver := &driver{}
	handle, err := driver.Open("default", map[string]any{"pool_size": 2, "dsn": "http://localhost:8888/druid/v2/sql"}, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)

	olap, ok := handle.AsOLAP("")
	require.True(t, ok)

	res, err := olap.Query(context.Background(), &drivers.Statement{
		Query: `
		select 
			cast(1 as boolean) as bool1,
			cast(1 as bigint) as bigint1
		union all
		select 
			cast(1 as boolean) as bool1,
			cast(1 as bigint) as bigint1
		`,
	})
	require.NoError(t, err)
	require.True(t, len(res.Schema.Fields) > 0)

	data, err := rowsToData(res)

	require.NoError(t, err)

	require.Equal(t, 2, len(data))
	require.True(t, data[0].Fields["bool1"].GetBoolValue())
	require.Equal(t, 1.0, data[0].Fields["bigint1"].GetNumberValue())
	require.True(t, data[1].Fields["bool1"].GetBoolValue())
	require.Equal(t, 1.0, data[1].Fields["bigint1"].GetNumberValue())

}

func Ignore_TestDriver_error(t *testing.T) {
	driver := &driver{}
	handle, err := driver.Open("default", map[string]any{"pool_size": 2, "dsn": "http://localhost:8888/druid/v2/sql"}, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)

	olap, ok := handle.AsOLAP("")
	require.True(t, ok)

	_, err = olap.Query(context.Background(), &drivers.Statement{
		Query: `select select`,
	})
	require.Error(t, err)
	require.True(t, strings.HasPrefix(err.Error(), `"error":"druidException"`))
}

func TestDriver_correctURL(t *testing.T) {
	s, err := correctURL("https://localhost/druid/sql/v2/avatica-protobuf")
	require.NoError(t, err)
	require.Equal(t, "https://localhost/druid/v2/sql", s)

	s, err = correctURL("https://localhost/druid/sql/v2/avatica-protobuf?authentication=BASIC&avaticaUser=user1&avaticaPassword=pass%40rd")
	require.NoError(t, err)
	require.Equal(t, "https://user1:pass%40rd@localhost/druid/v2/sql", s)

	s, err = correctURL("https://localhost:8888/druid/sql/v2/avatica-protobuf?authentication=BASIC&avaticaUser=user1&avaticaPassword=pass%40rd")
	require.NoError(t, err)
	require.Equal(t, "https://user1:pass%40rd@localhost:8888/druid/v2/sql", s)
}

func rowsToData(rows *drivers.Result) ([]*structpb.Struct, error) {
	var data []*structpb.Struct
	for rows.Next() {
		rowMap := make(map[string]any)
		err := rows.MapScan(rowMap)
		if err != nil {
			return nil, err
		}

		rowStruct, err := pbutil.ToStruct(rowMap, rows.Schema)
		if err != nil {
			return nil, err
		}

		data = append(data, rowStruct)
	}

	err := rows.Err()
	if err != nil {
		return nil, err
	}

	return data, nil
}
