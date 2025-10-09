package parser

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
		// S3 source, with a dedicated connector
		`sources/foo-dedicated.yaml`: `
connector: s3-dedicated
uri: s3://path/to/foo
`,
		// DuckDB source, referencing a tertiary connector
		`sources/foobar.yaml`: `
connector: duckdb
sql: SELECT * FROM read_csv('s3://bucket/path.csv')
`,
		"/alerts/a1.yaml": `
type: alert
display_name: Test Alert
refs:
- type: MetricsView
  name: mv1
watermark: inherit
intervals:
  duration: P1D
query:
  name: MetricsViewAggregation
  args:
    metrics_view: mv1
    dimensions:
    - name: country
    measures:
    - name: measure_0
    time_range:
      iso_duration: P1W
    having:
      cond:
        op: OPERATION_GTE
        exprs:
        - ident: measure_0
        - val: 4
notify:
  slack:
    channels:
    - channel-test
    users:
    - user@example.com
`,
		// Dedicated S3 connector
		"/connectors/s3-dedicated.yaml": `
driver: s3
name: s3-dedicated
region: us-west-2
`,
	})

	p, err := Parse(ctx, repo, "", "", "duckdb")
	require.NoError(t, err)

	cs := p.AnalyzeConnectors(ctx)
	for _, c := range cs {
		if c.Err != nil {
			require.NoError(t, c.Err)
		}
	}

	require.Len(t, cs, 7)

	c := cs[0]
	require.Len(t, c.Resources, 4)
	require.Equal(t, "druid", c.Name)
	require.Equal(t, "druid", c.Driver)
	require.Equal(t, false, c.AnonymousAccess)
	require.Equal(t, drivers.Connectors["druid"].Spec(), *c.Spec)

	c = cs[1]
	require.Len(t, c.Resources, 1)
	require.Equal(t, "duckdb", c.Name)
	require.Equal(t, "duckdb", c.Driver)
	require.Equal(t, false, c.AnonymousAccess)
	require.Equal(t, drivers.Connectors["duckdb"].Spec(), *c.Spec)

	c = cs[2]
	require.Len(t, c.Resources, 1)
	require.Equal(t, "gcs", c.Name)
	require.Equal(t, "gcs", c.Driver)
	require.Equal(t, false, c.AnonymousAccess)
	require.Equal(t, drivers.Connectors["gcs"].Spec(), *c.Spec)

	c = cs[3]
	require.Len(t, c.Resources, 1)
	require.Equal(t, "my-s3", c.Name)
	require.Equal(t, "s3", c.Driver)
	require.Equal(t, false, c.AnonymousAccess)
	require.Equal(t, drivers.Connectors["s3"].Spec(), *c.Spec)

	c = cs[4]
	require.Len(t, c.Resources, 1)
	require.Equal(t, "s3", c.Name)
	require.Equal(t, "s3", c.Driver)
	require.Equal(t, false, c.AnonymousAccess)
	require.Equal(t, drivers.Connectors["s3"].Spec(), *c.Spec)

	c = cs[5]
	require.Len(t, c.Resources, 1)
	require.Equal(t, "s3-dedicated", c.Name)
	require.Equal(t, "s3", c.Driver)
	require.Equal(t, false, c.AnonymousAccess)
	require.Equal(t, drivers.Connectors["s3"].Spec(), *c.Spec)
	require.Equal(t, c.DefaultConfig["region"], "us-west-2")

	c = cs[6]
	require.Len(t, c.Resources, 1)
	require.Equal(t, "slack", c.Name)
	require.Equal(t, "slack", c.Driver)
	require.Equal(t, false, c.AnonymousAccess)
	require.Equal(t, drivers.Connectors["slack"].Spec(), *c.Spec)
}

