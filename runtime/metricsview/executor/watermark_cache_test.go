package executor

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestWatermarkCache(t *testing.T) {
	t.Run("key format", func(t *testing.T) {
		key := watermarkCacheKey("inst1", "mydb", "myschema", "mytable")
		require.Equal(t, "inst1:mydb:myschema:mytable", key)
	})

	t.Run("set and get", func(t *testing.T) {
		key := watermarkCacheKey("test", "", "", "cache_test_table")
		mn := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
		mx := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
		setWatermark(key, mn, mx)

		wm, ok := getWatermark(key, time.Hour)
		require.True(t, ok)
		require.Equal(t, mn, wm.min)
		require.Equal(t, mx, wm.max)
	})

	t.Run("miss", func(t *testing.T) {
		_, ok := getWatermark("nonexistent:key", time.Hour)
		require.False(t, ok)
	})

	t.Run("expired", func(t *testing.T) {
		key := watermarkCacheKey("test", "", "", "expired_table")
		mn := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
		mx := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)

		// Manually insert an entry with a past fetchedAt
		watermarkCache.mu.Lock()
		watermarkCache.items[key] = watermarkEntry{
			min:       mn,
			max:       mx,
			fetchedAt: time.Now().Add(-2 * time.Hour),
		}
		watermarkCache.mu.Unlock()

		// TTL of 1 hour; entry was fetched 2 hours ago
		_, ok := getWatermark(key, time.Hour)
		require.False(t, ok)

		// Verify it was cleaned up
		watermarkCache.mu.Lock()
		_, exists := watermarkCache.items[key]
		watermarkCache.mu.Unlock()
		require.False(t, exists)
	})
}
