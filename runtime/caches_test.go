package runtime

import (
	"context"
	"sync"
	"testing"
	"time"

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
	qc := newQueryCache(int64(datasize.MB * 100))
	defer qc.cache.Close()

	qc.add(queryCacheKey{"1", "1", "1"}.String(), "value", 1)
	qc.cache.Wait()
	v, ok := qc.get(queryCacheKey{"1", "1", "1"}.String())
	require.Equal(t, "value", v)
	require.True(t, ok)

	qc.add(queryCacheKey{"1", "1", "1"}.String(), nil, 1)
	qc.cache.Wait()
	v, ok = qc.get(queryCacheKey{"1", "1", "1"}.String())
	require.Nil(t, v)
	require.True(t, ok)

	v, ok = qc.get(queryCacheKey{"nosuch", "nosuch", "nosuch"}.String())
	require.Nil(t, v)
	require.False(t, ok)
}

func Test_queryCache_getOrLoad(t *testing.T) {
	qc := newQueryCache(int64(datasize.MB))
	defer qc.cache.Close()

	f := func(ctx context.Context) (interface{}, error) {
		for {
			select {
			case <-ctx.Done():
				// Handle context cancellation
				return nil, ctx.Err()
			case <-time.After(200 * time.Millisecond):
				// Simulate some work
				return &QueryResult{Value: "hello"}, nil
			}
		}
	}
	errs := make([]error, 5)
	values := make([]interface{}, 5)
	cached := make([]bool, 5)
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			var ctx context.Context
			var cancel context.CancelFunc
			if i%2 == 0 {
				// cancel all even requests
				ctx, cancel = context.WithTimeout(context.Background(), 100*time.Millisecond)
			} else {
				ctx, cancel = context.WithCancel(context.TODO())
			}
			defer cancel()
			defer wg.Done()
			values[i], cached[i], errs[i] = qc.getOrLoad(ctx, "key", "query", f)

		}(i)
		time.Sleep(10 * time.Millisecond) // ensure that first goroutine starts the work
	}
	wg.Wait()

	require.False(t, cached[0])
	require.Error(t, errs[0])
	for i := 1; i < 5; i++ {
		if i%2 == 0 {
			require.Error(t, errs[i])
		} else {
			require.True(t, cached[i])
			require.NoError(t, errs[i])
			require.Equal(t, values[i], "hello")
		}
	}
}
