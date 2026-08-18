package ai_test

import (
	"testing"

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
