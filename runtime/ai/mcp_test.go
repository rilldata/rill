package ai_test

import (
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/google/uuid"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/ai"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

// TestMCPToolSpecs asserts that tool specs can be built without a runtime,
// which is the invariant ai.MCPToolSpecs relies on to serve the admin service's unified MCP server.
func TestMCPToolSpecs(t *testing.T) {
	specs := ai.MCPToolSpecs()
	require.NotEmpty(t, specs)
	for name, spec := range specs {
		require.Equal(t, name, spec.Name)
		require.NotEmpty(t, spec.Description, "tool %q", name)
		require.NotNil(t, spec.InputSchema, "tool %q", name)
	}
	require.Contains(t, specs, ai.QueryMetricsViewName)

	// The specs are copies: extending one does not leak into the next call.
	schema := specs[ai.QueryMetricsViewName].InputSchema.(*jsonschema.Schema)
	schema.Properties["project"] = &jsonschema.Schema{Type: "string"}
	schema.Required = append(schema.Required, "project")
	specs[ai.QueryMetricsViewName].Meta["ui"] = map[string]any{"resourceUri": "ui://x"}
	again := ai.MCPToolSpecs()[ai.QueryMetricsViewName]
	require.NotContains(t, again.InputSchema.(*jsonschema.Schema).Properties, "project")
	require.NotContains(t, again.InputSchema.(*jsonschema.Schema).Required, "project")
	require.NotContains(t, again.Meta, "ui")
	require.Contains(t, again.Meta, "openai/toolInvocation/invoking", "the tool's own meta is kept")

	// Tools may advertise a different definition on MCP servers than in-app (see ai.MCPSpecProvider):
	// create_chart takes the Flint shape externally and keeps Rill's own chart spec for the in-app chat.
	external := specs[ai.CreateChartName].InputSchema.(*jsonschema.Schema)
	require.Contains(t, external.Properties, "chart_spec")
	require.Contains(t, external.Properties, "data")
	require.Contains(t, external.Properties, "chart_type", "the previous shape stays valid for clients with a cached tool list")
	require.Empty(t, external.Required)
	require.Contains(t, external.Defs, "Expression", "the data block reuses the filter expression schema")
	require.Contains(t, specs[ai.CreateChartName].Description, "Waterfall Chart")
	require.NotNil(t, specs[ai.CreateChartName].OutputSchema, "the MCP spec inherits the inferred output schema")
	require.Contains(t, specs[ai.ChartDataName].InputSchema.(*jsonschema.Schema).Properties, "chart_spec")
	inApp, err := ai.NewRunner(nil, nil).Tools[ai.CreateChartName].AsProto()
	require.NoError(t, err)
	require.Contains(t, inApp.InputSchema, `"chart_type"`)
	require.NotContains(t, inApp.InputSchema, `"chart_spec"`)
	require.NotContains(t, inApp.Description, "Waterfall Chart")

	// The tools that take part in the MCP Apps widget are known without a runtime too.
	require.False(t, ai.IsMCPAppOnlyTool(ai.CreateChartName), "create_chart is called by the model")
	require.True(t, ai.IsMCPAppOnlyTool(ai.ChartDataName))
	require.False(t, ai.IsMCPAppOnlyTool(ai.QueryMetricsViewName))
	require.False(t, ai.IsMCPAppOnlyTool("unknown"))
	require.Nil(t, ai.MCPAppToolMeta(ai.QueryMetricsViewName, nil))
	require.Nil(t, ai.MCPAppToolMeta(ai.CreateChartName, nil), "create_chart links to the widget, which requires the asset")
	require.Equal(t, []any{"app"}, ai.MCPAppToolMeta(ai.ChartDataName, nil)["ui"].(map[string]any)["visibility"], "app-only visibility does not need the asset")
}

// TestSessionCreateIfNotExists asserts that a caller can address a session by an ID of its own choosing,
// which the admin service's unified MCP server relies on to keep one session per project.
func TestSessionCreateIfNotExists(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{})
	claims := &runtime.SecurityClaims{UserID: uuid.NewString(), Permissions: []runtime.Permission{runtime.UseAI}}
	r := ai.NewRunner(rt, activity.NewNoopClient())
	sessionID := uuid.NewString()

	// The session does not exist yet, so it is created with the given ID.
	s, err := r.Session(t.Context(), &ai.SessionOptions{
		InstanceID:        instanceID,
		SessionID:         sessionID,
		CreateIfNotExists: true,
		Claims:            claims,
		UserAgent:         "mcp-client",
	})
	require.NoError(t, err)
	require.Equal(t, sessionID, s.ID())
	require.NoError(t, s.Flush(t.Context()))

	// A second call loads the same session instead of creating another one.
	s2, err := r.Session(t.Context(), &ai.SessionOptions{
		InstanceID:        instanceID,
		SessionID:         sessionID,
		CreateIfNotExists: true,
		Claims:            claims,
	})
	require.NoError(t, err)
	require.Equal(t, sessionID, s2.ID())

	// Without CreateIfNotExists, an unknown session ID is still an error.
	_, err = r.Session(t.Context(), &ai.SessionOptions{
		InstanceID: instanceID,
		SessionID:  uuid.NewString(),
		Claims:     claims,
	})
	require.Error(t, err)
}

// TestDeveloperToolsMCPAccess verifies that the leaf developer tools are exposed to external
// MCP clients (non-rill user agents) for callers with EditRepo, while the developer agents remain
// restricted to first-party Rill clients.
func TestDeveloperToolsMCPAccess(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{})

	newMCPSession := func(t *testing.T, permissions ...runtime.Permission) *ai.Session {
		claims := &runtime.SecurityClaims{
			UserID:      uuid.NewString(),
			SkipChecks:  false,
			Permissions: permissions,
		}
		r := ai.NewRunner(rt, activity.NewNoopClient())
		s, err := r.Session(t.Context(), &ai.SessionOptions{
			InstanceID: instanceID,
			Claims:     claims,
			UserAgent:  "mcp-client", // Non-rill user agent, i.e. an external MCP client
		})
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, s.Flush(t.Context()))
		})
		return s
	}

	assertAccess := func(t *testing.T, s *ai.Session, name string, want bool) {
		t.Helper()
		tool, ok := s.Tool(name)
		require.True(t, ok, "tool %q should be registered", name)
		allowed, err := tool.CheckAccess(ai.WithSession(t.Context(), s))
		require.NoError(t, err)
		require.Equal(t, want, allowed, "tool %q access", name)
	}

	leafTools := []string{ai.WriteFileName, ai.ReadFileName, ai.ListFilesName, ai.SearchFilesName}
	agentTools := []string{ai.DevelopFileName, ai.DeveloperAgentName}

	t.Run("with EditRepo", func(t *testing.T) {
		s := newMCPSession(t, runtime.UseAI, runtime.EditRepo)
		for _, name := range leafTools {
			assertAccess(t, s, name, true)
		}
		for _, name := range agentTools {
			assertAccess(t, s, name, false)
		}
	})

	t.Run("without EditRepo", func(t *testing.T) {
		s := newMCPSession(t, runtime.UseAI)
		for _, name := range leafTools {
			assertAccess(t, s, name, false)
		}
	})
}
