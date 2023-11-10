package queries

import (
	"fmt"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/duration"

	// Load IANA time zone data
	_ "time/tzdata"
)

func TruncateTime(start time.Time, tg runtimev1.TimeGrain, tz *time.Location, firstDay, firstMonth int) time.Time {
	switch tg {
	case runtimev1.TimeGrain_TIME_GRAIN_MILLISECOND:
		return start.Truncate(time.Millisecond)
	case runtimev1.TimeGrain_TIME_GRAIN_SECOND:
		return start.Truncate(time.Second)
	case runtimev1.TimeGrain_TIME_GRAIN_MINUTE:
		return start.Truncate(time.Minute)
	case runtimev1.TimeGrain_TIME_GRAIN_HOUR:
		start = start.In(tz)
		start = time.Date(start.Year(), start.Month(), start.Day(), start.Hour(), 0, 0, 0, tz)
		return start.In(time.UTC)
	case runtimev1.TimeGrain_TIME_GRAIN_DAY:
		start = start.In(tz)
		start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, tz)
		return start.In(time.UTC)
	case runtimev1.TimeGrain_TIME_GRAIN_WEEK:
		start = start.In(tz)
		weekday := int(start.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		if firstDay < 1 {
			firstDay = 1
		}
		if firstDay > 7 {
			firstDay = 7
		}

		daysToSubtract := -(weekday - firstDay)
		if weekday < firstDay {
			daysToSubtract = -7 + daysToSubtract
		}
		start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, tz)
		start = start.AddDate(0, 0, daysToSubtract)
		return start.In(time.UTC)
	case runtimev1.TimeGrain_TIME_GRAIN_MONTH:
		start = start.In(tz)
		start = time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, tz)
		start = start.In(time.UTC)
		return start
	case runtimev1.TimeGrain_TIME_GRAIN_QUARTER:
		monthsToSubtract := (3 + int(start.Month()) - firstMonth%3) % 3
		start = start.In(tz)
		start = time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, tz)
		start = start.AddDate(0, -monthsToSubtract, 0)
		return start.In(time.UTC)
	case runtimev1.TimeGrain_TIME_GRAIN_YEAR:
		start = start.In(tz)
		year := start.Year()
		if int(start.Month()) < firstMonth {
			year = start.Year() - 1
		}

		start = time.Date(year, time.Month(firstMonth), 1, 0, 0, 0, 0, tz)
		return start.In(time.UTC)
	}

	return start
}

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
			start = d.Add(start)
		}
		if !end.IsZero() {
			end = d.Add(end)
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
			start = TruncateTime(start, tr.RoundToGrain, tz, fdow, fmoy)
		}
		if !end.IsZero() {
			end = TruncateTime(end, tr.RoundToGrain, tz, fdow, fmoy)
		}
	}

	return start, end, nil
}
