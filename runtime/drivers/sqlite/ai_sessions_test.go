package sqlite

import (
	"testing"
	"time"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
)

func TestDeleteAISession(t *testing.T) {
	ctx := t.Context()
	c := openTestConnection(t)
	catalog, ok := c.AsCatalogStore("inst")
	require.True(t, ok)
	otherCatalog, ok := c.AsCatalogStore("other")
	require.True(t, ok)
	now := time.Now().UTC()

	require.NoError(t, catalog.InsertAISession(ctx, &drivers.AISession{ID: "s1", InstanceID: "inst", OwnerID: "a", CreatedOn: now, UpdatedOn: now}))
	require.NoError(t, catalog.InsertAIMessage(ctx, &drivers.AIMessage{ID: "m1", SessionID: "s1", CreatedOn: now, UpdatedOn: now}))
	require.NoError(t, catalog.InsertAISession(ctx, &drivers.AISession{ID: "s2", InstanceID: "inst", OwnerID: "a", CreatedOn: now, UpdatedOn: now}))
	require.NoError(t, catalog.InsertAIMessage(ctx, &drivers.AIMessage{ID: "m2", SessionID: "s2", CreatedOn: now, UpdatedOn: now}))

	// A session in another instance cannot be deleted.
	require.ErrorIs(t, otherCatalog.DeleteAISession(ctx, "s1"), drivers.ErrNotFound)
	_, err := catalog.FindAISession(ctx, "s1")
	require.NoError(t, err)

	// Deleting a session removes it and its messages, leaving other sessions intact.
	require.NoError(t, catalog.DeleteAISession(ctx, "s1"))
	_, err = catalog.FindAISession(ctx, "s1")
	require.ErrorIs(t, err, drivers.ErrNotFound)
	msgs, err := catalog.FindAIMessages(ctx, "s1")
	require.NoError(t, err)
	require.Empty(t, msgs)
	msgs, err = catalog.FindAIMessages(ctx, "s2")
	require.NoError(t, err)
	require.Len(t, msgs, 1)

	// Deleting a missing session returns ErrNotFound.
	require.ErrorIs(t, catalog.DeleteAISession(ctx, "s1"), drivers.ErrNotFound)

	// Writes to a deleted session return ErrNotFound and leave no orphaned rows.
	require.ErrorIs(t, catalog.UpdateAISession(ctx, &drivers.AISession{ID: "s1", Title: "x"}), drivers.ErrNotFound)
	require.ErrorIs(t, catalog.InsertAIMessage(ctx, &drivers.AIMessage{ID: "m3", SessionID: "s1", CreatedOn: now, UpdatedOn: now}), drivers.ErrNotFound)
	msgs, err = catalog.FindAIMessages(ctx, "s1")
	require.NoError(t, err)
	require.Empty(t, msgs)
}
