package server_test

import (
	"context"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/ratelimit"
	"github.com/rilldata/rill/runtime/server"
	"github.com/rilldata/rill/runtime/server/auth"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestServer_TestSimpleSQLQueryResolver(t *testing.T) {
	t.Parallel()

	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": ``,
			// Model
			"ad_bids.sql": `SELECT now() AS time, 'DA' AS country, 3 as price`,
			// Duckdb connector
			"custom_duckdb.yaml": `
type: connector
driver: duckdb
`,
		},
	})

	// Create some externally managed tables
	olapExecAdhoc(t, rt, instanceID, "duckdb", "CREATE TABLE IF NOT EXISTS foo AS SELECT now() AS time, 'DA' AS country, 3.141 as price")
	olapExecAdhoc(t, rt, instanceID, "custom_duckdb", "CREATE TABLE IF NOT EXISTS foo AS SELECT now() AS time, 'DA' AS country, 3.141 as price")

	ctx, cancel := context.WithTimeout(testCtx(), 25*time.Second)
	defer cancel()

	_, release, err := rt.Repo(ctx, instanceID)
	require.NoError(t, err)
	defer release()

	server, err := server.NewServer(ctx, &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	tt := []struct {
		name               string             // Test case name
		resolver           string             // Resolver name (e.g. ad_bids.sql - ad_bids)
		resolverProperties *structpb.Struct   // Resolver properties
		resolverArgs       *structpb.Struct   // Resolver arguments
		schema             []string           // Expected schema
		data               []*structpb.Struct // Expected data
		expectError        bool               // Expect an error
		code               codes.Code         // Expected gRPC error code
	}{
		{
			name:        "should fail with invalid resolver",
			resolver:    "invalid_resolver",
			expectError: true,
			code:        codes.Internal,
		},
		{
			name:        "should fail with missing sql query",
			resolver:    "sql",
			expectError: true,
			code:        codes.Internal, // Update the expected error code
		},
		{
			name:     "should succeed with a simple SQL query",
			resolver: "sql",
			resolverProperties: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"sql": structpb.NewStringValue("SELECT country FROM foo limit 1"),
				},
			},
			resolverArgs: &structpb.Struct{},
			schema:       []string{"country"},
			data: []*structpb.Struct{
				{
					Fields: map[string]*structpb.Value{
						"country": structpb.NewStringValue("DA"),
					},
				},
			},
		},
		// Test multiple columns
		{
			name:     "should succeed with multiple columns",
			resolver: "sql",
			resolverProperties: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"sql": structpb.NewStringValue("SELECT country, price FROM foo limit 1"),
				},
			},
			resolverArgs: &structpb.Struct{},
			schema:       []string{"country", "price"},
			data: []*structpb.Struct{
				{
					Fields: map[string]*structpb.Value{
						"country": structpb.NewStringValue("DA"),
						"price":   structpb.NewNumberValue(3.141),
					},
				},
			},
		},
		{
			name:     "should succeed with a simple SQL query with a WHERE clause",
			resolver: "sql",
			resolverProperties: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"sql": structpb.NewStringValue("SELECT country FROM foo WHERE country = 'DA'"),
				},
			},
			resolverArgs: &structpb.Struct{},
			schema:       []string{"country"},
			data: []*structpb.Struct{
				{
					Fields: map[string]*structpb.Value{
						"country": structpb.NewStringValue("DA"),
					},
				},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(testCtx(), 25*time.Second)
			defer cancel()

			req := &runtimev1.QueryResolverRequest{
				InstanceId:         instanceID,
				Resolver:           tc.resolver,
				ResolverProperties: tc.resolverProperties,
				ResolverArgs:       tc.resolverArgs,
			}
			res, err := server.QueryResolver(ctx, req)
			if tc.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			data := res.GetData()
			schema := res.GetSchema()

			// Check expected schema
			require.Equal(t, len(schema.Fields), len(tc.schema))
			for i, s := range tc.schema {
				require.Equal(t, schema.Fields[i].Name, s)
			}

			// Check expected data
			require.Equal(t, len(data), len(tc.data))
			for i, d := range data {
				require.Equal(t, d, tc.data[i])
			}
		})
	}
}

