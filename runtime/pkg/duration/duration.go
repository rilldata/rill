// Package duration copied as it is from github.com/senseyeio/duration
package duration

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

type Duration interface {
	Add(t time.Time) time.Time
	Sub(t time.Time) time.Time
}

// StandardDuration represents an ISO8601 duration with Rill-specific extensions.
// See ParseISO8601 for details.
type StandardDuration struct {
	// If Inf is true, the other components should be ignored
	Inf bool
	// Date component
	Year  int
	Month int
	Week  int
	Day   int
	// Time Component
	Hour   int
	Minute int
	Second int
}

// Regexes used by ParseISO8601
var (
	infPattern         = regexp.MustCompile("^(?i)inf$")
	durationPattern    = regexp.MustCompile(`^P((?P<year>\d+)Y)?((?P<month>\d+)M)?((?P<week>\d+)W)?((?P<day>\d+)D)?(T((?P<hour>\d+)H)?((?P<minute>\d+)M)?((?P<second>\d+)S)?)?$`)
	daxToDateNotations = map[string]runtimev1.TimeGrain{
		// Pulled from https://www.daxpatterns.com/standard-time-related-calculations/
		// Add more here once support in UI has been added
		"MTD": runtimev1.TimeGrain_TIME_GRAIN_MONTH,
		"QTD": runtimev1.TimeGrain_TIME_GRAIN_QUARTER,
		"YTD": runtimev1.TimeGrain_TIME_GRAIN_YEAR,
	}
	daxOffsetNotations = map[string]StandardDuration{
		"PP": {},
		"PM": {Month: 1},
		"PQ": {Month: 3},
		"PY": {Year: 1},
	}
)

// ParseISO8601 parses an ISO8601 duration as well as some Rill-specific extensions.
// (Section 3.7 of the standard supposedly allows extensions that do not interfere with the standard.)
// The only current extension is "inf", for representing an unbounded duration of time.
func ParseISO8601(from string) (Duration, error) {
	// Try parsing for "inf"
	if infPattern.MatchString(from) {
		return StandardDuration{Inf: true}, nil
	}

	rillDur := strings.Replace(from, "rill-", "", 1)
	if a, ok := daxToDateNotations[rillDur]; ok {
		return ToDateDuration{anchor: a}, nil
	}
	if o, ok := daxOffsetNotations[rillDur]; ok {
		return o, nil
	}

	// Parse as a regular ISO8601 duration
	if !durationPattern.MatchString(from) {
		return StandardDuration{}, fmt.Errorf("string %q is not a valid ISO 8601 duration", from)
	}

	var d StandardDuration
	match := durationPattern.FindStringSubmatch(from)
	for i, name := range durationPattern.SubexpNames() {
		part := match[i]
		if i == 0 || name == "" || part == "" {
			continue
		}

		val, err := strconv.Atoi(part)
		if err != nil {
			return StandardDuration{}, err
		}
		switch name {
		case "year":
			d.Year = val
		case "month":
			d.Month = val
		case "week":
			d.Week = val
		case "day":
			d.Day = val
		case "hour":
			d.Hour = val
		case "minute":
			d.Minute = val
		case "second":
			d.Second = val
		default:
			return d, fmt.Errorf("unexpected field %q in duration", name)
		}
	}

	return d, nil
}

// Add adds the duration to a timestamp
func (d StandardDuration) Add(t time.Time) time.Time {
	if d.Inf {
		return time.Time{}
	}

	days := 7*d.Week + d.Day
	t = t.AddDate(d.Year, d.Month, days)

	td := time.Duration(d.Second)*time.Second + time.Duration(d.Minute)*time.Minute + time.Duration(d.Hour)*time.Hour
	return t.Add(td)
}

// Sub subtracts the duration from a timestamp
func (d StandardDuration) Sub(t time.Time) time.Time {
	if d.Inf {
		return time.Time{}
	}

	days := 7*d.Week + d.Day
	t = t.AddDate(-d.Year, -d.Month, -days)

	td := time.Duration(d.Second)*time.Second + time.Duration(d.Minute)*time.Minute + time.Duration(d.Hour)*time.Hour
	return t.Add(-td)
}

type ToDateDuration struct {
	anchor runtimev1.TimeGrain
}

func (d ToDateDuration) Add(t time.Time) time.Time {
	return time.Time{}
}

func (d ToDateDuration) Sub(t time.Time) time.Time {
	return TruncateTime(t, d.anchor, t.Location(), 1, 1) // TODO: get first day and month
}
