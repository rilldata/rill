package s3_test

import (
	"testing"

	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/rilldata/rill/runtime/testruntime/testmode"
)

func TestOLAPToObjectStoreS3(t *testing.T) {
	testmode.Expensive(t)

	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		TestConnectors: []string{"clickhouse", "s3"},
		Files: map[string]string{
			"rill.yaml": "olap_connector: clickhouse",
			"connectors/s3.yaml": `
type: connector
driver: s3
region: us-east-1
aws_access_key_id: "{{ .env.connector.s3.aws_access_key_id }}"
aws_secret_access_key: "{{ .env.connector.s3.aws_secret_access_key }}"
`,
			"models/export.yaml": `
type: model
sql: SELECT number FROM numbers(16)
output:
  connector: s3
  path: s3://integration-test.rilldata.com/export_test
`,
		},
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 3, 0, 0)

}

func TestOLAPToObjectStoreS3FixedPath(t *testing.T) {
	testmode.Expensive(t)

	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		TestConnectors: []string{"clickhouse", "s3"},
		Files: map[string]string{
			"rill.yaml": "olap_connector: clickhouse",
			"connectors/s3.yaml": `
type: connector
driver: s3
region: us-east-1
aws_access_key_id: "{{ .env.connector.s3.aws_access_key_id }}"
aws_secret_access_key: "{{ .env.connector.s3.aws_secret_access_key }}"
`,
			"models/export.yaml": `
type: model
sql: SELECT number FROM numbers(16)
output:
  connector: s3
  path: s3://integration-test.rilldata.com/export_test/fixed.parquet
`,
		},
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 3, 0, 0)

}

func TestOLAPToObjectStoreGCS(t *testing.T) {
	testmode.Expensive(t)

	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		TestConnectors: []string{"clickhouse", "gcs_s3_compat"},
		Files: map[string]string{
			"rill.yaml": "olap_connector: clickhouse",
			"connectors/s3.yaml": `
type: connector
driver: s3
region: auto
endpoint: "https://storage.googleapis.com"
aws_access_key_id: "{{ .env.connector.gcs_s3_compat.key_id }}"
aws_secret_access_key: "{{ .env.connector.gcs_s3_compat.secret }}"
`,
			"models/export.yaml": `
type: model
sql: SELECT number FROM numbers(16)
output:
  connector: s3
  path: gs://integration-test.rilldata.com/export_test
`,
		},
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 3, 0, 0)
}
