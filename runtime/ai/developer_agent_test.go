package ai_test

import (
	"strings"
	"testing"

	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/ai"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func TestDeveloperShopify(t *testing.T) {
	// Setup a basic empty project
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		AIConnector: "openai",
		Files: map[string]string{
			"rill.yaml": `
olap_connector: duckdb
`,
			"connectors/duckdb.yaml": `
type: connector
driver: duckdb
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 2, 0, 0)

	// Initialize eval
	s := newEval(t, rt, instanceID)

	// Ask it to add a Shopify dashboard
	var res *ai.RouterAgentResult
	_, err := s.CallTool(t.Context(), ai.RoleUser, ai.RouterAgentName, &res, ai.RouterAgentArgs{
		Prompt: "Develop a dashboard of Shopify orders using mock data. Please proceed without asking clarifying questions.",
	})
	require.NoError(t, err)
	require.Equal(t, ai.DeveloperAgentName, res.Agent)

	// Verify it created a Shopify orders model
	msg, ok := s.LatestMessage(
		ai.FilterByTool(ai.DevelopFileName),
		ai.FilterByType(ai.MessageTypeCall),
	)
	require.True(t, ok)
	args := s.MustUnmarshalMessageContent(msg).(*ai.DevelopFileArgs)
	require.Contains(t, []string{"explore", "canvas", "metrics_view"}, args.Type)

	// Check that it doesn't have any parse or reconcile errors.
	testruntime.RequireReconcileState(t, rt, instanceID, -1, 0, 0)

	// Check there's a model and metrics view created related to shopify
	ctrl, err := rt.Controller(t.Context(), instanceID)
	require.NoError(t, err)
	models, err := ctrl.List(t.Context(), runtime.ResourceKindModel, "", false)
	require.NoError(t, err)
	metricsViews, err := ctrl.List(t.Context(), runtime.ResourceKindMetricsView, "", false)
	require.NoError(t, err)

	foundModel := false
	for _, m := range models {
		if strings.Contains(m.Meta.Name.Name, "shopify") {
			foundModel = true
			break
		}
	}
	require.True(t, foundModel, "expected a model related to shopify")

	foundMV := false
	for _, mv := range metricsViews {
		if strings.Contains(mv.Meta.Name.Name, "shopify") {
			foundMV = true
			break
		}
	}
	require.True(t, foundMV, "expected a metrics view related to shopify")
}

func TestClickhousePlayground(t *testing.T) {
	// Setup a basic empty project
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		AIConnector: "openai",
		Files: map[string]string{
			"rill.yaml": ``,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 1, 0, 0)

	// Initialize eval
	s := newEval(t, rt, instanceID)

	// Ask it to build something with the Clickhouse playground
	var res *ai.RouterAgentResult
	_, err := s.CallTool(t.Context(), ai.RoleUser, ai.RouterAgentName, &res, ai.RouterAgentArgs{
		Prompt: "Clickhouse has a public playground database on play.clickhouse.com:9440 (username `play` and no password). Can you set it up as the OLAP connector and develop a dashboard based on one of its tables? Please proceed without asking clarifying questions.",
	})
	require.NoError(t, err)
	require.Equal(t, ai.DeveloperAgentName, res.Agent)

	// Check that it doesn't have any parse or reconcile errors.
	testruntime.RequireReconcileState(t, rt, instanceID, -1, 0, 0)

	// Check there's a metrics view created
	ctrl, err := rt.Controller(t.Context(), instanceID)
	require.NoError(t, err)
	metricsViews, err := ctrl.List(t.Context(), runtime.ResourceKindMetricsView, "", false)
	require.NoError(t, err)
	require.Greater(t, len(metricsViews), 0)
}

func TestS3Model(t *testing.T) {
	// Setup a basic empty project
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		AIConnector:    "openai",
		TestConnectors: []string{"s3"}, // Add environment variables for the test S3 connector
		Files: map[string]string{
			"rill.yaml": `
olap_connector: duckdb
`,
			"connectors/duckdb.yaml": `
type: connector
driver: duckdb
managed: true
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 2, 0, 0)

	// Initialize eval
	s := newEval(t, rt, instanceID)

	// Ask it to build a DuckDB model for CSV files in S3
	var res *ai.RouterAgentResult
	_, err := s.CallTool(t.Context(), ai.RoleUser, ai.RouterAgentName, &res, ai.RouterAgentArgs{
		Prompt: "I have some CSV files in S3. Can you create a connector for S3 and a DuckDB model that loads the data at s3://integration-test.rilldata.com/glob_test/y=*/*.csv? I've already added environment variables for S3 access. Please proceed without asking clarifying questions.",
	})
	require.NoError(t, err)
	require.Equal(t, ai.DeveloperAgentName, res.Agent)

	// Check that it doesn't have any parse or reconcile errors.
	testruntime.RequireReconcileState(t, rt, instanceID, -1, 0, 0)

	// Check there's a DuckDB and S3 connector
	ctrl, err := rt.Controller(t.Context(), instanceID)
	require.NoError(t, err)
	connectors, err := ctrl.List(t.Context(), runtime.ResourceKindConnector, "", false)
	require.NoError(t, err)
	var connectorNames []string
	for _, c := range connectors {
		connectorNames = append(connectorNames, c.Meta.Name.Name)
	}
	require.Contains(t, connectorNames, "duckdb")
	require.Contains(t, connectorNames, "s3")

	// Check there's a model created
	models, err := ctrl.List(t.Context(), runtime.ResourceKindModel, "", false)
	require.NoError(t, err)
	require.Greater(t, len(models), 0)
}

