package sqlite

import (
	"testing"
	"time"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestCatalogModelPartitionRetryFields(t *testing.T) {
	cfg := map[string]any{"dsn": ":memory:"}
	h, err := driver{}.Open("", cfg, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	defer h.Close()
	require.NoError(t, h.Migrate(t.Context()))

	catalog, ok := h.AsCatalogStore("test-instance")
	require.True(t, ok)

	now := time.Now().UTC().Truncate(time.Millisecond)
	p := drivers.ModelPartition{
		Key:        "p1",
		DataJSON:   []byte(`{"day":"2026-01-01"}`),
		Index:      1,
		Watermark:  &now,
		ExecutedOn: &now,
		Error:      "failed",
		Elapsed:    2 * time.Second,
		RetryUsed:  2,
		RetryMax:   3,
	}

	require.NoError(t, catalog.InsertModelPartition(t.Context(), "m1", p))
	got, err := catalog.FindModelPartitionsByKeys(t.Context(), "m1", []string{"p1"})
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, uint32(2), got[0].RetryUsed)
	require.Equal(t, uint32(3), got[0].RetryMax)

	p.RetryUsed = 1
	p.RetryMax = 5
	require.NoError(t, catalog.UpdateModelPartition(t.Context(), "m1", p))
	got, err = catalog.FindModelPartitionsByKeys(t.Context(), "m1", []string{"p1"})
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, uint32(1), got[0].RetryUsed)
	require.Equal(t, uint32(5), got[0].RetryMax)
}
