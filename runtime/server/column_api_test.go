package server

import (
	"context"
	"github.com/rilldata/rill/runtime/api"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestServer_GetTopK(t *testing.T) {
	server, instanceId, err := GetTestServer()

	server.QueryDirect(context.Background(), &api.QueryDirectRequest{
		InstanceId: instanceId,
		Sql:        "CREATE TABLE test AS (SELECT 'abc' AS col, 1 AS val UNION ALL SELECT 'def' AS col, 5 AS val UNION ALL SELECT 'abc' AS col, 3 AS val)",
		Args:       nil,
	})

	if err != nil {
		t.Fatal(err)
	}
	if server == nil {
		t.Fatal("server is nil")
	}
	if instanceId == "" {
		t.Fatal("instanceId is empty")
	}

	res, err := server.GetTopK(context.Background(), &api.TopKRequest{InstanceId: instanceId, TableName: "test", ColumnName: "col"})
	if err != nil {
		t.Error(err)
	}
	require.NotEmpty(t, res)
	require.Equal(t, 2, len(res.Data))
	require.Equal(t, "abc", res.Data[0].Fields["value"].GetStringValue())
	require.Equal(t, 2, int(res.Data[0].Fields["count"].GetNumberValue()))
	require.Equal(t, "def", res.Data[1].Fields["value"].GetStringValue())
	require.Equal(t, 1, int(res.Data[1].Fields["count"].GetNumberValue()))

	agg := "sum(val)"
	res, err = server.GetTopK(context.Background(), &api.TopKRequest{InstanceId: instanceId, TableName: "test", ColumnName: "col", Agg: &agg})
	if err != nil {
		t.Error(err)
	}
	require.NotEmpty(t, res)
	require.Equal(t, 2, len(res.Data))
	require.Equal(t, "def", res.Data[0].Fields["value"].GetStringValue())
	require.Equal(t, 5, int(res.Data[0].Fields["count"].GetNumberValue()))
	require.Equal(t, "abc", res.Data[1].Fields["value"].GetStringValue())
	require.Equal(t, 4, int(res.Data[1].Fields["count"].GetNumberValue()))

	k := int32(1)
	res, err = server.GetTopK(context.Background(), &api.TopKRequest{InstanceId: instanceId, TableName: "test", ColumnName: "col", K: &k})
	if err != nil {
		t.Error(err)
	}
	require.NotEmpty(t, res)
	require.Equal(t, 1, len(res.Data))
	require.Equal(t, "abc", res.Data[0].Fields["value"].GetStringValue())
	require.Equal(t, 2, int(res.Data[0].Fields["count"].GetNumberValue()))
}