func TestAnonSlackConnector(t *testing.T) {
	ctx := context.Background()
	repo := makeRepo(t, map[string]string{
		`rill.yaml`: ``,
		"/alerts/a1.yaml": `
type: alert
display_name: Test Alert
refs:
- type: MetricsView
  name: mv1
watermark: inherit
intervals:
  duration: P1D
query:
  name: MetricsViewAggregation
  args:
    metrics_view: mv1
    dimensions:
    - name: country
    measures:
    - name: measure_0
    time_range:
      iso_duration: P1W
    having:
      cond:
        op: OPERATION_GTE
        exprs:
        - ident: measure_0
        - val: 4
notify:
  slack:
    webhooks:
    - https://hooks.slack.com/services/123/456/789
`,
	})

	p, err := Parse(ctx, repo, "", "", "duckdb")
	require.NoError(t, err)

	cs := p.AnalyzeConnectors(ctx)
	for _, c := range cs {
		if c.Err != nil {
			require.NoError(t, c.Err)
		}
	}

	require.Len(t, cs, 1)

	c := cs[0]
	require.Len(t, c.Resources, 1)
	require.Equal(t, "slack", c.Name)
	require.Equal(t, "slack", c.Driver)
	require.Equal(t, true, c.AnonymousAccess)
	require.Equal(t, drivers.Connectors["slack"].Spec(), *c.Spec)
}

func TestManagedConnectorPropagation(t *testing.T) {
	ctx := context.Background()
	repo := makeRepo(t, map[string]string{
		`rill.yaml`: `
olap_connector: duckdb
`,
		"/connectors/duckdb.yaml": `
type: connector
driver: duckdb
managed: true
`,
	})

	p, err := Parse(ctx, repo, "", "", "duckdb")
	require.NoError(t, err)

	// Verify that the managed field is properly set in ConnectorSpec
	require.Len(t, p.Resources, 1)

	// Get the resource from the map
	var r *Resource
	for _, res := range p.Resources {
		r = res
		break
	}
	require.NotNil(t, r)
	require.Equal(t, "duckdb", r.Name.Name)
	require.True(t, r.ConnectorSpec.Provision, "Provision should be true when managed: true")

	// Verify that the managed field is also in the Properties map so drivers can access it
	require.NotNil(t, r.ConnectorSpec.Properties)
	require.Equal(t, "true", r.ConnectorSpec.Properties["managed"], "managed field should be in Properties map for driver access")
}

func TestManagedConnectorFalse(t *testing.T) {
	ctx := context.Background()
	repo := makeRepo(t, map[string]string{
		`rill.yaml`: `
olap_connector: duckdb
`,
		"/connectors/duckdb.yaml": `
type: connector
driver: duckdb
managed: false
`,
	})

	p, err := Parse(ctx, repo, "", "", "duckdb")
	require.NoError(t, err)

	// Verify that the managed field is properly set in ConnectorSpec
	require.Len(t, p.Resources, 1)

	// Get the resource from the map
	var r *Resource
	for _, res := range p.Resources {
		r = res
		break
	}
	require.NotNil(t, r)
	require.Equal(t, "duckdb", r.Name.Name)
	require.False(t, r.ConnectorSpec.Provision, "Provision should be false when managed: false")

	// Verify that the managed field is also in the Properties map
	require.NotNil(t, r.ConnectorSpec.Properties)
	require.Equal(t, "false", r.ConnectorSpec.Properties["managed"], "managed field should be in Properties map for driver access")
}

func TestManagedConnectorWithArgs(t *testing.T) {
	ctx := context.Background()
	repo := makeRepo(t, map[string]string{
		`rill.yaml`: `
olap_connector: duckdb
`,
		"/connectors/duckdb.yaml": `
type: connector
driver: duckdb
managed:
  arg1: value1
  arg2: value2
`,
	})

	p, err := Parse(ctx, repo, "", "", "duckdb")
	require.NoError(t, err)

	// Verify that the managed field with args is properly set
	require.Len(t, p.Resources, 1)

	// Get the resource from the map
	var r *Resource
	for _, res := range p.Resources {
		r = res
		break
	}
	require.NotNil(t, r)
	require.Equal(t, "duckdb", r.Name.Name)
	require.True(t, r.ConnectorSpec.Provision, "Provision should be true when managed has args")
	require.NotNil(t, r.ConnectorSpec.ProvisionArgs, "ProvisionArgs should be set when managed has args")

	// Verify that the managed field is also in the Properties map
	require.NotNil(t, r.ConnectorSpec.Properties)
	require.Equal(t, "true", r.ConnectorSpec.Properties["managed"], "managed field should be 'true' in Properties map when args are provided")
}
