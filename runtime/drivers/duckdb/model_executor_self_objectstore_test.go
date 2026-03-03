package duckdb_test

import (
	"context"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	activity "github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/rilldata/rill/runtime/testruntime/testmode"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestDuckDBToObjectStoreS3(t *testing.T) {
	testmode.Expensive(t)

	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		TestConnectors: []string{"s3"},
		Files: map[string]string{
			"connectors/s3.yaml": `
type: connector
driver: s3
region: us-east-1
aws_access_key_id: "{{ .env.connector.s3.aws_access_key_id }}"
aws_secret_access_key: "{{ .env.connector.s3.aws_secret_access_key }}"
`,
			"models/export.yaml": `
type: model
sql: SELECT * FROM range(16)
output:
  connector: s3
  path: s3://integration-test.rilldata.com/export_test
`,
		},
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 3, 0, 0)

	testExportedObjectExists(t, "s3", rt, id)
}

func TestDuckDBToObjectStoreS3FixedPath(t *testing.T) {
	testmode.Expensive(t)

	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		TestConnectors: []string{"s3"},
		Files: map[string]string{
			"connectors/s3.yaml": `
type: connector
driver: s3
region: us-east-1
aws_access_key_id: "{{ .env.connector.s3.aws_access_key_id }}"
aws_secret_access_key: "{{ .env.connector.s3.aws_secret_access_key }}"
`,
			"models/export.yaml": `
type: model
sql: SELECT * FROM range(16)
output:
  connector: s3
  path: s3://integration-test.rilldata.com/export_test/fixed.parquet
`,
		},
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 3, 0, 0)

	testExportedObjectExists(t, "s3", rt, id)
}

func TestDuckDBToObjectStoreGCS(t *testing.T) {
	testmode.Expensive(t)

	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		TestConnectors: []string{"gcs_s3_compat"},
		Files: map[string]string{
			"connectors/gcs.yaml": `
type: connector
driver: gcs
key_id: "{{ .env.connector.gcs_s3_compat.key_id }}"
secret: "{{ .env.connector.gcs_s3_compat.secret }}"
`,
			"models/export.yaml": `
type: model
sql: SELECT * FROM range(16)
output:
  connector: gcs
  path: gs://integration-test.rilldata.com/export_test
`,
		},
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 3, 0, 0)

	testExportedObjectExists(t, "gcs", rt, id)
}

func testExportedObjectExists(t *testing.T, driver string, rt *runtime.Runtime, id string) {
	r := testruntime.GetResource(t, rt, id, runtime.ResourceKindModel, "export")
	require.NotNil(t, r, "export")
	path := r.GetModel().State.ResultProperties.AsMap()["path"].(string)

	ctx := context.Background()
	handle, _, err := rt.AcquireHandle(ctx, id, driver)
	conn, err := drivers.Open(driver, "default", handle.Config(), storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })

	objectStore, ok := conn.AsObjectStore()
	require.True(t, ok)

	uri, err := url.Parse(path)
	require.NoError(t, err)
	objects, nextToken, err := objectStore.ListObjects(t.Context(), uri.Host, strings.TrimPrefix(uri.Path, "/"), "/", 100, "")

	require.NoError(t, err)
	require.Empty(t, nextToken)
	require.Len(t, objects, 1)
	require.WithinDuration(t, objects[0].UpdatedOn, time.Now(), 30*time.Second)
}
