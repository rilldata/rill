package rilltime

import (
	"strings"
	"time"

	"github.com/alecthomas/participle/v2"
	"github.com/rilldata/rill/runtime/pkg/timeutil"
)

var (
	rillTimeParserFinal = participle.MustBuild[ExpressionFinal](
		participle.Lexer(rillTimeLexer),
		participle.Elide("Whitespace"),
		participle.UseLookahead(-1), // TODO: try and avoid this
	)
)

type ExpressionFinal struct {
	Interval       *Interval         `parser:"@@"`
	AnchorOverride *GrainPointInTime `parser:"(As Of @@)?"`
	Grain          *string           `parser:"(By @Grain)?"`
	TimeZone       *string           `parser:"('@' @TimeZone)?"`

	timeZone *time.Location
}

type Interval struct {
	AnchoredDuration *AnchoredDurationInterval `parser:"( @@"`
	Ordinal          *OrdinalInterval          `parser:"| @@"`
	StartEnd         *StartEndInterval         `parser:"| @@"`
	Interval         *GrainToInterval          `parser:"| @@)"`
}

type AnchoredDurationInterval struct {
	Duration    *GrainDur `parser:"@@"`
	Starting    bool      `parser:"( @Starting"`
	Ending      bool      `parser:"| @Ending)"`
	PointInTime *Point    `parser:"@@"`
}

type OrdinalInterval struct {
	Ordinal *OrdinalDuration    `parser:"@@"`
	End     *OrdinalIntervalEnd `parser:"(Of @@)?"`
}

type OrdinalIntervalEnd struct {
	Grains      *GrainToInterval  `parser:"( @@"`
	Interval    *StartEndInterval `parser:"| @@"`
	SingleGrain *string           `parser:"| @Grain)"`
}

type StartEndInterval struct {
	Start *Point `parser:"@@"`
	End   *Point `parser:"To @@"`
}

type GrainToInterval struct {
	Interval *GrainPointInTime `parser:"@@ Interval"`
}

type Point struct {
	Ordinal *OrdinalPointInTime `parser:"( @@"`
	Grain   *GrainPointInTime   `parser:"| @@"`
	Labeled *LabeledPointInTime `parser:"| @@)"`
}

type OrdinalPointInTime struct {
	Ordinal *Ord             `parser:"@@"`
	Suffix  string           `parser:"@Suffix"`
	Rest    *OrdinalDuration `parser:"@@?"`
}

type GrainPointInTime struct {
	Parts []*GrainPointInTimePart `parser:"@@ @@*"`
}

type GrainPointInTimePart struct {
	Prefix   *string   `parser:"@Prefix?"`
	Duration *GrainDur `parser:"@@"`
	Snap     *string   `parser:"(Snap @Grain)?"`
	Suffix   *string   `parser:"@Suffix?"`
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
	Ordinal       *Ord      `parser:"( @@"`
	Snap          *string   `parser:"| @SnapPrefix"`
	GrainDuration *GrainDur `parser:"  @@)"`
}

type Ord struct {
	Grain string `parser:"@Grain"`
	Num   int    `parser:"@Number"`
}

type GrainDur struct {
	Parts []*GrainDurPart `parser:"@@ @@*"`
}

type GrainDurPart struct {
	Num   *int   `parser:"@Number?"`
	Grain string `parser:"@Grain"`
}

func ParseFinal(from string, parseOpts ParseOptions) (*ExpressionFinal, error) {
	//tokens, err := rillTimeParserFinal.Lex("", strings.NewReader(from))
	//if err != nil {
	//	return nil, err
	//}
	//syms := rillTimeParserFinal.Lexer().Symbols()
	//for _, token := range tokens {
	//	typeName := ""
	//	for tn, tt := range syms {
	//		if tt == token.Type {
	//			typeName = tn
	//			break
	//		}
	//	}
	//
	//	fmt.Println(typeName, token.Value)
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
	} else if e.TimeZone != nil {
		var err error
		e.timeZone, err = time.LoadLocation(strings.Trim(*e.TimeZone, "{}"))
		if err != nil {
			return err
		}
	} else if parseOpts.DefaultTimeZone != nil {
		e.timeZone = parseOpts.DefaultTimeZone
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

	start := evalOpts.Watermark

	if e.Interval != nil {
		return e.Interval.eval(evalOpts, start, e.timeZone)
	}

	return start, start, timeutil.TimeGrainUnspecified
}

/* Intervals */

func (i *Interval) eval(evalOpts EvalOptions, start time.Time, tz *time.Location) (time.Time, time.Time, timeutil.TimeGrain) {
	if i.AnchoredDuration != nil {
		return i.AnchoredDuration.eval(evalOpts, start, tz)
	} else if i.Ordinal != nil {
		return i.Ordinal.eval(evalOpts, start, tz)
	} else if i.StartEnd != nil {
		return i.StartEnd.eval(evalOpts, start, tz)
	} else if i.Interval != nil {
		return i.Interval.eval(evalOpts, start, tz)
	}
	return start, start, timeutil.TimeGrainUnspecified
}

