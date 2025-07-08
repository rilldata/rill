package duckdb_test

import (
	"testing"

	"github.com/rilldata/rill/runtime/testruntime"

	_ "github.com/rilldata/rill/runtime/resolvers"
)

func TestCreateSecrets(t *testing.T) {
	files := map[string]string{
		".env": `
s3_key_id: foo
s3_secret: bar
`,
		"s3.yaml": `
type: connector
driver: s3
region: eu-north-1
aws_access_key_id: "{{.env.s3_key_id}}"
aws_secret_access_key: "{{.env.s3_secret}}"
`,
		"duckdb.yaml": `
type: connector
driver: duckdb
secrets: s3
`,
		"secrets.yaml": `
type: model
materialize: true
connector: duckdb
sql: > 
  SELECT
    regexp_extract(secret_string, 'region=([^;]+);', 1) as region,
    regexp_extract(secret_string, 'endpoint=([^;]+);', 1) as endpoint,
    regexp_extract(secret_string, 'key_id=([^;]+);', 1) as key_id,
    regexp_extract(secret_string, 'secret_key=([^;]+);', 1) as secret_key
  FROM duckdb_secrets()
`,
	}

	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files:             files,
		DisableHostAccess: true,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 1, 0)

	testruntime.RequireResolve(t, rt, id, &testruntime.RequireResolveOptions{
		Resolver:   "sql",
		Properties: map[string]any{"sql": `SELECT * FROM secrets`},
		Result: []map[string]any{{
			"region":     "eu-north-1",
			"endpoint":   "",
			"key_id":     "foo",
			"secret_key": "", // Gets redacted
		}},
	})

}
