package metricsview

import "time"

type TimestampsResult struct {
	Min       time.Time
	Max       time.Time
	Watermark time.Time
	Now       time.Time
	Rollups   map[string]TimestampsResult // keyed by rollup table name; nil when no rollups or for non-primary time dimension
}
