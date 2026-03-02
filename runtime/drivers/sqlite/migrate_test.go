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

func TestDeleteExpiredAISessions(t *testing.T) {
	now := time.Now().UTC()
	old := now.Add(-aiSessionTTL - 24*time.Hour)

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	storageDir := filepath.Join(tmpDir, "storage")
	cfg := map[string]any{"dsn": dbPath}

	// Open the database, run migrations, and seed test data.
	h, err := driver{}.Open("", cfg, storage.MustNew(storageDir, nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	require.NoError(t, h.Migrate(t.Context()))

	catalog, ok := h.AsCatalogStore("inst")
	require.True(t, ok)

	// Session 1: old session with only old messages (should be deleted).
	require.NoError(t, catalog.InsertAISession(t.Context(), &drivers.AISession{ID: "s1", CreatedOn: old, UpdatedOn: old}))
	require.NoError(t, catalog.InsertAIMessage(t.Context(), &drivers.AIMessage{ID: "m1", SessionID: "s1", CreatedOn: old, UpdatedOn: old}))

	// Session 2: old session with no messages (should be deleted).
	require.NoError(t, catalog.InsertAISession(t.Context(), &drivers.AISession{ID: "s2", CreatedOn: old, UpdatedOn: old}))

	// Session 3: old session with a recent message (should be kept).
	require.NoError(t, catalog.InsertAISession(t.Context(), &drivers.AISession{ID: "s3", CreatedOn: old, UpdatedOn: now}))
	require.NoError(t, catalog.InsertAIMessage(t.Context(), &drivers.AIMessage{ID: "m3", SessionID: "s3", CreatedOn: now, UpdatedOn: now}))

	// Session 4: new session with no messages (should be kept).
	require.NoError(t, catalog.InsertAISession(t.Context(), &drivers.AISession{ID: "s4", CreatedOn: now, UpdatedOn: now}))

	// Close the handle, then re-open and migrate to trigger the TTL cleanup.
	require.NoError(t, h.Close())
	h, err = driver{}.Open("", cfg, storage.MustNew(storageDir, nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	defer h.Close()
	require.NoError(t, h.Migrate(t.Context()))

	// Query the database directly to verify results.
	db := h.(*connection).db

	var sessionIDs []string
	require.NoError(t, db.SelectContext(t.Context(), &sessionIDs, `SELECT id FROM ai_sessions ORDER BY id`))
	require.Equal(t, []string{"s3", "s4"}, sessionIDs)

	var messageIDs []string
	require.NoError(t, db.SelectContext(t.Context(), &messageIDs, `SELECT id FROM ai_messages ORDER BY id`))
	require.Equal(t, []string{"m3"}, messageIDs)
}
