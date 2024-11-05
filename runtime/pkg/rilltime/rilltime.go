package rilltime

import (
	"fmt"
	"time"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/rilldata/rill/runtime/pkg/timeutil"
)

var (
	rillTimeLexer = lexer.MustSimple([]lexer.SimpleRule{
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
	rillTimeParser = participle.MustBuild[RillTime](
		participle.Lexer(rillTimeLexer),
		participle.Elide("Whitespace"),
	)
	daxOffsetRangeNotations = map[string]timeutil.TimeGrain{
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
	Start      *TimeModifier `parser:"  @@"`
	End        *TimeModifier `parser:"(',' @@)?"`
	Modifiers  *Modifiers    `parser:"(':' @@)?"`
	Grain      timeutil.TimeGrain
	IsComplete bool
	TimeZone   *time.Location
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

func Parse(timeRange string) (*RillTime, error) {
	ast, err := rillTimeParser.ParseString("", timeRange)
	if err != nil {
		return nil, err
	}

	ast.TimeZone = time.UTC
	if ast.Modifiers != nil {
		if ast.Modifiers.Grain != nil {
			ast.Grain = daxOffsetRangeNotations[ast.Modifiers.Grain.Grain]
			// TODO: non-1 grains
		} else if ast.Modifiers.CompleteGrain != nil {
			ast.Grain = daxOffsetRangeNotations[ast.Modifiers.CompleteGrain.Grain]
			// TODO: non-1 grains
			ast.IsComplete = true
		}

		if ast.Modifiers.At != nil {
			if ast.Modifiers.At.TimeZone != nil {
				var err error
				ast.TimeZone, err = time.LoadLocation(*ast.Modifiers.At.TimeZone)
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
	grain := r.Grain

	if tm.Now {
		t = resolverCtx.Now.In(r.TimeZone)
		isTruncate = r.IsComplete
	} else if tm.Earliest {
		t = resolverCtx.MinTime.In(r.TimeZone)
	} else if tm.Latest {
		t = resolverCtx.MaxTime.In(r.TimeZone)
		isTruncate = r.IsComplete
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

		t = t.In(r.TimeZone)
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

		grain = daxOffsetRangeNotations[g]
	}

	if isTruncate {
		return timeutil.TruncateTime(t, grain, r.TimeZone, resolverCtx.FirstDay, resolverCtx.FirstMonth)
	}
	return timeutil.CeilTime(t, grain, r.TimeZone, resolverCtx.FirstDay, resolverCtx.FirstMonth)
}
