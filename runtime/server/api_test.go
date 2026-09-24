package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/httputil"
	"github.com/rilldata/rill/runtime/pkg/ratelimit"
	"github.com/rilldata/rill/runtime/server/auth"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestAPIHandlerSecurity(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml":    "",
			"models/m.sql": "SELECT 1 AS a, 'x' AS d",
			"metrics/restricted_mv.yaml": `
type: metrics_view
model: m
dimensions:
  - column: d
measures:
  - name: count
    expression: count(*)
security:
  access: '{{ .user.admin }}'
`,
			"apis/enterprise_only.yaml": `
type: api
sql: SELECT 'sensitive' AS value
security:
  access: '{{ eq .user.tier "enterprise" }}'
`,
			"apis/enterprise_skip.yaml": `
type: api
metrics_sql: SELECT count FROM restricted_mv
security:
  access: '{{ eq .user.tier "enterprise" }}'
skip_nested_security: true
`,
			"apis/open.yaml": `
type: api
sql: SELECT 'open' AS value
`,
			"apis/open_mv.yaml": `
type: api
metrics_sql: SELECT count FROM restricted_mv
`,
			"apis/proxy.yaml": `
type: api
api: enterprise_only
`,
			"apis/proxy_skip.yaml": `
type: api
api: enterprise_only
security:
  access: true
skip_nested_security: true
`,
			"apis/admin_only.yaml": `
type: api
sql: SELECT 'admin' AS value
security:
  access: '{{ .user.admin }}'
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 11, 0, 0)

	srv, err := NewServer(context.Background(), &Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	call := func(name string, claims *runtime.SecurityClaims) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(auth.WithClaims(context.Background(), claims))
		req.SetPathValue("instance_id", instanceID)
		req.SetPathValue("name", name)
		httputil.Handler(srv.apiHandler).ServeHTTP(rec, req)
		return rec
	}

	free := &runtime.SecurityClaims{
		UserAttributes: map[string]any{"tier": "free", "admin": false},
		Permissions:    []runtime.Permission{runtime.ReadAPI},
	}
	enterprise := &runtime.SecurityClaims{
		UserAttributes: map[string]any{"tier": "enterprise", "admin": false},
		Permissions:    []runtime.Permission{runtime.ReadAPI},
	}

	t.Run("api access rule denies", func(t *testing.T) {
		rec := call("enterprise_only", free)
		require.Equal(t, http.StatusForbidden, rec.Code)
		require.NotContains(t, rec.Body.String(), "sensitive")
	})

	t.Run("api access rule allows", func(t *testing.T) {
		rec := call("enterprise_only", enterprise)
		require.Equal(t, http.StatusOK, rec.Code)
		require.JSONEq(t, `[{"value":"sensitive"}]`, rec.Body.String())
	})

	t.Run("denied again after allowed call on same runtime", func(t *testing.T) {
		rec := call("enterprise_only", free)
		require.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("no security block allows", func(t *testing.T) {
		rec := call("open", free)
		require.Equal(t, http.StatusOK, rec.Code)
		require.JSONEq(t, `[{"value":"open"}]`, rec.Body.String())
	})

	t.Run("skip_nested_security does not skip the api's own rule", func(t *testing.T) {
		rec := call("enterprise_skip", free)
		require.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("skip_nested_security skips nested metrics view rule for authorized caller", func(t *testing.T) {
		rec := call("enterprise_skip", enterprise)
		require.Equal(t, http.StatusOK, rec.Code)
		require.JSONEq(t, `[{"count":1}]`, rec.Body.String())
	})

	t.Run("nested metrics view denial returns 403", func(t *testing.T) {
		rec := call("open_mv", free)
		require.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("proxied api enforces target rule", func(t *testing.T) {
		rec := call("proxy", free)
		require.Equal(t, http.StatusForbidden, rec.Code)
		require.NotContains(t, rec.Body.String(), "sensitive")
	})

	t.Run("proxied api with skip_nested_security bypasses target rule", func(t *testing.T) {
		rec := call("proxy_skip", free)
		require.Equal(t, http.StatusOK, rec.Code)
		require.JSONEq(t, `[{"value":"sensitive"}]`, rec.Body.String())
	})

	t.Run("exclusive additional rules deny unlisted api", func(t *testing.T) {
		claims := &runtime.SecurityClaims{
			Permissions: []runtime.Permission{runtime.ReadAPI},
			AdditionalRules: []*runtimev1.SecurityRule{{Rule: &runtimev1.SecurityRule_Access{Access: &runtimev1.SecurityRuleAccess{
				Allow:              true,
				Exclusive:          true,
				ConditionResources: []*runtimev1.ResourceName{{Kind: runtime.ResourceKindExplore, Name: "e"}},
			}}}},
		}
		rec := call("open", claims)
		require.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("exclusive additional rules allow listed api", func(t *testing.T) {
		claims := &runtime.SecurityClaims{
			Permissions: []runtime.Permission{runtime.ReadAPI},
			AdditionalRules: []*runtimev1.SecurityRule{{Rule: &runtimev1.SecurityRule_Access{Access: &runtimev1.SecurityRuleAccess{
				Allow:              true,
				Exclusive:          true,
				ConditionResources: []*runtimev1.ResourceName{{Kind: runtime.ResourceKindAPI, Name: "open"}},
			}}}},
		}
		rec := call("open", claims)
		require.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("policy evaluation error fails closed", func(t *testing.T) {
		claims := &runtime.SecurityClaims{
			UserAttributes: map[string]any{"admin": "notabool"},
			Permissions:    []runtime.Permission{runtime.ReadAPI},
		}
		rec := call("admin_only", claims)
		require.NotEqual(t, http.StatusOK, rec.Code)
		require.NotContains(t, rec.Body.String(), `"value"`)
	})

	t.Run("skip checks allows everything", func(t *testing.T) {
		claims := &runtime.SecurityClaims{SkipChecks: true}
		for _, name := range []string{"enterprise_only", "enterprise_skip", "open", "open_mv", "proxy", "proxy_skip", "admin_only"} {
			rec := call(name, claims)
			require.Equal(t, http.StatusOK, rec.Code, name)
		}
	})

	t.Run("missing ReadAPI permission denies", func(t *testing.T) {
		rec := call("open", &runtime.SecurityClaims{UserAttributes: map[string]any{"tier": "enterprise"}})
		require.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("openapi spec lists all apis regardless of claims", func(t *testing.T) {
		for _, claims := range []*runtime.SecurityClaims{free, enterprise} {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(auth.WithClaims(context.Background(), claims))
			req.SetPathValue("instance_id", instanceID)
			httputil.Handler(srv.combinedOpenAPISpec).ServeHTTP(rec, req)
			require.Equal(t, http.StatusOK, rec.Code)
			require.Contains(t, rec.Body.String(), "/enterprise_only")
			require.Contains(t, rec.Body.String(), "/open")
		}
	})
}
