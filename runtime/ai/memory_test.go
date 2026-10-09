package ai_test

import (
	"fmt"
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
		Action:   ai.MemoryOpAdd,
		Category: ai.MemoryCategoryPreference,
		Content:  "Prefers tables over charts.",
	})
	require.NoError(t, err)
	require.Len(t, res.Ops, 1)
	require.Equal(t, ai.MemoryOpAdd, res.Ops[0].Op)
	require.NotEmpty(t, res.Ops[0].MemoryID)
	memoryID := res.Ops[0].MemoryID

	// The session sees the new memory immediately
	require.Len(t, s.UserMemories(), 1)
	require.Equal(t, "Prefers tables over charts.", s.UserMemories()[0].Content)
	require.Equal(t, drivers.AIMemorySourceExplicit, s.UserMemories()[0].Source)
	require.Equal(t, s.ID(), s.UserMemories()[0].SourceSessionID)
	require.NotEmpty(t, s.UserMemories()[0].SourceMessageID)

	// Duplicate add is a noop
	_, err = s.CallTool(t.Context(), ai.RoleAssistant, ai.UpdateMemoryName, &res, &ai.UpdateMemoryArgs{
		Action:   ai.MemoryOpAdd,
		Category: ai.MemoryCategoryPreference,
		Content:  "prefers tables over charts",
	})
	require.NoError(t, err)
	require.Equal(t, ai.MemoryOpNoop, res.Ops[0].Op)
	require.Equal(t, memoryID, res.Ops[0].MemoryID)
	require.Len(t, s.UserMemories(), 1)

	// Update
	_, err = s.CallTool(t.Context(), ai.RoleAssistant, ai.UpdateMemoryName, &res, &ai.UpdateMemoryArgs{
		Action:   ai.MemoryOpUpdate,
		MemoryID: memoryID,
		Category: ai.MemoryCategoryPreference,
		Content:  "Prefers tables over charts, except for time series.",
	})
	require.NoError(t, err)
	require.Equal(t, ai.MemoryOpUpdate, res.Ops[0].Op)
	require.Equal(t, "Prefers tables over charts, except for time series.", s.UserMemories()[0].Content)

	// Unknown ID is a noop
	_, err = s.CallTool(t.Context(), ai.RoleAssistant, ai.UpdateMemoryName, &res, &ai.UpdateMemoryArgs{
		Action:   ai.MemoryOpDelete,
		MemoryID: "does-not-exist",
	})
	require.NoError(t, err)
	require.Equal(t, ai.MemoryOpNoop, res.Ops[0].Op)
	require.Equal(t, "memory not found", res.Ops[0].Reason)

	// Delete is permanent; the op echoes what was removed
	_, err = s.CallTool(t.Context(), ai.RoleAssistant, ai.UpdateMemoryName, &res, &ai.UpdateMemoryArgs{
		Action:   ai.MemoryOpDelete,
		MemoryID: memoryID,
	})
	require.NoError(t, err)
	require.Equal(t, ai.MemoryOpDelete, res.Ops[0].Op)
	require.Equal(t, "Prefers tables over charts, except for time series.", res.Ops[0].Content)
	require.Empty(t, s.UserMemories())

	catalog, release, err := rt.Catalog(t.Context(), instanceID)
	require.NoError(t, err)
	defer release()
	_, err = catalog.FindAIMemory(t.Context(), memoryID)
	require.ErrorIs(t, err, drivers.ErrNotFound)

	// Adding the same content again creates a new memory.
	// The category is omitted here to check that it gets a default.
	_, err = s.CallTool(t.Context(), ai.RoleAssistant, ai.UpdateMemoryName, &res, &ai.UpdateMemoryArgs{
		Action:  ai.MemoryOpAdd,
		Content: "Prefers tables over charts, except for time series.",
	})
	require.NoError(t, err)
	require.Equal(t, ai.MemoryOpAdd, res.Ops[0].Op)
	require.NotEqual(t, memoryID, res.Ops[0].MemoryID)
	require.Equal(t, ai.MemoryCategoryContext, res.Ops[0].Category)
	require.Len(t, s.UserMemories(), 1)
	require.Equal(t, ai.MemoryCategoryContext, s.UserMemories()[0].Category)

	// Secrets are refused
	_, err = s.CallTool(t.Context(), ai.RoleAssistant, ai.UpdateMemoryName, &res, &ai.UpdateMemoryArgs{
		Action:   ai.MemoryOpAdd,
		Category: ai.MemoryCategoryContext,
		Content:  "Their API key is sk-123",
	})
	require.NoError(t, err)
	require.Equal(t, ai.MemoryOpNoop, res.Ops[0].Op)
	require.Contains(t, res.Ops[0].Reason, "secret")
	require.Len(t, s.UserMemories(), 1)

	// Unknown categories fall back to context
	_, err = s.CallTool(t.Context(), ai.RoleAssistant, ai.UpdateMemoryName, &res, &ai.UpdateMemoryArgs{
		Action:   ai.MemoryOpAdd,
		Category: "whatever",
		Content:  "Leads the growth team",
	})
	require.NoError(t, err)
	require.Equal(t, ai.MemoryOpAdd, res.Ops[0].Op)
	require.Equal(t, ai.MemoryCategoryContext, res.Ops[0].Category)
}

