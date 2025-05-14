package rilltime

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/alecthomas/participle/v2/lexer"
	"github.com/rilldata/rill/runtime/pkg/duration"
	"github.com/rilldata/rill/runtime/pkg/timeutil"
)

var (
	infPattern      = regexp.MustCompile("^(?i)inf$")
	durationPattern = regexp.MustCompile(`^P((?P<year>\d+)Y)?((?P<month>\d+)M)?((?P<week>\d+)W)?((?P<day>\d+)D)?(T((?P<hour>\d+)H)?((?P<minute>\d+)M)?((?P<second>\d+)S)?)?$`)
	isoTimePattern  = `(?P<year>\d{4})(-(?P<month>\d{2})(-(?P<day>\d{2})(T(?P<hour>\d{2})(:(?P<minute>\d{2})(:(?P<second>\d{2})Z)?)?)?)?)?`
	isoTimeRegex    = regexp.MustCompile(isoTimePattern)
	// nolint:govet // This is suggested usage by the docs.
	rillTimeLexer = lexer.MustSimple([]lexer.SimpleRule{
		{"Earliest", "earliest"},
		{"Now", "now"},
		{"Latest", "latest"},
		{"Watermark", "watermark"},
		{"Starting", "starting"},
		{"Ending", "ending"},
		// this needs to be after Now and Latest to match to them
		{"WeekSnapGrain", `[qQMyY][wW]`},
		{"PeriodToGrain", `[sSmhHdDwWqQMyY]T[sSmhHdDwWqQMyY]`},
		{"Grain", `[sSmhHdDwWqQMyY]`},
		// this has to be at the end
		{"TimeZone", `{.+?}`},
		{"ISOTime", isoTimePattern},
		{"Prefix", `[+\-]`},
		{"Suffix", `[\^\$]`},
		{"SnapPrefix", `[<>]`},
		{"AnchorPrefix", `[+\-<>]`},
		{"Current", "[~]"},
		{"Number", `\d+`},
		{"Snap", `[/]`},
		{"Interval", `[!]`},
		{"To", `(?i)to`},
		{"By", `(?i)by`},
		{"Of", `(?i)of`},
		{"As", `(?i)as`},
		// needed for misc. direct character references used
		{"Punct", `[-[!@#$%^&*()+_={}\|:;"'<,>.?/]]`},
		{"Whitespace", `[ \t]+`},
	})
	daxNotations = map[string]string{
		// Mapping for our old rill-<DAX> syntax
		"TD":  "D~",
		"WTD": "WTD",
		"MTD": "MTD",
		"QTD": "QTD",
		"YTD": "YTD",
		"PDC": "D",
		"PWC": "W",
		"PMC": "M",
		"PQC": "Q",
		"PYC": "Y",
		// TODO: previous period is contextual. should be handled in UI
		"PP": "",
		"PD": "-1D to D~",
		"PW": "-1W to W~",
		"PM": "-1M to M~",
		"PQ": "-1Q to Q~",
		"PY": "-1Y to Y~",
	}
	grainMap = map[string]timeutil.TimeGrain{
		"s": timeutil.TimeGrainSecond,
		"S": timeutil.TimeGrainSecond,
		"m": timeutil.TimeGrainMinute,
		"h": timeutil.TimeGrainHour,
		"H": timeutil.TimeGrainHour,
		"d": timeutil.TimeGrainDay,
		"D": timeutil.TimeGrainDay,
		"w": timeutil.TimeGrainWeek,
		"W": timeutil.TimeGrainWeek,
		"q": timeutil.TimeGrainQuarter,
		"Q": timeutil.TimeGrainQuarter,
		"M": timeutil.TimeGrainMonth,
		"y": timeutil.TimeGrainYear,
		"Y": timeutil.TimeGrainYear,
	}
	higherOrderMap = map[timeutil.TimeGrain]timeutil.TimeGrain{
		timeutil.TimeGrainSecond:  timeutil.TimeGrainMinute,
		timeutil.TimeGrainMinute:  timeutil.TimeGrainHour,
		timeutil.TimeGrainHour:    timeutil.TimeGrainDay,
		timeutil.TimeGrainDay:     timeutil.TimeGrainMonth,
		timeutil.TimeGrainWeek:    timeutil.TimeGrainMonth,
		timeutil.TimeGrainMonth:   timeutil.TimeGrainYear,
		timeutil.TimeGrainQuarter: timeutil.TimeGrainYear,
	}
	lowerOrderMap = map[timeutil.TimeGrain]timeutil.TimeGrain{
		timeutil.TimeGrainSecond:  timeutil.TimeGrainMillisecond,
		timeutil.TimeGrainMinute:  timeutil.TimeGrainSecond,
		timeutil.TimeGrainHour:    timeutil.TimeGrainMinute,
		timeutil.TimeGrainDay:     timeutil.TimeGrainHour,
		timeutil.TimeGrainWeek:    timeutil.TimeGrainDay,
		timeutil.TimeGrainMonth:   timeutil.TimeGrainDay,
		timeutil.TimeGrainQuarter: timeutil.TimeGrainMonth,
		timeutil.TimeGrainYear:    timeutil.TimeGrainMonth,
	}
)

