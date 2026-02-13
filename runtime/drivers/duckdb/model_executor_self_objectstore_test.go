package duckdb_test

import (
	"testing"

	"github.com/rilldata/rill/runtime/testruntime"
)

func TestDuckDBToObjectStoreS3(t *testing.T) {
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

}

func TestDuckDBToObjectStoreS3FixedPath(t *testing.T) {
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

}

func TestDuckDBToObjectStoreGCS(t *testing.T) {
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

}
