package duckdb_test

import (
	"testing"

	"github.com/rilldata/rill/runtime/drivers"
	activity "github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestMotherDuckModeEnforcement(t *testing.T) {
	t.Run("read mode blocks model execution", func(t *testing.T) {
		cfg := testruntime.AcquireConnector(t, "motherduck")
		cfg["mode"] = "read"

		handle, err := drivers.Open("motherduck", "test", cfg,
			storage.MustNew(t.TempDir(), nil),
			activity.NewNoopClient(),
			zap.NewNop())
		require.NoError(t, err)
		defer handle.Close()

		// Test AsModelExecutor
		opts := &drivers.ModelExecutorOptions{
			InputHandle:  handle,
			OutputHandle: handle,
		}
		executor, err := handle.AsModelExecutor("test", opts)
		require.ErrorContains(t, err, "model execution is disabled")
		require.Nil(t, executor)

		// Test AsModelManager
		manager, ok := handle.AsModelManager("test")
		require.False(t, ok)
		require.Nil(t, manager)
	})

	t.Run("readwrite mode allows model execution", func(t *testing.T) {
		cfg := testruntime.AcquireConnector(t, "motherduck")
		cfg["mode"] = "readwrite"

		handle, err := drivers.Open("motherduck", "test", cfg,
			storage.MustNew(t.TempDir(), nil),
			activity.NewNoopClient(),
			zap.NewNop())
		require.NoError(t, err)
		defer handle.Close()

		// Test AsModelExecutor
		opts := &drivers.ModelExecutorOptions{
			InputHandle:  handle,
			OutputHandle: handle,
		}
		executor, err := handle.AsModelExecutor("test", opts)
		require.NoError(t, err)
		require.NotNil(t, executor)

		// Test AsModelManager
		manager, ok := handle.AsModelManager("test")
		require.True(t, ok)
		require.NotNil(t, manager)
	})
}