func TestUpdateMemoryLimit(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{})
	s := newSession(t, rt, instanceID)

	var res *ai.MemoryUpdateResult
	for i := range ai.MaxMemories {
		_, err := s.CallTool(t.Context(), ai.RoleAssistant, ai.UpdateMemoryName, &res, &ai.UpdateMemoryArgs{
			Action:   ai.MemoryOpAdd,
			Category: ai.MemoryCategoryContext,
			Content:  strings.Repeat("x", i+1),
		})
		require.NoError(t, err)
		require.Equal(t, ai.MemoryOpAdd, res.Ops[0].Op)
	}
	require.Len(t, s.UserMemories(), ai.MaxMemories)

	_, err := s.CallTool(t.Context(), ai.RoleAssistant, ai.UpdateMemoryName, &res, &ai.UpdateMemoryArgs{
		Action:   ai.MemoryOpAdd,
		Category: ai.MemoryCategoryContext,
		Content:  "one too many",
	})
	require.NoError(t, err)
	require.Equal(t, ai.MemoryOpNoop, res.Ops[0].Op)
	require.Contains(t, res.Ops[0].Reason, "limit")
}

func TestMemoryTurnedOff(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{})
	claims := &runtime.SecurityClaims{UserID: uuid.NewString(), SkipChecks: true}

	catalog, release, err := rt.Catalog(t.Context(), instanceID)
	require.NoError(t, err)
	defer release()
	require.NoError(t, catalog.InsertAIMemory(t.Context(), &drivers.AIMemory{
		ID: uuid.NewString(), InstanceID: instanceID, OwnerID: claims.UserID, Category: ai.MemoryCategoryPreference,
		Content: "Prefers tables", Source: drivers.AIMemorySourceManual,
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
		Action:   ai.MemoryOpAdd,
		Category: ai.MemoryCategoryPreference,
		Content:  "Prefers charts",
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
		Action:   ai.MemoryOpAdd,
		Category: ai.MemoryCategoryPreference,
		Content:  "Prefers tables",
	})
	require.ErrorContains(t, err, "access denied")
}

// TestSessionRepeatedFlush covers the flush path that background memory extraction relies on:
// it appends messages to a session that has already been flushed and flushes it again.
func TestSessionRepeatedFlush(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{})
	s := newSession(t, rt, instanceID)

	var res *ai.MemoryUpdateResult
	_, err := s.CallTool(t.Context(), ai.RoleAssistant, ai.UpdateMemoryName, &res, &ai.UpdateMemoryArgs{
		Action:   ai.MemoryOpAdd,
		Category: ai.MemoryCategoryPreference,
		Content:  "Prefers tables over charts.",
	})
	require.NoError(t, err)
	require.NoError(t, s.Flush(t.Context()))

	// A second turn on the same session must not re-insert the messages the first flush already persisted.
	_, err = s.CallTool(t.Context(), ai.RoleAssistant, ai.UpdateMemoryName, &res, &ai.UpdateMemoryArgs{
		Action:   ai.MemoryOpAdd,
		Category: ai.MemoryCategoryPreference,
		Content:  "Reports in UTC.",
	})
	require.NoError(t, err)
	require.NoError(t, s.Flush(t.Context()))

	catalog, release, err := rt.Catalog(t.Context(), instanceID)
	require.NoError(t, err)
	defer release()
	msgs, err := catalog.FindAIMessages(t.Context(), s.ID())
	require.NoError(t, err)
	require.Len(t, msgs, len(s.Messages()))

	// Indexes stay unique and ordered across both flushes.
	seen := make(map[int]bool, len(msgs))
	for i, m := range msgs {
		require.False(t, seen[m.Index], "duplicate message index %d", m.Index)
		seen[m.Index] = true
		if i > 0 {
			require.Greater(t, m.Index, msgs[i-1].Index)
		}
	}
}

// TestMemoryOpsPerTurnLimit checks that ops beyond the per-turn budget are reported as rejected, not dropped.
func TestMemoryOpsPerTurnLimit(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{})
	s := newSession(t, rt, instanceID)

	ops := make([]ai.MemoryOp, 0, ai.MaxMemoryOpsPerTurn+2)
	for i := range ai.MaxMemoryOpsPerTurn + 2 {
		ops = append(ops, ai.MemoryOp{
			Op:       ai.MemoryOpAdd,
			Category: ai.MemoryCategoryContext,
			Content:  fmt.Sprintf("Owns dashboard number %d", i),
		})
	}

	applied, err := ai.ApplyMemoryOpsForTest(t.Context(), s, ops)
	require.NoError(t, err)
	require.Len(t, applied, len(ops))
	for i, op := range applied {
		if i < ai.MaxMemoryOpsPerTurn {
			require.Equal(t, ai.MemoryOpAdd, op.Op)
		} else {
			require.Equal(t, ai.MemoryOpNoop, op.Op)
			require.Contains(t, op.Reason, "limit")
		}
	}
	require.Len(t, s.UserMemories(), ai.MaxMemoryOpsPerTurn)
}
