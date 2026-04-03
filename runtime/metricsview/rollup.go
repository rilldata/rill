package metricsview

import (
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

// grainOrder defines the numeric ordering of grains for derivability checks.
// Two branches diverge from day: day->week and day->month->quarter->year.
var grainOrder = map[runtimev1.TimeGrain]int{
	runtimev1.TimeGrain_TIME_GRAIN_MILLISECOND: 0,
	runtimev1.TimeGrain_TIME_GRAIN_SECOND:      1,
	runtimev1.TimeGrain_TIME_GRAIN_MINUTE:      2,
	runtimev1.TimeGrain_TIME_GRAIN_HOUR:        3,
	runtimev1.TimeGrain_TIME_GRAIN_DAY:         4,
	runtimev1.TimeGrain_TIME_GRAIN_WEEK:        5,
	runtimev1.TimeGrain_TIME_GRAIN_MONTH:       6,
	runtimev1.TimeGrain_TIME_GRAIN_QUARTER:     7,
	runtimev1.TimeGrain_TIME_GRAIN_YEAR:        8,
}

// grainBranch assigns each grain to a branch.
// Sub-day grains and day are on branch 0.
// Week is on branch 1 (diverges from day).
// Month, quarter, year are on branch 2 (diverges from day).
const (
	branchSubDay   = 0
	branchWeek     = 1
	branchCalendar = 2
)

var grainBranch = map[runtimev1.TimeGrain]int{
	runtimev1.TimeGrain_TIME_GRAIN_MILLISECOND: branchSubDay,
	runtimev1.TimeGrain_TIME_GRAIN_SECOND:      branchSubDay,
	runtimev1.TimeGrain_TIME_GRAIN_MINUTE:      branchSubDay,
	runtimev1.TimeGrain_TIME_GRAIN_HOUR:        branchSubDay,
	runtimev1.TimeGrain_TIME_GRAIN_DAY:         branchSubDay, // day is the common ancestor
	runtimev1.TimeGrain_TIME_GRAIN_WEEK:        branchWeek,
	runtimev1.TimeGrain_TIME_GRAIN_MONTH:       branchCalendar,
	runtimev1.TimeGrain_TIME_GRAIN_QUARTER:     branchCalendar,
	runtimev1.TimeGrain_TIME_GRAIN_YEAR:        branchCalendar,
}

// GrainDerivableFrom returns true if queryGrain can be computed by
// re-aggregating data stored at rollupGrain.
//
// The grain hierarchy has two branches diverging from day:
//
//	ms -> s -> min -> hour -> day -> week
//	                             \-> month -> quarter -> year
//
// A coarser grain is derivable from a finer grain only if they are on the
// same branch (or one is sub-day/day and the other is on a branch rooted at day).
// For example, month is derivable from day, but not from week.
func GrainDerivableFrom(queryGrain, rollupGrain runtimev1.TimeGrain) bool {
	if queryGrain == runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED || rollupGrain == runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED {
		return false
	}
	if queryGrain == rollupGrain {
		return true
	}

	qOrder := grainOrder[queryGrain]
	rOrder := grainOrder[rollupGrain]

	// Query grain must be coarser (higher order) than rollup grain
	if qOrder <= rOrder {
		return false
	}

	qBranch := grainBranch[queryGrain]
	rBranch := grainBranch[rollupGrain]

	// Same branch: always derivable (finer -> coarser)
	if qBranch == rBranch {
		return true
	}

	// Sub-day/day (branch 0) can feed into either branch
	if rBranch == branchSubDay {
		return true
	}

	// Cross-branch (week -> month or month -> week): not derivable
	return false
}

// TimeRangeAligned returns true if start and end are aligned to the boundaries
// of the given grain. For sub-day grains, alignment is checked in UTC. For
// day and coarser, alignment is checked in the given timezone.
// For week grain, firstDayOfWeek (1=Monday, 7=Sunday) is used.
func TimeRangeAligned(start, end time.Time, grain runtimev1.TimeGrain, tz *time.Location, firstDayOfWeek uint32) bool {
	if tz == nil {
		tz = time.UTC
	}
	return isAligned(start, grain, tz, firstDayOfWeek) && isAligned(end, grain, tz, firstDayOfWeek)
}

func isAligned(t time.Time, grain runtimev1.TimeGrain, tz *time.Location, firstDayOfWeek uint32) bool {
	if t.IsZero() {
		return true
	}
	switch grain {
	case runtimev1.TimeGrain_TIME_GRAIN_MILLISECOND:
		// Millisecond: microseconds and nanoseconds must be zero
		return t.Nanosecond()%1_000_000 == 0
	case runtimev1.TimeGrain_TIME_GRAIN_SECOND:
		return t.Nanosecond() == 0
	case runtimev1.TimeGrain_TIME_GRAIN_MINUTE:
		return t.Second() == 0 && t.Nanosecond() == 0
	case runtimev1.TimeGrain_TIME_GRAIN_HOUR:
		return t.Minute() == 0 && t.Second() == 0 && t.Nanosecond() == 0
	case runtimev1.TimeGrain_TIME_GRAIN_DAY:
		lt := t.In(tz)
		return lt.Hour() == 0 && lt.Minute() == 0 && lt.Second() == 0 && lt.Nanosecond() == 0
	case runtimev1.TimeGrain_TIME_GRAIN_WEEK:
		lt := t.In(tz)
		if lt.Hour() != 0 || lt.Minute() != 0 || lt.Second() != 0 || lt.Nanosecond() != 0 {
			return false
		}
		// Check weekday alignment
		fdow := int(firstDayOfWeek)
		if fdow < 1 || fdow > 7 {
			fdow = 1 // default to Monday
		}
		// Go: Sunday=0, Monday=1, ..., Saturday=6
		// ISO: Monday=1, ..., Sunday=7
		goWeekday := int(lt.Weekday())
		if goWeekday == 0 {
			goWeekday = 7 // Sunday=7
		}
		return goWeekday == fdow
	case runtimev1.TimeGrain_TIME_GRAIN_MONTH:
		lt := t.In(tz)
		return lt.Day() == 1 && lt.Hour() == 0 && lt.Minute() == 0 && lt.Second() == 0 && lt.Nanosecond() == 0
	case runtimev1.TimeGrain_TIME_GRAIN_QUARTER:
		lt := t.In(tz)
		return lt.Day() == 1 && (lt.Month()-1)%3 == 0 && lt.Hour() == 0 && lt.Minute() == 0 && lt.Second() == 0 && lt.Nanosecond() == 0
	case runtimev1.TimeGrain_TIME_GRAIN_YEAR:
		lt := t.In(tz)
		return lt.Month() == 1 && lt.Day() == 1 && lt.Hour() == 0 && lt.Minute() == 0 && lt.Second() == 0 && lt.Nanosecond() == 0
	default:
		return false
	}
}
