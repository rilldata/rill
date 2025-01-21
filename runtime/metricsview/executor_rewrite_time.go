package metricsview

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rilldata/rill/runtime/pkg/duration"
	"github.com/rilldata/rill/runtime/pkg/rilltime"
	"github.com/rilldata/rill/runtime/pkg/timeutil"
)

// rewriteQueryTimeRanges rewrites the time ranges in the query to fixed start/end timestamps.
func (e *Executor) rewriteQueryTimeRanges(ctx context.Context, qry *Query, executionTime *time.Time) error {
	if e.metricsView.TimeDimension == "" {
		return nil
	}

	ts, err := e.Timestamps(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch timestamps: %w", err)
	}

	tz := time.UTC
	if qry.TimeZone != "" {
		var err error
		tz, err = time.LoadLocation(qry.TimeZone)
		if err != nil {
			return fmt.Errorf("invalid time zone %q: %w", qry.TimeZone, err)
		}
	}

	maxTime := ts.Max
	watermark := ts.Watermark
	now := time.Now()
	if executionTime != nil {
		// all end anchors should use execution time when provided
		maxTime = *executionTime
		watermark = *executionTime
		now = *executionTime
	}

	err = e.resolveTimeRange(qry.TimeRange, tz, ts.Min, maxTime, watermark, now)
	if err != nil {
		return fmt.Errorf("failed to resolve time range: %w", err)
	}

	err = e.resolveTimeRange(qry.ComparisonTimeRange, tz, ts.Min, maxTime, watermark, now)
	if err != nil {
		return fmt.Errorf("failed to resolve comparison time range: %w", err)
	}

	return nil
}

// resolveTimeRange resolves the given time range, ensuring only its Start and End properties are populated.
func (e *Executor) resolveTimeRange(tr *TimeRange, tz *time.Location, minTime, maxTime, watermark, now time.Time) error {
	if tr == nil || tr.IsZero() {
		return nil
	}

	if tr.Expression == "" {
		return e.resolveISOTimeRange(tr, tz, watermark)
	}
	if !tr.Start.IsZero() || !tr.End.IsZero() || tr.IsoDuration != "" || tr.IsoOffset != "" || tr.RoundToGrain != TimeGrainUnspecified {
		return errors.New("other fields are not supported when expression is provided")
	}

	rillTime, err := rilltime.Parse(tr.Expression, rilltime.ParseOptions{})
	if err != nil {
		return err
	}

	tr.Start, tr.End, err = rillTime.Eval(rilltime.EvalOptions{
		Now:        now,
		MinTime:    minTime,
		MaxTime:    maxTime,
		Watermark:  watermark,
		FirstDay:   int(e.metricsView.FirstDayOfWeek),
		FirstMonth: int(e.metricsView.FirstMonthOfYear),
	})
	if err != nil {
		return err
	}

	// Clear all other fields than Start and End
	tr.Expression = ""
	tr.IsoDuration = ""
	tr.IsoOffset = ""
	tr.RoundToGrain = TimeGrainUnspecified

	return nil
}

// resolveISOTimeRange resolves the given time range where either only start/end is specified along with ISO duration/offset, ensuring only its Start and End properties are populated.
func (e *Executor) resolveISOTimeRange(tr *TimeRange, tz *time.Location, watermark time.Time) error {
	if tr.Start.IsZero() && tr.End.IsZero() {
		tr.End = watermark
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
		if tr.RoundToGrain != TimeGrainUnspecified {
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
	tr.RoundToGrain = TimeGrainUnspecified

	return nil
}
