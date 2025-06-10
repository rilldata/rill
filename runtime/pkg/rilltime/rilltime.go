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
		{"WallClock", "wallclock"},
		{"Latest", "latest"},
		{"Watermark", "watermark"},
		{"Starting", "starting"},
		{"Ending", "ending"},
		// this needs to be after Now and Latest to match to them
		{"WeekSnapGrain", `[qQMyY][wW]`},
		{"PeriodToGrain", `[sSmhHdDwWqQMyY]TD`},
		{"Grain", `[sSmhHdDwWqQMyY]`},
		// this has to be at the end
		{"TimeZone", `{.+?}`},
		{"ISOTime", isoTimePattern},
		{"Prefix", `[+\-]`},
		{"Suffix", `[\^\$]`},
		{"SnapPrefix", `[<>]`},
		{"Number", `\d+`},
		{"Snap", `[/]`},
		{"Interval", `[#]`},
		{"Ceil", `[!]`},
		{"To", `(?i)to`},
		{"ToDate", `(?i)TD`},
		{"Thru", `(?i)thru`},
		{"Offset", `(?i)offset`},
		{"By", `(?i)by`},
		{"Of", `(?i)of`},
		{"As", `(?i)as`},
		{"In", `(?i)in`},
		// needed for misc. direct character references used
		{"Punct", `[-[!@#$%^&*()+_={}\|:;"'<,>.?/]]`},
		{"Whitespace", `[ \t]+`},
	})
	rillTimeParser = participle.MustBuild[Expression](
		participle.Lexer(rillTimeLexer),
		participle.Elide("Whitespace"),
		participle.UseLookahead(-1), // We need this to disambiguate certain cases. Mainly for something like `-4d#`
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
	simplifiedSnapMap = map[timeutil.TimeGrain]timeutil.TimeGrain{
		timeutil.TimeGrainSecond:  timeutil.TimeGrainSecond,
		timeutil.TimeGrainMinute:  timeutil.TimeGrainMinute,
		timeutil.TimeGrainHour:    timeutil.TimeGrainHour,
		timeutil.TimeGrainDay:     timeutil.TimeGrainHour,
		timeutil.TimeGrainWeek:    timeutil.TimeGrainDay,
		timeutil.TimeGrainMonth:   timeutil.TimeGrainDay,
		timeutil.TimeGrainQuarter: timeutil.TimeGrainDay,
		timeutil.TimeGrainYear:    timeutil.TimeGrainDay,
	}
)

type Expression struct {
	Range      *Range          `parser:"@@"`
	Offsets    []*OffsetClause `parser:"@@*"`
	AsOfClause *AsOfClause     `parser:"@@?"`

	isNewFormat bool
	timeZone    *time.Location
	isoDuration *duration.StandardDuration
	grain       string
}

type Range struct {
	Ordinal         *Ordinal         `parser:"( @@"`
	ToRange         *ToRange         `parser:"| @@"`
	ThruRange       *ThruRange       `parser:"| @@"`
	PointInTime     *PointInTime     `parser:"| @@"`
	UnitToDateRange *UnitToDateRange `parser:"| @@"`
	UnitRange       *UnitRange       `parser:"| @@"`
	DurationRange   *DurationRange   `parser:"| @@)"`
}

type NestedRange struct {
	SubExpr *Expression `parser:"'(' @@ ')'"`
}

type RangeSide struct {
	PointInTime *PointInTime `parser:"( @@"`
	Ordinal     *Ordinal     `parser:"| @@"`
	NestedRange *NestedRange `parser:"| ( @@"`
	Snap        *string      `parser:"    @Suffix?))"`
}

// ToRange has inclusive start and exclusive end
type ToRange struct {
	Start *RangeSide `parser:"@@"`
	End   *RangeSide `parser:"To @@"`
}

// ThruRange has inclusive start and inclusive end (TODO: merge with ToRange?)
type ThruRange struct {
	Start *RangeSide `parser:"@@"`
	End   *RangeSide `parser:"Thru @@"`
}

// DurationRange allows to start or end a duration on a point
// EG: 24h starting -1d/d^
type DurationRange struct {
	Duration  *Duration  `parser:"@@"`
	Starting  bool       `parser:"( @Starting"`
	Ending    bool       `parser:"| @Ending)"`
	RangeSide *RangeSide `parser:"@@"`
}

type AsOfClause struct {
	WallClock      bool            `parser:"As Of ( @WallClock"`
	Latest         bool            `parser:"| @Latest"`
	Watermark      bool            `parser:"| @Watermark"`
	PointInTime    *PointInTime    `parser:"| @@"`
	OffsetSequence *OffsetSequence `parser:"| @@)"`
}

