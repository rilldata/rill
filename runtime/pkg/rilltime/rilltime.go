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
		{"PeriodToGrain", `[sSmhHdDwWqQMyY]T[sSmhHdDwWqQMyY]`},
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
		"PD": "-1D to D",
		"PW": "-1W to W",
		"PM": "-1M to M",
		"PQ": "-1Q to Q",
		"PY": "-1Y to Y",
	}
	rillTimeParser = participle.MustBuild[Expression](
		participle.Lexer(rillTimeLexer),
		participle.Elide("Whitespace"),
		participle.UseLookahead(2),
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
	From           *Link          `parser:"@@"`
	To             *Link          `parser:"(To @@)?"`
	Grain          *string        `parser:"(By @Grain)?"`
	AnchorOverride *LabeledAnchor `parser:"('@' @@)?"`
	TimeZone       *string        `parser:"('@' @TimeZone)?"`

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
	Ordinal       *Ordinal       `parser:"( @@"`
	PeriodToGrain *PeriodToGrain `parser:"| @@"`
	Anchor        *TimeAnchor    `parser:"| @@"`
	AbsoluteTime  *AbsoluteTime  `parser:"| @@"`
	LabeledAnchor *LabeledAnchor `parser:"| @@)"`
}

type LabeledAnchor struct {
	Earliest  bool `parser:"( @Earliest"`
	Now       bool `parser:"| @Now"`
	Latest    bool `parser:"| @Latest"`
	Watermark bool `parser:"| @Watermark)"`
}

// Ordinal represent a particular sequence of a grain in the next order grain.
// EG: W2 - week 2 of the month.
// EG: M5 - month 5 of the year.
type Ordinal struct {
	Grain string `parser:"@Grain"`
	Num   int    `parser:"@Number"`
}

type PeriodToGrain struct {
	Prefix         *string `parser:"@AnchorPrefix?"`
	Num            *int    `parser:"@Number?"`
	PeriodToGrain  string  `parser:"@PeriodToGrain"`
	IncludeCurrent bool    `parser:"@Current?"`

	from timeutil.TimeGrain
	to   timeutil.TimeGrain
}

type TimeAnchor struct {
	Prefix         *string `parser:"@AnchorPrefix?"`
	Num            *int    `parser:"@Number?"`
	Grain          string  `parser:"@Grain"`
	IncludeCurrent bool    `parser:"@Current?"`
}

type AbsoluteTime struct {
	ISO        string `parser:"@ISOTime"`
	year       int
	month      int
	week       int
	day        int
	hour       int
	minute     int
	second     int
	smallestTg timeutil.TimeGrain
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

	if rt.From != nil {
		err = rt.From.parse()
		if err != nil {
			return nil, err
		}
	} else if rt.isoDuration == nil {
		return nil, errors.New("invalid range: missing from")
	}

	if rt.To != nil {
		err = rt.To.parse()
		if err != nil {
			return nil, err
		}
	}

	rt.timeZone = time.UTC
	if parseOpts.TimeZoneOverride != nil {
		rt.timeZone = parseOpts.TimeZoneOverride
	} else if rt.TimeZone != nil {
		rt.timeZone, err = time.LoadLocation(strings.Trim(*rt.TimeZone, "{}"))
		if err != nil {
			return nil, err
		}
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
	}

	return start, end, tg
}

