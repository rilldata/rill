package metricsview

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/duration"
	"github.com/rilldata/rill/runtime/pkg/rilltime"
	"github.com/rilldata/rill/runtime/pkg/timeutil"
)

// rewriteQueryTimeRanges rewrites the time ranges in the query to fixed start/end timestamps.
func (e *Executor) rewriteQueryTimeRanges(ctx context.Context, qry *Query, executionTime *time.Time) error {
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

	return nil
}

// resolveTimeRange resolves the given time range, ensuring only its Start and End properties are populated.
func (e *Executor) resolveTimeRange(ctx context.Context, tr *TimeRange, tz *time.Location, executionTime *time.Time) error {
	if tr == nil || tr.IsZero() {
		return nil
	}

	if tr.RillTime == "" {
		return e.resolveTimeRangeLegacy(ctx, tr, tz, executionTime)
	}

	rt, err := rilltime.Parse(tr.RillTime)
	if err != nil {
		return err
	}

	t, err := e.loadWatermark(ctx, executionTime)
	if err != nil {
		return err
	}

	tr.Start, tr.End, err = rt.Resolve(rilltime.ResolverContext{
		Now:        time.Now(),
		MinTime:    t,
		MaxTime:    time.Time{}, // TODO
		FirstDay:   int(e.metricsView.FirstDayOfWeek),
		FirstMonth: int(e.metricsView.FirstMonthOfYear),
	})
	if err != nil {
		return err
	}

	// Clear all other fields than Start and End
	tr.RillTime = ""
	tr.IsoDuration = ""
	tr.IsoOffset = ""
	tr.RoundToGrain = TimeGrainUnspecified

	return nil
}

// resolveTimeRange resolves the given time range, ensuring only its Start and End properties are populated.
func (e *Executor) resolveTimeRangeLegacy(ctx context.Context, tr *TimeRange, tz *time.Location, executionTime *time.Time) error {
	if tr.Start.IsZero() && tr.End.IsZero() {
		t, err := e.loadWatermark(ctx, executionTime)
		if err != nil {
			return err
		}
		tr.End = t
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
	tr.RillTime = ""
	tr.IsoDuration = ""
	tr.IsoOffset = ""
	tr.RoundToGrain = TimeGrainUnspecified

	return nil
}

// resolveWatermark resolves the metric view's watermark expression.
// If the resolved time is zero, it defaults to the current time.
func (e *Executor) loadWatermark(ctx context.Context, executionTime *time.Time) (time.Time, error) {
	if executionTime != nil {
		return *executionTime, nil
	}

	if !e.watermark.IsZero() {
		return e.watermark, nil
	}

	dialect := e.olap.Dialect()

	var expr string
	if e.metricsView.WatermarkExpression != "" {
		expr = e.metricsView.WatermarkExpression
	} else if e.metricsView.TimeDimension != "" {
		expr = fmt.Sprintf("MAX(%s)", dialect.EscapeIdentifier(e.metricsView.TimeDimension))
	} else {
		return time.Time{}, errors.New("cannot determine time anchor for relative time range")
	}

	sql := fmt.Sprintf("SELECT %s FROM %s", expr, dialect.EscapeTable(e.metricsView.Database, e.metricsView.DatabaseSchema, e.metricsView.Table))

	res, err := e.olap.Execute(ctx, &drivers.Statement{
		Query:            sql,
		Priority:         e.priority,
		ExecutionTimeout: defaultInteractiveTimeout,
	})
	if err != nil {
		return time.Time{}, err
	}
	defer res.Close()

	var t time.Time
	if res.Next() {
		if err := res.Scan(&t); err != nil {
			return time.Time{}, fmt.Errorf("failed to scan time anchor: %w", err)
		}
	}

	if t.IsZero() {
		t = time.Now()
	}

	e.watermark = t
	return t, nil
}
