package sqlite

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func openTestConnection(t *testing.T) *connection {
	tmpDir := t.TempDir()
	cfg := map[string]any{"dsn": filepath.Join(tmpDir, "test.db")}
	h, err := driver{}.Open(t.Context(), "", "", cfg, storage.MustNew(filepath.Join(tmpDir, "storage"), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	t.Cleanup(func() { _ = h.Close() })
	require.NoError(t, h.Migrate(t.Context()))
	return h.(*connection)
}

func TestAIMemoriesCRUD(t *testing.T) {
	ctx := t.Context()
	c := openTestConnection(t)
	catalog, ok := c.AsCatalogStore("inst")
	require.True(t, ok)
	now := time.Now().UTC().Truncate(time.Second)

	m1 := &drivers.AIMemory{
		InstanceID:      "inst",
		OwnerID:         "alice",
		Agent:           "analyst_agent",
		Content:         "Prefers tables over charts",
		SourceSessionID: "s1",
		CreatedOn:       now,
		UpdatedOn:       now,
	}
	require.NoError(t, catalog.InsertAIMemory(ctx, m1))
	m2 := &drivers.AIMemory{
		InstanceID: "inst", OwnerID: "alice", Agent: "developer_agent", Content: "Names models in snake_case",
		SourceSessionID: "s1", CreatedOn: now, UpdatedOn: now,
	}
	require.NoError(t, catalog.InsertAIMemory(ctx, m2))
	m3 := &drivers.AIMemory{InstanceID: "inst", OwnerID: "bob", Content: "Works in finance", CreatedOn: now, UpdatedOn: now}
	require.NoError(t, catalog.InsertAIMemory(ctx, m3))

	// IDs are assigned on insert and increase
	require.NotZero(t, m1.ID)
	require.Greater(t, m2.ID, m1.ID)
	require.Greater(t, m3.ID, m2.ID)

	// Find by owner
	res, err := catalog.FindAIMemories(ctx, "alice")
	require.NoError(t, err)
	require.Len(t, res, 2)
	require.Equal(t, m1.ID, res[0].ID)
	require.Equal(t, "analyst_agent", res[0].Agent)
	require.Equal(t, "Prefers tables over charts", res[0].Content)
	require.Equal(t, "s1", res[0].SourceSessionID)

	res, err = catalog.FindAIMemories(ctx, "bob")
	require.NoError(t, err)
	require.Len(t, res, 1)
	require.Equal(t, m3.ID, res[0].ID)
	require.Empty(t, res[0].Agent)
	require.Empty(t, res[0].SourceSessionID)

	// Find one
	got, err := catalog.FindAIMemory(ctx, m1.ID)
	require.NoError(t, err)
	require.Equal(t, "alice", got.OwnerID)
	_, err = catalog.FindAIMemory(ctx, m3.ID+1)
	require.ErrorIs(t, err, drivers.ErrNotFound)

	// Instance isolation
	other, ok := c.AsCatalogStore("other")
	require.True(t, ok)
	res, err = other.FindAIMemories(ctx, "alice")
	require.NoError(t, err)
	require.Empty(t, res)
	_, err = other.FindAIMemory(ctx, m1.ID)
	require.ErrorIs(t, err, drivers.ErrNotFound)

	// Update
	got.Content = "Prefers tables"
	got.SourceSessionID = "s2"
	got.Agent = ""
	require.NoError(t, catalog.UpdateAIMemory(ctx, got))
	got, err = catalog.FindAIMemory(ctx, m1.ID)
	require.NoError(t, err)
	require.Equal(t, "Prefers tables", got.Content)
	require.Equal(t, "s2", got.SourceSessionID)
	require.Empty(t, got.Agent)
	require.False(t, got.UpdatedOn.Before(now))

	// Delete leaves other memories alone
	require.NoError(t, catalog.DeleteAIMemory(ctx, m1.ID))
	_, err = catalog.FindAIMemory(ctx, m1.ID)
	require.ErrorIs(t, err, drivers.ErrNotFound)
	res, err = catalog.FindAIMemories(ctx, "alice")
	require.NoError(t, err)
	require.Len(t, res, 1)
	require.Equal(t, m2.ID, res[0].ID)

	// Deleting the newest memory does not free its ID for the next insert
	require.NoError(t, catalog.DeleteAIMemory(ctx, m3.ID))
	m4 := &drivers.AIMemory{InstanceID: "inst", OwnerID: "bob", Content: "Works in sales", CreatedOn: now, UpdatedOn: now}
	require.NoError(t, catalog.InsertAIMemory(ctx, m4))
	require.Greater(t, m4.ID, m3.ID)
}

func TestAIMemorySettings(t *testing.T) {
	ctx := t.Context()
	c := openTestConnection(t)
	catalog, ok := c.AsCatalogStore("inst")
	require.True(t, ok)

	// There are no settings until the owner changes them
	_, err := catalog.FindAIMemorySettings(ctx, "alice")
	require.ErrorIs(t, err, drivers.ErrNotFound)

	// Upsert twice
	require.NoError(t, catalog.UpsertAIMemorySettings(ctx, &drivers.AIMemorySettings{OwnerID: "alice", Enabled: false}))
	s, err := catalog.FindAIMemorySettings(ctx, "alice")
	require.NoError(t, err)
	require.False(t, s.Enabled)
	require.Equal(t, "alice", s.OwnerID)
	require.NoError(t, catalog.UpsertAIMemorySettings(ctx, &drivers.AIMemorySettings{OwnerID: "alice", Enabled: true}))
	s, err = catalog.FindAIMemorySettings(ctx, "alice")
	require.NoError(t, err)
	require.True(t, s.Enabled)
}

func TestAISessionCleanupKeepsMemories(t *testing.T) {
	ctx := t.Context()
	c := openTestConnection(t)
	catalog, ok := c.AsCatalogStore("inst")
	require.True(t, ok)
	old := time.Now().UTC().Add(-aiSessionTTL - 24*time.Hour)

	require.NoError(t, catalog.InsertAISession(ctx, &drivers.AISession{ID: "s1", InstanceID: "inst", OwnerID: "a", CreatedOn: old, UpdatedOn: old}))
	m := &drivers.AIMemory{InstanceID: "inst", OwnerID: "a", Agent: "analyst_agent", Content: "x", SourceSessionID: "s1", CreatedOn: old, UpdatedOn: old}
	require.NoError(t, catalog.InsertAIMemory(ctx, m))

	require.NoError(t, c.deleteExpiredAISessions(ctx))

	_, err := catalog.FindAISession(ctx, "s1")
	require.ErrorIs(t, err, drivers.ErrNotFound)
	got, err := catalog.FindAIMemory(ctx, m.ID)
	require.NoError(t, err)
	require.Equal(t, "s1", got.SourceSessionID)
}
