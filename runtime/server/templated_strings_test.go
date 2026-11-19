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

// TestResolveTemplatedString is a basic integration test that verifies the RPC delegates to the text resolver.
// Comprehensive test coverage lives in runtime/resolvers/text_test.go.
func TestResolveTemplatedString(t *testing.T) {
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
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 3, 0, 0)

	srv, err := server.NewServer(t.Context(), &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	res, err := srv.ResolveTemplatedString(testCtx(), &runtimev1.ResolveTemplatedStringRequest{
		InstanceId: instanceID,
		Body:       `Revenue: {{ metrics_sql "SELECT total_revenue FROM mv" }}`,
	})
	require.NoError(t, err)
	require.Equal(t, "Revenue: 300", res.Body)
}

// TestResolveTemplatedString_ProjectVariables verifies that the request's template cannot read project variables,
// which commonly hold secrets such as connector credentials, whoever the caller is.
func TestResolveTemplatedString_ProjectVariables(t *testing.T) {
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
		Variables: map[string]string{"secret": secret},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 3, 0, 0)

	srv, err := server.NewServer(t.Context(), &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	// A viewer has the permissions of an embed token: it can query data, but not read sensitive instance info.
	viewerCtx := auth.WithClaims(context.Background(), &runtime.SecurityClaims{
		Permissions: []runtime.Permission{runtime.ReadAPI, runtime.ReadMetrics},
	})

	// Every spelling a caller could use to reach a variable, including a nested metrics_sql call,
	// is rejected with an explicit error that does not echo the value.
	for _, body := range []string{
		`{{ .env.secret }}`,
		`{{ .vars.secret }}`,
		`{{ index .env "secret" }}`,
		`{{ env "secret" }}`,
		`{{ toJson .env }}`,
		`{{ .env.secret | b64enc }}`,
		`{{ . }}`,
		`{{ metrics_sql "SELECT total_revenue FROM mv WHERE country = '{{ .env.secret }}'" }}`,
		`{{ metrics_sql "SELECT {{ .env.secret }} FROM mv" }}`,
	} {
		t.Run(body, func(t *testing.T) {
			_, err := srv.ResolveTemplatedString(viewerCtx, &runtimev1.ResolveTemplatedStringRequest{
				InstanceId: instanceID,
				Body:       body,
			})
			require.ErrorContains(t, err, `project variables (".env") and the full template data ("." or "$") are only available to templates in connectors and models`)
			require.NotContains(t, err.Error(), secret)
		})
	}

	// Templates that do not reference variables still resolve for a viewer.
	res, err := srv.ResolveTemplatedString(viewerCtx, &runtimev1.ResolveTemplatedStringRequest{
		InstanceId: instanceID,
		Body:       `Revenue: {{ metrics_sql "SELECT total_revenue FROM mv" }}`,
	})
	require.NoError(t, err)
	require.Equal(t, "Revenue: 300", res.Body)

	// Not even a caller that could read the variables through GetInstance: the template is not declared in the project.
	_, err = srv.ResolveTemplatedString(testCtx(), &runtimev1.ResolveTemplatedStringRequest{
		InstanceId: instanceID,
		Body:       `{{ .env.secret }}`,
	})
	require.ErrorContains(t, err, `project variables (".env") and the full template data ("." or "$") are only available to templates in connectors and models`)
}
