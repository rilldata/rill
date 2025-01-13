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
	// nolint:govet // This is suggested usage by the docs.
	rillTimeLexer = lexer.MustSimple([]lexer.SimpleRule{
		{"Earliest", "earliest"},
		{"Now", "now"},
		{"Latest", "latest"},
		{"Watermark", "watermark"},
		// this needs to be after Now and Latest to match to them
		{"Grain", `[smhdDWQMY]`},
		// this has to be at the end
		{"TimeZone", `{.+?}`},
		{"AbsoluteTime", `\d{4}-\d{2}-\d{2} \d{2}:\d{2}`},
		{"AbsoluteDate", `\d{4}-\d{2}-\d{2}`},
		{"Number", `[-+]?\d+`},
		// needed for misc. direct character references used
		{"Punct", `[-[!@#$%^&*()+_={}\|:;"'<,>.?/]|]`},
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
	grainMap = map[string]timeutil.TimeGrain{
		"s": timeutil.TimeGrainSecond,
		"m": timeutil.TimeGrainMinute,
		"h": timeutil.TimeGrainHour,
		"d": timeutil.TimeGrainDay,
		"D": timeutil.TimeGrainDay,
		"W": timeutil.TimeGrainWeek,
		"Q": timeutil.TimeGrainQuarter,
		"M": timeutil.TimeGrainMonth,
		"Y": timeutil.TimeGrainYear,
	}
)

type Expression struct {
	Start         *TimeAnchor  `parser:"  @@"`
	End           *TimeAnchor  `parser:"(',' @@)?"`
	Modifiers     *Modifiers   `parser:"(':' @@)?"`
	AtModifiers   *AtModifiers `parser:"('@' @@)?"`
	isNewFormat   bool
	grain         *Grain
	truncateGrain timeutil.TimeGrain
	isComplete    bool
	timeZone      *time.Location
}

type TimeAnchor struct {
	Grain       *Grain  `parser:"( @@"`
	AbsDate     *string `parser:"| @AbsoluteDate"`
	AbsTime     *string `parser:"| @AbsoluteTime"`
	Earliest    bool    `parser:"| @Earliest"`
	Now         bool    `parser:"| @Now"`
	Latest      bool    `parser:"| @Latest"`
	Watermark   bool    `parser:"| @Watermark)"`
	Offset      *Grain  `parser:"@@?"`
	Trunc       *string `parser:"  ('/' @Grain)?"`
	isoDuration *duration.StandardDuration
}

type Modifiers struct {
	Grain         *Grain `parser:"( @@"`
	CompleteGrain *Grain `parser:"| '|' @@ '|')?"`
}

type Grain struct {
	Num   *int   `parser:"@Number?"`
	Grain string `parser:"@Grain"`
}

type AtModifiers struct {
	Offset   *TimeAnchor `parser:"@@?"`
	TimeZone *string     `parser:"@TimeZone?"`
}

type EvalOptions struct {
	Now        time.Time
	MinTime    time.Time
	MaxTime    time.Time
	Watermark  time.Time
	FirstDay   int
	FirstMonth int
}

func Parse(from string) (*Expression, error) {
	var rt *Expression
	var err error

	rt, err = ParseISO(from, false)
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

	if rt.End == nil {
		rt.End = &TimeAnchor{
			Now: true,
		}
	}

	return rt, nil
}

func ParseISO(from string, strict bool) (*Expression, error) {
	// Try parsing for "inf"
	if infPattern.MatchString(from) {
		return &Expression{
			Start: &TimeAnchor{Earliest: true},
			End:   &TimeAnchor{Latest: true},
		}, nil
	}

	if strings.HasPrefix(from, "rill-") {
		// We are using "rill-" as a prefix to DAX notation so that it doesn't interfere with ISO8601 standard.
		// Pulled from https://www.daxpatterns.com/standard-time-related-calculations/
		rillDur := strings.Replace(from, "rill-", "", 1)
		if t, ok := daxNotations[rillDur]; ok {
			return Parse(t)
		}
	}

	// Parse as a regular ISO8601 duration
	if !durationPattern.MatchString(from) {
		if !strict {
			return nil, nil
		}
		return nil, fmt.Errorf("string %q is not a valid ISO 8601 duration", from)
	}

	rt := &Expression{
		Start: &TimeAnchor{},
		End:   &TimeAnchor{Now: true},
		// mirrors old UI behaviour
		isComplete: false,
	}
	d, err := duration.ParseISO8601(from)
	if err != nil {
		if !strict {
			return nil, nil
		}
		return nil, err
	}
	sd, ok := d.(duration.StandardDuration)
	if !ok {
		if !strict {
			return nil, nil
		}
		return nil, fmt.Errorf("duration %q is invalid iso format", from)
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

func ParseCompatibility(timeRange, offset string) error {
	isNewFormat := false
	if timeRange != "" {
		rt, err := Parse(timeRange)
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
	start := evalOpts.Now
	if e.End != nil && e.End.Latest {
		// if end has latest mentioned then start also should be relative to latest.
		start = evalOpts.MaxTime
	}

	if e.Start != nil {
		start = e.Modify(evalOpts, e.Start, start, true)
	}

	end := evalOpts.Now
	if e.End != nil {
		end = e.Modify(evalOpts, e.End, end, true)
	}

	return start, end, nil
}

func (e *Expression) Modify(evalOpts EvalOptions, ta *TimeAnchor, tm time.Time, addOffset bool) time.Time {
	isTruncate := true
	truncateGrain := e.truncateGrain
	isBoundary := false

	if ta.isoDuration != nil {
		// handling for old iso format
		return ta.isoDuration.Sub(evalOpts.MinTime.In(e.timeZone))
	}

	if ta.Now {
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

	timeBeforeOffset := tm
	if ta.Offset != nil {
		tm = ta.Offset.offset(tm)
	}
	if addOffset && e.AtModifiers != nil && e.AtModifiers.Offset != nil {
		tm = e.Modify(evalOpts, e.AtModifiers.Offset, tm, false)
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

	if isBoundary && modifiedTime.Equal(timeBeforeOffset) {
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

func (g *Grain) offset(tm time.Time) time.Time {
	n := 0
	if g.Num != nil {
		n = *g.Num
	}

	switch g.Grain {
	case "s":
		tm = tm.Add(time.Duration(n) * time.Second)
	case "m":
		tm = tm.Add(time.Duration(n) * time.Minute)
	case "h":
		tm = tm.Add(time.Duration(n) * time.Hour)
	case "d":
		tm = tm.AddDate(0, 0, n)
	case "D":
		tm = tm.AddDate(0, 0, n)
	case "W":
		tm = tm.AddDate(0, 0, n*7)
	case "M":
		tm = tm.AddDate(0, n, 0)
	case "Q":
		tm = tm.AddDate(0, n*3, 0)
	case "Y":
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