// TestQueryResolver_ProjectVariables verifies that the request's resolver properties cannot read project variables,
// which commonly hold secrets such as connector credentials, unless the caller may read sensitive instance info.
func TestQueryResolver_ProjectVariables(t *testing.T) {
	t.Parallel()

	const secret = "s3cret-value"
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
			"model.sql": `
SELECT 'US' AS country, 100 AS revenue
UNION ALL
SELECT 'UK' AS country, 200 AS revenue
`,
			"mv.yaml": `
type: metrics_view
version: 1
model: model
dimensions:
- column: country
measures:
- name: total_revenue
  expression: SUM(revenue)
`,
		},
		Variables: map[string]string{"secret": secret, "country": "US"},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 3, 0, 0)

	srv, err := server.NewServer(t.Context(), &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	metricsSQL := func(sql string) *runtimev1.QueryResolverRequest {
		return &runtimev1.QueryResolverRequest{
			InstanceId: instanceID,
			Resolver:   "metrics_sql",
			ResolverProperties: &structpb.Struct{
				Fields: map[string]*structpb.Value{"sql": structpb.NewStringValue(sql)},
			},
		}
	}

	// A viewer has the permissions of an embed token: it can query metrics, but not read sensitive instance info.
	viewerCtx := auth.WithClaims(context.Background(), &runtime.SecurityClaims{
		Permissions: []runtime.Permission{runtime.ReadMetrics},
	})

	// Templates that reference variables are rejected with an explicit error instead of resolving to empty values,
	// so neither the result nor the error can be used to learn a variable's value.
	for _, sql := range []string{
		`SELECT total_revenue FROM mv WHERE country = '{{ .env.country }}'`,
		`SELECT {{ .env.secret }} FROM mv`,
	} {
		_, err = srv.QueryResolver(viewerCtx, metricsSQL(sql))
		require.ErrorContains(t, err, "project variables are only available to templates declared in the project files")
		require.NotContains(t, err.Error(), secret)
	}

	// Templates that do not reference variables still resolve for a viewer.
	res, err := srv.QueryResolver(viewerCtx, metricsSQL(`SELECT total_revenue FROM mv WHERE country = 'US'`))
	require.NoError(t, err)
	require.Len(t, res.Data, 1)
	require.Equal(t, float64(100), res.Data[0].Fields["total_revenue"].GetNumberValue())

	// A caller that can read sensitive instance info can already read the variables through GetInstance, so templates resolve them.
	res, err = srv.QueryResolver(testCtx(), metricsSQL(`SELECT total_revenue FROM mv WHERE country = '{{ .env.country }}'`))
	require.NoError(t, err)
	require.Len(t, res.Data, 1)
	require.Equal(t, float64(100), res.Data[0].Fields["total_revenue"].GetNumberValue())
}

// TestQueryResolver_Component verifies that a request may name the component its metrics_sql was taken from,
// so that a project-declared canvas chart keeps its access to project variables while the runtime never returns the resolved SQL.
func TestQueryResolver_Component(t *testing.T) {
	t.Parallel()

	const secret = "s3cret-value"
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
			"model.sql": `
SELECT 'US' AS country, 100 AS revenue
UNION ALL
SELECT 'UK' AS country, 200 AS revenue
`,
			"mv.yaml": `
type: metrics_view
version: 1
model: model
dimensions:
- column: country
measures:
- name: total_revenue
  expression: SUM(revenue)
`,
			"project.yaml": `
type: canvas
rows:
- items:
  - custom_chart:
      metrics_sql:
      - "SELECT total_revenue FROM mv WHERE country = '{{ .env.country }}'"
      - "SELECT total_revenue FROM mv"
      vega_spec: "{}"
`,
			// A personal canvas created through the admin API by the viewer, so its queries were written by the viewer.
			"personal.yaml": `
type: canvas
annotations:
  admin_managed: "true"
  admin_owner_user_id: "viewer"
rows:
- items:
  - custom_chart:
      metrics_sql: "SELECT total_revenue FROM mv WHERE country = '{{ .env.secret }}'"
      vega_spec: "{}"
`,
		},
		Variables: map[string]string{"country": "US", "secret": secret},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 7, 0, 0)

	srv, err := server.NewServer(t.Context(), &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	req := func(component, resolver, sql string) *runtimev1.QueryResolverRequest {
		return &runtimev1.QueryResolverRequest{
			InstanceId:         instanceID,
			Resolver:           resolver,
			ResolverProperties: &structpb.Struct{Fields: map[string]*structpb.Value{"sql": structpb.NewStringValue(sql)}},
			Component:          component,
		}
	}
	viewerCtx := auth.WithClaims(context.Background(), &runtime.SecurityClaims{
		UserAttributes: map[string]any{"admin": false, "id": "viewer"},
		Permissions:    []runtime.Permission{runtime.ReadMetrics},
	})

	// A query taken from a project-declared component resolves the variable for a viewer.
	res, err := srv.QueryResolver(viewerCtx, req("project--component-0-0", "metrics_sql", `SELECT total_revenue FROM mv WHERE country = '{{ .env.country }}'`))
	require.NoError(t, err)
	require.Len(t, res.Data, 1)
	require.Equal(t, float64(100), res.Data[0].Fields["total_revenue"].GetNumberValue())

	// The query must match one of the component's saved queries, so the component cannot be used to run other templates.
	_, err = srv.QueryResolver(viewerCtx, req("project--component-0-0", "metrics_sql", `SELECT total_revenue FROM mv WHERE country = '{{ .env.secret }}'`))
	require.Equal(t, codes.InvalidArgument, status.Code(err))
	require.NotContains(t, err.Error(), secret)

	// Only metrics_sql queries can be taken from a component. (Other resolvers require ReadResolvers, which the test context has.)
	_, err = srv.QueryResolver(testCtx(), req("project--component-0-0", "sql", `SELECT '{{ .env.country }}' AS v`))
	require.Equal(t, codes.InvalidArgument, status.Code(err))

	// A query written by the viewer in their personal canvas never sees the variables.
	_, err = srv.QueryResolver(viewerCtx, req("personal--component-0-0", "metrics_sql", `SELECT total_revenue FROM mv WHERE country = '{{ .env.secret }}'`))
	require.ErrorContains(t, err, "project variables are only available to templates declared in the project files")
	require.NotContains(t, err.Error(), secret)
}
