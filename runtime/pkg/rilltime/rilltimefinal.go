package rilltime

import (
	"fmt"
	"time"

	"github.com/alecthomas/participle/v2"
	"github.com/rilldata/rill/runtime/pkg/timeutil"
)

var (
	rillTimeParserFinal = participle.MustBuild[ExpressionFinal](
		participle.Lexer(rillTimeLexer),
		participle.Elide("Whitespace"),
	)
	snapToStart = "^"
	snapToEnd   = "$"
)

type ExpressionFinal struct {
	Interval *Interval `parser:"@@"`

	timeZone *time.Location
}

type Interval struct {
	Ordinal  *OrdinalInterval  `parser:"( @@"`
	StartEnd *StartEndInterval `parser:"| @@)"`
}

type OrdinalInterval struct {
	Parts             []*OrdinalIntervalPart `parser:"@@ (Of @@)*"`
	IntervalAnchor    *StartEndInterval      `parser:"( Of @@"`
	PointInTimeAnchor *PointInTime           `parser:"| As Of @@)"`
}

type OrdinalIntervalPart struct {
	Grain string `parser:"@Grain"`
	Num   int    `parser:"@Number"`
}

type StartEndInterval struct {
	Start      *PointInTime `parser:"@@"`
	EndOfStart bool         `parser:"( '!'"`
	End        *PointInTime `parser:"| To @@)"`
}

type PointInTime struct {
	Relative *RelativePointInTime `parser:"( @@"`
	Labelled *LabelledPointInTime `parser:"| @@)"`
}

type RelativePointInTime struct {
	Prefix   *string   `parser:"@Prefix?"`
	Duration *Duration `parser:"@@"`
	Snap     *string   `parser:"(Snap @Grain)?"`
	Suffix   string    `parser:"@Suffix"`
}

type LabelledPointInTime struct {
	Earliest  bool `parser:"( @Earliest"`
	Now       bool `parser:"| @Now"`
	Latest    bool `parser:"| @Latest"`
	Watermark bool `parser:"| @Watermark)"`
}

type Duration struct {
	Num   *int   `parser:"@Number?"`
	Grain string `parser:"@Grain"`
}

func ParseFinal(from string, parseOpts ParseOptions) (*ExpressionFinal, error) {
	//tokens, err := rillTimeParserFinal.Lex("", strings.NewReader(from))
	//if err != nil {
	//	return nil, err
	//}
	//for _, token := range tokens {
	//	fmt.Println(token.Type, token.Value)
	//}

	rt, err := rillTimeParserFinal.ParseString("", from)
	if err != nil {
		return nil, err
	}

	err = rt.parse(parseOpts)
	if err != nil {
		return nil, err
	}

	return rt, nil
}

func (e *ExpressionFinal) parse(parseOpts ParseOptions) error {
	e.timeZone = time.UTC
	if parseOpts.TimeZoneOverride != nil {
		e.timeZone = parseOpts.TimeZoneOverride
	} else if parseOpts.DefaultTimeZone != nil {
		e.timeZone = parseOpts.DefaultTimeZone
	}

	if e.Interval != nil {
		return e.Interval.parse()
	}
	return nil
}

func (e *ExpressionFinal) Eval(evalOpts EvalOptions) (time.Time, time.Time, timeutil.TimeGrain) {
	if evalOpts.FirstDay == 0 {
		evalOpts.FirstDay = 1
	}
	if evalOpts.FirstMonth == 0 {
		evalOpts.FirstMonth = 1
	}

	cur := evalOpts.Watermark

	if e.Interval != nil {
		return e.Interval.eval(evalOpts, cur, e.timeZone)
	}

	return cur, cur, timeutil.TimeGrainUnspecified
}

func (i *Interval) parse() error {
	if i.StartEnd != nil {
		return i.StartEnd.parse()
	}
	return nil
}

func (i *Interval) eval(evalOpts EvalOptions, tm time.Time, tz *time.Location) (time.Time, time.Time, timeutil.TimeGrain) {
	if i.StartEnd != nil {
		return i.StartEnd.eval(evalOpts, tm, tz)
	} else if i.Ordinal != nil {
		return i.Ordinal.eval(evalOpts, tm, tz)
	}
	return tm, tm, timeutil.TimeGrainUnspecified
}

