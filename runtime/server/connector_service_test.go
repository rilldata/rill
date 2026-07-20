package server_test

import (
	"context"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/ratelimit"
	"github.com/rilldata/rill/runtime/server"
	"github.com/rilldata/rill/runtime/server/auth"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestConnectorServiceAuth(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithModel(t, "foo", "SELECT 1 AS a")

	srv, err := server.NewServer(context.Background(), &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	// Claims matching a Rill Cloud project viewer; must not grant access to connector introspection.
	viewerCtx := auth.WithClaims(context.Background(), &runtime.SecurityClaims{
		Permissions: []runtime.Permission{runtime.ReadAPI, runtime.ReadMetrics, runtime.ReadObjects, runtime.UseAI},
	})

	// Claims with ReadInstance, as granted to users with status/manage access.
	adminCtx := auth.WithClaims(context.Background(), &runtime.SecurityClaims{
		Permissions: []runtime.Permission{runtime.ReadInstance},
	})

	t.Run("ListBuckets", func(t *testing.T) {
		_, err := srv.ListBuckets(viewerCtx, &runtimev1.ListBucketsRequest{InstanceId: instanceID, Connector: "duckdb"})
		require.ErrorIs(t, err, server.ErrForbidden)
	})

	t.Run("ListObjects", func(t *testing.T) {
		_, err := srv.ListObjects(viewerCtx, &runtimev1.ListObjectsRequest{InstanceId: instanceID, Connector: "duckdb"})
		require.ErrorIs(t, err, server.ErrForbidden)
	})

	t.Run("OLAPListTables", func(t *testing.T) {
		_, err := srv.OLAPListTables(viewerCtx, &runtimev1.OLAPListTablesRequest{InstanceId: instanceID})
		require.ErrorIs(t, err, server.ErrForbidden)

		res, err := srv.OLAPListTables(adminCtx, &runtimev1.OLAPListTablesRequest{InstanceId: instanceID})
		require.NoError(t, err)
		require.Len(t, res.Tables, 1)
		require.Equal(t, "foo", res.Tables[0].Name)
	})

	t.Run("OLAPGetTable", func(t *testing.T) {
		_, err := srv.OLAPGetTable(viewerCtx, &runtimev1.OLAPGetTableRequest{InstanceId: instanceID, Table: "foo"})
		require.ErrorIs(t, err, server.ErrForbidden)

		res, err := srv.OLAPGetTable(adminCtx, &runtimev1.OLAPGetTableRequest{InstanceId: instanceID, Table: "foo"})
		require.NoError(t, err)
		require.Len(t, res.Schema.Fields, 1)
	})

	t.Run("ListDatabaseSchemas", func(t *testing.T) {
		_, err := srv.ListDatabaseSchemas(viewerCtx, &runtimev1.ListDatabaseSchemasRequest{InstanceId: instanceID, Connector: "duckdb"})
		require.ErrorIs(t, err, server.ErrForbidden)

		res, err := srv.ListDatabaseSchemas(adminCtx, &runtimev1.ListDatabaseSchemasRequest{InstanceId: instanceID, Connector: "duckdb"})
		require.NoError(t, err)
		require.NotEmpty(t, res.DatabaseSchemas)
	})

	t.Run("ListTables", func(t *testing.T) {
		_, err := srv.ListTables(viewerCtx, &runtimev1.ListTablesRequest{InstanceId: instanceID, Connector: "duckdb"})
		require.ErrorIs(t, err, server.ErrForbidden)
	})

	t.Run("GetTable", func(t *testing.T) {
		_, err := srv.GetTable(viewerCtx, &runtimev1.GetTableRequest{InstanceId: instanceID, Connector: "duckdb", Table: "foo"})
		require.ErrorIs(t, err, server.ErrForbidden)
	})
}