func (o *AnchoredDurationInterval) eval(evalOpts EvalOptions, tm time.Time, tz *time.Location) (time.Time, time.Time, timeutil.TimeGrain) {
	start, _ := o.PointInTime.eval(evalOpts, tm, tz)
	end := start

	tg := timeutil.TimeGrainUnspecified
	if o.Starting {
		end, tg = o.Duration.offset(start, 1)
	} else if o.Ending {
		start, tg = o.Duration.offset(end, -1)
	}

	return start, end, tg
}

func (o *OrdinalInterval) eval(evalOpts EvalOptions, start time.Time, tz *time.Location) (time.Time, time.Time, timeutil.TimeGrain) {
	if o.End != nil {
		start, _, _ = o.End.eval(evalOpts, start, tz)
	}

	// TODO: should end from above be passed here?
	start, end, tg := o.Ordinal.eval(evalOpts, start, tz)

	return start, end, tg
}

func (o *OrdinalIntervalEnd) eval(evalOpts EvalOptions, start time.Time, tz *time.Location) (time.Time, time.Time, timeutil.TimeGrain) {
	if o.Grains != nil {
		return o.Grains.eval(evalOpts, start, tz)
	} else if o.Interval != nil {
		return o.Interval.eval(evalOpts, start, tz)
	} else if o.SingleGrain != nil {
		tg := grainMap[*o.SingleGrain]
		end := timeutil.CeilTime(start, tg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)
		start = truncateWithCorrection(start, tg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)
		return start, end, tg
	}
	return start, start, timeutil.TimeGrainUnspecified
}

func (o *StartEndInterval) eval(evalOpts EvalOptions, tm time.Time, tz *time.Location) (time.Time, time.Time, timeutil.TimeGrain) {
	start, startTg := o.Start.eval(evalOpts, tm, tz)
	end, endTg := o.End.eval(evalOpts, tm, tz)

	tg := endTg
	if endTg == timeutil.TimeGrainUnspecified {
		tg = startTg
	}

	return start, end, tg
}

func (o *GrainToInterval) eval(evalOpts EvalOptions, start time.Time, tz *time.Location) (time.Time, time.Time, timeutil.TimeGrain) {
	if len(o.Interval.Parts) == 0 {
		return start, start, timeutil.TimeGrainUnspecified
	}

	start, tg := o.Interval.eval(evalOpts, start, tz)

	end := timeutil.CeilTime(start, tg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)
	start = truncateWithCorrection(start, tg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)

	return start, end, tg
}

/* Point in times */

func (p *Point) eval(evalOpts EvalOptions, start time.Time, tz *time.Location) (time.Time, timeutil.TimeGrain) {
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
	tg := timeutil.TimeGrainUnspecified
	if o.Rest != nil {
		start, _, tg = o.Rest.eval(evalOpts, start, tz)
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
	sign := -1
	if g.Prefix != nil && *g.Prefix == "+" {
		sign = 1
	}
	tm, tg := g.Duration.offset(start, sign)

	if g.Snap != nil {
		tg = grainMap[*g.Snap]
	}

	if g.Suffix != nil {
		// `$` suffix means snap to end. So add 1 to the offset before truncating.
		if *g.Suffix == "$" {
			tm = timeutil.OffsetTime(tm, tg, 1)
		}
		tm = truncateWithCorrection(tm, tg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)
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

func (o *OrdinalDuration) eval(evalOpts EvalOptions, start time.Time, tz *time.Location) (time.Time, time.Time, timeutil.TimeGrain) {
	end := start
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
		return o.Ordinal.eval(evalOpts, start, tz)
	}

	if o.Snap == nil || o.GrainDuration == nil {
		return time.Time{}, time.Time{}, timeutil.TimeGrainUnspecified
	}

	tg := timeutil.TimeGrainUnspecified
	if *o.Snap == "<" {
		// Anchor the range to the beginning of the higher order start
		// EG: <4d of M : gives 1st 4 days of the current month regardless of current date.
		end, tg = o.GrainDuration.offset(start, 1)
	} else {
		// Anchor the range to the end of the higher order end
		// EG: >4d of M : gives last 4 days of the current month regardless of current date.
		start, tg = o.GrainDuration.offset(end, -1)
	}
	return start, end, tg
}

func (o *Ord) eval(evalOpts EvalOptions, start time.Time, tz *time.Location) (time.Time, time.Time, timeutil.TimeGrain) {
	tg := grainMap[o.Grain]
	offset := o.Num - 1

	start = timeutil.OffsetTime(start, tg, offset)
	start = truncateWithCorrection(start, tg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)

	end := timeutil.OffsetTime(start, tg, 1)

	return start, end, tg
}

func (g *GrainDur) offset(tm time.Time, sign int) (time.Time, timeutil.TimeGrain) {
	tg := timeutil.TimeGrainUnspecified
	for _, part := range g.Parts {
		tm, tg = part.offset(tm, sign)
	}
	return tm, tg
}

func (g *GrainDurPart) offset(tm time.Time, sign int) (time.Time, timeutil.TimeGrain) {
	tg := grainMap[g.Grain]
	offset := 0
	if g.Num != nil {
		offset = *g.Num
	}
	offset *= sign

	return timeutil.OffsetTime(tm, tg, offset), tg
}
