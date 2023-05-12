package runtime

import (
	"context"
	"testing"

	"github.com/c2h5oh/datasize"
	"github.com/google/uuid"
	_ "github.com/rilldata/rill/runtime/drivers/sqlite"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
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
	qc := newQueryCache(int64(datasize.MB * 10))

	qc.add(queryCacheKey{"1", "1", "1"}.String(), "value")
	qc.cache.Wait()
	v, ok := qc.get(queryCacheKey{"1", "1", "1"}.String())
	require.Equal(t, "value", v)
	require.True(t, ok)

	qc.add(queryCacheKey{"1", "1", "1"}.String(), nil)
	qc.cache.Wait()
	v, ok = qc.get(queryCacheKey{"1", "1", "1"}.String())
	require.Nil(t, v)
	require.True(t, ok)

	v, ok = qc.get(queryCacheKey{"nosuch", "nosuch", "nosuch"}.String())
	require.Nil(t, v)
	require.False(t, ok)
}
