package metricsresolver

import (
	"context"
	"time"
)

func (r *Resolver) ResolveTimeRange(_ context.Context, tr *TimeRange) (col string, start, end *time.Time, ok bool, err error) {
	if tr == nil {
		return "", nil, nil, false, nil
	}

	// TODO: Use qry.time_zone, r.executionTime, if necessary, resolve the watermark or start/end time of the MV
	// TODO: If resolving watermark, cache it to avoid repeat for comparison time range
	panic("not implemented")
}

// func ResolveTimeRange(tr *runtimev1.TimeRange, mv *runtimev1.MetricsViewSpec) (time.Time, time.Time, error) {
// 	tz := time.UTC

// 	if tr.TimeZone != "" {
// 		var err error
// 		tz, err = time.LoadLocation(tr.TimeZone)
// 		if err != nil {
// 			return time.Time{}, time.Time{}, fmt.Errorf("invalid time_range.time_zone %q: %w", tr.TimeZone, err)
// 		}
// 	}

// 	var start, end time.Time
// 	if tr.Start != nil {
// 		start = tr.Start.AsTime().In(tz)
// 	}
// 	if tr.End != nil {
// 		end = tr.End.AsTime().In(tz)
// 	}

// 	isISO := false

// 	if tr.IsoDuration != "" {
// 		if !start.IsZero() && !end.IsZero() {
// 			return time.Time{}, time.Time{}, fmt.Errorf("only two of time_range.{start,end,iso_duration} can be specified")
// 		}

// 		d, err := duration.ParseISO8601(tr.IsoDuration)
// 		if err != nil {
// 			return time.Time{}, time.Time{}, fmt.Errorf("invalid iso_duration %q: %w", tr.IsoDuration, err)
// 		}

// 		if !start.IsZero() {
// 			end = d.Add(start)
// 		} else if !end.IsZero() {
// 			start = d.Sub(end)
// 		} else {
// 			return time.Time{}, time.Time{}, fmt.Errorf("one of time_range.{start,end} must be specified with time_range.iso_duration")
// 		}

// 		isISO = true
// 	}

// 	if tr.IsoOffset != "" {
// 		d, err := duration.ParseISO8601(tr.IsoOffset)
// 		if err != nil {
// 			return time.Time{}, time.Time{}, fmt.Errorf("invalid iso_offset %q: %w", tr.IsoOffset, err)
// 		}

// 		if !start.IsZero() {
// 			start = d.Sub(start)
// 		}
// 		if !end.IsZero() {
// 			end = d.Sub(end)
// 		}

// 		isISO = true
// 	}

// 	// Only modify the start and end if ISO duration or offset was sent.
// 	// This is to maintain backwards compatibility for calls from the UI.
// 	if isISO {
// 		fdow := int(mv.FirstDayOfWeek)
// 		if mv.FirstDayOfWeek > 7 || mv.FirstDayOfWeek <= 0 {
// 			fdow = 1
// 		}
// 		fmoy := int(mv.FirstMonthOfYear)
// 		if mv.FirstMonthOfYear > 12 || mv.FirstMonthOfYear <= 0 {
// 			fmoy = 1
// 		}
// 		if !start.IsZero() {
// 			start = timeutil.TruncateTime(start, convTimeGrain(tr.RoundToGrain), tz, fdow, fmoy)
// 		}
// 		if !end.IsZero() {
// 			end = timeutil.TruncateTime(end, convTimeGrain(tr.RoundToGrain), tz, fdow, fmoy)
// 		}
// 	}

// 	return start, end, nil
// }
