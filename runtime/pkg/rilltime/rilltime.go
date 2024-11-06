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
		{"Now", `now`},
		{"Latest", `latest`},
		{"Earliest", `earliest`},
		// this needs to be after Now and Latest to match to them
		{"Grain", `[smhdDWQMY]`},
		// this has to be at the end
		{"String", `[a-zA-Z]\w*`},
		{"Number", `[-+]?\d+`},
		// needed for random chars used
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
	Start       *TimeModifier `parser:"  @@"`
	End         *TimeModifier `parser:"(',' @@)?"`
	Modifiers   *Modifiers    `parser:"(':' @@)?"`
	IsNewFormat bool
	grain       timeutil.TimeGrain
	isComplete  bool
	timeZone    *time.Location
}

type TimeModifier struct {
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
	Offset   *TimeModifier `parser:"@@?"`
	TimeZone *string       `parser:"('{' @String '}')?"`
}

type ResolverContext struct {
	Now        time.Time
	MinTime    time.Time
	MaxTime    time.Time
	FirstDay   int
	FirstMonth int
}

func Parse(from string) (*RillTime, error) {
	isoRange, err := ParseISO(from, false)
	if err != nil {
		return nil, err
	}
	if isoRange != nil {
		return isoRange, nil
	}

	ast, err := rillTimeParser.ParseString("", from)
	if err != nil {
		return nil, err
	}
	ast.IsNewFormat = true

	ast.timeZone = time.UTC
	if ast.Modifiers != nil {
		if ast.Modifiers.Grain != nil {
			ast.grain = grainMap[ast.Modifiers.Grain.Grain]
			// TODO: non-1 grains
		} else if ast.Modifiers.CompleteGrain != nil {
			ast.grain = grainMap[ast.Modifiers.CompleteGrain.Grain]
			// TODO: non-1 grains
			ast.isComplete = true
		}

		if ast.Modifiers.At != nil {
			if ast.Modifiers.At.TimeZone != nil {
				var err error
				ast.timeZone, err = time.LoadLocation(*ast.Modifiers.At.TimeZone)
				if err != nil {
					return nil, fmt.Errorf("invalid time zone %q: %w", *ast.Modifiers.At.TimeZone, err)
				}
			}
		}
	}

	if ast.End == nil {
		ast.End = &TimeModifier{
			Now: true,
		}
	}

	return ast, nil
}

func ParseISO(from string, strict bool) (*RillTime, error) {
	// Try parsing for "inf"
	if infPattern.MatchString(from) {
		return &RillTime{
			Start: &TimeModifier{Earliest: true},
			End:   &TimeModifier{Latest: true},
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
		Start: &TimeModifier{},
		End:   &TimeModifier{Now: true},
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

	start := resolverCtx.MaxTime
	if r.Start != nil {
		start = r.ModifyTime(resolverCtx, start, r.Start)
	}

	end := resolverCtx.MaxTime
	if r.End != nil {
		end = r.ModifyTime(resolverCtx, end, r.End)
	}

	return start, end, nil
}

func (r *RillTime) ModifyTime(resolverCtx ResolverContext, t time.Time, tm *TimeModifier) time.Time {
	isTruncate := true
	grain := r.grain

	if tm.Now {
		t = resolverCtx.Now.In(r.timeZone)
		isTruncate = r.isComplete
	} else if tm.Earliest {
		t = resolverCtx.MinTime.In(r.timeZone)
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

		grain = grainMap[g]
	}

	if isTruncate {
		return timeutil.TruncateTime(t, grain, r.timeZone, resolverCtx.FirstDay, resolverCtx.FirstMonth)
	}
	return timeutil.CeilTime(t, grain, r.timeZone, resolverCtx.FirstDay, resolverCtx.FirstMonth)
}
