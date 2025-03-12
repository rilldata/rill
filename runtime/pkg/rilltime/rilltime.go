package rilltime

import (
	"fmt"
	"regexp"
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
		{"AbsoluteTime", `\d{4}-\d{2}-\d{2} \d{2}:\d{2}`},
		{"AbsoluteDate", `\d{4}-\d{2}-\d{2}`},
		{"ISOTime", isoTimePattern},
		{"AnchorPrefix", `[+\-<>]`},
		{"Sign", `[+-]`},
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
		"TD":  "0d,latest",
		"WTD": "0W,latest",
		"MTD": "0M,latest",
		"QTD": "0Q,latest",
		"YTD": "0Y,latest",
		"PDC": "-1d,0d",
		"PWC": "-1W,0W",
		"PMC": "-1M,0M",
		"PQC": "-1Q,0Q",
		"PYC": "-1Y,0Y",
		// TODO: previous period is contextual. should be handled in UI
		"PP": "",
		"PD": "-1d,0d",
		"PW": "-1W,0W",
		"PM": "-1M,0M",
		"PQ": "-1Q,0Q",
		"PY": "-1Y,0Y",
	}
	rillTimeParser = participle.MustBuild[Expression](
		participle.Lexer(rillTimeLexer),
		participle.Elide("Whitespace"),
	)
	rillTimeV2Parser = participle.MustBuild[ExpressionV2](
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
	Start         *TA          `parser:"  @@"`
	End           *TA          `parser:"(',' @@)?"`
	Modifiers     *Modifiers   `parser:"(':' @@)?"`
	AtModifiers   *AtModifiers `parser:"('@' @@)?"`
	isNewFormat   bool
	grain         *Grain
	truncateGrain timeutil.TimeGrain
	isComplete    bool
	timeZone      *time.Location
}

type TA struct {
	Grain       *Grain  `parser:"( @@"`
	AbsDate     *string `parser:"| @AbsoluteDate"`
	AbsTime     *string `parser:"| @AbsoluteTime"`
	Earliest    bool    `parser:"| @Earliest"`
	Now         bool    `parser:"| @Now"`
	Latest      bool    `parser:"| @Latest"`
	Watermark   bool    `parser:"| @Watermark)"`
	Trunc       *string `parser:"  ('/' @Grain)?"`
	Offset      *Grain  `parser:"@@?"`
	isoDuration *duration.StandardDuration
}

type Modifiers struct {
	Grain         *Grain `parser:"( @@"`
	CompleteGrain *Grain `parser:"| '|' @@ '|')?"`
}

type Grain struct {
	Sign  *string `parser:"@Sign?"`
	Num   *int    `parser:"@Number?"`
	Grain string  `parser:"@Grain"`
}

type AtModifiers struct {
	AnchorOverride *TA     `parser:"@@?"`
	TimeZone       *string `parser:"@TimeZone?"`
}

