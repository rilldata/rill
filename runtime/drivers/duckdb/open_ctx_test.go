package duckdb

import (
	"context"
	"testing"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// TestOpenCtxNotRetained verifies that the handle does not retain the ctx passed to Open.
// The connection cache cancels that ctx as soon as Open returns,
// so anything that outlives Open (notably the per-connection init queries, which run for every connection the pool opens later)
// must not observe its cancellation.
func TestOpenCtxNotRetained(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	handle, err := Driver{}.Open(ctx, "", "default", map[string]any{"data_dir": t.TempDir(), "pool_size": 4}, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	defer handle.Close()
	require.NoError(t, handle.Migrate(ctx))

	// Mimic the connection cache, which cancels the open ctx immediately after Open returns.
	cancel()

	olap, ok := handle.AsOLAP("default")
	require.True(t, ok)

	// Query concurrently so the pool has to open connections it didn't already have at the time of the cancellation.
	var g errgroup.Group
	for range 8 {
		g.Go(func() error {
			res, err := olap.Query(context.Background(), &drivers.Statement{Query: "SELECT 1"})
			if err != nil {
				return err
			}
			return res.Close()
		})
	}
	require.NoError(t, g.Wait())
}
