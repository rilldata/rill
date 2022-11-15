package server

import (
	"context"
	"github.com/rilldata/rill/runtime/api"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestServer_GetTopK(t *testing.T) {
	server, instanceId, err := getTestServer(t)

	_, err = server.QueryDirect(context.Background(), &api.QueryDirectRequest{
		InstanceId: instanceId,
		Sql:        "CREATE TABLE test AS (SELECT 'abc' AS col, 1 AS val UNION ALL SELECT 'def' AS col, 5 AS val UNION ALL SELECT 'abc' AS col, 3 AS val)",
		Args:       nil,
	})
	require.NoError(t, err)

	res, err := server.GetTopK(context.Background(), &api.TopKRequest{InstanceId: instanceId, TableName: "test", ColumnName: "col"})
	if err != nil {
		t.Error(err)
	}
	require.NotEmpty(t, res)
	require.Equal(t, 2, len(res.TopKResponse.Entries))
	require.Equal(t, "abc", *res.TopKResponse.Entries[0].Value)
	require.Equal(t, 2, int(res.TopKResponse.Entries[0].Count))
	require.Equal(t, "def", *res.TopKResponse.Entries[1].Value)
	require.Equal(t, 1, int(res.TopKResponse.Entries[1].Count))

	agg := "sum(val)"
	res, err = server.GetTopK(context.Background(), &api.TopKRequest{InstanceId: instanceId, TableName: "test", ColumnName: "col", Agg: &agg})
	if err != nil {
		t.Error(err)
	}
	require.NotEmpty(t, res)
	require.Equal(t, 2, len(res.TopKResponse.Entries))
	require.Equal(t, "def", *res.TopKResponse.Entries[0].Value)
	require.Equal(t, 5, int(res.TopKResponse.Entries[0].Count))
	require.Equal(t, "abc", *res.TopKResponse.Entries[1].Value)
	require.Equal(t, 4, int(res.TopKResponse.Entries[1].Count))

	k := int32(1)
	res, err = server.GetTopK(context.Background(), &api.TopKRequest{InstanceId: instanceId, TableName: "test", ColumnName: "col", K: &k})
	if err != nil {
		t.Error(err)
	}
	require.NotEmpty(t, res)
	require.Equal(t, 1, len(res.TopKResponse.Entries))
	require.Equal(t, "abc", *res.TopKResponse.Entries[0].Value)
	require.Equal(t, 2, int(res.TopKResponse.Entries[0].Count))
}
