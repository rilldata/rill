package server

import (
	"testing"

	"github.com/jmoiron/sqlx"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func TestServer_TableCardinality(t *testing.T) {
	server, instanceId := getTableTestServer(t)
	cr, err := server.GetTableCardinality(testCtx(), &runtimev1.GetTableCardinalityRequest{
		InstanceId: instanceId,
		TableName:  "test",
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), cr.Cardinality)
}

func TestServer_TableCardinality_EmptyModel(t *testing.T) {
	server, instanceId := getTableTestServerWithEmptyModel(t)
	cr, err := server.GetTableCardinality(testCtx(), &runtimev1.GetTableCardinalityRequest{
		InstanceId: instanceId,
		TableName:  "test",
	})
	require.NoError(t, err)
	require.Equal(t, int64(0), cr.Cardinality)
}

func TestServer_ProfileColumns(t *testing.T) {
	server, instanceId := getTableTestServer(t)
	cr, err := server.ProfileColumns(testCtx(), &runtimev1.ProfileColumnsRequest{
		InstanceId: instanceId,
		TableName:  "test",
	})
	require.NoError(t, err)
	require.Equal(t, 2, len(cr.GetProfileColumns()))
	require.Equal(t, "a", cr.GetProfileColumns()[0].Name)
	require.Equal(t, "INTEGER", cr.GetProfileColumns()[0].Type)
	//require.Equal(t, int32(1), cr.GetProfileColumns()[0].LargestStringLength)

	require.Equal(t, "b\"b", cr.GetProfileColumns()[1].Name)
	require.Equal(t, "INTEGER", cr.GetProfileColumns()[1].Type)
	//require.Equal(t, int32(len("10")), cr.GetProfileColumns()[1].LargestStringLength)
}

func TestServer_ProfileColumns_DuplicateNames(t *testing.T) {
	server, instanceId := getTableTestServerWithSql(t, "select * from (select 1 as a) a join (select 1 as a) b on a.a = b.a")
	cr, err := server.ProfileColumns(testCtx(), &runtimev1.ProfileColumnsRequest{
		InstanceId: instanceId,
		TableName:  "test",
	})
	require.NoError(t, err)
	require.Equal(t, 2, len(cr.GetProfileColumns()))
	require.Equal(t, "a", cr.GetProfileColumns()[0].Name)
	require.Equal(t, "INTEGER", cr.GetProfileColumns()[0].Type)

	require.Equal(t, "a:1", cr.GetProfileColumns()[1].Name)
	require.Equal(t, "INTEGER", cr.GetProfileColumns()[1].Type)
}

func TestServer_ProfileColumns_EmptyModel(t *testing.T) {
	server, instanceId := getTableTestServerWithEmptyModel(t)
	cr, err := server.ProfileColumns(testCtx(), &runtimev1.ProfileColumnsRequest{
		InstanceId: instanceId,
		TableName:  "test",
	})
	require.NoError(t, err)
	require.Equal(t, 2, len(cr.GetProfileColumns()))
	require.Equal(t, "a", cr.GetProfileColumns()[0].Name)
	require.Equal(t, "INTEGER", cr.GetProfileColumns()[0].Type)
	require.Equal(t, int32(0), cr.GetProfileColumns()[0].LargestStringLength)

	require.Equal(t, "b\"b", cr.GetProfileColumns()[1].Name)
	require.Equal(t, "INTEGER", cr.GetProfileColumns()[1].Type)
	require.Equal(t, int32(0), cr.GetProfileColumns()[1].LargestStringLength)
}

func TestServer_TableRows(t *testing.T) {
	server, instanceId := getTableTestServer(t)
	cr, err := server.GetTableRows(testCtx(), &runtimev1.GetTableRowsRequest{
		InstanceId: instanceId,
		TableName:  "test",
		Limit:      1,
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(cr.Data))
}

func TestServer_TableRows_EmptyModel(t *testing.T) {
	server, instanceId := getTableTestServerWithEmptyModel(t)
	cr, err := server.GetTableRows(testCtx(), &runtimev1.GetTableRowsRequest{
		InstanceId: instanceId,
		TableName:  "test",
		Limit:      1,
	})
	require.NoError(t, err)
	require.Equal(t, 0, len(cr.Data))
}

func getTableTestServer(t *testing.T) (*Server, string) {
	rt, instanceID := testruntime.NewInstanceWithModel(t, "test", `
		SELECT 1::int AS a, 10::int AS "b""b"
	`)

	server, err := NewServer(&Options{}, rt, nil)
	require.NoError(t, err)

	return server, instanceID
}

func getTableTestServerWithSql(t *testing.T, sql string) (*Server, string) {
	rt, instanceID := testruntime.NewInstanceWithModel(t, "test", sql)

	server, err := NewServer(&Options{}, rt, nil)
	require.NoError(t, err)

	return server, instanceID
}

func getSingleValue(t *testing.T, rows *sqlx.Rows) int {
	var val int
	if rows.Next() {
		err := rows.Scan(&val)
		require.NoError(t, err)
	}
	rows.Close()
	return val
}

func getTableTestServerWithEmptyModel(t *testing.T) (*Server, string) {
	rt, instanceID := testruntime.NewInstanceWithModel(t, "test", `
		SELECT 1::int AS a, 10::int AS "b""b" where 1<>1
	`)

	server, err := NewServer(&Options{}, rt, nil)
	require.NoError(t, err)

	return server, instanceID
}
