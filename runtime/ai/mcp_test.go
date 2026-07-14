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
