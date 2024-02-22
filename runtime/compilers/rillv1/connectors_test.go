package rillv1

import (
	"context"
	"testing"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"

	_ "github.com/rilldata/rill/runtime/drivers/druid"
	_ "github.com/rilldata/rill/runtime/drivers/duckdb"
	_ "github.com/rilldata/rill/runtime/drivers/gcs"
	_ "github.com/rilldata/rill/runtime/drivers/s3"
)

func TestAnalyzeConnectors(t *testing.T) {
	ctx := context.Background()
	repo := makeRepo(t, map[string]string{
		`rill.yaml`: `
olap_connector: druid
connectors:
- name: my-s3
  type: s3
`,
		// GCS source, not configured with a custom name in rill.yaml
		`sources/bar.yaml`: `
connector: gcs
uri: gs://path/to/bar
`,
		// S3 source, with a custom name in rill.yaml
		`sources/foo.yaml`: `
connector: my-s3
uri: s3://path/to/foo
`,
		// DuckDB source, referencing a tertiary connector
		`sources/foobar.yaml`: `
connector: duckdb
sql: SELECT * FROM read_csv('s3://bucket/path.csv')
`,
	})

	p, err := Parse(ctx, repo, "", "", "duckdb")
	require.NoError(t, err)

	cs, err := p.AnalyzeConnectors(ctx)
	require.NoError(t, err)

	require.Len(t, cs, 5)

	c := cs[0]
	require.Len(t, c.Resources, 3)
	require.Equal(t, "druid", c.Name)
	require.Equal(t, "druid", c.Driver)
	require.Equal(t, false, c.AnonymousAccess)
	require.Equal(t, drivers.Connectors["druid"].Spec(), c.Spec)

	c = cs[1]
	require.Len(t, c.Resources, 1)
	require.Equal(t, "duckdb", c.Name)
	require.Equal(t, "duckdb", c.Driver)
	require.Equal(t, false, c.AnonymousAccess)
	require.Equal(t, drivers.Connectors["duckdb"].Spec(), c.Spec)

	c = cs[2]
	require.Len(t, c.Resources, 1)
	require.Equal(t, "gcs", c.Name)
	require.Equal(t, "gcs", c.Driver)
	require.Equal(t, false, c.AnonymousAccess)
	require.Equal(t, drivers.Connectors["gcs"].Spec(), c.Spec)

	c = cs[3]
	require.Len(t, c.Resources, 1)
	require.Equal(t, "my-s3", c.Name)
	require.Equal(t, "s3", c.Driver)
	require.Equal(t, false, c.AnonymousAccess)
	require.Equal(t, drivers.Connectors["s3"].Spec(), c.Spec)

	c = cs[4]
	require.Len(t, c.Resources, 1)
	require.Equal(t, "s3", c.Name)
	require.Equal(t, "s3", c.Driver)
	require.Equal(t, false, c.AnonymousAccess)
	require.Equal(t, drivers.Connectors["s3"].Spec(), c.Spec)
}
