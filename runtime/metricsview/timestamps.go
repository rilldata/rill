package metricsview

import "time"

type TimestampsResult struct {
	Min       time.Time
	Max       time.Time
	Watermark time.Time
	Now       time.Time
}
