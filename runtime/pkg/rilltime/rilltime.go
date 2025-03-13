package rilltime

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/alecthomas/participle/v2"
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
		// this needs to be after Now and Latest to match to them
		{"Grain", `[sSmhHdDwWqQMyY]`},
		// this has to be at the end
		{"TimeZone", `{.+?}`},
		{"ISOTime", isoTimePattern},
		{"AnchorPrefix", `[+\-<>]`},
		{"Current", "[~]"},
		{"Number", `\d+`},
		{"To", `(?i)to`},
		{"By", `(?i)by`},
		{"Of", `(?i)of`},
		// needed for misc. direct character references used
		{"Punct", `[-[!@#$%^&*()+_={}\|:;"'<,>.?/]]`},
		{"Whitespace", `[ \t\n\r]+`},
	})
	daxNotations = map[string]string{
		// Mapping for our old rill-<DAX> syntax
		"TD":  "D~",
		"WTD": "W~",
		"MTD": "M~",
		"QTD": "Q~",
		"YTD": "Y~",
		"PDC": "D",
		"PWC": "W",
		"PMC": "M",
		"PQC": "Q",
		"PYC": "Y",
		// TODO: previous period is contextual. should be handled in UI
		"PP": "",
		"PD": "-1D to D",
		"PW": "-1W to W",
		"PM": "-1M to M",
		"PQ": "-1Q to Q",
		"PY": "-1Y to Y",
	}
	rillTimeParser = participle.MustBuild[Expression](
		participle.Lexer(rillTimeLexer),
		participle.Elide("Whitespace"),
	)
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
		timeutil.TimeGrainDay:     timeutil.TimeGrainWeek,
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
		timeutil.TimeGrainMonth:   timeutil.TimeGrainWeek,
		timeutil.TimeGrainQuarter: timeutil.TimeGrainMonth,
		timeutil.TimeGrainYear:    timeutil.TimeGrainMonth,
	}
)

type Expression struct {
	From           *Link            `parser:"@@"`
	To             *Link            `parser:"(To @@)?"`
	Grain          *string          `parser:"(By @Grain)?"`
	AnchorOverride *HardcodedAnchor `parser:"('@' @@)?"`
	TimeZone       *string          `parser:"('@' @TimeZone)?"`

	isNewFormat bool
	timeZone    *time.Location
	isoDuration *duration.StandardDuration
}

// Link represents a link of grains specifying the customisable anchors.
// EG: 7d of -1M : The 7day period of last month. 7day is relative to watermark unless something else is specified.
type Link struct {
	Parts []*LinkPart `parser:"@@ (Of @@)*"`
}

type LinkPart struct {
	Pos lexer.Position

	Ordinal         *Ordinal         `parser:"( @@"`
	Anchor          *TimeAnchor      `parser:"| @@"`
	AbsoluteTime    *AbsoluteTime    `parser:"| @@"`
	HardcodedAnchor *HardcodedAnchor `parser:"| @@)"`
}

type HardcodedAnchor struct {
	Earliest  bool `parser:"( @Earliest"`
	Now       bool `parser:"| @Now"`
	Latest    bool `parser:"| @Latest"`
	Watermark bool `parser:"| @Watermark)"`
}

type TimeAnchor struct {
	Pos lexer.Position

	Prefix    *string `parser:"@AnchorPrefix?"`
	Num       *int    `parser:"@Number?"`
	Grain     string  `parser:"@Grain"`
	IsCurrent bool    `parser:"@Current?"`
}

// Ordinal represent a particular sequence of a grain in the next order grain.
// EG: W2 - week 2 of the month.
//     M5 - month 5 of the year.
type Ordinal struct {
	Grain string `parser:"@Grain"`
	Num   int    `parser:"@Number"`
}

type AbsoluteTime struct {
	ISO    string `parser:"@ISOTime"`
	year   int
	month  int
	week   int
	day    int
	hour   int
	minute int
	second int
	tg     timeutil.TimeGrain
}

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

