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
	Num      *int    `parser:"( @Number?"`
	Grain    *string `parser:"  @Grain"`
	Now      bool    `parser:"| @Now"`
	Latest   bool    `parser:"| @Latest"`
	Earliest bool    `parser:"| @Earliest)"`
	Trunc    *string `parser:"  ('/' @Grain)?"`
}

type Modifiers struct {
	Grain         *GrainModifier `parser:"( @@"`
	CompleteGrain *GrainModifier `parser:"| '|' @@ '|')?"`
	At            *AtModifiers   `parser:"( '@' @@)?"`
}

type GrainModifier struct {
	Num   *int   `parser:"(@Number)?"`
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
		Start: &TimeAnchor{},
		End:   &TimeAnchor{Now: true},
	}
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
		rt.Start.Num = &val
		var g string
		switch name {
		case "year":
			g = "Y"
		case "month":
			g = "M"
		case "week":
			g = "W"
		case "day":
			g = "d"
		case "hour":
			g = "h"
		case "minute":
			g = "m"
		case "second":
			g = "s"
		default:
			return nil, fmt.Errorf("unexpected field %q in duration", name)
		}
		rt.Start.Grain = &g
	}

	return rt, nil
}

func (r *RillTime) Resolve(resolverCtx ResolverContext) (time.Time, time.Time, error) {
	if r.Modifiers != nil && r.Modifiers.At != nil && r.Modifiers.At.Offset != nil {
		resolverCtx.Now = r.ModifyTime(resolverCtx, resolverCtx.Now, r.Modifiers.At.Offset)
		resolverCtx.MinTime = r.ModifyTime(resolverCtx, resolverCtx.MinTime, r.Modifiers.At.Offset)
		resolverCtx.MaxTime = r.ModifyTime(resolverCtx, resolverCtx.MaxTime, r.Modifiers.At.Offset)
	}

	start := resolverCtx.Now
	if r.End != nil && r.End.Latest {
		// if end has latest mentioned then start also should be relative to latest.
		start = resolverCtx.MaxTime
	}

	if r.Start != nil {
		start = r.ModifyTime(resolverCtx, start, r.Start)
	}

	end := resolverCtx.Now
	if r.End != nil {
		end = r.ModifyTime(resolverCtx, end, r.End)
	}

	return start, end, nil
}

func (r *RillTime) ModifyTime(resolverCtx ResolverContext, t time.Time, tm *TimeAnchor) time.Time {
	isTruncate := true
	truncateGrain := r.grain

	if tm.Now {
		t = resolverCtx.Now.In(r.timeZone)
		isTruncate = r.isComplete
	} else if tm.Earliest {
		t = resolverCtx.MinTime.In(r.timeZone)
		isTruncate = true
	} else if tm.Latest {
		t = resolverCtx.MaxTime.In(r.timeZone)
		isTruncate = r.isComplete
	} else {
		n := 0
		if tm.Num != nil {
			n = *tm.Num
		}
		// TODO: what should the defaults here be?
		g := "s"
		if tm.Grain != nil {
			g = *tm.Grain
		}

		t = t.In(r.timeZone)
		switch g {
		case "s":
			t = t.Add(time.Duration(n) * time.Second)
		case "m":
			t = t.Add(time.Duration(n) * time.Minute)
		case "h":
			t = t.Add(time.Duration(n) * time.Hour)
		case "d":
			t = t.AddDate(0, 0, n)
		case "D":
			t = t.AddDate(0, 0, n)
		case "W":
			t = t.AddDate(0, 0, n*7)
		case "M":
			t = t.AddDate(0, n, 0)
		case "Q":
			t = t.AddDate(0, n*3, 0)
		case "Y":
			t = t.AddDate(n, 0, 0)
		}

		truncateGrain = grainMap[g]
		isTruncate = true
	}

	if tm.Trunc != nil {
		truncateGrain = grainMap[*tm.Trunc]
		isTruncate = true
	}

	if isTruncate {
		return timeutil.TruncateTime(t, truncateGrain, r.timeZone, resolverCtx.FirstDay, resolverCtx.FirstMonth)
	}
	return timeutil.CeilTime(t, truncateGrain, r.timeZone, resolverCtx.FirstDay, resolverCtx.FirstMonth)
}