// ParseOptions allows for additional options that could probably not be added to the time range itself
type ParseOptions struct {
	DefaultTimeZone  *time.Location
	TimeZoneOverride *time.Location
}

type EvalOptions struct {
	Now        time.Time
	MinTime    time.Time
	MaxTime    time.Time
	Watermark  time.Time
	FirstDay   int
	FirstMonth int
}

func ParseCompatibility(timeRange, offset string) error {
	isNewFormat := false
	if timeRange != "" {
		// ParseCompatibility is called for time ranges.
		// All parse options should be part of the time range syntax there.
		rt, err := Parse(timeRange, ParseOptions{})
		if err != nil {
			return fmt.Errorf("invalid comparison range %q: %w", timeRange, err)
		}
		isNewFormat = rt.isNewFormat
	}
	if offset != "" {
		if isNewFormat {
			return fmt.Errorf("offset cannot be provided along with rill time range")
		}
		if err := duration.ValidateISO8601(offset, false, false); err != nil {
			return fmt.Errorf("invalid comparison offset %q: %w", offset, err)
		}
	}
	return nil
}

func parseISOFinal(from string, parseOpts ParseOptions) (*ExpressionFinal, error) {
	// Try parsing for "inf"
	if infPattern.MatchString(from) {
		return &ExpressionFinal{
			Interval: &Interval{
				StartEnd: &StartEndInterval{
					Start: &Point{
						Labeled: &LabeledPointInTime{
							Earliest: true,
						},
					},
					End: &Point{
						Labeled: &LabeledPointInTime{
							Latest: true,
						},
					},
				},
			},
		}, nil
	}

	if strings.HasPrefix(from, "rill-") {
		// We are using "rill-" as a prefix to DAX notation so that it doesn't interfere with ISO8601 standard.
		// Pulled from https://www.daxpatterns.com/standard-time-related-calculations/
		rillDur := strings.Replace(from, "rill-", "", 1)
		if t, ok := daxNotations[rillDur]; ok {
			return Parse(t, parseOpts)
		}
	}

	// Parse as a regular ISO8601 duration
	if !durationPattern.MatchString(from) {
		return nil, nil
	}

	rt := &ExpressionFinal{}
	d, err := duration.ParseISO8601(from)
	if err != nil {
		return nil, nil
	}
	sd, ok := d.(duration.StandardDuration)
	if !ok {
		return nil, nil
	}
	rt.isoDuration = &sd
	minGrain := getMinGrain(sd)
	if minGrain != "" {
		rt.Grain = &minGrain
	}

	return rt, nil
}

func getMinGrain(d duration.StandardDuration) string {
	if d.Second != 0 {
		return "s"
	} else if d.Minute != 0 {
		return "m"
	} else if d.Hour != 0 {
		return "h"
	} else if d.Day != 0 {
		return "D"
	} else if d.Week != 0 {
		return "W"
	} else if d.Month != 0 {
		return "M"
	} else if d.Year != 0 {
		return "Y"
	}
	return ""
}

// truncateWithCorrection truncates time by a grain but corrects for https://en.wikipedia.org/wiki/ISO_week_date#First_week
// TODO: will adding this directly to timeutil.TruncateTime break anything?
func truncateWithCorrection(tm time.Time, tg timeutil.TimeGrain, tz *time.Location, firstDay, firstMonth int) time.Time {
	weekday := int(tm.Weekday())
	tm = timeutil.TruncateTime(tm, tg, tz, firstDay, firstMonth)

	if tg == timeutil.TimeGrainWeek {
		if weekday == 0 {
			// time package's week starts on sunday
			weekday = 7
		}
		if weekday >= 5 {
			tm = timeutil.OffsetTime(tm, tg, 1)
		}
	}

	return tm
}

// ceilWithCorrection ceils time by a grain but corrects for https://en.wikipedia.org/wiki/ISO_week_date#First_week
// TODO: will adding this directly to timeutil.CeilTime break anything?
func ceilWithCorrection(tm time.Time, tg timeutil.TimeGrain, tz *time.Location, firstDay, firstMonth int) time.Time {
	weekday := int(tm.Weekday())
	tm = timeutil.CeilTime(tm, tg, tz, firstDay, firstMonth)

	if tg == timeutil.TimeGrainWeek {
		if weekday == 0 {
			// time package's week starts on sunday
			weekday = 7
		}
		if weekday < 5 {
			tm = timeutil.OffsetTime(tm, tg, -1)
		}
	}

	return tm
}
