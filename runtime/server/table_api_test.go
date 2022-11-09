package server

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"

	"github.com/rilldata/rill/runtime/api"
	"github.com/rilldata/rill/runtime/drivers"
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
	server, instanceId, err := getTestServer()
	require.NoError(t, err)
	rows := createTestTable(server, instanceId, t)
	require.Equal(t, 1, getSingleValue(t, rows))
	rows, err = server.query(context.Background(), instanceId, &drivers.Statement{
		Query: "select count(*) from test",
	})
	require.NoError(t, err)
	require.Equal(t, 1, getSingleValue(t, rows))
}

func createTestTable(server *Server, instanceId string, t *testing.T) *sqlx.Rows {
	rows, err := server.query(context.Background(), instanceId, &drivers.Statement{
		Query: "create table test (a int)",
	})
	require.NoError(t, err)
	rows.Close()
	rows, _ = server.query(context.Background(), instanceId, &drivers.Statement{
		Query: "insert into test values (1)",
	})
	rows.Close()
	rows, err = server.query(context.Background(), instanceId, &drivers.Statement{
		Query: "select count(*) from test",
	})
	require.NoError(t, err)
	return rows
}

func TestServer_TableCardinality(t *testing.T) {
	server, instanceId, err := getTestServer()
	require.NoError(t, err)
	rows := createTestTable(server, instanceId, t)
	rows.Close()
	cr, err := server.TableCardinality(context.Background(), &api.CardinalityRequest{
		InstanceId: instanceId,
		TableName:  "test",
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), cr.Cardinality)
}

func TestServer_ProfileColumns(t *testing.T) {
	server, instanceId, err := getTestServer()
	require.NoError(t, err)
	rows := createTestTable(server, instanceId, t)
	rows.Close()
	cr, err := server.ProfileColumns(context.Background(), &api.ProfileColumnsRequest{
		InstanceId: instanceId,
		TableName:  "test",
	})
	require.NoError(t, err)
	require.Equal(t, 0, len(cr.ProfileColumn))
}

func TestServer_TableRows(t *testing.T) {
	server, instanceId, err := getTestServer()
	require.NoError(t, err)
	rows := createTestTable(server, instanceId, t)
	rows.Close()
	cr, err := server.TableRows(context.Background(), &api.RowsRequest{
		InstanceId: instanceId,
		TableName:  "test",
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(cr.Data))
}