type OffsetClause struct {
	OffsetSequence *OffsetSequence `parser:"Offset @@"`
}

type PointInTime struct {
	Now            *string         `parser:"(( @Now"`
	NowSnap        *string         `parser:"   @Suffix?)"`
	Iso            *AbsoluteTime   `parser:"| @@"`
	OffsetSequence *OffsetSequence `parser:"| @@"`
	UnitStartOrEnd *UnitStartOrEnd `parser:"| @@)"`
}

type UnitToDateRange struct {
	Unit string `parser:"@Grain ToDate"`
}

type UnitRange struct {
	Unit string `parser:"@Grain"`
}

type UnitStartOrEnd struct {
	UnitRange *UnitRange `parser:"@@"`
	SnapTo    string     `parser:"@Suffix"`
}

type Ordinal struct {
	Grain   string `parser:"@Grain"`
	Number  int    `parser:"@Number"`
	OfRange *Range `parser:"(Of @@)?"`
}

type OffsetSequence struct {
	Offsets []*Offset `parser:"@@ @@*"`
}

type Offset struct {
	SignedDuration *SignedDuration `parser:"@@"`
	UnitSnapping   *UnitSnapping   `parser:"@@?"`
}

type UnitSnapping struct {
	Unit   string  `parser:"'/' @Grain"`
	SnapTo *string `parser:"@Suffix?"`
}

type SignedDuration struct {
	Sign     string    `parser:"@Prefix"`
	Duration *Duration `parser:"@@"`
}

type Duration struct {
	Terms []*Term `parser:"@@ @@*"`
}

type Term struct {
	Number int    `parser:"@Number"`
	Unit   string `parser:"@Grain"`
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
	Now           time.Time
	MinTime       time.Time
	MaxTime       time.Time
	Watermark     time.Time
	FirstDay      int
	FirstMonth    int
	SmallestGrain timeutil.TimeGrain

	ref time.Time
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

		err = rt.parse()
		if err != nil {
			return nil, err
		}
	}

	rt.timeZone = time.UTC
	if parseOpts.TimeZoneOverride != nil {
		rt.timeZone = parseOpts.TimeZoneOverride
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
	if evalOpts.SmallestGrain == timeutil.TimeGrainUnspecified {
		evalOpts.SmallestGrain = timeutil.TimeGrainMillisecond
	}
	if evalOpts.ref.IsZero() {
		// If the ref is not set at all, set it to watermark
		evalOpts.ref = evalOpts.Watermark
	}

	if e.AsOfClause != nil {
		// Update the ref based on as of clause.
		evalOpts.ref = e.AsOfClause.eval(evalOpts, evalOpts.ref, e.timeZone)
	}

	if e.isoDuration != nil {
		// handling for old iso format
		isoStart := e.isoDuration.Sub(evalOpts.ref.In(e.timeZone))
		isoEnd := evalOpts.ref
		tg := timeutil.TimeGrainUnspecified
		if e.grain != "" {
			tg = grainMap[e.grain]
			isoStart = timeutil.TruncateTime(isoStart, tg, e.timeZone, evalOpts.FirstDay, evalOpts.FirstMonth)
			isoEnd = timeutil.TruncateTime(isoEnd, tg, e.timeZone, evalOpts.FirstDay, evalOpts.FirstMonth)
		}

		return isoStart, isoEnd, tg
	}

	start, end := e.Range.eval(evalOpts, evalOpts.ref, e.timeZone)

	// TODO: offsets

	return start, end, timeutil.TimeGrainUnspecified
}

// Parse functions. We only need parse for iso right now.
// Since there will be arbitrary nesting, we need all these individual parse functions.