func (l *Link) parse() error {
	for _, part := range l.Parts {
		err := part.parse()
		if err != nil {
			return err
		}
	}
	return nil
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

func (l *LinkPart) parse() error {
	if l.PeriodToGrain != nil {
		return l.PeriodToGrain.parse()
	} else if l.AbsoluteTime != nil {
		return l.AbsoluteTime.parse()
	}
	return nil
}

func (l *LinkPart) time(evalOpts EvalOptions, start, end time.Time, tz *time.Location, tg timeutil.TimeGrain, isFirstPart bool) (time.Time, time.Time, timeutil.TimeGrain) {
	if l.PeriodToGrain != nil {
		return l.PeriodToGrain.time(evalOpts, start, tz)
	} else if l.Anchor != nil {
		return l.Anchor.time(evalOpts, start, end, tz, tg, isFirstPart)
	} else if l.Ordinal != nil {
		return l.Ordinal.time(evalOpts, start, tz, tg, isFirstPart)
	} else if l.AbsoluteTime != nil {
		return l.AbsoluteTime.time(tz, isFirstPart)
	} else if l.LabeledAnchor != nil {
		tm := l.LabeledAnchor.anchor(evalOpts)
		return tm, tm, tg
	}
	return time.Time{}, time.Time{}, tg
}

func (a *LabeledAnchor) anchor(evalOpts EvalOptions) time.Time {
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

func (p *PeriodToGrain) parse() error {
	grains := strings.Split(p.PeriodToGrain, "T")
	if len(grains) != 2 {
		return fmt.Errorf("invalid period grain format: %s", p.PeriodToGrain)
	}
	p.from = grainMap[grains[0]]
	p.to = grainMap[grains[1]]
	// TODO: from should be smaller than to

	return nil
}

func (p *PeriodToGrain) time(evalOpts EvalOptions, start time.Time, tz *time.Location) (time.Time, time.Time, timeutil.TimeGrain) {
	num := 1
	if p.Num != nil {
		num = *p.Num
	}

	if p.IncludeCurrent {
		num--
	}

	ptgStart := timeutil.TruncateTime(start, p.from, tz, evalOpts.FirstDay, evalOpts.FirstMonth)
	if num > 1 && p.Prefix != nil {
		switch *p.Prefix {
		case "-":
			ptgStart = timeutil.OffsetTime(ptgStart, p.from, -num+1)

		case "+":
			ptgStart = timeutil.OffsetTime(ptgStart, p.from, num-1)
		}
	}

	ptgEnd := timeutil.CeilTime(start, p.to, tz, evalOpts.FirstDay, evalOpts.FirstMonth)

	return ptgStart, ptgEnd, p.to
}

func (t *TimeAnchor) time(evalOpts EvalOptions, start, end time.Time, tz *time.Location, higherTg timeutil.TimeGrain, isFirstPart bool) (time.Time, time.Time, timeutil.TimeGrain) {
	num := 1
	if t.Num != nil {
		num = *t.Num
	}

	if t.IncludeCurrent {
		num--
	}

	curTg := grainMap[t.Grain]
	if higherTg == timeutil.TimeGrainUnspecified {
		higherTg = higherOrderMap[curTg]
	}

	if t.Prefix == nil {
		if !isFirstPart && num == 1 {
			// For anchors not in the 1st part of a link, M & 0M are the same so do not offset by setting num to 0
			num = 0
		}
		if num > 0 {
			start = timeutil.OffsetTime(start, curTg, -num)
		}

		start = timeutil.TruncateTime(start, curTg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)

		if !t.IncludeCurrent {
			end = timeutil.TruncateTime(end, curTg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)
		} else {
			end = timeutil.CeilTime(end, curTg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)
		}
	} else {
		switch *t.Prefix {
		// -<grain> is used as an offset rather than a range.
		// So we subtract <num> from start and <num-1> from end.
		case "-":
			start = timeutil.OffsetTime(start, curTg, -num)
			if isFirstPart {
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
			if isFirstPart {
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

	nextTg := curTg
	// If this is first part in a link and a single unit of grain is requested, return a grain less that the part's grain
	if isFirstPart && num <= 1 {
		nextTg = lowerOrderMap[nextTg]
	}

	return start, end, nextTg
}

func (o *Ordinal) time(evalOpts EvalOptions, start time.Time, tz *time.Location, higherTg timeutil.TimeGrain, isFirstPart bool) (time.Time, time.Time, timeutil.TimeGrain) {
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

	nextTg := curTg
	// If this is first part in a link, return a grain less that the part's grain
	if isFirstPart {
		nextTg = lowerOrderMap[curTg]
	}

	return start, end, nextTg
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
			a.smallestTg = timeutil.TimeGrainYear
		case "month":
			a.month = val
			a.smallestTg = timeutil.TimeGrainMonth
		case "week":
			a.week = val
			a.smallestTg = timeutil.TimeGrainWeek
		case "day":
			a.day = val
			a.smallestTg = timeutil.TimeGrainDay
		case "hour":
			a.hour = val
			a.smallestTg = timeutil.TimeGrainHour
		case "minute":
			a.minute = val
			a.smallestTg = timeutil.TimeGrainMinute
		case "second":
			a.second = val
			a.smallestTg = timeutil.TimeGrainSecond
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

func (a *AbsoluteTime) time(tz *time.Location, isFirstPart bool) (time.Time, time.Time, timeutil.TimeGrain) {
	start := time.Date(a.year, time.Month(a.month), a.day, a.hour, a.minute, a.second, 0, tz)
	end := start

	if isFirstPart {
		end = timeutil.OffsetTime(start, a.smallestTg, 1)
	}

	nextTg := a.smallestTg
	// If this is the first part in the link then return a grain lower than return grain lower than the smallest time grain
	if isFirstPart {
		nextTg = lowerOrderMap[nextTg]
	}

	return start, end, nextTg
}

func parseISO(from string, parseOpts ParseOptions) (*Expression, error) {
	// Try parsing for "inf"
	if infPattern.MatchString(from) {
		return &Expression{
			From: &Link{
				Parts: []*LinkPart{
					{LabeledAnchor: &LabeledAnchor{Earliest: true}},
				},
			},
			To: &Link{
				Parts: []*LinkPart{
					{LabeledAnchor: &LabeledAnchor{Latest: true}},
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
