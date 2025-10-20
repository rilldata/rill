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
	isoTimePattern  = `(?P<year>\d{4})(-(?P<month>\d{2})(-(?P<day>\d{2})(T(?P<hour>\d{2})(:(?P<minute>\d{2})(:(?P<second>\d{2})(\.((?P<milli>\d{3})|(?P<micro>\d{6})|(?P<nano>\d{9})))?Z)?)?)?)?)?`
	isoTimeRegex    = regexp.MustCompile(isoTimePattern)
	// nolint:govet // This is suggested usage by the docs.
	rillTimeLexer = lexer.MustSimple([]lexer.SimpleRule{
		{"Ref", "ref"},
		{"Earliest", "earliest"},
		{"Now", "now"},
		{"Latest", "latest"},
		{"Watermark", "watermark"},
		{"PreviousPeriod", "(?i)p"},
		{"Offset", `(?i)offset`},
		// this needs to be after Now and Latest to match to them
		{"PeriodToGrain", `[sSmhHdDwWqQMyY]TD`},
		{"Grain", `[sSmhHdDwWqQMyY]`},
		{"ISOTime", isoTimePattern},
		{"Prefix", `[+\-]`},
		{"Number", `\d+`},
		{"Snap", `[/]`},
		// this has to be at the end
		{"TimeZone", `[0-9a-zA-Z/+\-_]{3,}`},
		{"To", `(?i)to`},
		{"By", `(?i)by`},
		{"Of", `(?i)of`},
		{"As", `(?i)as`},
		{"Tz", `(?i)tz`},
		// Separate entry is needed outside of Punct
		{"RangeSeparator", `[,]`},
		// needed for misc. direct character references used
		{"Punct", `[-[!@#$%^&*()+_={}\|:;"'<>.?/]]`},
		{"Whitespace", `[ \t]+`},
	})
	rillTimeParser = participle.MustBuild[Expression](
		participle.Lexer(rillTimeLexer),
		participle.Elide("Whitespace"),
		participle.UseLookahead(2), // Needed to disambiguate `offset -1P` vs `offset -1M`
	)
	daxNotations = map[string]string{
		// Mapping for our old rill-<DAX> syntax
		"TD":  "ref/D to ref as of watermark",
		"WTD": "ref/W to ref as of watermark",
		"MTD": "ref/M to ref as of watermark",
		"QTD": "ref/Q to ref as of watermark",
		"YTD": "ref/Y to ref as of watermark",
		"PDC": "-1D/D to ref/D as of watermark",
		"PWC": "-1W/W to ref/W as of watermark",
		"PMC": "-1M/M to ref/M as of watermark",
		"PQC": "-1Q/Q to ref/Q as of watermark",
		"PYC": "-1Y/Y to ref/Y as of watermark",
		// TODO: previous period is contextual. should be handled in UI
		"PP": "",
		"PD": "-1D/D to ref/D as of watermark",
		"PW": "-1W/W to ref/W as of watermark",
		"PM": "-1M/M to ref/M as of watermark",
		"PQ": "-1Q/Q to ref/Q as of watermark",
		"PY": "-1Y/Y to ref/Y as of watermark",
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
	reverseGrainMap = map[timeutil.TimeGrain]string{
		timeutil.TimeGrainUnspecified: "s",
		timeutil.TimeGrainMillisecond: "s",
		timeutil.TimeGrainSecond:      "s",
		timeutil.TimeGrainMinute:      "m",
		timeutil.TimeGrainHour:        "h",
		timeutil.TimeGrainDay:         "D",
		timeutil.TimeGrainWeek:        "W",
		timeutil.TimeGrainMonth:       "M",
		timeutil.TimeGrainQuarter:     "Q",
		timeutil.TimeGrainYear:        "Y",
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
	Interval        *Interval      `parser:"@@"`
	AnchorOverrides []*PointInTime `parser:"(As Of @@)*"`
	Grain           *string        `parser:"(By @Grain)?"`
	TimeZone        *string        `parser:"(Tz @Whitespace @TimeZone)?"`
	Offset          *Offset        `parser:"(Offset @@)?"`

	isNewFormat bool
	tz          *time.Location
	isoDuration *duration.StandardDuration
}

type Interval struct {
	Shorthand     *ShorthandInterval     `parser:"( @@"`
	PeriodToGrain *PeriodToGrainInterval `parser:"| @@"`
	StartEnd      *StartEndInterval      `parser:"| @@"`
	Ordinal       *OrdinalInterval       `parser:"| @@"`
	Iso           *IsoInterval           `parser:"| @@)"`
}

// ShorthandInterval is a convenience shorthand syntax for the advanced StartEndInterval
// <num><grain> maps to -<num><grain> to ref
type ShorthandInterval struct {
	Parts []*GrainDurationPart `parser:"@@ @@*"`
}

// PeriodToGrainInterval is a convenience syntax for specifying <grain> to ref
// <grain>TD maps to ref/<grain> to ref
type PeriodToGrainInterval struct {
	PeriodToGrain string `parser:"@PeriodToGrain"`
}

// OrdinalInterval is an interval formed with a chain of ordinals.
type OrdinalInterval struct {
	Ordinals []*Ordinal `parser:"@@ (Of @@)*"`
}

// StartEndInterval is a basic interval with a start and an end.
type StartEndInterval struct {
	Start *PointInTime `parser:"@@"`
	End   *PointInTime `parser:"To @@"`
}

// IsoInterval is an interval formed by ISO timestamps. Allows for partial timestamps in ISOPointInTime.
type IsoInterval struct {
	Start *ISOPointInTime `parser:"@@"`
	End   *ISOPointInTime `parser:"((To | '/' | RangeSeparator) @@)?"`
}

type PointInTime struct {
	Points []*PointInTimeWithSnap `parser:"@@ @@*"`
}

type PointInTimeWithSnap struct {
	Grain   *GrainPointInTime   `parser:"( @@"`
	Labeled *LabeledPointInTime `parser:"| @@"`
	ISO     *ISOPointInTime     `parser:"| @@)"`

	Snap *string `parser:"(Snap @Grain"`
	// A secondary snap after the above snap. This allows specifying a time range bucketed by week but snapped by a higher order grain.
	// EG: `Y/Y/W` or `Y/Y/W + 1Y` snaps to the beginning of the 1st week of the year or the beginning of the 1st week of next year (to include the last week of the year)
	//     `Y/Y` or `Y/Y + 1Y` instead gives 1st day of the year or 1st day of next year.
	SecondarySnap *string `parser:"(Snap @Grain)?)?"`
}

type GrainPointInTime struct {
	Parts []*GrainPointInTimePart `parser:"@@ @@*"`
}

type GrainPointInTimePart struct {
	Prefix   string         `parser:"@Prefix"`
	Duration *GrainDuration `parser:"@@"`
}

type LabeledPointInTime struct {
	Ref       bool `parser:"( @Ref"`
	Earliest  bool `parser:"| @Earliest"`
	Now       bool `parser:"| @Now"`
	Latest    bool `parser:"| @Latest"`
	Watermark bool `parser:"| @Watermark)"`
}

type ISOPointInTime struct {
	ISO string `parser:"@ISOTime"`

	year   int
	month  int
	week   int
	day    int
	hour   int
	minute int
	second int
	nano   int

	tg timeutil.TimeGrain
}

type Offset struct {
	PreviousPeriod *PreviousPeriod       `parser:"( @@"`
	Grain          *GrainPointInTimePart `parser:"| @@)"`
}

type PreviousPeriod struct {
	Prefix string `parser:"@Prefix"`
	Num    int    `parser:"@Number PreviousPeriod"`
}

type Ordinal struct {
	Grain string `parser:"@Grain"`
	Num   int    `parser:"@Number"`
}

type GrainDuration struct {
	Parts []*GrainDurationPart `parser:"@@ @@*"`
}

type GrainDurationPart struct {
	Num   int    `parser:"@Number"`
	Grain string `parser:"@Grain"`
}

// ParseOptions allows for additional options that could probably not be added to the time range itself
type ParseOptions struct {
	DefaultTimeZone  *time.Location
	TimeZoneOverride *time.Location
	// TODO: the correct way is perhaps add a keyword in syntax to reference smallest grain.
	SmallestGrain timeutil.TimeGrain
}

type EvalOptions struct {
	Now        time.Time
	MinTime    time.Time
	MaxTime    time.Time
	Watermark  time.Time
	FirstDay   int
	FirstMonth int

	ref          time.Time
	truncatedRef bool
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

		if rt.Interval != nil {
			err := rt.Interval.parse()
			if err != nil {
				return nil, err
			}
		}

		for _, override := range rt.AnchorOverrides {
			err := override.parse()
			if err != nil {
				return nil, err
			}
		}
	}

	rt.tz = time.UTC
	if parseOpts.TimeZoneOverride != nil {
		rt.tz = parseOpts.TimeZoneOverride
	} else if rt.TimeZone != nil {
		var err error
		rt.tz, err = time.LoadLocation(strings.TrimSpace(*rt.TimeZone))
		if err != nil {
			return nil, err
		}
	} else if parseOpts.DefaultTimeZone != nil {
		rt.tz = parseOpts.DefaultTimeZone
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

	// Update the ref so that anchor override can use it if needed. (EG: `-2Y^ to Y^ as of now`)
	if evalOpts.ref.IsZero() {
		evalOpts.ref = evalOpts.Now
	}
	i := len(e.AnchorOverrides) - 1
	for i >= 0 {
		evalOpts.ref, _ = e.AnchorOverrides[i].eval(evalOpts, evalOpts.ref, e.tz)
		if e.AnchorOverrides[i].truncates() {
			evalOpts.truncatedRef = true
		}
		i--
	}

	if e.isoDuration != nil {
		// handling for old iso format. all the times are relative to watermark for old format.
		isoStart := e.isoDuration.Sub(evalOpts.Watermark.In(e.tz))
		isoEnd := evalOpts.Watermark
		tg := timeutil.TimeGrainUnspecified
		if e.Grain != nil {
			tg = grainMap[*e.Grain]

			// ISO durations are mapped to `ref-iso to ref as of watermark/grain+1grain`
			isoStart = timeutil.OffsetTime(isoStart, tg, 1, e.tz)
			isoStart = timeutil.TruncateTime(isoStart, tg, e.tz, evalOpts.FirstDay, evalOpts.FirstMonth)
			isoEnd = timeutil.OffsetTime(isoEnd, tg, 1, e.tz)
			isoEnd = timeutil.TruncateTime(isoEnd, tg, e.tz, evalOpts.FirstDay, evalOpts.FirstMonth)
		}

		return isoStart, isoEnd, tg
	}

	start, end, tg := e.Interval.eval(evalOpts, evalOpts.ref, e.tz)

	if e.Offset != nil {
		start, end = e.Offset.eval(evalOpts, start, end, e.Interval, e.tz)
	}

	if e.Grain != nil {
		tg = grainMap[*e.Grain]
	} else {
		tg = getLowerOrderGrain(start, end, tg, e.tz)
	}

	return start.In(time.UTC), end.In(time.UTC), tg
}

/* Intervals */

func (i *Interval) parse() error {
	if i.StartEnd != nil {
		return i.StartEnd.parse()
	} else if i.Shorthand != nil {
		// Shorthand syntax that maps to StartEndInterval.
		i.StartEnd = i.Shorthand.expand()
	} else if i.PeriodToGrain != nil {
		// Period-to-date syntax maps to StartEndInterval as well.
		i.StartEnd = i.PeriodToGrain.expand()
	} else if i.Iso != nil {
		return i.Iso.parse()
	}
	return nil
}

func (i *Interval) eval(evalOpts EvalOptions, start time.Time, tz *time.Location) (time.Time, time.Time, timeutil.TimeGrain) {
	if i.Ordinal != nil {
		return i.Ordinal.eval(evalOpts, start, tz)
	} else if i.StartEnd != nil {
		return i.StartEnd.eval(evalOpts, start, tz)
	} else if i.Iso != nil {
		return i.Iso.eval(tz)
	}
	return start, start, timeutil.TimeGrainUnspecified
}

func (i *Interval) previousPeriod(evalOpts EvalOptions, tm time.Time, tz *time.Location) (time.Time, time.Time) {
	var start time.Time
	end := tm
	if i.Ordinal != nil {
		o := i.Ordinal.Ordinals[0]
		tg := grainMap[o.Grain]
		start = timeutil.OffsetTime(tm, tg, -1, tz)
	} else if i.StartEnd != nil {
		evalOpts.ref = tm
		start, _, _ = i.StartEnd.eval(evalOpts, tm, tz)
	} else if i.Iso != nil {
		return i.Iso.previousPeriod(tm, tz)
	}
	return start, end
}

func (s *ShorthandInterval) expand() *StartEndInterval {
	return &StartEndInterval{
		Start: &PointInTime{
			Points: []*PointInTimeWithSnap{
				{
					Grain: &GrainPointInTime{
						Parts: []*GrainPointInTimePart{
							{
								Prefix: "-",
								Duration: &GrainDuration{
									Parts: s.Parts,
								},
							},
						},
					},
				},
			},
		},
		End: &PointInTime{
			Points: []*PointInTimeWithSnap{
				{
					Labeled: &LabeledPointInTime{Ref: true},
				},
			},
		},
	}
}

func (p *PeriodToGrainInterval) expand() *StartEndInterval {
	fromTg := string(p.PeriodToGrain[0])
	return &StartEndInterval{
		Start: &PointInTime{
			Points: []*PointInTimeWithSnap{
				{
					Labeled: &LabeledPointInTime{Ref: true},
					Snap:    &fromTg,
				},
			},
		},
		End: &PointInTime{
			Points: []*PointInTimeWithSnap{
				{
					Labeled: &LabeledPointInTime{Ref: true},
				},
			},
		},
	}
}

func (o *OrdinalInterval) eval(evalOpts EvalOptions, start time.Time, tz *time.Location) (time.Time, time.Time, timeutil.TimeGrain) {
	end := start
	tg := timeutil.TimeGrainUnspecified

	if len(o.Ordinals) == 0 {
		return start, end, tg
	}

	if !evalOpts.truncatedRef {
		tg = grainMap[o.Ordinals[len(o.Ordinals)-1].Grain]
		start = truncateWithCorrection(start, higherOrderMap[tg], tz, evalOpts.FirstDay, evalOpts.FirstMonth)
	}

	i := len(o.Ordinals) - 1
	for i >= 0 {
		start, end, tg = o.Ordinals[i].eval(evalOpts, start, tz)
		i--
	}

	return start, end, tg
}

func (o *StartEndInterval) parse() error {
	err := o.Start.parse()
	if err != nil {
		return err
	}
	return o.End.parse()
}

func (o *StartEndInterval) eval(evalOpts EvalOptions, tm time.Time, tz *time.Location) (time.Time, time.Time, timeutil.TimeGrain) {
	start, startTg := o.Start.eval(evalOpts, tm, tz)
	end, endTg := o.End.eval(evalOpts, tm, tz)

	tg := endTg
	if endTg == timeutil.TimeGrainUnspecified || startTg > endTg {
		tg = startTg
	}

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

func (i *IsoInterval) previousPeriod(tm time.Time, tz *time.Location) (time.Time, time.Time) {
	if i.End != nil {
		start, end, _ := i.eval(tz)
		diff := end.Sub(start)

		end = start
		start = end.Add(-diff)
		return start, end
	}

	end := tm
	start := timeutil.OffsetTime(end, i.Start.tg, -1, tz)
	return start, end
}

/* Points in time */

func (p *PointInTime) parse() error {
	for _, point := range p.Points {
		err := point.parse()
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *PointInTime) eval(evalOpts EvalOptions, tm time.Time, tz *time.Location) (time.Time, timeutil.TimeGrain) {
	tg := timeutil.TimeGrainUnspecified
	for _, point := range p.Points {
		tm, tg = point.eval(evalOpts, tm, tz)
	}
	return tm, tg
}

func (p *PointInTime) truncates() bool {
	return len(p.Points) > 0 && p.Points[len(p.Points)-1].Snap != nil
}

func (p *PointInTimeWithSnap) parse() error {
	if p.ISO != nil {
		return p.ISO.parse()
	}
	return nil
}

func (p *PointInTimeWithSnap) eval(evalOpts EvalOptions, tm time.Time, tz *time.Location) (time.Time, timeutil.TimeGrain) {
	tg := timeutil.TimeGrainUnspecified
	if p.Grain != nil {
		tm, tg = p.Grain.eval(tm, tz)
	} else if p.Labeled != nil {
		tm = p.Labeled.eval(evalOpts)
	} else if p.ISO != nil {
		tm, _, tg = p.ISO.eval(tz)
	}

	if p.Snap != nil {
		tg = grainMap[*p.Snap]

		secondarySnap := timeutil.TimeGrainUnspecified
		if p.SecondarySnap != nil {
			// SecondarySnap is a special case, allows snap by a grain and then by another grain.
			// The only case where this will be different is when weeks are involved.
			secondarySnap = grainMap[*p.SecondarySnap]
		}

		tm = timeutil.TruncateTime(tm, tg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)

		if secondarySnap != timeutil.TimeGrainUnspecified {
			// If there is a secondary snap, then apply it after the primary snap has happened.
			// These need week correction since that is the primary goal of this syntax.
			tm = truncateWithCorrection(tm, secondarySnap, tz, evalOpts.FirstDay, evalOpts.FirstMonth)
		}
	}

	return tm, tg
}

func (g *GrainPointInTime) eval(tm time.Time, tz *time.Location) (time.Time, timeutil.TimeGrain) {
	tg := timeutil.TimeGrainUnspecified
	for _, part := range g.Parts {
		tm, tg = part.eval(tm, tz)
	}
	return tm, tg
}

func (g *GrainPointInTimePart) eval(tm time.Time, tz *time.Location) (time.Time, timeutil.TimeGrain) {
	dir := -1
	if g.Prefix == "+" {
		dir = 1
	}
	// Offset the time based on duration. Direction is specified here rather in Duration.
	return g.Duration.offset(tm, dir, tz)
}

func (l *LabeledPointInTime) eval(evalOpts EvalOptions) time.Time {
	if l.Ref {
		return evalOpts.ref
	} else if l.Earliest {
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

// TODO: reuse code from duration.ParseISO8601
func (a *ISOPointInTime) parse() error {
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
		case "milli":
			a.nano = val * 1000 * 1000
			a.tg = timeutil.TimeGrainMillisecond
		case "micro":
			a.nano = val * 1000
			a.tg = timeutil.TimeGrainMillisecond // We dont go below milli
		case "nano":
			a.nano = val
			a.tg = timeutil.TimeGrainMillisecond // We dont go below milli
		default:
			return fmt.Errorf("unexpected field %q in time", name)
		}
	}

	return nil
}

func (a *ISOPointInTime) eval(tz *time.Location) (time.Time, time.Time, timeutil.TimeGrain) {
	// Since we use this to build a time, month and day cannot be zero, hence the max(1, xx)
	absStart := time.Date(a.year, time.Month(max(1, a.month)), max(1, a.day), a.hour, a.minute, a.second, a.nano, tz)
	absEnd := timeutil.OffsetTime(absStart, a.tg, 1, tz)

	return absStart, absEnd, a.tg
}

func (o *Offset) eval(evalOpts EvalOptions, start, end time.Time, mainInterval *Interval, tz *time.Location) (time.Time, time.Time) {
	if o.PreviousPeriod != nil {
		start, end = o.PreviousPeriod.eval(evalOpts, start, mainInterval, tz)
	} else if o.Grain != nil {
		start, _ = o.Grain.eval(start, tz)
		end, _ = o.Grain.eval(end, tz)
	}

	return start, end
}

func (p *PreviousPeriod) eval(evalOpts EvalOptions, start time.Time, mainInterval *Interval, tz *time.Location) (time.Time, time.Time) {
	// TODO: things other than -1 period
	return mainInterval.previousPeriod(evalOpts, start, tz)
}

/* Durations */

func (o *Ordinal) eval(evalOpts EvalOptions, start time.Time, tz *time.Location) (time.Time, time.Time, timeutil.TimeGrain) {
	tg := grainMap[o.Grain]
	offset := o.Num - 1

	start = timeutil.OffsetTime(start, tg, offset, tz)
	start = truncateWithCorrection(start, tg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)

	end := timeutil.OffsetTime(start, tg, 1, tz)

	return start, end, tg
}

func (g *GrainDuration) offset(tm time.Time, dir int, tz *time.Location) (time.Time, timeutil.TimeGrain) {
	tg := timeutil.TimeGrainUnspecified
	i := len(g.Parts) - 1
	for i >= 0 {
		tm, tg = g.Parts[i].offset(tm, dir, tz)
		i--
	}
	return tm, tg
}

func (g *GrainDurationPart) offset(tm time.Time, dir int, tz *time.Location) (time.Time, timeutil.TimeGrain) {
	tg := grainMap[g.Grain]
	offset := g.Num
	offset *= dir

	return timeutil.OffsetTime(tm, tg, offset, tz), tg
}

func parseISO(from string, parseOpts ParseOptions) (*Expression, error) {
	// Try parsing for "inf"
	if infPattern.MatchString(from) {
		grainAlias := reverseGrainMap[parseOpts.SmallestGrain]
		return &Expression{
			Interval: &Interval{
				StartEnd: &StartEndInterval{
					Start: &PointInTime{
						Points: []*PointInTimeWithSnap{
							{
								Labeled: &LabeledPointInTime{
									Earliest: true,
								},
							},
						},
					},
					End: &PointInTime{
						Points: []*PointInTimeWithSnap{
							{
								Labeled: &LabeledPointInTime{
									Latest: true,
								},
								Snap: &grainAlias,
							},
							{
								Grain: &GrainPointInTime{
									Parts: []*GrainPointInTimePart{
										{
											Prefix:   "+",
											Duration: &GrainDuration{Parts: []*GrainDurationPart{{Grain: grainAlias, Num: 1}}},
										},
									},
								},
							},
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
// In most scenarios we need straight forward truncation. So this is not directly incorporated into timeutil.TruncateTime
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
			newTm = timeutil.OffsetTime(newTm, tg, 1, tz)
		}
	}

	return newTm
}

// getLowerOrderGrain returns the lowest grain where 2 periods can fit between start and end. Uses lowerOrderMap to get the lower grain.
func getLowerOrderGrain(start, end time.Time, tg timeutil.TimeGrain, tz *time.Location) timeutil.TimeGrain {
	for tg > timeutil.TimeGrainMillisecond {
		twoLower := timeutil.OffsetTime(end, tg, -2, tz)
		// if start < end - 2*grain, then we can return the grain.
		if start.Before(twoLower) || start.Equal(twoLower) {
			return tg
		}
		// else check the lower order grain
		tg = lowerOrderMap[tg]
	}
	return tg
}