func TestS3Introspection(t *testing.T) {
	// Setup a project with the S3 connector
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		AIConnector:    "openai",
		TestConnectors: []string{"s3"}, // Add environment variables for the test S3 connector
		Files: map[string]string{
			"rill.yaml": `
olap_connector: duckdb
`,
			"connectors/duckdb.yaml": `
type: connector
driver: duckdb
managed: true
`,
			"connectors/s3.yaml": `
type: connector
driver: s3
aws_access_key_id: "{{ .env.connector.s3.aws_access_key_id }}"
aws_secret_access_key: "{{ .env.connector.s3.aws_secret_access_key }}"
region: us-east-1
path_prefixes: [s3://integration-test.rilldata.com]
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 3, 0, 0)

	// Initialize eval
	s := newEval(t, rt, instanceID)

	// Ask it to describe the S3 data
	var res *ai.RouterAgentResult
	_, err := s.CallTool(t.Context(), ai.RoleUser, ai.RouterAgentName, &res, ai.RouterAgentArgs{
		Prompt: "I have some data in S3. Can you tell me what buckets are available? And also show me a little preview of what files are available at s3://integration-test.rilldata.com/glob_test/?",
		Agent:  ai.DeveloperAgentName,
	})
	require.NoError(t, err)
	require.Equal(t, ai.DeveloperAgentName, res.Agent)

	// Check that it doesn't change anything in the project.
	testruntime.RequireReconcileState(t, rt, instanceID, 3, 0, 0)
}

func TestFixMetricsViewBug(t *testing.T) {
	// Setup a project with a reconcile error
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		AIConnector:    "openai",
		TestConnectors: []string{"s3"}, // Add environment variables for the test S3 connector
		Files: map[string]string{
			"rill.yaml": `olap_connector: duckdb`,
			`connectors/duckdb.yaml`: `
type: connector
driver: duckdb
managed: true
`,
			`models/mock_data.yaml`: `
type: model
materialize: true
sql: |
  SELECT '2025-01-01T00:00:00Z'::TIMESTAMP AS event_time, 'Denmark' AS country, 100 AS value
  UNION ALL
  SELECT '2025-01-02T00:00:00Z'::TIMESTAMP AS event_time, 'Sweden' AS country, 150 AS value
  UNION ALL
  SELECT '2025-01-03T00:00:00Z'::TIMESTAMP AS event_time, 'Norway' AS country, 200 AS value
`,
			`metrics/mock_data.yaml`: `
type: metrics_view
model: mock_data
timeseries: event_time
dimensions:
  - column: countr
measures:
  - name: total_value
    expression: SUM(value)
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 5, 2, 0)

	// Initialize eval
	s := newEval(t, rt, instanceID)

	// Ask it to build a DuckDB model for CSV files in S3
	var res *ai.RouterAgentResult
	_, err := s.CallTool(t.Context(), ai.RoleUser, ai.RouterAgentName, &res, ai.RouterAgentArgs{
		Prompt: "This metrics view has an error. Can you fix it?",
		DeveloperAgentArgs: &ai.DeveloperAgentArgs{
			CurrentFilePath: "/metrics/mock_data.yaml",
		},
	})
	require.NoError(t, err)
	require.Equal(t, ai.DeveloperAgentName, res.Agent)

	// Check that it doesn't have any parse or reconcile errors.
	testruntime.RequireReconcileState(t, rt, instanceID, 5, 0, 0)
}