func Parse(from string, parseOpts ParseOptions) (*Expression, error) {
	var rt *Expression
	var err error

	rt, err = parseISO(from, parseOpts)
	if err != nil {
		return nil, err
	}

	if rt == nil {
		rt, err = rillTimeParser.ParseString("", from)
		if err != nil {
			return nil, err
		}
		rt.isNewFormat = true
	}

	// TODO: validation per link and link-part
	if rt.From != nil {
		for _, part := range rt.From.Parts {
			if part.AbsoluteTime != nil {
				err = part.AbsoluteTime.parse()
				if err != nil {
					return nil, err
				}
			}
		}
	} else if rt.isoDuration == nil {
		return nil, errors.New("invalid range: missing from")
	}

	if rt.To != nil {
		for _, part := range rt.To.Parts {
			if part.AbsoluteTime != nil {
				err = part.AbsoluteTime.parse()
				if err != nil {
					return nil, err
				}
			}
		}
	}

	rt.timeZone = time.UTC
	if parseOpts.TimeZoneOverride != nil {
		rt.timeZone = parseOpts.TimeZoneOverride
	} else if rt.TimeZone != nil {
		rt.timeZone, err = time.LoadLocation(strings.Trim(*rt.TimeZone, "{}"))
	} else if parseOpts.DefaultTimeZone != nil {
		rt.timeZone = parseOpts.DefaultTimeZone
	}

	return rt, nil
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

func (e *Expression) Eval(evalOpts EvalOptions) (time.Time, time.Time, timeutil.TimeGrain) {
	anchor := evalOpts.Watermark
	if e.AnchorOverride != nil {
		anchor = e.AnchorOverride.anchor(evalOpts)
	}

	if e.isoDuration != nil {
		// handling for old iso format
		start := e.isoDuration.Sub(evalOpts.MaxTime.In(e.timeZone))
		end := anchor
		tg := timeutil.TimeGrainUnspecified
		if e.Grain != nil {
			tg = grainMap[*e.Grain]
			start = timeutil.TruncateTime(start, tg, e.timeZone, evalOpts.FirstDay, evalOpts.FirstMonth)
			end = timeutil.TruncateTime(anchor, tg, e.timeZone, evalOpts.FirstDay, evalOpts.FirstMonth)
		}

		return start, end, tg
	}

	start, end, tg := e.From.time(evalOpts, anchor, anchor, e.timeZone)
	if e.To != nil {
		_, end, _ = e.To.time(evalOpts, anchor, anchor, e.timeZone)
	}

	if e.Grain != nil {
		tg = grainMap[*e.Grain]
	} else {
		tg = lowerOrderMap[tg]
	}

	return start, end, tg
}

func (l *Link) time(evalOpts EvalOptions, start, end time.Time, tz *time.Location) (time.Time, time.Time, timeutil.TimeGrain) {
	tg := timeutil.TimeGrainUnspecified
	i := len(l.Parts) - 1
	for i >= 0 {
		start, end, tg = l.Parts[i].time(evalOpts, start, end, tz, tg, i == 0)
		i--
	}

	return start, end, tg
}

func (l *LinkPart) time(evalOpts EvalOptions, start, end time.Time, tz *time.Location, tg timeutil.TimeGrain, isFinal bool) (time.Time, time.Time, timeutil.TimeGrain) {
	if l.Anchor != nil {
		return l.Anchor.time(evalOpts, start, end, tz, tg, isFinal)
	} else if l.Ordinal != nil {
		return l.Ordinal.time(evalOpts, start, tz, tg)
	} else if l.AbsoluteTime != nil {
		return l.AbsoluteTime.time(tz, isFinal)
	} else if l.HardcodedAnchor != nil {
		tm := l.HardcodedAnchor.anchor(evalOpts)
		return tm, tm, tg
	}
	return time.Time{}, time.Time{}, tg
}

func (a *HardcodedAnchor) anchor(evalOpts EvalOptions) time.Time {
	if a.Earliest {
		return evalOpts.MinTime
	} else if a.Now {
		return evalOpts.Now
	} else if a.Latest {
		return evalOpts.MaxTime
	} else if a.Watermark {
		return evalOpts.Watermark
	}
	return time.Time{}
}

func (t *TimeAnchor) time(evalOpts EvalOptions, start, end time.Time, tz *time.Location, higherTg timeutil.TimeGrain, isFinal bool) (time.Time, time.Time, timeutil.TimeGrain) {
	num := 1
	if t.Num != nil {
		num = *t.Num
	}

	if t.IsCurrent {
		num--
	}

	curTg := grainMap[t.Grain]
	if higherTg == timeutil.TimeGrainUnspecified {
		higherTg = higherOrderMap[curTg]
	}

	if t.Prefix == nil {
		if num > 0 {
			start = timeutil.OffsetTime(start, curTg, -num)
		}

		start = timeutil.TruncateTime(start, curTg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)

		if !t.IsCurrent {
			end = timeutil.TruncateTime(end, curTg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)
		} else if isFinal {
			end = timeutil.OffsetTime(end, timeutil.TimeGrainSecond, 1)
		}
	} else {
		switch *t.Prefix {
		// -<grain> is used as an offset rather than a range.
		// So we subtract <num> from start and <num-1> from end.
		case "-":
			start = timeutil.OffsetTime(start, curTg, -num)
			if isFinal {
				start = timeutil.TruncateTime(start, curTg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)

				end = timeutil.OffsetTime(end, curTg, -num+1)
				end = timeutil.TruncateTime(end, curTg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)
			} else {
				end = timeutil.OffsetTime(end, curTg, -num)
			}

		// Same with +<grain> is used as an offset.
		// So we add <num-1> to start and <num> to end.
		case "+":
			end = timeutil.OffsetTime(end, curTg, num)
			if isFinal {
				start = timeutil.OffsetTime(start, curTg, num-1)
				start = timeutil.CeilTime(start, curTg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)
				end = timeutil.CeilTime(end, curTg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)
			} else {
				start = timeutil.OffsetTime(start, curTg, num)
			}

		// Anchor the range to the beginning of the higher order grain
		// EG: <4d of M : gives 1st 4 days of the current month regardless of current date.
		case "<":
			start = timeutil.TruncateTime(start, higherTg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)

			end = timeutil.OffsetTime(start, curTg, num)
			end = timeutil.TruncateTime(end, curTg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)

		// Anchor the range to the end of the higher order grain
		// EG: >4d of M : gives last 4 days of the current month regardless of current date.
		case ">":
			end = timeutil.CeilTime(end, higherTg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)

			start = timeutil.OffsetTime(end, curTg, -num)
			start = timeutil.TruncateTime(start, curTg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)
		}
	}

	return start, end, curTg
}

func (o *Ordinal) time(evalOpts EvalOptions, start time.Time, tz *time.Location, higherTg timeutil.TimeGrain) (time.Time, time.Time, timeutil.TimeGrain) {
	curTg := grainMap[o.Grain]
	if higherTg == timeutil.TimeGrainUnspecified {
		higherTg = higherOrderMap[curTg]
	}

	start = timeutil.TruncateTime(start, higherTg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)

	offset := o.Num - 1
	if curTg == timeutil.TimeGrainWeek {
		// https://en.wikipedia.org/wiki/ISO_week_date#First_week
		if start.Weekday() >= 5 {
			offset++
		}

		start = timeutil.TruncateTime(start, timeutil.TimeGrainWeek, tz, evalOpts.FirstDay, evalOpts.FirstMonth)
	}

	start = timeutil.OffsetTime(start, curTg, offset)
	end := timeutil.OffsetTime(start, curTg, 1)

	return start, end, curTg
}

// TODO: reuse code from duration.ParseISO8601
func (a *AbsoluteTime) parse() error {
	match := isoTimeRegex.FindStringSubmatch(a.ISO)

	for i, name := range isoTimeRegex.SubexpNames() {
		part := match[i]
		if i == 0 || name == "" || part == "" {
			continue
		}

		val, err := strconv.Atoi(part)
		if err != nil {
			return err
		}
		switch name {
		case "year":
			a.year = val
			a.tg = timeutil.TimeGrainYear
		case "month":
			a.month = val
			a.tg = timeutil.TimeGrainMonth
		case "week":
			a.week = val
			a.tg = timeutil.TimeGrainWeek
		case "day":
			a.day = val
			a.tg = timeutil.TimeGrainDay
		case "hour":
			a.hour = val
			a.tg = timeutil.TimeGrainHour
		case "minute":
			a.minute = val
			a.tg = timeutil.TimeGrainMinute
		case "second":
			a.second = val
			a.tg = timeutil.TimeGrainSecond
		default:
			return fmt.Errorf("unexpected field %q in duration", name)
		}
	}

	// Since we use this to build a time, month and day cannot be zero
	if a.month == 0 {
		a.month = 1
	}
	if a.day == 0 {
		a.day = 1
	}

	return nil
}

func (a *AbsoluteTime) time(tz *time.Location, isFinal bool) (time.Time, time.Time, timeutil.TimeGrain) {
	start := time.Date(a.year, time.Month(a.month), a.day, a.hour, a.minute, a.second, 0, tz)
	end := start

	if isFinal {
		end = timeutil.OffsetTime(start, a.tg, 1)
	}

	return start, end, a.tg
}

func parseISO(from string, parseOpts ParseOptions) (*Expression, error) {
	// Try parsing for "inf"
	if infPattern.MatchString(from) {
		return &Expression{
			From: &Link{
				Parts: []*LinkPart{
					{HardcodedAnchor: &HardcodedAnchor{Earliest: true}},
				},
			},
			To: &Link{
				Parts: []*LinkPart{
					{HardcodedAnchor: &HardcodedAnchor{Latest: true}},
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

	rt := &Expression{}
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