func (e *Expression) parse() error {
	err := e.Range.parse()
	if err != nil {
		return err
	}

	if e.AsOfClause != nil {
		err := e.AsOfClause.parse()
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Range) parse() error {
	if r.PointInTime != nil {
		return r.PointInTime.parse()
	} else if r.ToRange != nil {
		return r.ToRange.parse()
	} else if r.ThruRange != nil {
		return r.ThruRange.parse()
	} else if r.DurationRange != nil {
		return r.DurationRange.parse()
	} else if r.Ordinal != nil {
		return r.Ordinal.parse()
	}
	return nil
}

func (n *NestedRange) parse() error {
	return n.SubExpr.parse()
}

func (r *RangeSide) parse() error {
	if r.PointInTime != nil {
		return r.PointInTime.parse()
	} else if r.Ordinal != nil {
		return r.Ordinal.parse()
	} else if r.NestedRange != nil {
		return r.NestedRange.parse()
	}
	return nil
}

func (t *ToRange) parse() error {
	err := t.Start.parse()
	if err != nil {
		return err
	}
	return t.End.parse()
}

func (t *ThruRange) parse() error {
	err := t.Start.parse()
	if err != nil {
		return err
	}
	return t.End.parse()
}

func (d *DurationRange) parse() error {
	return d.RangeSide.parse()
}

func (a *AsOfClause) parse() error {
	if a.PointInTime != nil {
		return a.PointInTime.parse()
	}
	return nil
}

func (p *PointInTime) parse() error {
	if p.Iso != nil {
		return p.Iso.parse()
	}
	return nil
}

func (o *Ordinal) parse() error {
	if o.OfRange != nil {
		err := o.OfRange.parse()
		if err != nil {
			return err
		}
	}
	return nil
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

// Eval functions

func (r *Range) eval(evalOpts EvalOptions, tm time.Time, tz *time.Location) (time.Time, time.Time) {
	if r.PointInTime != nil {
		tm = r.PointInTime.eval(evalOpts, tm, tz)
		return tm, tm
	} else if r.UnitToDateRange != nil {
		return r.UnitToDateRange.eval(evalOpts, tm, tz)
	} else if r.UnitRange != nil {
		return r.UnitRange.eval(evalOpts, tm, tz)
	} else if r.ToRange != nil {
		return r.ToRange.eval(evalOpts, tm, tz)
	} else if r.ThruRange != nil {
		return r.ThruRange.eval(evalOpts, tm, tz)
	} else if r.DurationRange != nil {
		return r.DurationRange.eval(evalOpts, tm, tz)
	} else if r.Ordinal != nil {
		return r.Ordinal.eval(evalOpts, tm, tz)
	}
	return time.Time{}, time.Time{}
}

func (n *NestedRange) eval(evalOpts EvalOptions, tm time.Time, tz *time.Location) (time.Time, time.Time) {
	newEvalOpts := EvalOptions{
		Now:        evalOpts.Now,
		MinTime:    evalOpts.MinTime,
		MaxTime:    evalOpts.MaxTime,
		Watermark:  evalOpts.Watermark,
		FirstDay:   evalOpts.FirstDay,
		FirstMonth: evalOpts.FirstMonth,
		ref:        tm,
	}
	start, end, _ := n.SubExpr.Eval(newEvalOpts)
	return start, end
}

func (r *RangeSide) eval(evalOpts EvalOptions, tm time.Time, tz *time.Location) time.Time {
	if r.PointInTime != nil {
		return r.PointInTime.eval(evalOpts, tm, tz)
	} else if r.Ordinal != nil {
		tm, _ = r.Ordinal.eval(evalOpts, tm, tz)
		return tm
	} else if r.NestedRange != nil {
		start, end := r.NestedRange.eval(evalOpts, tm, tz)
		if r.Snap == nil || *r.Snap == "$" {
			return end
		}
		return start
	}
	return time.Time{}
}

func (t *ToRange) eval(evalOpts EvalOptions, tm time.Time, tz *time.Location) (time.Time, time.Time) {
	start := t.Start.eval(evalOpts, tm, tz)
	end := t.End.eval(evalOpts, tm, tz)
	return start, end
}

func (t *ThruRange) eval(evalOpts EvalOptions, tm time.Time, tz *time.Location) (time.Time, time.Time) {
	start := t.Start.eval(evalOpts, tm, tz)
	end := t.End.eval(evalOpts, tm, tz)
	return start, end
}

func (d *DurationRange) eval(evalOpts EvalOptions, tm time.Time, tz *time.Location) (time.Time, time.Time) {
	tm = d.RangeSide.eval(evalOpts, tm, tz)

	if d.Starting {
		end := d.Duration.eval(tm, 1)
		return tm, end
	} else {
		start := d.Duration.eval(tm, -1)
		return start, tm
	}
}

func (a *AsOfClause) eval(evalOpts EvalOptions, tm time.Time, tz *time.Location) time.Time {
	if a.WallClock {
		return evalOpts.Now
	} else if a.Latest {
		return evalOpts.MaxTime
	} else if a.Watermark {
		return evalOpts.Watermark
	} else if a.PointInTime != nil {
		return a.PointInTime.eval(evalOpts, tm, tz)
	} else if a.OffsetSequence != nil {
		return a.OffsetSequence.eval(evalOpts, tm, tz)
	}
	return time.Time{}
}

func (o *OffsetClause) eval(evalOpts EvalOptions, tm time.Time, tz *time.Location) (time.Time, time.Time) {
	return time.Time{}, time.Time{}
}

func (p *PointInTime) eval(evalOpts EvalOptions, tm time.Time, tz *time.Location) time.Time {
	if p.Now != nil {
		tm = evalOpts.ref
		// What does now-snap do here?
		return tm
	} else if p.Iso != nil {
		tm, _, _ := p.Iso.eval(tz)
		return tm
	} else if p.OffsetSequence != nil {
		return p.OffsetSequence.eval(evalOpts, tm, tz)
	} else if p.UnitStartOrEnd != nil {
		return p.UnitStartOrEnd.eval(evalOpts, tm, tz)
	}
	return time.Time{}
}

func (u *UnitToDateRange) eval(evalOpts EvalOptions, tm time.Time, tz *time.Location) (time.Time, time.Time) {
	return time.Time{}, time.Time{}
}

func (u *UnitRange) eval(evalOpts EvalOptions, tm time.Time, tz *time.Location) (time.Time, time.Time) {
	tg := grainMap[u.Unit]
	start := timeutil.TruncateTime(tm, tg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)
	end := timeutil.OffsetTime(start, tg, 1)
	return start, end
}

func (u *UnitStartOrEnd) eval(evalOpts EvalOptions, tm time.Time, tz *time.Location) time.Time {
	start, end := u.UnitRange.eval(evalOpts, tm, tz)
	if u.SnapTo == "^" {
		return start
	} else {
		return end
	}
}

func (o *Ordinal) eval(evalOpts EvalOptions, tm time.Time, tz *time.Location) (time.Time, time.Time) {
	start := tm
	tg := grainMap[o.Grain]

	if o.OfRange != nil {
		start, _ = o.OfRange.eval(evalOpts, tm, tz)
		// TODO: we might need the grain from range here to truncate
	} else {
		highTg := higherOrderMap[tg]
		start = truncateWithCorrection(tm, highTg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)
	}

	start = truncateWithCorrection(start, tg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)
	if o.Number-1 > 0 {
		start = timeutil.OffsetTime(start, tg, o.Number-1)
	}
	end := timeutil.OffsetTime(start, tg, 1)

	return start, end
}

func (o *OffsetSequence) eval(evalOpts EvalOptions, tm time.Time, tz *time.Location) time.Time {
	for _, o := range o.Offsets {
		tm = o.eval(evalOpts, tm, tz)
	}
	return tm
}

func (o *Offset) eval(evalOpts EvalOptions, tm time.Time, tz *time.Location) time.Time {
	tm = o.SignedDuration.eval(tm)
	if o.UnitSnapping != nil {
		tm = o.UnitSnapping.eval(evalOpts, tm, tz)
	}
	return tm
}

func (u *UnitSnapping) eval(evalOpts EvalOptions, tm time.Time, tz *time.Location) time.Time {
	tg := grainMap[u.Unit]
	if u.SnapTo != nil && *u.SnapTo == "$" {
		tm = timeutil.OffsetTime(tm, tg, 1)
	}
	tm = timeutil.TruncateTime(tm, tg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)

	return tm
}

func (s *SignedDuration) eval(tm time.Time) time.Time {
	dir := 1
	if s.Sign == "-" {
		dir = -1
	}
	return s.Duration.eval(tm, dir)
}

func (d *Duration) eval(tm time.Time, dir int) time.Time {
	for _, t := range d.Terms {
		tg := grainMap[t.Unit]
		tm = timeutil.OffsetTime(tm, tg, t.Number*dir)
	}
	return tm
}

func (a *AbsoluteTime) eval(tz *time.Location) (time.Time, time.Time, timeutil.TimeGrain) {
	// Since we use this to build a time, month and day cannot be zero, hence the max(1, xx)
	absStart := time.Date(a.year, time.Month(max(1, a.month)), max(1, a.day), a.hour, a.minute, a.second, 0, tz)
	absEnd := timeutil.OffsetTime(absStart, a.tg, 1)

	return absStart, absEnd, a.tg
}

func parseISO(from string, parseOpts ParseOptions) (*Expression, error) {
	// Try parsing for "inf"
	if infPattern.MatchString(from) {
		return &Expression{
			// TODO: inf
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
	rt.grain = getMinGrain(sd)

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
