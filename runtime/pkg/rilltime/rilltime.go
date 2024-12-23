package rilltime

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/rilldata/rill/runtime/pkg/timeutil"
)

var (
	infPattern      = regexp.MustCompile("^(?i)inf$")
	durationPattern = regexp.MustCompile(`^P((?P<year>\d+)Y)?((?P<month>\d+)M)?((?P<week>\d+)W)?((?P<day>\d+)D)?(T((?P<hour>\d+)H)?((?P<minute>\d+)M)?((?P<second>\d+)S)?)?$`)
	rillTimeLexer   = lexer.MustSimple([]lexer.SimpleRule{
		{"Now", "now"},
		{"Latest", "latest"},
		{"Earliest", "earliest"},
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
		"TD":  "",
		"WTD": "",
		"MTD": "",
		"QTD": "",
		"YTD": "",
		"PP":  "",
		"PD":  "",
		"PW":  "",
		"PM":  "",
		"PQ":  "",
		"PY":  "",
		"PDC": "",
		"PWC": "",
		"PMC": "",
		"PQC": "",
		"PYC": "",
	}
	rillTimeParser = participle.MustBuild[RillTime](
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

type RillTime struct {
	Start       *TimeAnchor `parser:"  @@"`
	End         *TimeAnchor `parser:"(',' @@)?"`
	Modifiers   *Modifiers  `parser:"(':' @@)?"`
	IsNewFormat bool
	grain       timeutil.TimeGrain
	isComplete  bool
	timeZone    *time.Location
}

type TimeAnchor struct {
	Grain    *Grain  `parser:"( @@"`
	AbsDate  *string `parser:"| @AbsoluteDate"`
	AbsTime  *string `parser:"| @AbsoluteTime"`
	Now      bool    `parser:"| @Now"`
	Latest   bool    `parser:"| @Latest"`
	Earliest bool    `parser:"| @Earliest)"`
	Trunc    *string `parser:"  ('/' @Grain)?"`
}

type Modifiers struct {
	Grain         *Grain       `parser:"( @@"`
	CompleteGrain *Grain       `parser:"| '|' @@ '|')?"`
	At            *AtModifiers `parser:"( '@' @@)?"`
}

type Grain struct {
	Num   *int   `parser:"@Number?"`
	Grain string `parser:"@Grain"`
}

type AtModifiers struct {
	Offset   *TimeAnchor `parser:"@@?"`
	TimeZone *string     `parser:"@TimeZone?"`
}

type ResolverContext struct {
	Now        time.Time
	MinTime    time.Time
	MaxTime    time.Time
	FirstDay   int
	FirstMonth int
}

func Parse(from string) (*RillTime, error) {
	var rt *RillTime
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
		rt.IsNewFormat = true
	}

	rt.timeZone = time.UTC
	if rt.Modifiers != nil {
		if rt.Modifiers.Grain != nil {
			rt.grain = grainMap[rt.Modifiers.Grain.Grain]
			// TODO: non-1 grains
		} else if rt.Modifiers.CompleteGrain != nil {
			rt.grain = grainMap[rt.Modifiers.CompleteGrain.Grain]
			// TODO: non-1 grains
			rt.isComplete = true
		}

		if rt.Modifiers.At != nil {
			if rt.Modifiers.At.TimeZone != nil {
				var err error
				rt.timeZone, err = time.LoadLocation(strings.Trim(*rt.Modifiers.At.TimeZone, "{}"))
				if err != nil {
					return nil, fmt.Errorf("invalid time zone %q: %w", *rt.Modifiers.At.TimeZone, err)
				}
			}
		}
	}

	if rt.End == nil {
		rt.End = &TimeAnchor{
			Now: true,
		}
	}

	return rt, nil
}

func ParseISO(from string, strict bool) (*RillTime, error) {
	// Try parsing for "inf"
	if infPattern.MatchString(from) {
		return &RillTime{
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

	rt := &RillTime{
		Start: &TimeAnchor{Grain: &Grain{}},
		End:   &TimeAnchor{Now: true},
	}
	// TODO: we do not need name based matching here since we just map the grain
	match := durationPattern.FindStringSubmatch(from)
	for i, name := range durationPattern.SubexpNames() {
		part := match[i]
		if i == 0 || name == "" || part == "" {
			continue
		}

		val, err := strconv.Atoi(part)
		if err != nil {
			return nil, err
		}
		rt.Start.Grain.Num = &val
		switch name {
		case "year":
			rt.Start.Grain.Grain = "Y"
		case "month":
			rt.Start.Grain.Grain = "M"
		case "week":
			rt.Start.Grain.Grain = "W"
		case "day":
			rt.Start.Grain.Grain = "d"
		case "hour":
			rt.Start.Grain.Grain = "h"
		case "minute":
			rt.Start.Grain.Grain = "m"
		case "second":
			rt.Start.Grain.Grain = "s"
		default:
			return nil, fmt.Errorf("unexpected field %q in duration", name)
		}
	}

	return rt, nil
}

func (r *RillTime) Resolve(resolverCtx ResolverContext) (time.Time, time.Time, error) {
	if r.Modifiers != nil && r.Modifiers.At != nil && r.Modifiers.At.Offset != nil {
		resolverCtx.Now = r.Modifiers.At.Offset.Modify(resolverCtx, resolverCtx.Now, r.grain, r.timeZone, r.isComplete)
		resolverCtx.MinTime = r.Modifiers.At.Offset.Modify(resolverCtx, resolverCtx.MinTime, r.grain, r.timeZone, r.isComplete)
		resolverCtx.MaxTime = r.Modifiers.At.Offset.Modify(resolverCtx, resolverCtx.MaxTime, r.grain, r.timeZone, r.isComplete)
	}

	start := resolverCtx.Now
	if r.End != nil && r.End.Latest {
		// if end has latest mentioned then start also should be relative to latest.
		start = resolverCtx.MaxTime
	}

	if r.Start != nil {
		start = r.Start.Modify(resolverCtx, start, r.grain, r.timeZone, r.isComplete)
	}

	end := resolverCtx.Now
	if r.End != nil {
		end = r.End.Modify(resolverCtx, end, r.grain, r.timeZone, r.isComplete)
	}

	return start, end, nil
}

func (t *TimeAnchor) Modify(resolverCtx ResolverContext, tm time.Time, tg timeutil.TimeGrain, tz *time.Location, isComplete bool) time.Time {
	isTruncate := true
	truncateGrain := tg

	if t.Now {
		tm = resolverCtx.Now.In(tz)
		isTruncate = isComplete
	} else if t.Earliest {
		tm = resolverCtx.MinTime.In(tz)
		isTruncate = true
	} else if t.Latest {
		tm = resolverCtx.MaxTime.In(tz)
		isTruncate = isComplete
	} else if t.AbsDate != nil {
		absTm, _ := time.Parse(time.DateOnly, *t.AbsDate)
		tm = absTm.In(tz)
	} else if t.AbsTime != nil {
		absTm, _ := time.Parse("2006-01-02 15:04", *t.AbsTime)
		tm = absTm.In(tz)
	} else if t.Grain != nil {
		n := 0
		if t.Grain.Num != nil {
			n = *t.Grain.Num
		}

		tm = tm.In(tz)
		switch t.Grain.Grain {
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

		truncateGrain = grainMap[t.Grain.Grain]
		isTruncate = true
	} else {
		return tm.In(tz)
	}

	if t.Trunc != nil {
		truncateGrain = grainMap[*t.Trunc]
		isTruncate = true
	}

	if isTruncate {
		return timeutil.TruncateTime(tm, truncateGrain, tz, resolverCtx.FirstDay, resolverCtx.FirstMonth)
	}
	return timeutil.CeilTime(tm, truncateGrain, tz, resolverCtx.FirstDay, resolverCtx.FirstMonth)
}
