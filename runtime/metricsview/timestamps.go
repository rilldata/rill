package metricsview

import "time"

type TimestampsResult struct {
	BaseTable TimestampsResultEntry
	Rollups   map[string]TimestampsResultEntry
}

type TimestampsResultEntry struct {
	Min       time.Time
	Max       time.Time
	Watermark time.Time
	Now       time.Time
}
