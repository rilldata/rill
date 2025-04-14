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
		"PD": "-1D to D~",
		"PW": "-1W to W~",
		"PM": "-1M to M~",
		"PQ": "-1Q to Q~",
		"PY": "-1Y to Y~",
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
	Parts       []*LinkPart `parser:"@@ (Of @@)*"`
	anchorParts []linkPartAnchor
}

type LinkPart struct {
	Ordinal       *Ordinal       `parser:"( @@"`
	Anchor        *TimeAnchor    `parser:"| @@"`
	AbsoluteTime  *AbsoluteTime  `parser:"| @@"`
	LabeledAnchor *LabeledAnchor `parser:"| @@)"`
}

type linkPartAnchor interface {
	parse() error
	eval(evalOpts EvalOptions, start, cur, end time.Time, tz *time.Location) (time.Time, time.Time, time.Time)
	// Returns the grain for the part assuming it is first part in the link.
	grain() timeutil.TimeGrain
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

	offset int
	tg     timeutil.TimeGrain
}

type TimeAnchor struct {
	Prefix         *string `parser:"@AnchorPrefix?"`
	Num            *int    `parser:"@Number?"`
	Grain          *string `parser:"( @Grain"`
	PeriodToGrain  *string `parser:"| @PeriodToGrain)"`
	IncludeCurrent bool    `parser:"@Current?"`

	offset   int
	from, tg timeutil.TimeGrain
}

type AbsoluteTime struct {
	ISO string `parser:"@ISOTime"`

	year   int
	month  int
	week   int
	day    int
	hour   int
	minute int
	second int

	tg timeutil.TimeGrain
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
	if evalOpts.FirstDay == 0 {
		evalOpts.FirstDay = 1
	}
	if evalOpts.FirstMonth == 0 {
		evalOpts.FirstMonth = 1
	}

	cur := evalOpts.Watermark
	if e.AnchorOverride != nil {
		cur, _, _ = e.AnchorOverride.eval(evalOpts, time.Time{}, time.Time{}, time.Time{}, e.timeZone)
	}

	if e.isoDuration != nil {
		// handling for old iso format
		start := e.isoDuration.Sub(evalOpts.MaxTime.In(e.timeZone))
		end := cur
		tg := timeutil.TimeGrainUnspecified
		if e.Grain != nil {
			tg = grainMap[*e.Grain]
			start = timeutil.TruncateTime(start, tg, e.timeZone, evalOpts.FirstDay, evalOpts.FirstMonth)
			end = timeutil.TruncateTime(cur, tg, e.timeZone, evalOpts.FirstDay, evalOpts.FirstMonth)
		}

		return start, end, tg
	}

	start, end, tg := e.From.time(evalOpts, cur, e.timeZone)
	if e.To != nil {
		// we take the start of right of "to" as end
		end, _, _ = e.To.time(evalOpts, cur, e.timeZone)
	}

	if e.Grain != nil {
		tg = grainMap[*e.Grain]
	}

	return start, end, tg
}

func (l *Link) parse() error {
	l.anchorParts = make([]linkPartAnchor, len(l.Parts))

	for i, part := range l.Parts {
		lpa, err := part.parse()
		if err != nil {
			return err
		}
		l.anchorParts[i] = lpa
	}

	return nil
}

func (l *Link) time(evalOpts EvalOptions, cur time.Time, tz *time.Location) (time.Time, time.Time, timeutil.TimeGrain) {
	var start time.Time
	var end time.Time

	i := len(l.anchorParts) - 1
	for i >= 0 {
		start, cur, end = l.anchorParts[i].eval(evalOpts, start, cur, end, tz)
		i--
	}

	tg := timeutil.TimeGrainUnspecified
	if len(l.anchorParts) > 0 {
		tg = l.anchorParts[0].grain()
	}

	return start, end, tg
}

func (l *LinkPart) parse() (linkPartAnchor, error) {
	var lpa linkPartAnchor
	if l.Ordinal != nil {
		lpa = l.Ordinal
	} else if l.Anchor != nil {
		lpa = l.Anchor
	} else if l.AbsoluteTime != nil {
		lpa = l.AbsoluteTime
	} else if l.LabeledAnchor != nil {
		lpa = l.LabeledAnchor
	}

	if lpa == nil {
		return nil, fmt.Errorf("invalid link part: atleast one of ordinal, anchor, absolute time or labeled anchor must be specified")
	}

	err := lpa.parse()
	if err != nil {
		return nil, err
	}

	return lpa, nil
}

func (a *LabeledAnchor) parse() error {
	return nil
}

func (a *LabeledAnchor) eval(evalOpts EvalOptions, start, cur, end time.Time, tz *time.Location) (time.Time, time.Time, time.Time) {
	var tm time.Time
	if a.Earliest {
		tm = evalOpts.MinTime
	} else if a.Now {
		tm = evalOpts.Now
	} else if a.Latest {
		tm = evalOpts.MaxTime
	} else if a.Watermark {
		tm = evalOpts.Watermark
	}

	tm = tm.In(tz)

	return tm, tm, tm
}

func (a *LabeledAnchor) grain() timeutil.TimeGrain {
	return timeutil.TimeGrainUnspecified
}

func (o *Ordinal) parse() error {
	o.offset = o.Num - 1

	o.tg = grainMap[o.Grain]

	return nil
}

func (o *Ordinal) eval(evalOpts EvalOptions, start, cur, end time.Time, tz *time.Location) (time.Time, time.Time, time.Time) {
	if start.IsZero() {
		start = timeutil.TruncateTime(cur, higherOrderMap[o.tg], tz, evalOpts.FirstDay, evalOpts.FirstMonth)
	}

	offset := o.Num - 1

	start = timeutil.OffsetTime(start, o.tg, offset)
	start = truncateWithCorrection(start, o.tg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)

	// update cur to match the shifted ordinal
	cur = timeutil.CopyTimeComponentsUntil(start, cur, o.tg)

	updatedEnd := timeutil.OffsetTime(start, o.tg, 1)

	return start, cur, updatedEnd
}

