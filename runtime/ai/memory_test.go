package ai_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/ai"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func TestUpdateMemoryTool(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{})
	s := newSession(t, rt, instanceID)
	require.True(t, s.MemoryEnabled())
	require.Empty(t, s.UserMemories())

	// Add
	var res *ai.MemoryUpdateResult
	_, err := s.CallTool(t.Context(), ai.RoleAssistant, ai.UpdateMemoryName, &res, &ai.UpdateMemoryArgs{
		Action:  ai.MemoryOpAdd,
		Content: "Prefers tables over charts.",
	})
	require.NoError(t, err)
	require.Equal(t, ai.MemoryOpAdd, res.Op)
	require.NotEmpty(t, res.MemoryID)
	memoryID := res.MemoryID

	// The session sees the new memory immediately
	require.Len(t, s.UserMemories(), 1)
	require.Equal(t, "Prefers tables over charts.", s.UserMemories()[0].Content)
	require.Equal(t, s.ID(), s.UserMemories()[0].SourceSessionID)
	// The tool was called outside any agent, so the memory applies to every agent
	require.Empty(t, s.UserMemories()[0].Agent)

	// Duplicate add is a noop
	_, err = s.CallTool(t.Context(), ai.RoleAssistant, ai.UpdateMemoryName, &res, &ai.UpdateMemoryArgs{
		Action:  ai.MemoryOpAdd,
		Content: "prefers tables over charts",
	})
	require.NoError(t, err)
	require.Equal(t, ai.MemoryOpNoop, res.Op)
	require.Equal(t, memoryID, res.MemoryID)
	require.Len(t, s.UserMemories(), 1)

	// Update by the short ID shown in the prompt; the result carries the full ID
	_, err = s.CallTool(t.Context(), ai.RoleAssistant, ai.UpdateMemoryName, &res, &ai.UpdateMemoryArgs{
		Action:   ai.MemoryOpUpdate,
		MemoryID: memoryID[:8],
		Content:  "Prefers tables over charts, except for time series.",
	})
	require.NoError(t, err)
	require.Equal(t, ai.MemoryOpUpdate, res.Op)
	require.Equal(t, memoryID, res.MemoryID)
	require.Equal(t, "Prefers tables over charts, except for time series.", s.UserMemories()[0].Content)

	// Unknown ID is a noop
	_, err = s.CallTool(t.Context(), ai.RoleAssistant, ai.UpdateMemoryName, &res, &ai.UpdateMemoryArgs{
		Action:   ai.MemoryOpDelete,
		MemoryID: "does-not-exist",
	})
	require.NoError(t, err)
	require.Equal(t, ai.MemoryOpNoop, res.Op)
	require.Equal(t, "memory not found", res.Reason)

	// Delete is permanent; the result echoes what was removed
	_, err = s.CallTool(t.Context(), ai.RoleAssistant, ai.UpdateMemoryName, &res, &ai.UpdateMemoryArgs{
		Action:   ai.MemoryOpDelete,
		MemoryID: memoryID,
	})
	require.NoError(t, err)
	require.Equal(t, ai.MemoryOpDelete, res.Op)
	require.Equal(t, "Prefers tables over charts, except for time series.", res.Content)
	require.Empty(t, s.UserMemories())

	catalog, release, err := rt.Catalog(t.Context(), instanceID)
	require.NoError(t, err)
	defer release()
	_, err = catalog.FindAIMemory(t.Context(), memoryID)
	require.ErrorIs(t, err, drivers.ErrNotFound)

	// Adding the same content again creates a new memory
	_, err = s.CallTool(t.Context(), ai.RoleAssistant, ai.UpdateMemoryName, &res, &ai.UpdateMemoryArgs{
		Action:  ai.MemoryOpAdd,
		Content: "Prefers tables over charts, except for time series.",
	})
	require.NoError(t, err)
	require.Equal(t, ai.MemoryOpAdd, res.Op)
	require.NotEqual(t, memoryID, res.MemoryID)
	require.Len(t, s.UserMemories(), 1)

	// Content is sanitized on write: whitespace is collapsed and angle brackets are stripped
	_, err = s.CallTool(t.Context(), ai.RoleAssistant, ai.UpdateMemoryName, &res, &ai.UpdateMemoryArgs{
		Action:  ai.MemoryOpAdd,
		Content: "  Reports  in\n<UTC>  ",
	})
	require.NoError(t, err)
	require.Equal(t, ai.MemoryOpAdd, res.Op)
	require.Equal(t, "Reports in UTC", res.Content)
	require.Len(t, s.UserMemories(), 2)

	// Secrets are refused
	_, err = s.CallTool(t.Context(), ai.RoleAssistant, ai.UpdateMemoryName, &res, &ai.UpdateMemoryArgs{
		Action:  ai.MemoryOpAdd,
		Content: "Their API key is sk-123",
	})
	require.NoError(t, err)
	require.Equal(t, ai.MemoryOpNoop, res.Op)
	require.Contains(t, res.Reason, "secret")
	require.Len(t, s.UserMemories(), 2)
}

