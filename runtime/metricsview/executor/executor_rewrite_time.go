package executor

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/pkg/rilltime"
	"github.com/rilldata/rill/runtime/pkg/timeutil"
)

// rewriteQueryTimeRanges rewrites the time ranges in the query to fixed start/end timestamps.
func (e *Executor) rewriteQueryTimeRanges(ctx context.Context, qry *metricsview.Query, executionTime *time.Time) error {
	if e.metricsView.TimeDimension == "" && (qry.TimeRange == nil || qry.TimeRange.TimeDimension == "") {
		return nil
	}

	tz := time.UTC
	if qry.TimeZone != "" {
		var err error
		tz, err = time.LoadLocation(qry.TimeZone)
		if err != nil {
			return fmt.Errorf("invalid time zone %q: %w", qry.TimeZone, err)
		}
	}

	err := e.resolveTimeRange(ctx, qry.TimeRange, tz, executionTime)
	if err != nil {
		return fmt.Errorf("failed to resolve time range: %w", err)
	}

	err = e.resolveTimeRange(ctx, qry.ComparisonTimeRange, tz, executionTime)
	if err != nil {
		return fmt.Errorf("failed to resolve comparison time range: %w", err)
	}

	// If time range is specified in the spine, resolve it.
	if qry.Spine != nil && qry.Spine.TimeRange != nil {
		var computedTimeDims []*metricsview.Dimension
		for _, d := range qry.Dimensions {
			if d.Compute != nil && d.Compute.TimeFloor != nil {
				computedTimeDims = append(computedTimeDims, &d)
			}
		}

		if len(computedTimeDims) != 1 {
			return errors.New("spine time range is only supported with a single time dimension")
		}

		qry.Spine.TimeRange.Start = qry.TimeRange.Start
		qry.Spine.TimeRange.End = qry.TimeRange.End
		qry.Spine.TimeRange.Grain = computedTimeDims[0].Compute.TimeFloor.Grain
		qry.Spine.TimeRange.TimeDimension = computedTimeDims[0].Compute.TimeFloor.Dimension
	}

	return nil
}

// resolveTimeRange resolves the given time range, ensuring only its Start and End properties are populated.
func (e *Executor) resolveTimeRange(ctx context.Context, tr *metricsview.TimeRange, tz *time.Location, executionTime *time.Time) error {
	if tr == nil || tr.IsZero() {
		return nil
	}

	// Start and End are hardcoded, skip the rest of the code
	if !tr.Start.IsZero() && !tr.End.IsZero() {
		// Clear all other fields than Start and End
		tr.Expression = ""
		tr.IsoDuration = ""
		tr.IsoOffset = ""
		tr.RoundToGrain = metricsview.TimeGrainUnspecified
		return nil
	}

	if tr.Expression != "" && (!tr.Start.IsZero() || !tr.End.IsZero() || tr.IsoDuration != "" || tr.IsoOffset != "" || tr.RoundToGrain != metricsview.TimeGrainUnspecified) {
		return errors.New("other fields are not supported when expression is provided")
	}

	// TODO: Implement lazy evaluation where we only evaluate timestamps if required for the time expression.
	ts, err := e.Timestamps(ctx, tr.TimeDimension)
	if err != nil {
		return fmt.Errorf("failed to fetch timestamps: %w", err)
	}
	if executionTime != nil {
		// If provided, all the end anchors should use the execution time.
		ts.Watermark = *executionTime
		ts.Max = *executionTime
		ts.Now = *executionTime
	}

	var rt *rilltime.Expression
	if tr.IsoDuration != "" || tr.IsoOffset != "" || tr.RoundToGrain != metricsview.TimeGrainUnspecified {
		rt, err = rilltime.ParseISO(tr.IsoDuration, tr.IsoOffset, tr.End, tr.RoundToGrain.ToTimeutil(), rilltime.ParseOptions{
			DefaultTimeZone: tz,
			SmallestGrain:   timeutil.TimeGrainFromAPI(e.metricsView.SmallestTimeGrain),
		})
	} else {
		rt, err = rilltime.Parse(tr.Expression, rilltime.ParseOptions{
			DefaultTimeZone: tz,
			SmallestGrain:   timeutil.TimeGrainFromAPI(e.metricsView.SmallestTimeGrain),
		})
	}
	if err != nil {
		return err
	}

	// TODO: use grain when we have timeseries from metrics_view_aggregation
	tr.Start, tr.End, _ = rt.Eval(rilltime.EvalOptions{
		Now:        ts.Now,
		MinTime:    ts.Min,
		MaxTime:    ts.Max,
		Watermark:  ts.Watermark,
		FirstDay:   int(e.metricsView.FirstDayOfWeek),
		FirstMonth: int(e.metricsView.FirstMonthOfYear),
	})

	// Clear all other fields than Start and End
	tr.Expression = ""
	tr.IsoDuration = ""
	tr.IsoOffset = ""
	tr.RoundToGrain = metricsview.TimeGrainUnspecified

	return nil
}