func (o *Ordinal) grain() timeutil.TimeGrain {
	// Return a grain less that the part's grain for ordinals
	return lowerOrderMap[o.tg]
}

func (t *TimeAnchor) parse() error {
	if t.Num != nil {
		t.offset = *t.Num
		if t.offset == 0 {
			// if something like 0D is specified we still offset start by 1
			t.offset = 1
		}
	} else {
		// if something like D is specified we still offset start by 1
		t.offset = 1
	}

	if t.Grain != nil {
		t.tg = grainMap[*t.Grain]
	} else if t.PeriodToGrain != nil {
		grains := strings.Split(*t.PeriodToGrain, "T")
		if len(grains) != 2 {
			return fmt.Errorf("invalid period grain format: %s", *t.PeriodToGrain)
		}
		t.from = grainMap[grains[0]]
		t.tg = grainMap[grains[1]]
	} else {
		return fmt.Errorf("neither grain nor period-to-grain specified")
	}

	return nil
}

func (t *TimeAnchor) eval(evalOpts EvalOptions, start, cur, end time.Time, tz *time.Location) (time.Time, time.Time, time.Time) {
	if start.IsZero() {
		start = timeutil.TruncateTime(cur, higherOrderMap[t.tg], tz, evalOpts.FirstDay, evalOpts.FirstMonth)
	}
	if end.IsZero() {
		end = timeutil.CeilTime(cur, higherOrderMap[t.tg], tz, evalOpts.FirstDay, evalOpts.FirstMonth)
	}

	if t.PeriodToGrain != nil {
		start = timeutil.TruncateTime(cur, t.from, tz, evalOpts.FirstDay, evalOpts.FirstMonth)
		if t.offset > 1 {
			offset := t.offset
			if t.Prefix == nil || *t.Prefix == "-" {
				offset = -offset + 1
			} else if *t.Prefix == "+" {
				offset--
			}
			start = timeutil.OffsetTime(start, t.from, offset)
		}

		// X-to-week should give buckets in week. Should also follow week rules https://en.wikipedia.org/wiki/ISO_week_date#First_week
		start = truncateWithCorrection(start, t.tg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)

		end := timeutil.TruncateTime(cur, t.tg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)
		if t.IncludeCurrent {
			end = timeutil.OffsetTime(end, t.tg, 1)
		}

		return start, cur, end
	}

	if t.Prefix == nil {
		// Without a prefix of either +/- we actually need a time range.
		// EG: 7D is 7 day period.

		end = timeutil.TruncateTime(cur, t.tg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)
		if t.IncludeCurrent {
			end = timeutil.OffsetTime(end, t.tg, 1)
			cur = timeutil.OffsetTime(cur, t.tg, 1)
		}

		start = timeutil.OffsetTime(end, t.tg, -t.offset)

		return start, cur, end
	}

	switch *t.Prefix {
	case "-":
		offset := t.offset
		if t.IncludeCurrent {
			offset--
		}

		start = timeutil.TruncateTime(cur, t.tg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)
		start = timeutil.OffsetTime(start, t.tg, -offset)

		end = timeutil.OffsetTime(start, t.tg, 1)

		cur = timeutil.OffsetTime(cur, t.tg, -offset)

	case "+":
		start = timeutil.TruncateTime(cur, t.tg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)
		start = timeutil.OffsetTime(start, t.tg, t.offset)

		end = timeutil.OffsetTime(start, t.tg, 1)

		cur = timeutil.OffsetTime(cur, t.tg, t.offset)

	case "<":
		// Anchor the range to the beginning of the higher order start
		// EG: <4d of M : gives 1st 4 days of the current month regardless of current date.

		// Anchoring to start should follow week rules https://en.wikipedia.org/wiki/ISO_week_date#First_week
		start = truncateWithCorrection(start, t.tg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)
		end = timeutil.OffsetTime(start, t.tg, t.offset)
		end = timeutil.TruncateTime(end, t.tg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)

		cur = start

	case ">":
		// Anchor the range to the end of the higher order end
		// EG: >4d of M : gives last 4 days of the current month regardless of current date.
		end = timeutil.CeilTime(end, t.tg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)
		start = timeutil.TruncateTime(end, t.tg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)
		start = timeutil.OffsetTime(start, t.tg, -t.offset)

		cur = end
	}

	return start, cur, end
}

func (t *TimeAnchor) grain() timeutil.TimeGrain {
	// If a single unit of grain is requested, return a grain less that the part's grain
	// But only applies to grains and not period-to-grains
	if t.offset <= 1 && t.Grain != nil {
		return lowerOrderMap[t.tg]
	}
	return t.tg
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

	return nil
}

func (a *AbsoluteTime) eval(evalOpts EvalOptions, start, cur, end time.Time, tz *time.Location) (time.Time, time.Time, time.Time) {
	// Since we use this to build a time, month and day cannot be zero, hence the max(1, xx)
	absStart := time.Date(a.year, time.Month(max(1, a.month)), max(1, a.day), a.hour, a.minute, a.second, 0, tz)

	absEnd := timeutil.OffsetTime(absStart, a.tg, 1)

	// update cur to match the absolute time
	absCur := timeutil.CopyTimeComponentsUntil(absStart, cur, a.tg)

	return absStart, absCur, absEnd
}

func (a *AbsoluteTime) grain() timeutil.TimeGrain {
	// Return a grain lower than the smallest time grain
	return lowerOrderMap[a.tg]
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
