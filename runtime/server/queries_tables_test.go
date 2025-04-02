package server_test

import (
	"context"
	"testing"

	"github.com/rilldata/rill/runtime/pkg/ratelimit"
	"github.com/rilldata/rill/runtime/server"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func TestServer_TableCardinality(t *testing.T) {
	t.Parallel()
	server, instanceId := getTableTestServer(t)
	cr, err := server.TableCardinality(testCtx(), &runtimev1.TableCardinalityRequest{
		InstanceId: instanceId,
		TableName:  "test",
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), cr.Cardinality)
}

func TestServer_TableCardinality_EmptyModel(t *testing.T) {
	t.Parallel()
	server, instanceId := getTableTestServerWithEmptyModel(t)
	cr, err := server.TableCardinality(testCtx(), &runtimev1.TableCardinalityRequest{
		InstanceId: instanceId,
		TableName:  "test",
	})
	require.NoError(t, err)
	require.Equal(t, int64(0), cr.Cardinality)
}

func TestServer_TableColumns(t *testing.T) {
	t.Parallel()
	server, instanceId := getTableTestServer(t)
	cr, err := server.TableColumns(testCtx(), &runtimev1.TableColumnsRequest{
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

func TestServer_TableColumns_DuplicateNames(t *testing.T) {
	t.Parallel()
	server, instanceId := getTableTestServerWithSql(t, "select * from (select 1 as a) a join (select 1 as a) b on a.a = b.a")
	cr, err := server.TableColumns(testCtx(), &runtimev1.TableColumnsRequest{
		InstanceId: instanceId,
		TableName:  "test",
	})
	require.NoError(t, err)
	require.Equal(t, 2, len(cr.GetProfileColumns()))
	require.Equal(t, "a", cr.GetProfileColumns()[0].Name)
	require.Equal(t, "INTEGER", cr.GetProfileColumns()[0].Type)

	require.Equal(t, "a_1", cr.GetProfileColumns()[1].Name)
	require.Equal(t, "INTEGER", cr.GetProfileColumns()[1].Type)
}

func TestServer_TableColumns_EmptyModel(t *testing.T) {
	t.Parallel()
	server, instanceId := getTableTestServerWithEmptyModel(t)
	cr, err := server.TableColumns(testCtx(), &runtimev1.TableColumnsRequest{
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
	t.Parallel()
	server, instanceId := getTableTestServer(t)
	cr, err := server.TableRows(testCtx(), &runtimev1.TableRowsRequest{
		InstanceId: instanceId,
		TableName:  "test",
		Limit:      1,
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(cr.Data))
}

func TestServer_TableRows_EmptyModel(t *testing.T) {
	t.Parallel()
	server, instanceId := getTableTestServerWithEmptyModel(t)
	cr, err := server.TableRows(testCtx(), &runtimev1.TableRowsRequest{
		InstanceId: instanceId,
		TableName:  "test",
		Limit:      1,
	})
	require.NoError(t, err)
	require.Equal(t, 0, len(cr.Data))
}

func getTableTestServer(t *testing.T) (*server.Server, string) {
	rt, instanceID := testruntime.NewInstanceWithModel(t, "test", `
		SELECT 1::int AS a, 10::int AS "b""b"
	`)

	server, err := server.NewServer(context.Background(), &server.Options{}, rt, nil, ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	return server, instanceID
}

func getTableTestServerWithSql(t *testing.T, sql string) (*server.Server, string) {
	rt, instanceID := testruntime.NewInstanceWithModel(t, "test", sql)

	server, err := server.NewServer(context.Background(), &server.Options{}, rt, nil, ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	return server, instanceID
}

func getTableTestServerWithEmptyModel(t *testing.T) (*server.Server, string) {
	rt, instanceID := testruntime.NewInstanceWithModel(t, "test", `
		SELECT 1::int AS a, 10::int AS "b""b" where 1<>1
	`)

	server, err := server.NewServer(context.Background(), &server.Options{}, rt, nil, ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	return server, instanceID
}