// ParseOptions allows for additional options that could probably not be added to the time range itself
type ParseOptions struct {
	DefaultTimeZone *time.Location
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

	rt.timeZone = time.UTC
	if parseOpts.DefaultTimeZone != nil {
		rt.timeZone = parseOpts.DefaultTimeZone
	}
	if rt.Modifiers != nil {
		if rt.Modifiers.Grain != nil {
			rt.truncateGrain = grainMap[rt.Modifiers.Grain.Grain]
			rt.grain = rt.Modifiers.Grain
			// TODO: non-1 grains
		} else if rt.Modifiers.CompleteGrain != nil {
			rt.truncateGrain = grainMap[rt.Modifiers.CompleteGrain.Grain]
			rt.grain = rt.Modifiers.CompleteGrain
			// TODO: non-1 grains
			rt.isComplete = true
		}
	}
	if rt.AtModifiers != nil && rt.AtModifiers.TimeZone != nil {
		var err error
		rt.timeZone, err = time.LoadLocation(strings.Trim(*rt.AtModifiers.TimeZone, "{}"))
		if err != nil {
			return nil, fmt.Errorf("invalid time zone %q: %w", *rt.AtModifiers.TimeZone, err)
		}
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

func (e *Expression) Eval(evalOpts EvalOptions) (time.Time, time.Time, error) {
	anchor, fallbackEndAnchor := e.getAnchor(evalOpts)

	start := anchor
	if e.Start != nil {
		start = e.modify(evalOpts, e.Start, anchor)
	}

	endAnchor := e.End
	if e.End == nil {
		endAnchor = fallbackEndAnchor
	}
	end := e.modify(evalOpts, endAnchor, anchor)

	return start, end, nil
}

func (e *Expression) modify(evalOpts EvalOptions, ta *TA, tm time.Time) time.Time {
	isTruncate := true
	truncateGrain := e.truncateGrain
	isBoundary := false

	if ta.isoDuration != nil {
		// handling for old iso format
		tm = ta.isoDuration.Sub(evalOpts.MaxTime.In(e.timeZone))
		isTruncate = true
		if e.grain != nil && e.grain.Grain != "" {
			truncateGrain = grainMap[e.grain.Grain]
		}
	} else if ta.Now {
		tm = evalOpts.Now.In(e.timeZone)
		isTruncate = e.isComplete
		isBoundary = true
	} else if ta.Earliest {
		tm = evalOpts.MinTime.In(e.timeZone)
		isTruncate = true
	} else if ta.Latest {
		tm = evalOpts.MaxTime.In(e.timeZone)
		isTruncate = e.isComplete
		isBoundary = true
	} else if ta.Watermark {
		tm = evalOpts.Watermark.In(e.timeZone)
		isTruncate = e.isComplete
		isBoundary = true
	} else if ta.AbsDate != nil {
		absTm, _ := time.Parse(time.DateOnly, *ta.AbsDate)
		tm = absTm.In(e.timeZone)
	} else if ta.AbsTime != nil {
		absTm, _ := time.Parse("2006-01-02 15:04", *ta.AbsTime)
		tm = absTm.In(e.timeZone)
	} else if ta.Grain != nil {
		tm = ta.Grain.offset(tm.In(e.timeZone))

		truncateGrain = grainMap[ta.Grain.Grain]
		isTruncate = true
	} else {
		return tm.In(e.timeZone)
	}

	if ta.Trunc != nil {
		truncateGrain = grainMap[*ta.Trunc]
		isTruncate = true
	}

	modifiedTime := tm
	if isTruncate {
		modifiedTime = timeutil.TruncateTime(tm, truncateGrain, e.timeZone, evalOpts.FirstDay, evalOpts.FirstMonth)
	} else {
		modifiedTime = timeutil.CeilTime(tm, truncateGrain, e.timeZone, evalOpts.FirstDay, evalOpts.FirstMonth)
	}

	// Add local offset after truncate. This allows for `0W+1D`
	if ta.Offset != nil {
		modifiedTime = ta.Offset.offset(modifiedTime)
		modifiedTime = timeutil.TruncateTime(modifiedTime, grainMap[ta.Offset.Grain], e.timeZone, evalOpts.FirstDay, evalOpts.FirstMonth)
	}

	// Add global offset from AtModifiers after truncate
	// Only grain offset is applied here. Anchor offsets like `@ now` or `@ latest` are applied to `tm` param
	if e.AtModifiers != nil && e.AtModifiers.AnchorOverride != nil && e.AtModifiers.AnchorOverride.Grain != nil {
		modifiedTime = e.AtModifiers.AnchorOverride.Grain.offset(modifiedTime)
	}

	if isBoundary && modifiedTime.Equal(tm) && (e.Modifiers == nil || e.Modifiers.CompleteGrain == nil) {
		// edge case where the end time falls on a boundary. add +1grain to make sure the last data point is included
		n := 1
		g := &Grain{
			Num:   &n,
			Grain: "s",
		}
		if e.grain != nil {
			g.Grain = e.grain.Grain
		}
		modifiedTime = g.offset(modifiedTime)
	}

	return modifiedTime
}

func (e *Expression) getAnchor(evalOpts EvalOptions) (time.Time, *TA) {
	if e.AtModifiers != nil && e.AtModifiers.AnchorOverride != nil {
		if e.AtModifiers.AnchorOverride.Now {
			return evalOpts.Now, e.AtModifiers.AnchorOverride
		}
		if e.AtModifiers.AnchorOverride.Latest {
			return evalOpts.MaxTime, e.AtModifiers.AnchorOverride
		}
		if e.AtModifiers.AnchorOverride.Earliest {
			return evalOpts.MinTime, e.AtModifiers.AnchorOverride
		}
		if e.AtModifiers.AnchorOverride.AbsDate != nil {
			absTm, _ := time.Parse(time.DateOnly, *e.AtModifiers.AnchorOverride.AbsDate)
			return absTm, e.AtModifiers.AnchorOverride
		}
		if e.AtModifiers.AnchorOverride.AbsTime != nil {
			absTm, _ := time.Parse("2006-01-02 15:04", *e.AtModifiers.AnchorOverride.AbsTime)
			return absTm, e.AtModifiers.AnchorOverride
		}
	}

	if e.End == nil {
		return evalOpts.Watermark, &TA{
			Watermark: true,
		}
	}

	if e.End.Latest {
		// if end has latest mentioned then start also should be relative to latest.
		return evalOpts.MaxTime, e.End
	}
	if e.End.Now {
		// if end has now mentioned then start also should be relative to latest.
		return evalOpts.Now, e.End
	}
	return evalOpts.Watermark, e.End
}

func parseISO(from string, parseOpts ParseOptions) (*Expression, error) {
	// Try parsing for "inf"
	if infPattern.MatchString(from) {
		return &Expression{
			Start: &TA{Earliest: true},
			End:   &TA{Latest: true},
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

	rt := &Expression{
		Start: &TA{},
		End:   &TA{Latest: true},
		// mirrors old UI behaviour
		isComplete: false,
	}
	d, err := duration.ParseISO8601(from)
	if err != nil {
		return nil, nil
	}
	sd, ok := d.(duration.StandardDuration)
	if !ok {
		return nil, nil
	}
	rt.Start.isoDuration = &sd
	minGrain := getMinGrain(sd)
	if minGrain != "" {
		rt.grain = &Grain{
			Grain: minGrain,
		}
	}

	return rt, nil
}

func (g *Grain) offset(tm time.Time) time.Time {
	n := 0
	if g.Num != nil {
		n = *g.Num

		if g.Sign != nil && *g.Sign == "-" {
			n = -n
		}
	}

	switch g.Grain {
	case "s", "S":
		tm = tm.Add(time.Duration(n) * time.Second)
	case "m":
		tm = tm.Add(time.Duration(n) * time.Minute)
	case "h", "H":
		tm = tm.Add(time.Duration(n) * time.Hour)
	case "d", "D":
		tm = tm.AddDate(0, 0, n)
	case "W", "w":
		tm = tm.AddDate(0, 0, n*7)
	case "M":
		tm = tm.AddDate(0, n, 0)
	case "Q", "q":
		tm = tm.AddDate(0, n*3, 0)
	case "Y", "y":
		tm = tm.AddDate(n, 0, 0)
	}

	return tm
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
