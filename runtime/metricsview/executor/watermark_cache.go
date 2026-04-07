package executor

import (
	"context"
	"errors"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
)

// fetchRollupWatermark returns the min/max time of the rollup table via the metrics_watermark resolver.
func (e *Executor) fetchRollupWatermark(ctx context.Context, rollup *runtimev1.MetricsViewSpec_RollupTable) (time.Time, time.Time, error) {
	return e.resolveWatermark(ctx, rollup.Table, rollup.Database, rollup.DatabaseSchema)
}

// fetchBaseWatermark returns the min/max time of the base table via the metrics_watermark resolver.
func (e *Executor) fetchBaseWatermark(ctx context.Context) (time.Time, time.Time, error) {
	return e.resolveWatermark(ctx, "", "", "")
}

// resolveWatermark calls the metrics_watermark resolver to fetch min/max timestamps.
func (e *Executor) resolveWatermark(ctx context.Context, table, database, databaseSchema string) (time.Time, time.Time, error) {
	args := map[string]any{
		"priority": e.priority,
	}
	if table != "" {
		args["table"] = table
	}
	if database != "" {
		args["database"] = database
	}
	if databaseSchema != "" {
		args["database_schema"] = databaseSchema
	}

	res, _, err := e.rt.Resolve(ctx, &runtime.ResolveOptions{
		InstanceID: e.instanceID,
		Resolver:   "metrics_watermark",
		ResolverProperties: map[string]any{
			"metrics_view": e.metricsViewName,
		},
		Args:   args,
		Claims: &runtime.SecurityClaims{SkipChecks: true},
	})
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	defer res.Close()

	row, err := res.Next()
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	mn, err := toTime(row["min"])
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	mx, err := toTime(row["max"])
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	return mn, mx, nil
}

// toTime converts an any value to time.Time (handles nil, time.Time, and string).
func toTime(v any) (time.Time, error) {
	if v == nil {
		return time.Time{}, nil
	}
	switch t := v.(type) {
	case time.Time:
		return t, nil
	case string:
		return time.Parse(time.RFC3339Nano, t)
	default:
		return time.Time{}, errors.New("unexpected type for time value")
	}
}
