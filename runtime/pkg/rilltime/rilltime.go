package rilltime

import (
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
	rillTimeParser = participle.MustBuild[Expression](
		participle.Lexer(rillTimeLexer),
		participle.Elide("Whitespace"),
		participle.UseLookahead(25), // We need this to disambiguate certain cases. Mainly for something like `-4d!`
	)
	daxNotations = map[string]string{
		// Mapping for our old rill-<DAX> syntax
		"TD":  "D^ to D$",
		"WTD": "W^ to D$",
		"MTD": "M^ to D$",
		"QTD": "Q^ to D$",
		"YTD": "Y^ to D$",
		"PDC": "-1D^ to D^",
		"PWC": "-1W^ to W^",
		"PMC": "-1M^ to M^",
		"PQC": "-1Q^ to Q^",
		"PYC": "-1Y^ to Y^",
		// TODO: previous period is contextual. should be handled in UI
		"PP": "",
		"PD": "-1D^ to D^",
		"PW": "-1W^ to W^",
		"PM": "-1M^ to M^",
		"PQ": "-1Q^ to Q^",
		"PY": "-1Y^ to Y^",
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

type Expression struct {
	Interval       *Interval       `parser:"@@"`
	AnchorOverride *AnchorOverride `parser:"(As Of @@)?"`
	Grain          *string         `parser:"(By @Grain)?"`
	TimeZone       *string         `parser:"('@' @TimeZone)?"`

	isNewFormat bool
	timeZone    *time.Location
	isoDuration *duration.StandardDuration
}

type Interval struct {
	AnchoredDuration *AnchoredDurationInterval `parser:"( @@"`
	Ordinal          *OrdinalInterval          `parser:"| @@"`
	StartEnd         *StartEndInterval         `parser:"| @@"`
	Interval         *GrainToInterval          `parser:"| @@"`
	Iso              *IsoInterval              `parser:"| @@)"`
}

// AnchoredDurationInterval anchors a duration either starting or ending at a point in time.
// EG: `2D starting -2D!`
type AnchoredDurationInterval struct {
	Duration    *GrainDuration `parser:"@@"`
	Starting    bool           `parser:"( @Starting"`
	Ending      bool           `parser:"| @Ending)"`
	PointInTime *PointInTime   `parser:"@@"`
}

// OrdinalInterval is an interval formed with a chain of ordinals ended by an interval.
// EG: `W2 of Q2 of -2Y!`
type OrdinalInterval struct {
	Ordinal *OrdinalDuration    `parser:"@@"`
	End     *OrdinalIntervalEnd `parser:"(Of @@)?"`
}

// OrdinalIntervalEnd marks the end of OrdinalInterval with a non-ordinal interval.
type OrdinalIntervalEnd struct {
	Grains   *GrainToInterval  `parser:"( @@"`
	StartEnd *StartEndInterval `parser:"| @@"`
	// `SingleGrain` supports simplified syntax like W1 of Y for getting an ordinal of the current period.
	SingleGrain *string `parser:"| @Grain)"`
}

// StartEndInterval is a basic interval with a start and an end.
type StartEndInterval struct {
	Start *PointInTime `parser:"@@"`
	End   *PointInTime `parser:"To @@"`
}

// GrainToInterval is a convenience syntax to easily convert a grain point in time to an interval. Uses the character `!`.
// EG: Convert -2D to interval using: `-2D!`
type GrainToInterval struct {
	Interval *GrainPointInTime `parser:"@@ Interval"`
}

// IsoInterval is an interval formed by ISO timestamps. Allows for partial timestamps in AbsoluteTime.
type IsoInterval struct {
	Start *AbsoluteTime `parser:"@@"`
	End   *AbsoluteTime `parser:"((To | '/') @@)?"`
}

// AnchorOverride allows overriding the default `watermark` anchor.
type AnchorOverride struct {
	Grain *GrainPointInTime   `parser:"( @@"`
	Label *LabeledPointInTime `parser:"| @@"`
	Abs   *AbsoluteTime       `parser:"| @@)"`
}

type PointInTime struct {
	Ordinal *OrdinalPointInTime `parser:"( @@"`
	Grain   *GrainPointInTime   `parser:"| @@"`
	Labeled *LabeledPointInTime `parser:"| @@)"`
}

type OrdinalPointInTime struct {
	Ordinal *Ordinal         `parser:"@@"`
	Suffix  string           `parser:"@Suffix"`
	Rest    *OrdinalDuration `parser:"@@?"`
}

type GrainPointInTime struct {
	Parts []*GrainPointInTimePart `parser:"@@ @@*"`
}

type GrainPointInTimePart struct {
	Prefix   *string        `parser:"@Prefix?"`
	Duration *GrainDuration `parser:"@@"`
	Snap     *string        `parser:"( Snap @Grain"`
	// Snap by a primary grain and then by week. This allows specifying a time range bucketed by week but snapped by a higher order grain.
	// EG: `Y/YW^` or `Y/YW$` snaps to the beginning of the 1st week of the year or the beginning of the 1st week of next year (to include the last week of the year)
	//     `Y/Y^` or `Y/Y$` instead gives 1st day of the year or 1st day of next year.
	WeekSnapGrain *string `parser:"| Snap @WeekSnapGrain)?"`
	Suffix        *string `parser:"@Suffix?"`
}

type LabeledPointInTime struct {
	Earliest  bool `parser:"( @Earliest"`
	Now       bool `parser:"| @Now"`
	Latest    bool `parser:"| @Latest"`
	Watermark bool `parser:"| @Watermark)"`
}

type OrdinalDuration struct {
	Durations []*OrdinalDurationPart `parser:"@@ (Of @@)*"`
}

type OrdinalDurationPart struct {
	Ordinal       *Ordinal           `parser:"( @@"`
	Snap          *string            `parser:"| @SnapPrefix"`
	GrainDuration *GrainDurationPart `parser:"  @@)"`
}

type Ordinal struct {
	Grain string `parser:"@Grain"`
	Num   int    `parser:"@Number"`
}

type GrainDuration struct {
	Parts []*GrainDurationPart `parser:"@@ @@*"`
}

type GrainDurationPart struct {
	Num   *int   `parser:"@Number?"`
	Grain string `parser:"@Grain"`
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

	err = rt.parse(parseOpts)
	if err != nil {
		return nil, err
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

	start := evalOpts.Watermark
	if e.AnchorOverride != nil {
		start = e.AnchorOverride.eval(evalOpts, start, e.timeZone)
	}

	if e.isoDuration != nil {
		// handling for old iso format
		isoStart := e.isoDuration.Sub(evalOpts.MaxTime.In(e.timeZone))
		isoEnd := start
		tg := timeutil.TimeGrainUnspecified
		if e.Grain != nil {
			tg = grainMap[*e.Grain]
			isoStart = timeutil.TruncateTime(isoStart, tg, e.timeZone, evalOpts.FirstDay, evalOpts.FirstMonth)
			isoEnd = timeutil.TruncateTime(start, tg, e.timeZone, evalOpts.FirstDay, evalOpts.FirstMonth)
		}

		return isoStart, isoEnd, tg
	}

	start, end, tg := e.Interval.eval(evalOpts, start, e.timeZone)

	if e.Grain != nil {
		tg = grainMap[*e.Grain]
	} else {
		tg = getLowerOrderGrain(start, end, tg)
	}

	return start, end, tg
}

func (e *Expression) parse(parseOpts ParseOptions) error {
	e.timeZone = time.UTC
	if parseOpts.TimeZoneOverride != nil {
		e.timeZone = parseOpts.TimeZoneOverride
	} else if e.TimeZone != nil {
		var err error
		e.timeZone, err = time.LoadLocation(strings.Trim(*e.TimeZone, "{}"))
		if err != nil {
			return err
		}
	} else if parseOpts.DefaultTimeZone != nil {
		e.timeZone = parseOpts.DefaultTimeZone
	}

	if e.Interval != nil {
		err := e.Interval.parse()
		if err != nil {
			return err
		}
	}

	if e.AnchorOverride != nil {
		err := e.AnchorOverride.parse()
		if err != nil {
			return err
		}
	}

	return nil
}

/* Intervals */

func (i *Interval) parse() error {
	if i.Iso != nil {
		return i.Iso.parse()
	}
	return nil
}

func (i *Interval) eval(evalOpts EvalOptions, start time.Time, tz *time.Location) (time.Time, time.Time, timeutil.TimeGrain) {
	if i.AnchoredDuration != nil {
		return i.AnchoredDuration.eval(evalOpts, start, tz)
	} else if i.Ordinal != nil {
		return i.Ordinal.eval(evalOpts, start, tz)
	} else if i.StartEnd != nil {
		return i.StartEnd.eval(evalOpts, start, tz)
	} else if i.Interval != nil {
		return i.Interval.eval(evalOpts, start, tz)
	} else if i.Iso != nil {
		return i.Iso.eval(tz)
	}
	return start, start, timeutil.TimeGrainUnspecified
}

func (o *AnchoredDurationInterval) eval(evalOpts EvalOptions, tm time.Time, tz *time.Location) (time.Time, time.Time, timeutil.TimeGrain) {
	// Apply the point in time to the `tm` argument. Duration will be anchored to this.
	tm, _ = o.PointInTime.eval(evalOpts, tm, tz)

	var start, end time.Time
	tg := timeutil.TimeGrainUnspecified
	if o.Starting {
		// Starting from the point in time, offset the duration in the positive direction.
		start = tm
		end, tg = o.Duration.offset(tm, 1)
	} else if o.Ending {
		// Starting from the point in time, offset the duration in the negative direction.
		start, tg = o.Duration.offset(tm, -1)
		end = tm
	}

	return start, end, tg
}

func (o *OrdinalInterval) eval(evalOpts EvalOptions, start time.Time, tz *time.Location) (time.Time, time.Time, timeutil.TimeGrain) {
	end := start
	if o.End != nil {
		start, end, _ = o.End.eval(evalOpts, start, tz)
	}

	start, end, tg := o.Ordinal.eval(evalOpts, start, end, tz)

	return start, end, tg
}

func (o *OrdinalIntervalEnd) eval(evalOpts EvalOptions, start time.Time, tz *time.Location) (time.Time, time.Time, timeutil.TimeGrain) {
	if o.Grains != nil {
		return o.Grains.eval(evalOpts, start, tz)
	} else if o.StartEnd != nil {
		return o.StartEnd.eval(evalOpts, start, tz)
	} else if o.SingleGrain != nil {
		tg := grainMap[*o.SingleGrain]

		end := timeutil.OffsetTime(start, tg, 1)
		end = timeutil.TruncateTime(start, tg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)

		start = truncateWithCorrection(start, tg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)
		return start, end, tg
	}
	return start, start, timeutil.TimeGrainUnspecified
}

func (o *StartEndInterval) eval(evalOpts EvalOptions, tm time.Time, tz *time.Location) (time.Time, time.Time, timeutil.TimeGrain) {
	start, startTg := o.Start.eval(evalOpts, tm, tz)
	end, endTg := o.End.eval(evalOpts, tm, tz)
	// Correction for labeled ends points. We need to add +1ms to make sure the final point is included as our end time is exclusive.
	if o.End.Labeled != nil {
		end = timeutil.OffsetTime(end, timeutil.TimeGrainMillisecond, 1)
	}

	tg := endTg
	if endTg == timeutil.TimeGrainUnspecified || startTg > endTg {
		tg = startTg
	}

	return start, end, tg
}

func (o *GrainToInterval) eval(evalOpts EvalOptions, start time.Time, tz *time.Location) (time.Time, time.Time, timeutil.TimeGrain) {
	if len(o.Interval.Parts) == 0 {
		return start, start, timeutil.TimeGrainUnspecified
	}

	start, tg := o.Interval.eval(evalOpts, start, tz)

	end := timeutil.OffsetTime(start, tg, 1)
	end = truncateWithCorrection(end, tg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)

	start = truncateWithCorrection(start, tg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)

	return start, end, tg
}

func (i *IsoInterval) parse() error {
	err := i.Start.parse()
	if err != nil {
		return err
	}

	if i.End != nil {
		return i.End.parse()
	}

	return nil
}

func (i *IsoInterval) eval(tz *time.Location) (time.Time, time.Time, timeutil.TimeGrain) {
	start, end, tg := i.Start.eval(tz)
	if i.End != nil {
		end, _, _ = i.End.eval(tz)
	}
	return start, end, tg
}

func (a *AnchorOverride) parse() error {
	if a.Abs != nil {
		return a.Abs.parse()
	}
	return nil
}

func (a *AnchorOverride) eval(evalOpts EvalOptions, tm time.Time, tz *time.Location) time.Time {
	if a.Grain != nil {
		tm, _ = a.Grain.eval(evalOpts, tm, tz)
	} else if a.Label != nil {
		tm = a.Label.eval(evalOpts)
	} else if a.Abs != nil {
		tm, _, _ = a.Abs.eval(tz)
	}

	return tm
}

/* Points in time */

func (p *PointInTime) eval(evalOpts EvalOptions, start time.Time, tz *time.Location) (time.Time, timeutil.TimeGrain) {
	if p.Ordinal != nil {
		return p.Ordinal.eval(evalOpts, start, tz)
	} else if p.Grain != nil {
		return p.Grain.eval(evalOpts, start, tz)
	} else if p.Labeled != nil {
		return p.Labeled.eval(evalOpts), timeutil.TimeGrainUnspecified
	}
	return start, timeutil.TimeGrainUnspecified
}

func (o *OrdinalPointInTime) eval(evalOpts EvalOptions, start time.Time, tz *time.Location) (time.Time, timeutil.TimeGrain) {
	if o.Rest != nil {
		start, _, _ = o.Rest.eval(evalOpts, start, start, tz)
	} else {
		tg := higherOrderMap[grainMap[o.Ordinal.Grain]]
		start = truncateWithCorrection(start, tg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)
	}

	start, end, tg := o.Ordinal.eval(evalOpts, start, tz)

	if o.Suffix == "$" {
		start = end
	}

	return start, tg
}

func (g *GrainPointInTime) eval(evalOpts EvalOptions, start time.Time, tz *time.Location) (time.Time, timeutil.TimeGrain) {
	tg := timeutil.TimeGrainUnspecified
	for _, part := range g.Parts {
		start, tg = part.eval(evalOpts, start, tz)
	}
	return start, tg
}

func (g *GrainPointInTimePart) eval(evalOpts EvalOptions, start time.Time, tz *time.Location) (time.Time, timeutil.TimeGrain) {
	dir := -1
	if g.Prefix != nil && *g.Prefix == "+" {
		dir = 1
	}
	// Offset the time based on duration. Direction is specified here rather in Duration.
	tm, tg := g.Duration.offset(start, dir)

	// If there is no suffix specified, we do not snap to start or end of grain and just return the offset duration.
	if g.Suffix == nil {
		return tm, tg
	}

	secondarySnap := timeutil.TimeGrainUnspecified

	if g.Snap != nil {
		// If the snap grain is overridden, use that over the duration's grain.
		tg = grainMap[*g.Snap]
	} else if g.WeekSnapGrain != nil {
		// WeekSnapGrain is a special case, allows snap by a grain and then by a week.
		tgs := strings.Split(*g.WeekSnapGrain, "")
		tg = grainMap[tgs[0]]
		secondarySnap = grainMap[tgs[1]]
	}

	// `$` suffix means snap to end. So add 1 to the offset before truncating.
	if *g.Suffix == "$" {
		tm = timeutil.OffsetTime(tm, tg, 1)
	}
	tm = timeutil.TruncateTime(tm, tg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)

	if secondarySnap != timeutil.TimeGrainUnspecified {
		// If there is a secondary snap, then apply it after the primary snap has happened.
		// These need week correction since that is the primary goal of this syntax.
		if *g.Suffix == "$" {
			tm = ceilWithCorrection(tm, secondarySnap, tz, evalOpts.FirstDay, evalOpts.FirstMonth)
		} else {
			tm = truncateWithCorrection(tm, secondarySnap, tz, evalOpts.FirstDay, evalOpts.FirstMonth)
		}
	}

	return tm, tg
}

func (l *LabeledPointInTime) eval(evalOpts EvalOptions) time.Time {
	if l.Earliest {
		return evalOpts.MinTime
	} else if l.Now {
		return evalOpts.Now
	} else if l.Latest {
		return evalOpts.MaxTime
	} else if l.Watermark {
		return evalOpts.Watermark
	}
	return time.Time{}
}

/* Durations */

func (o *OrdinalDuration) eval(evalOpts EvalOptions, start, end time.Time, tz *time.Location) (time.Time, time.Time, timeutil.TimeGrain) {
	tg := timeutil.TimeGrainUnspecified

	i := len(o.Durations) - 1
	for i >= 0 {
		start, end, tg = o.Durations[i].eval(evalOpts, start, end, tz)
		i--
	}

	return start, end, tg
}

func (o *OrdinalDurationPart) eval(evalOpts EvalOptions, start, end time.Time, tz *time.Location) (time.Time, time.Time, timeutil.TimeGrain) {
	if o.Ordinal != nil {
		if start.Equal(end) {
			// Start and will be equal when this is the 1st part of the ordinal chain. So truncate to the higher order grain to get the correct ordinal.
			// EG: W1 as of -1Y should be W1 of the month (higher order grain for week) exactly 1 year ago.
			//     W1 of year would need explicit syntax like W1 of -1Y! (-1Y! would be `OrdinalIntervalEnd` and truncate would be handled there)
			tg := higherOrderMap[grainMap[o.Ordinal.Grain]]
			start = truncateWithCorrection(start, tg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)
		}

		return o.Ordinal.eval(evalOpts, start, tz)
	}

	if o.Snap == nil || o.GrainDuration == nil {
		return time.Time{}, time.Time{}, timeutil.TimeGrainUnspecified
	}

	tg := grainMap[o.GrainDuration.Grain]
	if *o.Snap == "<" {
		// Anchor the range to the beginning of the higher order start
		// EG: <4d of M : gives 1st 4 days of the current month regardless of current date.
		start = truncateWithCorrection(start, tg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)
		end, _ = o.GrainDuration.offset(start, 1)
	} else {
		// Anchor the range to the end of the higher order end
		// EG: >4d of M : gives last 4 days of the current month regardless of current date.
		end = ceilWithCorrection(end, tg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)
		start, _ = o.GrainDuration.offset(end, -1)
	}
	return start, end, tg
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

func (a *AbsoluteTime) eval(tz *time.Location) (time.Time, time.Time, timeutil.TimeGrain) {
	// Since we use this to build a time, month and day cannot be zero, hence the max(1, xx)
	absStart := time.Date(a.year, time.Month(max(1, a.month)), max(1, a.day), a.hour, a.minute, a.second, 0, tz)
	absEnd := timeutil.OffsetTime(absStart, a.tg, 1)

	return absStart, absEnd, a.tg
}

func (o *Ordinal) eval(evalOpts EvalOptions, start time.Time, tz *time.Location) (time.Time, time.Time, timeutil.TimeGrain) {
	tg := grainMap[o.Grain]
	offset := o.Num - 1

	start = timeutil.OffsetTime(start, tg, offset)
	start = truncateWithCorrection(start, tg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)

	end := timeutil.OffsetTime(start, tg, 1)

	return start, end, tg
}

func (g *GrainDuration) offset(tm time.Time, dir int) (time.Time, timeutil.TimeGrain) {
	tg := timeutil.TimeGrainUnspecified
	i := len(g.Parts) - 1
	for i >= 0 {
		tm, tg = g.Parts[i].offset(tm, dir)
		i--
	}
	return tm, tg
}

func (g *GrainDurationPart) offset(tm time.Time, dir int) (time.Time, timeutil.TimeGrain) {
	tg := grainMap[g.Grain]
	offset := 0
	if g.Num != nil {
		offset = *g.Num
	}
	offset *= dir

	return timeutil.OffsetTime(tm, tg, offset), tg
}

func parseISO(from string, parseOpts ParseOptions) (*Expression, error) {
	// Try parsing for "inf"
	if infPattern.MatchString(from) {
		return &Expression{
			Interval: &Interval{
				StartEnd: &StartEndInterval{
					Start: &PointInTime{
						Labeled: &LabeledPointInTime{
							Earliest: true,
						},
					},
					End: &PointInTime{
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
	weekday := (7 + int(tm.Weekday()) - (firstDay - 1)) % 7
	newTm := timeutil.TruncateTime(tm, tg, tz, firstDay, firstMonth)
	if newTm.Equal(tm) {
		return newTm
	}

	if tg == timeutil.TimeGrainWeek {
		if weekday == 0 {
			// time package's week starts on sunday
			weekday = 7
		}
		if weekday >= 5 {
			newTm = timeutil.OffsetTime(newTm, tg, 1)
		}
	}

	return newTm
}

// ceilWithCorrection ceils time by a grain but corrects for https://en.wikipedia.org/wiki/ISO_week_date#First_week
func ceilWithCorrection(tm time.Time, tg timeutil.TimeGrain, tz *time.Location, firstDay, firstMonth int) time.Time {
	weekday := (7 + int(tm.Weekday()) - (firstDay - 1)) % 7
	newTm := timeutil.TruncateTime(tm, tg, tz, firstDay, firstMonth)
	if newTm.Equal(tm) {
		return newTm
	}

	newTm = timeutil.OffsetTime(newTm, tg, 1)

	if tg == timeutil.TimeGrainWeek {
		if weekday == 0 {
			// time package's week starts on sunday
			weekday = 7
		}
		if weekday < 5 {
			newTm = timeutil.OffsetTime(newTm, tg, -1)
		}
	}

	return newTm
}

// getLowerOrderGrain returns the lowest grain where 2 periods can fit between start and end. Uses lowerOrderMap to get the lower grain.
func getLowerOrderGrain(start, end time.Time, tg timeutil.TimeGrain) timeutil.TimeGrain {
	for tg > timeutil.TimeGrainMillisecond {
		twoLower := timeutil.OffsetTime(end, tg, -2)
		// if start < end - 2*grain, then we can return the grain.
		if start.Before(twoLower) || start.Equal(twoLower) {
			return tg
		}
		// else check the lower order grain
		tg = lowerOrderMap[tg]
	}
	return tg
}