func (o *OrdinalInterval) eval(evalOpts EvalOptions, tm time.Time, tz *time.Location) (time.Time, time.Time, timeutil.TimeGrain) {
	if o.PointInTimeAnchor != nil {
		tm, _ = o.PointInTimeAnchor.eval(evalOpts, tm, tz)
	} else if o.IntervalAnchor != nil {
		tm, _, _ = o.IntervalAnchor.eval(evalOpts, tm, tz)
	}

	start := tm
	for _, part := range o.Parts {
		start = part.eval(evalOpts, start, tz)
	}
	i := len(o.Parts) - 1
	for i >= 0 {
		start = o.Parts[i].eval(evalOpts, start, tz)
		i--
	}

	tg := grainMap[o.Parts[0].Grain]

	end := timeutil.OffsetTime(start, tg, 1)

	return start, end, lowerOrderMap[tg]
}

func (o *OrdinalIntervalPart) eval(evalOpts EvalOptions, tm time.Time, tz *time.Location) time.Time {
	tg := grainMap[o.Grain]
	offset := o.Num - 1

	tm = timeutil.OffsetTime(tm, tg, offset)
	tm = truncateWithCorrection(tm, tg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)

	return tm
}

func (s *StartEndInterval) parse() error {
	if !s.EndOfStart {
		return nil
	}

	if s.Start.Relative == nil {
		return fmt.Errorf("start must be relative")
	}

	s.Start.Relative.Snap = &snapToStart
	s.End = &PointInTime{
		Relative: &RelativePointInTime{
			Prefix:   s.Start.Relative.Prefix,
			Duration: s.Start.Relative.Duration,
			Snap:     &snapToEnd,
			Suffix:   s.Start.Relative.Suffix,
		},
	}

	return nil
}

func (s *StartEndInterval) eval(evalOpts EvalOptions, tm time.Time, tz *time.Location) (time.Time, time.Time, timeutil.TimeGrain) {
	start, startTg := s.Start.eval(evalOpts, tm, tz)
	end, endTg := s.End.eval(evalOpts, tm, tz)
	tg := endTg
	if endTg == timeutil.TimeGrainUnspecified {
		tg = startTg
	}
	return start, end, tg
}

func (p *PointInTime) eval(evalOpts EvalOptions, tm time.Time, tz *time.Location) (time.Time, timeutil.TimeGrain) {
	if p.Relative != nil {
		return p.Relative.eval(evalOpts, tm, tz)
	} else if p.Labelled != nil {
		return p.Labelled.eval(evalOpts, tm), timeutil.TimeGrainUnspecified
	}
	return tm, timeutil.TimeGrainUnspecified
}

func (r *RelativePointInTime) eval(evalOpts EvalOptions, tm time.Time, tz *time.Location) (time.Time, timeutil.TimeGrain) {
	sign := -1
	if r.Prefix != nil && *r.Prefix == "+" {
		sign = 1
	}
	tm = r.Duration.Offset(tm, sign)

	tg := grainMap[r.Duration.Grain]
	if r.Snap != nil {
		tg = grainMap[*r.Snap]
	}

	// `$` suffix means snap to end. So add 1 to the offset before truncating.
	if r.Suffix == "$" {
		tm = timeutil.OffsetTime(tm, tg, 1)
	}
	tm = truncateWithCorrection(tm, tg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)

	return tm, tg
}

func (l *LabelledPointInTime) eval(evalOpts EvalOptions, tm time.Time) time.Time {
	if l.Earliest {
		return evalOpts.MinTime
	} else if l.Now {
		return evalOpts.Now
	} else if l.Latest {
		return evalOpts.MaxTime
	} else if l.Watermark {
		return evalOpts.Watermark
	}
	return tm
}

func (d *Duration) Offset(tm time.Time, sign int) time.Time {
	n := 0
	if d.Num != nil {
		n = *d.Num
	}
	n *= sign

	tg := grainMap[d.Grain]
	return timeutil.OffsetTime(tm, tg, n)
}
