package runtime_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/ClickHouse/clickhouse-go/v2"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/rilldata/rill/runtime/testruntime/testmode"
	"github.com/stretchr/testify/require"
)

// TestClickHouseNamedCollectionLifecycle exercises the full lifecycle of a ClickHouse named
// collection that mirrors a Rill connector resource:
//
//  1. Create a connector resource (driver: s3) in a project whose OLAP is ClickHouse.
//  2. Verify the corresponding `rill_<connector_name>` named collection appears on the CH cluster.
//  3. Edit the connector's properties; verify the named collection is updated.
//  4. Add a model that references the named collection via `s3(rill_<conn>, ...)`; ensure model
//     reconcile does not fail (we don't actually have a real S3 bucket — the model SQL targets
//     `system.one` to keep the test offline-friendly while still exercising the auto-detection path).
//  5. Delete the connector resource; verify the named collection is dropped.
//
// The test uses the cluster fixture so we also exercise the `ON CLUSTER` code path.
func TestClickHouseNamedCollectionLifecycle(t *testing.T) {
	testmode.Expensive(t)

	const connectorName = "my_bucket"
	const collectionName = "rill_" + connectorName

	files := map[string]string{
		"rill.yaml": "olap_connector: clickhouse\n",
		"connectors/" + connectorName + ".yaml": `
type: connector
driver: s3
aws_access_key_id: AKIA_TEST
aws_secret_access_key: TEST_SECRET
region: us-east-1
`,
	}

	rt, instanceID, repoPath, dsn, clusterName := testruntime.NewInstanceWithClickhouseFiles(t, files)
	require.NotEmpty(t, clusterName)

	ctx := t.Context()
	testruntime.ReconcileParserAndWait(t, rt, instanceID)

	// The connector resource should have reconciled cleanly: no errors. There are 2 resources
	// (the project parser + the connector); 0 reconcile errors; 0 parse errors.
	testruntime.RequireReconcileState(t, rt, instanceID, 2, 0, 0)

	// Open a direct connection to the CH cluster so we can inspect server state independently.
	chDB := openClickHouseDB(t, dsn)
	defer chDB.Close()

	// Verify the named collection exists with the expected fields.
	requireNamedCollectionFields(t, chDB, collectionName, map[string]string{
		"access_key_id":     "AKIA_TEST",
		"secret_access_key": "TEST_SECRET",
		"region":            "us-east-1",
	})

	// Mutate the connector to add an endpoint and re-reconcile.
	testruntime.PutFiles(t, rt, instanceID, map[string]string{
		"connectors/" + connectorName + ".yaml": `
type: connector
driver: s3
aws_access_key_id: AKIA_TEST_2
aws_secret_access_key: TEST_SECRET_2
region: us-west-2
endpoint: https://example.invalid
`,
	})
	testruntime.ReconcileParserAndWait(t, rt, instanceID)
	testruntime.RequireReconcileState(t, rt, instanceID, 2, 0, 0)

	requireNamedCollectionFields(t, chDB, collectionName, map[string]string{
		"access_key_id":     "AKIA_TEST_2",
		"secret_access_key": "TEST_SECRET_2",
		"region":            "us-west-2",
		"endpoint":          "https://example.invalid",
	})

	// Add a model that references the named collection. It must reconcile without errors. We
	// can't query an actual S3 bucket from a test cluster, so the model body is a trivial query
	// that includes a comment matching the auto-detection pattern. This validates that the
	// auto-detection path runs and does not error out when the collection exists.
	testruntime.PutFiles(t, rt, instanceID, map[string]string{
		"models/uses_named_collection.sql": `-- references s3(rill_my_bucket, url='dummy')
SELECT 1 AS x`,
	})
	testruntime.ReconcileParserAndWait(t, rt, instanceID)

	// Should still be 0 reconcile errors.
	model, err := getResource(ctx, rt, instanceID, runtime.ResourceKindModel, "uses_named_collection")
	require.NoError(t, err)
	require.NotNil(t, model.GetModel())
	if errs := model.Meta.ReconcileError; errs != "" {
		t.Fatalf("model reconcile failed: %s", errs)
	}

	// Delete the connector file. Reconcile should drop the named collection.
	testruntime.DeleteFiles(t, rt, instanceID, "connectors/"+connectorName+".yaml")
	// Also delete the model so the reconciler isn't blocked on the missing connector reference.
	testruntime.DeleteFiles(t, rt, instanceID, "models/uses_named_collection.sql")
	testruntime.ReconcileParserAndWait(t, rt, instanceID)

	// Verify the named collection is gone from the server.
	require.False(t, namedCollectionExists(t, chDB, collectionName), "expected named collection %q to be dropped", collectionName)

	_ = repoPath // unused but kept for clarity / future extension
}

// openClickHouseDB opens a direct connection to a CH cluster using the same DSN format the
// driver uses. It uses the native protocol and `default` user (matching the test fixture).
func openClickHouseDB(t *testing.T, dsn string) *sql.DB {
	t.Helper()
	opts, err := clickhouse.ParseDSN(dsn)
	require.NoError(t, err)
	db := clickhouse.OpenDB(opts)
	require.NoError(t, db.Ping())
	return db
}

func namedCollectionExists(t *testing.T, db *sql.DB, name string) bool {
	t.Helper()
	row := db.QueryRow("SELECT count() FROM system.named_collections WHERE name = ?", name)
	var count uint64
	require.NoError(t, row.Scan(&count))
	return count > 0
}

// requireNamedCollectionFields asserts the named collection contains exactly the given fields.
// We read the `collection` Map column from `system.named_collections` directly. The test fixture
// has `show_named_collections_secrets=1` so secret values are returned in plaintext.
func requireNamedCollectionFields(t *testing.T, db *sql.DB, name string, want map[string]string) {
	t.Helper()
	row := db.QueryRow("SELECT collection FROM system.named_collections WHERE name = ?", name)
	got := map[string]string{}
	require.NoError(t, row.Scan(&got))
	for k, v := range want {
		actual, ok := got[k]
		require.True(t, ok, "expected field %q in named collection %q; got fields=%v", k, name, mapKeys(got))
		require.Equal(t, v, actual, "field %q value mismatch", k)
	}
	// Also verify there are no unexpected fields beyond the ones we set, to catch mapping drift.
	for k := range got {
		_, ok := want[k]
		require.True(t, ok, "unexpected field %q=%q in named collection %q", k, got[k], name)
	}
}

func mapKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func getResource(ctx context.Context, rt *runtime.Runtime, instanceID, kind, name string) (*runtimev1.Resource, error) {
	ctrl, err := rt.Controller(ctx, instanceID)
	if err != nil {
		return nil, err
	}
	res, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: kind, Name: name}, false)
	if err != nil {
		return nil, err
	}
	return res, nil
}
