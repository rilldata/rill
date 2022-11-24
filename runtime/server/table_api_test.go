package server

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
)

func getSingleValue(t *testing.T, rows *sqlx.Rows) int {
	var val int
	for rows.Next() {
		err := rows.Scan(&val)
		require.NoError(t, err)
	}
	rows.Close()
	return val
}

func TestServer_Database(t *testing.T) {
	server, instanceId := getTestServer(t)
	result := createTestTable(server, instanceId, t)
	require.Equal(t, 1, getSingleValue(t, result.Rows))
	result, err := server.query(context.Background(), instanceId, &drivers.Statement{
		Query: "select count(*) from test",
	})
	require.NoError(t, err)
	require.Equal(t, 1, getSingleValue(t, result.Rows))
}

func createTestTable(server *Server, instanceId string, t *testing.T) *drivers.Result {
	return createTable(server, instanceId, t, "test")
}

func createTable(server *Server, instanceId string, t *testing.T, tableName string) *drivers.Result {
	result, err := server.query(context.Background(), instanceId, &drivers.Statement{
		Query: "create table " + quoteName(tableName) + " (a int, \"b\"\"b\" int)",
	})
	require.NoError(t, err)
	result.Close()
	result, _ = server.query(context.Background(), instanceId, &drivers.Statement{
		Query: "insert into " + quoteName(tableName) + " values (1, 10)",
	})
	require.NoError(t, err)
	result.Close()
	result, err = server.query(context.Background(), instanceId, &drivers.Statement{
		Query: "select count(*) from " + quoteName(tableName),
	})
	require.NoError(t, err)
	return result
}

func TestServer_TableCardinality(t *testing.T) {
	server, instanceId := getTestServer(t)
	rows := createTestTable(server, instanceId, t)
	rows.Close()
	cr, err := server.GetTableCardinality(context.Background(), &runtimev1.GetTableCardinalityRequest{
		InstanceId: instanceId,
		TableName:  "test",
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), cr.Cardinality)

	rows = createTable(server, instanceId, t, "select")
	rows.Close()
	cr, err = server.GetTableCardinality(context.Background(), &runtimev1.GetTableCardinalityRequest{
		InstanceId: instanceId,
		TableName:  "select",
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), cr.Cardinality)
}

func TestServer_ProfileColumns(t *testing.T) {
	server, instanceId := getTestServer(t)
	rows := createTestTable(server, instanceId, t)
	rows.Close()
	cr, err := server.ProfileColumns(context.Background(), &runtimev1.ProfileColumnsRequest{
		InstanceId: instanceId,
		TableName:  "test",
	})
	require.NoError(t, err)
	require.Equal(t, 2, len(cr.GetProfileColumns()))
	require.Equal(t, "a", cr.GetProfileColumns()[0].Name)
	require.Equal(t, "INTEGER", cr.GetProfileColumns()[0].Type)
	require.Equal(t, int32(1), cr.GetProfileColumns()[0].LargestStringLength)

	require.Equal(t, "b\"b", cr.GetProfileColumns()[1].Name)
	require.Equal(t, "INTEGER", cr.GetProfileColumns()[1].Type)
	require.Equal(t, int32(len("10")), cr.GetProfileColumns()[1].LargestStringLength)
}

func TestServer_TableRows(t *testing.T) {
	server, instanceId := getTestServer(t)
	rows := createTestTable(server, instanceId, t)
	rows.Close()
	cr, err := server.GetTableRows(context.Background(), &runtimev1.GetTableRowsRequest{
		InstanceId: instanceId,
		TableName:  "test",
		Limit:      1,
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(cr.Data))
}
