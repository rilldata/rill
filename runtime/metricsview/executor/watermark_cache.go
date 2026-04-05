package executor

import (
	"context"
	"sync"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

const defaultWatermarkCacheTTL = 5 * time.Minute

type watermarkEntry struct {
	min       time.Time
	max       time.Time
	fetchedAt time.Time
}

var watermarkCache = struct {
	mu    sync.Mutex
	items map[string]watermarkEntry
}{items: make(map[string]watermarkEntry)}

func watermarkCacheKey(instanceID, db, schema, table string) string {
	return instanceID + ":" + db + ":" + schema + ":" + table
}

func getWatermark(key string, ttl time.Duration) (watermarkEntry, bool) {
	watermarkCache.mu.Lock()
	defer watermarkCache.mu.Unlock()
	wm, ok := watermarkCache.items[key]
	if !ok {
		return watermarkEntry{}, false
	}
	if time.Since(wm.fetchedAt) > ttl {
		delete(watermarkCache.items, key)
		return watermarkEntry{}, false
	}
	return wm, true
}

func setWatermark(key string, minTime, maxTime time.Time) {
	watermarkCache.mu.Lock()
	defer watermarkCache.mu.Unlock()
	watermarkCache.items[key] = watermarkEntry{
		min:       minTime,
		max:       maxTime,
		fetchedAt: time.Now(),
	}
}

// fetchRollupWatermark returns the min/max time of the rollup table, using a cache with TTL.
// Returns ok=false on errors (caller should skip the rollup).
// Prone to thundering herd issue but keeping it simple for now.
func (e *Executor) fetchRollupWatermark(ctx context.Context, rollup *runtimev1.MetricsViewSpec_RollupTable) (minTime, maxTime time.Time, ok bool) {
	ttl := defaultWatermarkCacheTTL
	if rollup.WatermarkCacheTtlSeconds > 0 {
		ttl = time.Duration(rollup.WatermarkCacheTtlSeconds) * time.Second
	}

	key := watermarkCacheKey(e.instanceID, rollup.Database, rollup.DatabaseSchema, rollup.Table)
	if wm, hit := getWatermark(key, ttl); hit {
		return wm.min, wm.max, true
	}

	mn, mx, err := e.fetchTimestamps(ctx, rollup.Database, rollup.DatabaseSchema, rollup.Table)
	if err != nil {
		return time.Time{}, time.Time{}, false
	}
	setWatermark(key, mn, mx)
	return mn, mx, true
}

// fetchBaseWatermark returns the min/max time of the base table, using the shared watermark cache.
// Returns ok=false on errors (caller should proceed without base timestamps).
// Prone to thundering herd issue but keeping it simple for now.
func (e *Executor) fetchBaseWatermark(ctx context.Context) (minTime, maxTime time.Time, ok bool) {
	key := watermarkCacheKey(e.instanceID, e.metricsView.Database, e.metricsView.DatabaseSchema, e.metricsView.Table)
	if wm, hit := getWatermark(key, defaultWatermarkCacheTTL); hit {
		return wm.min, wm.max, true
	}

	if e.olap == nil {
		return time.Time{}, time.Time{}, false
	}

	ts, err := e.Timestamps(ctx, "")
	if err != nil || ts.Min.IsZero() {
		return time.Time{}, time.Time{}, false
	}
	setWatermark(key, ts.Min, ts.Max)
	return ts.Min, ts.Max, true
}