// TestUpdateMemoryAgentScope checks that a memory learned during an agent's call is scoped to that agent.
func TestUpdateMemoryAgentScope(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{})
	s := newSession(t, rt, instanceID)

	call := s.AddMessage(&ai.AddMessageOptions{
		Role:        ai.RoleAssistant,
		Type:        ai.MessageTypeCall,
		Tool:        ai.AnalystAgentName,
		ContentType: ai.MessageContentTypeJSON,
		Content:     "{}",
	})
	var res *ai.MemoryUpdateResult
	_, err := s.WithParent(call.ID).CallTool(t.Context(), ai.RoleAssistant, ai.UpdateMemoryName, &res, &ai.UpdateMemoryArgs{
		Action:  ai.MemoryOpAdd,
		Content: "Prefers tables over charts.",
	})
	require.NoError(t, err)
	require.Equal(t, ai.MemoryOpAdd, res.Op)
	require.Len(t, s.UserMemories(), 1)
	require.Equal(t, ai.AnalystAgentName, s.UserMemories()[0].Agent)
}

func TestUpdateMemoryLimit(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{})
	s := newSession(t, rt, instanceID)

	var res *ai.MemoryUpdateResult
	for i := range ai.MaxMemories {
		_, err := s.CallTool(t.Context(), ai.RoleAssistant, ai.UpdateMemoryName, &res, &ai.UpdateMemoryArgs{
			Action:  ai.MemoryOpAdd,
			Content: strings.Repeat("x", i+1),
		})
		require.NoError(t, err)
		require.Equal(t, ai.MemoryOpAdd, res.Op)
	}
	require.Len(t, s.UserMemories(), ai.MaxMemories)

	_, err := s.CallTool(t.Context(), ai.RoleAssistant, ai.UpdateMemoryName, &res, &ai.UpdateMemoryArgs{
		Action:  ai.MemoryOpAdd,
		Content: "one too many",
	})
	require.NoError(t, err)
	require.Equal(t, ai.MemoryOpNoop, res.Op)
	require.Contains(t, res.Reason, "limit")
}

func TestMemoryTurnedOff(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{})
	claims := &runtime.SecurityClaims{UserID: uuid.NewString(), SkipChecks: true}

	catalog, release, err := rt.Catalog(t.Context(), instanceID)
	require.NoError(t, err)
	defer release()
	require.NoError(t, catalog.InsertAIMemory(t.Context(), &drivers.AIMemory{
		ID: uuid.NewString(), InstanceID: instanceID, OwnerID: claims.UserID, Content: "Prefers tables",
	}))

	r := ai.NewRunner(rt, activity.NewNoopClient())
	s, err := r.Session(t.Context(), &ai.SessionOptions{InstanceID: instanceID, Claims: claims, UserAgent: "rill/test"})
	require.NoError(t, err)
	require.Len(t, s.UserMemories(), 1)

	// Turning memory off disables it for new sessions: nothing is injected and nothing can be written.
	require.NoError(t, catalog.UpsertAIMemorySettings(t.Context(), &drivers.AIMemorySettings{OwnerID: claims.UserID, Enabled: false}))
	s, err = r.Session(t.Context(), &ai.SessionOptions{InstanceID: instanceID, Claims: claims, UserAgent: "rill/test"})
	require.NoError(t, err)
	require.False(t, s.MemoryEnabled())
	require.Empty(t, s.UserMemories())

	var res *ai.MemoryUpdateResult
	_, err = s.CallTool(t.Context(), ai.RoleAssistant, ai.UpdateMemoryName, &res, &ai.UpdateMemoryArgs{
		Action:  ai.MemoryOpAdd,
		Content: "Prefers charts",
	})
	require.ErrorContains(t, err, "access denied")

	// The memory itself is kept
	memories, err := catalog.FindAIMemories(t.Context(), claims.UserID)
	require.NoError(t, err)
	require.Len(t, memories, 1)
}

func TestMemoryGates(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{})
	r := ai.NewRunner(rt, activity.NewNoopClient())

	cases := []struct {
		name      string
		userAgent string
		claims    *runtime.SecurityClaims
		enabled   bool
	}{
		{"rill chat", "rill/0.1", &runtime.SecurityClaims{UserID: uuid.NewString(), Permissions: []runtime.Permission{runtime.UseAI}}, true},
		{"local developer", "rill/0.1", &runtime.SecurityClaims{SkipChecks: true}, true},
		{"anonymous cloud", "rill/0.1", &runtime.SecurityClaims{Permissions: []runtime.Permission{runtime.UseAI}}, false},
		{"no ai permission", "rill/0.1", &runtime.SecurityClaims{UserID: uuid.NewString()}, false},
		{"report", "rill/report", &runtime.SecurityClaims{UserID: uuid.NewString(), SkipChecks: true}, false},
		{"mcp client", "claude-desktop/1.0", &runtime.SecurityClaims{UserID: uuid.NewString(), SkipChecks: true}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s, err := r.Session(t.Context(), &ai.SessionOptions{InstanceID: instanceID, Claims: c.claims, UserAgent: c.userAgent})
			require.NoError(t, err)
			require.Equal(t, c.enabled, s.MemoryEnabled())
		})
	}
}

func TestMemoryFeatureFlagOff(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "features:\n  chat_memory: false\n",
		},
	})
	s := newSession(t, rt, instanceID)
	require.False(t, s.MemoryEnabled())

	var res *ai.MemoryUpdateResult
	_, err := s.CallTool(t.Context(), ai.RoleAssistant, ai.UpdateMemoryName, &res, &ai.UpdateMemoryArgs{
		Action:  ai.MemoryOpAdd,
		Content: "Prefers tables",
	})
	require.ErrorContains(t, err, "access denied")
}
