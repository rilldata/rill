package executor

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/pkg/duration"
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

		qry.Spine.TimeRange.Start = timeutil.TruncateTime(qry.TimeRange.Start, computedTimeDims[0].Compute.TimeFloor.Grain.ToTimeutil(), tz, 1, 1)
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

	if tr.Expression == "" {
		return e.resolveISOTimeRange(ctx, tr, tz, executionTime)
	}
	if !tr.Start.IsZero() || !tr.End.IsZero() || tr.IsoDuration != "" || tr.IsoOffset != "" || tr.RoundToGrain != metricsview.TimeGrainUnspecified {
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

	rillTime, err := rilltime.Parse(tr.Expression, rilltime.ParseOptions{
		SmallestGrain:   timeutil.TimeGrainFromAPI(e.metricsView.SmallestTimeGrain),
		DefaultTimeZone: tz,
	})
	if err != nil {
		return err
	}

	// TODO: use grain when we have timeseries from metrics_view_aggregation
	tr.Start, tr.End, _ = rillTime.Eval(rilltime.EvalOptions{
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

// resolveISOTimeRange resolves the given time range where either only start/end is specified along with ISO duration/offset, ensuring only its Start and End properties are populated.
func (e *Executor) resolveISOTimeRange(ctx context.Context, tr *metricsview.TimeRange, tz *time.Location, executionTime *time.Time) error {
	if tr.Start.IsZero() && tr.End.IsZero() {
		if executionTime == nil {
			ts, err := e.Timestamps(ctx, tr.TimeDimension)
			if err != nil {
				return fmt.Errorf("failed to fetch timestamps: %w", err)
			}
			executionTime = &ts.Watermark
		}

		tr.End = *executionTime
	}

	var isISO bool
	if tr.IsoDuration != "" {
		d, err := duration.ParseISO8601(tr.IsoDuration)
		if err != nil {
			return fmt.Errorf("invalid iso_duration %q: %w", tr.IsoDuration, err)
		}

		if !tr.Start.IsZero() && !tr.End.IsZero() {
			return errors.New(`cannot resolve "iso_duration" for a time range with fixed "start" and "end" timestamps`)
		} else if !tr.Start.IsZero() {
			tr.End = d.Add(tr.Start)
		} else if !tr.End.IsZero() {
			tr.Start = d.Sub(tr.End)
		} else {
			// In practice, this shouldn't happen since we resolve a time anchor dynamically if both start and end are zero.
			return errors.New(`cannot resolve "iso_duration" for a time range without "start" and "end" timestamps`)
		}

		isISO = true
	}

	if tr.IsoOffset != "" {
		d, err := duration.ParseISO8601(tr.IsoOffset)
		if err != nil {
			return fmt.Errorf("invalid iso_offset %q: %w", tr.IsoOffset, err)
		}

		if !tr.Start.IsZero() {
			tr.Start = d.Sub(tr.Start)
		}
		if !tr.End.IsZero() {
			tr.End = d.Sub(tr.End)
		}

		isISO = true
	}

	// Only modify the start and end if ISO duration or offset was sent.
	// This is to maintain backwards compatibility for calls from the UI.
	if isISO {
		fdow := int(e.metricsView.FirstDayOfWeek)
		if fdow > 7 || fdow <= 0 {
			fdow = 1
		}
		fmoy := int(e.metricsView.FirstMonthOfYear)
		if fmoy > 12 || fmoy <= 0 {
			fmoy = 1
		}
		if !tr.RoundToGrain.Valid() {
			return fmt.Errorf("invalid time grain %q", tr.RoundToGrain)
		}
		if tr.RoundToGrain != metricsview.TimeGrainUnspecified {
			if !tr.Start.IsZero() {
				tr.Start = timeutil.TruncateTime(tr.Start, tr.RoundToGrain.ToTimeutil(), tz, fdow, fmoy)
			}
			if !tr.End.IsZero() {
				tr.End = timeutil.TruncateTime(tr.End, tr.RoundToGrain.ToTimeutil(), tz, fdow, fmoy)
			}
		}
	}

	// Clear all other fields than Start and End
	tr.IsoDuration = ""
	tr.IsoOffset = ""
	tr.RoundToGrain = metricsview.TimeGrainUnspecified

	return nil
}
