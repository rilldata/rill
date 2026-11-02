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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
// which commonly hold secrets such as connector credentials, unless the caller may read sensitive instance info.
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
			require.ErrorContains(t, err, "project variables are only available to templates declared in the project files")
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

	// A caller that can read sensitive instance info can already read the variables through GetInstance, so templates resolve them.
	res, err = srv.ResolveTemplatedString(testCtx(), &runtimev1.ResolveTemplatedStringRequest{
		InstanceId: instanceID,
		Body:       `{{ .env.secret }}`,
	})
	require.NoError(t, err)
	require.Equal(t, secret, res.Body)
}

// TestResolveTemplatedString_Component verifies that a component reference resolves the saved content of a markdown component,
// which is declared in the project and may therefore use project variables, while enforcing the caller's access to the component.
func TestResolveTemplatedString_Component(t *testing.T) {
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
			"public.yaml": `
type: canvas
rows:
- items:
  - markdown:
      content: 'Region: {{ .env.region }}, total: {{ metrics_sql "SELECT total_revenue FROM mv" }}'
  - kpi:
      metrics_view: mv
      measure: total_revenue
`,
			"admin_only.yaml": `
type: canvas
security:
  access: '{{ .user.admin }}'
rows:
- items:
  - markdown:
      content: 'Secret: {{ .env.secret }}'
`,
			// A personal canvas created through the admin API by the viewer, so its content was written by the viewer.
			"personal.yaml": `
type: canvas
annotations:
  admin_managed: "true"
  admin_owner_user_id: "viewer"
rows:
- items:
  - markdown:
      content: 'Secret: {{ .env.secret }}'
`,
			// An inline component hidden from non-admins by its item's condition, which another item of the canvas references by name.
			"hidden.yaml": `
type: canvas
rows:
- items:
  - markdown:
      content: 'Admins only: {{ .env.secret }}'
    if: '{{ .user.admin }}'
  - component: hidden--component-0-0
`,
			// A canvas that references another canvas's inline component, e.g. to share it through a link that grants only this canvas.
			"wrapper.yaml": `
type: canvas
rows:
- items:
  - component: public--component-0-0
`,
		},
		Variables: map[string]string{"region": "US", "secret": secret},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 13, 0, 0)

	srv, err := server.NewServer(t.Context(), &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	// A viewer has the permissions of an embed token: it can query data, but not read sensitive instance info.
	viewerCtx := auth.WithClaims(context.Background(), &runtime.SecurityClaims{
		UserAttributes: map[string]any{"admin": false, "id": "viewer"},
		Permissions:    []runtime.Permission{runtime.ReadAPI, runtime.ReadMetrics},
	})

	// The component's content is declared in the project, so it resolves the variable even for a viewer.
	res, err := srv.ResolveTemplatedString(viewerCtx, &runtimev1.ResolveTemplatedStringRequest{
		InstanceId: instanceID,
		Component:  "public--component-0-0",
	})
	require.NoError(t, err)
	require.Equal(t, "Region: US, total: 300", res.Body)

	// The same content sent as a body comes from the caller, so it cannot reference the variable.
	_, err = srv.ResolveTemplatedString(viewerCtx, &runtimev1.ResolveTemplatedStringRequest{
		InstanceId: instanceID,
		Body:       `Region: {{ .env.region }}, total: {{ metrics_sql "SELECT total_revenue FROM mv" }}`,
	})
	require.ErrorContains(t, err, "project variables are only available to templates declared in the project files")

	// The viewer can access their personal canvas, but its content was written by the viewer, not declared in the project,
	// so it cannot reference the variable; not even for a caller that could read the variables, since the content's author decides.
	for _, ctx := range []context.Context{viewerCtx, testCtx()} {
		_, err = srv.ResolveTemplatedString(ctx, &runtimev1.ResolveTemplatedStringRequest{
			InstanceId: instanceID,
			Component:  "personal--component-0-0",
		})
		require.ErrorContains(t, err, "project variables are only available to templates declared in the project files")
		require.NotContains(t, err.Error(), secret)
	}

	// A component hidden from the viewer by its item's condition is rejected, even though another item of an accessible canvas references it.
	_, err = srv.ResolveTemplatedString(viewerCtx, &runtimev1.ResolveTemplatedStringRequest{
		InstanceId: instanceID,
		Component:  "hidden--component-0-0",
	})
	require.Equal(t, codes.PermissionDenied, status.Code(err))
	require.NotContains(t, err.Error(), secret)
	res, err = srv.ResolveTemplatedString(auth.WithClaims(context.Background(), &runtime.SecurityClaims{
		UserAttributes: map[string]any{"admin": true},
		Permissions:    []runtime.Permission{runtime.ReadAPI, runtime.ReadMetrics},
	}), &runtimev1.ResolveTemplatedStringRequest{
		InstanceId: instanceID,
		Component:  "hidden--component-0-0",
	})
	require.NoError(t, err)
	require.Equal(t, "Admins only: "+secret, res.Body)

	// A shared link's token is granted only the wrapper canvas, yet it can resolve the component the wrapper references,
	// with the component's own origin: the content is declared in the project, so the variable resolves.
	tokenCtx := auth.WithClaims(context.Background(), &runtime.SecurityClaims{
		Permissions: []runtime.Permission{runtime.ReadAPI, runtime.ReadMetrics},
		AdditionalRules: []*runtimev1.SecurityRule{{Rule: &runtimev1.SecurityRule_TransitiveAccess{TransitiveAccess: &runtimev1.SecurityRuleTransitiveAccess{
			Resource: &runtimev1.ResourceName{Kind: runtime.ResourceKindCanvas, Name: "wrapper"},
		}}}},
	})
	res, err = srv.ResolveTemplatedString(tokenCtx, &runtimev1.ResolveTemplatedStringRequest{
		InstanceId: instanceID,
		Component:  "public--component-0-0",
	})
	require.NoError(t, err)
	require.Equal(t, "Region: US, total: 300", res.Body)
	// The token cannot resolve a component that no canvas it is granted shows.
	_, err = srv.ResolveTemplatedString(tokenCtx, &runtimev1.ResolveTemplatedStringRequest{
		InstanceId: instanceID,
		Component:  "hidden--component-0-0",
	})
	require.Equal(t, codes.PermissionDenied, status.Code(err))

	// A component the viewer cannot access is rejected without resolving its content.
	_, err = srv.ResolveTemplatedString(viewerCtx, &runtimev1.ResolveTemplatedStringRequest{
		InstanceId: instanceID,
		Component:  "admin_only--component-0-0",
	})
	require.Equal(t, codes.PermissionDenied, status.Code(err))
	require.NotContains(t, err.Error(), secret)

	// Only markdown components have content to resolve.
	_, err = srv.ResolveTemplatedString(viewerCtx, &runtimev1.ResolveTemplatedStringRequest{
		InstanceId: instanceID,
		Component:  "public--component-0-1",
	})
	require.Equal(t, codes.InvalidArgument, status.Code(err))

	_, err = srv.ResolveTemplatedString(viewerCtx, &runtimev1.ResolveTemplatedStringRequest{
		InstanceId: instanceID,
		Component:  "missing",
	})
	require.Equal(t, codes.NotFound, status.Code(err))
}
