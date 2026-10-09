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

	m := &drivers.AIMemory{
		ID:              "m1",
		InstanceID:      "inst",
		OwnerID:         "alice",
		Category:        "preference",
		Content:         "Prefers tables over charts",
		Source:          drivers.AIMemorySourceExplicit,
		SourceSessionID: "s1",
		SourceMessageID: "msg1",
		CreatedOn:       now,
		UpdatedOn:       now,
	}
	require.NoError(t, catalog.InsertAIMemory(ctx, m))
	require.NoError(t, catalog.InsertAIMemory(ctx, &drivers.AIMemory{
		ID: "m2", InstanceID: "inst", OwnerID: "alice", Category: "definition", Content: "Revenue means net revenue",
		Source: drivers.AIMemorySourceExtracted, CreatedOn: now, UpdatedOn: now,
	}))
	require.NoError(t, catalog.InsertAIMemory(ctx, &drivers.AIMemory{
		ID: "m3", InstanceID: "inst", OwnerID: "bob", Category: "context", Content: "Works in finance",
		Source: drivers.AIMemorySourceManual, CreatedOn: now, UpdatedOn: now,
	}))

	// Find by owner
	res, err := catalog.FindAIMemories(ctx, "alice")
	require.NoError(t, err)
	require.Len(t, res, 2)
	require.Equal(t, "m1", res[0].ID)
	require.Equal(t, "Prefers tables over charts", res[0].Content)
	require.Equal(t, "s1", res[0].SourceSessionID)
	require.Equal(t, "msg1", res[0].SourceMessageID)

	res, err = catalog.FindAIMemories(ctx, "bob")
	require.NoError(t, err)
	require.Len(t, res, 1)
	require.Equal(t, "m3", res[0].ID)

	// Find one
	got, err := catalog.FindAIMemory(ctx, "m1")
	require.NoError(t, err)
	require.Equal(t, "alice", got.OwnerID)
	_, err = catalog.FindAIMemory(ctx, "missing")
	require.ErrorIs(t, err, drivers.ErrNotFound)

	// Instance isolation
	other, ok := c.AsCatalogStore("other")
	require.True(t, ok)
	res, err = other.FindAIMemories(ctx, "alice")
	require.NoError(t, err)
	require.Empty(t, res)
	_, err = other.FindAIMemory(ctx, "m1")
	require.ErrorIs(t, err, drivers.ErrNotFound)

	// Update
	got.Content = "Prefers tables"
	got.Category = "feedback"
	require.NoError(t, catalog.UpdateAIMemory(ctx, got))
	got, err = catalog.FindAIMemory(ctx, "m1")
	require.NoError(t, err)
	require.Equal(t, "Prefers tables", got.Content)
	require.Equal(t, "feedback", got.Category)
	require.False(t, got.UpdatedOn.Before(now))

	// Delete leaves other memories alone
	require.NoError(t, catalog.DeleteAIMemory(ctx, "m1"))
	_, err = catalog.FindAIMemory(ctx, "m1")
	require.ErrorIs(t, err, drivers.ErrNotFound)
	res, err = catalog.FindAIMemories(ctx, "alice")
	require.NoError(t, err)
	require.Len(t, res, 1)
	require.Equal(t, "m2", res[0].ID)
}

func TestAIMemorySettings(t *testing.T) {
	ctx := t.Context()
	c := openTestConnection(t)
	catalog, ok := c.AsCatalogStore("inst")
	require.True(t, ok)

	// Memory is on when the owner has never changed the settings
	s, err := catalog.FindAIMemorySettings(ctx, "alice")
	require.NoError(t, err)
	require.True(t, s.Enabled)
	require.Equal(t, "alice", s.OwnerID)

	// Upsert twice
	require.NoError(t, catalog.UpsertAIMemorySettings(ctx, &drivers.AIMemorySettings{OwnerID: "alice", Enabled: false}))
	s, err = catalog.FindAIMemorySettings(ctx, "alice")
	require.NoError(t, err)
	require.False(t, s.Enabled)
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
	require.NoError(t, catalog.InsertAIMemory(ctx, &drivers.AIMemory{ID: "m1", InstanceID: "inst", OwnerID: "a", Category: "context", Content: "x", Source: "explicit", SourceSessionID: "s1", CreatedOn: old, UpdatedOn: old}))

	require.NoError(t, c.deleteExpiredAISessions(ctx))

	_, err := catalog.FindAISession(ctx, "s1")
	require.ErrorIs(t, err, drivers.ErrNotFound)
	m, err := catalog.FindAIMemory(ctx, "m1")
	require.NoError(t, err)
	require.Equal(t, "s1", m.SourceSessionID)
}
