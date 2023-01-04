package runtime

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	_ "github.com/rilldata/rill/runtime/drivers/sqlite"
)

func TestConnectionCache(t *testing.T) {
	ctx := context.Background()
	id := uuid.NewString()

	c := newConnectionCache(10, zap.NewNop())
	conn1, err := c.get(ctx, id, "sqlite", ":memory:")
	require.NoError(t, err)
	require.NotNil(t, conn1)

	conn2, err := c.get(ctx, id, "sqlite", ":memory:")
	require.NoError(t, err)
	require.NotNil(t, conn2)

	conn3, err := c.get(ctx, uuid.NewString(), "sqlite", ":memory:")
	require.NoError(t, err)
	require.NotNil(t, conn3)

	require.True(t, conn1 == conn2)
	require.False(t, conn2 == conn3)
}

func TestNilValues(t *testing.T) {
	qc := newQueryCache(10)

	qc.add(queryCacheKey{"1", "1", "1"}, "value")
	v, ok := qc.get(queryCacheKey{"1", "1", "1"})
	require.Equal(t, "value", v)
	require.True(t, ok)

	qc.add(queryCacheKey{"1", "1", "1"}, nil)
	v, ok = qc.get(queryCacheKey{"1", "1", "1"})
	require.Nil(t, v)
	require.True(t, ok)

	v, ok = qc.get(queryCacheKey{"nosuch", "nosuch", "nosuch"})
	require.Nil(t, v)
	require.False(t, ok)
}
