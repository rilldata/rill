package queries

import (
	"fmt"
	"time"

	"github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/duration"
	"github.com/rilldata/rill/runtime/pkg/timeutil"
)

func convTimeGrain(tg runtimev1.TimeGrain) timeutil.TimeGrain {
	switch tg {
	case runtimev1.TimeGrain_TIME_GRAIN_MILLISECOND:
		return timeutil.TimeGrainMillisecond
	case runtimev1.TimeGrain_TIME_GRAIN_SECOND:
		return timeutil.TimeGrainSecond
	case runtimev1.TimeGrain_TIME_GRAIN_MINUTE:
		return timeutil.TimeGrainMinute
	case runtimev1.TimeGrain_TIME_GRAIN_HOUR:
		return timeutil.TimeGrainHour
	case runtimev1.TimeGrain_TIME_GRAIN_DAY:
		return timeutil.TimeGrainDay
	case runtimev1.TimeGrain_TIME_GRAIN_WEEK:
		return timeutil.TimeGrainWeek
	case runtimev1.TimeGrain_TIME_GRAIN_MONTH:
		return timeutil.TimeGrainMonth
	case runtimev1.TimeGrain_TIME_GRAIN_QUARTER:
		return timeutil.TimeGrainQuarter
	case runtimev1.TimeGrain_TIME_GRAIN_YEAR:
		return timeutil.TimeGrainYear
	}
	return timeutil.TimeGrainUnspecified
}

func timeGrainToDuration(tg runtimev1.TimeGrain) duration.Duration {
	switch tg {
	// not supported
	// case runtimev1.TimeGrain_TIME_GRAIN_MILLISECOND:
	case runtimev1.TimeGrain_TIME_GRAIN_SECOND:
		return duration.StandardDuration{Second: 1}
	case runtimev1.TimeGrain_TIME_GRAIN_MINUTE:
		return duration.StandardDuration{Minute: 1}
	case runtimev1.TimeGrain_TIME_GRAIN_HOUR:
		return duration.StandardDuration{Hour: 1}
	case runtimev1.TimeGrain_TIME_GRAIN_DAY:
		return duration.StandardDuration{Day: 1}
	case runtimev1.TimeGrain_TIME_GRAIN_WEEK:
		return duration.StandardDuration{Week: 1}
	case runtimev1.TimeGrain_TIME_GRAIN_MONTH:
		return duration.StandardDuration{Month: 1}
	case runtimev1.TimeGrain_TIME_GRAIN_QUARTER:
		return duration.StandardDuration{Month: 3}
	case runtimev1.TimeGrain_TIME_GRAIN_YEAR:
		return duration.StandardDuration{Year: 1}
	}

	return duration.InfDuration{}
}

// func ResolveToTime(t *timestamppb.Timestamp, timeZone string) (time.Time, error) {
// 	if timeZone != "" {
// 		var err error
// 		tz, err := time.LoadLocation(timeZone)
// 		if err != nil {
// 			return time.Time{}, fmt.Errorf("invalid time_range.time_zone %q: %w", timeZone, err)
// 		}
// 		return t.AsTime().In(tz), nil
// 	} else {
// 		return t.AsTime(), nil
// 	}
// }

func ResolveTimeRange(tr *runtimev1.TimeRange, mv *runtimev1.MetricsViewSpec) (time.Time, time.Time, error) {
	tz := time.UTC

	if tr.TimeZone != "" {
		var err error
		tz, err = time.LoadLocation(tr.TimeZone)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid time_range.time_zone %q: %w", tr.TimeZone, err)
		}
	}

	var start, end time.Time
	if tr.Start != nil {
		start = tr.Start.AsTime().In(tz)
	}
	if tr.End != nil {
		end = tr.End.AsTime().In(tz)
	}

	isISO := false

	if tr.IsoDuration != "" {
		if !start.IsZero() && !end.IsZero() {
			return time.Time{}, time.Time{}, fmt.Errorf("only two of time_range.{start,end,iso_duration} can be specified")
		}

		d, err := duration.ParseISO8601(tr.IsoDuration)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid iso_duration %q: %w", tr.IsoDuration, err)
		}

		if !start.IsZero() {
			end = d.Add(start)
		} else if !end.IsZero() {
			start = d.Sub(end)
		} else {
			return time.Time{}, time.Time{}, fmt.Errorf("one of time_range.{start,end} must be specified with time_range.iso_duration")
		}

		isISO = true
	}

	if tr.IsoOffset != "" {
		d, err := duration.ParseISO8601(tr.IsoOffset)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid iso_offset %q: %w", tr.IsoOffset, err)
		}

		if !start.IsZero() {
			start = d.Sub(start)
		}
		if !end.IsZero() {
			end = d.Sub(end)
		}

		isISO = true
	}

	// Only modify the start and end if ISO duration or offset was sent.
	// This is to maintain backwards compatibility for calls from the UI.
	if isISO {
		fdow := int(mv.FirstDayOfWeek)
		if mv.FirstDayOfWeek > 7 || mv.FirstDayOfWeek <= 0 {
			fdow = 1
		}
		fmoy := int(mv.FirstMonthOfYear)
		if mv.FirstMonthOfYear > 12 || mv.FirstMonthOfYear <= 0 {
			fmoy = 1
		}
		if !start.IsZero() {
			start = timeutil.TruncateTime(start, convTimeGrain(tr.RoundToGrain), tz, fdow, fmoy)
		}
		if !end.IsZero() {
			end = timeutil.TruncateTime(end, convTimeGrain(tr.RoundToGrain), tz, fdow, fmoy)
		}
	}

	return start, end, nil
}
