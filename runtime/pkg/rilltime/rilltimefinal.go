package rilltime

import (
	"time"

	"github.com/alecthomas/participle/v2"
	"github.com/rilldata/rill/runtime/pkg/timeutil"
)

var (
	rillTimeParserFinal = participle.MustBuild[ExpressionFinal](
		participle.Lexer(rillTimeLexer),
		participle.Elide("Whitespace"),
		participle.UseLookahead(3), // TODO: try and avoid this
	)
)

type ExpressionFinal struct {
	Interval *Interval `parser:"@@"`

	timeZone *time.Location
}

type Interval struct {
	Ordinal        *OrdinalInterval        `parser:"( @@"`
	StartingEnding *StartingEndingInterval `parser:"| @@"`
	StartEnd       *StartEndInterval       `parser:"| @@)"`
}

type OrdinalInterval struct {
	Parts             []*IntervalPart   `parser:"@@ (Of @@)*"`
	IntervalAnchor    *StartEndInterval `parser:"( Of @@"`
	PointInTimeAnchor *PointInTimeList  `parser:"| As Of @@)"`
}

type StartEndInterval struct {
	Start    *PointInTimeList    `parser:"( @@"`
	End      *PointInTimeList    `parser:"To @@"`
	Duration *DurationToInterval `parser:"| @@)"`
}

type StartingEndingInterval struct {
	Duration    *Duration        `parser:"@@"`
	Starting    bool             `parser:"( @Starting"`
	Ending      bool             `parser:"| @Ending)"`
	PointInTime *PointInTimeList `parser:"@@"`
}

type IntervalPart struct {
	Ordinal *OrdinalIntervalPart `parser:"( @@"`
	Snapped *SnappedIntervalPart `parser:"| @@)"`
}

type OrdinalIntervalPart struct {
	Grain string `parser:"@Grain"`
	Num   int    `parser:"@Number"`
}

type SnappedIntervalPart struct {
	Prefix string `parser:"@SnapPrefix"`
	Num    int    `parser:"@Number"`
	Grain  string `parser:"@Grain"`
}

type DurationToInterval struct {
	Prefix   *string   `parser:"@Prefix?"`
	Duration *Duration `parser:"@@ Interval"`
}

type PointInTimeList struct {
	PointInTimes []*PointInTime `parser:"@@ @@*"`
}

type PointInTime struct {
	Relative *RelativePointInTime `parser:"( @@"`
	Labelled *LabelledPointInTime `parser:"| @@)"`
}

type RelativePointInTime struct {
	Prefix   *string   `parser:"@Prefix?"`
	Duration *Duration `parser:"@@"`
	Snap     *string   `parser:"(Snap @Grain)?"`
	Suffix   *string   `parser:"@Suffix?"`
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

	cur := evalOpts.Watermark

	if e.Interval != nil {
		return e.Interval.eval(evalOpts, cur, e.timeZone)
	}

	return cur, cur, timeutil.TimeGrainUnspecified
}

func (i *Interval) eval(evalOpts EvalOptions, tm time.Time, tz *time.Location) (time.Time, time.Time, timeutil.TimeGrain) {
	if i.Ordinal != nil {
		return i.Ordinal.eval(evalOpts, tm, tz)
	} else if i.StartEnd != nil {
		return i.StartEnd.eval(evalOpts, tm, tz)
	} else if i.StartingEnding != nil {
		return i.StartingEnding.eval(evalOpts, tm, tz)
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
	end := tm
	i := len(o.Parts) - 1
	for i >= 0 {
		start, end = o.Parts[i].eval(evalOpts, start, end, tz)
		i--
	}

	tg := o.Parts[0].grain()

	return start, end, lowerOrderMap[tg]
}

func (s *StartEndInterval) eval(evalOpts EvalOptions, tm time.Time, tz *time.Location) (time.Time, time.Time, timeutil.TimeGrain) {
	if s.Duration != nil {
		return s.Duration.eval(evalOpts, tm, tz)
	}

	start, startTg := s.Start.eval(evalOpts, tm, tz)
	end, endTg := s.End.eval(evalOpts, tm, tz)
	tg := endTg
	if endTg == timeutil.TimeGrainUnspecified {
		tg = startTg
	}
	return start, end, tg
}

func (p *PointInTimeList) eval(evalOpts EvalOptions, tm time.Time, tz *time.Location) (time.Time, timeutil.TimeGrain) {
	tg := timeutil.TimeGrainUnspecified
	for _, pt := range p.PointInTimes {
		tm, tg = pt.eval(evalOpts, tm, tz)
	}
	return tm, tg
}

func (p *PointInTime) eval(evalOpts EvalOptions, tm time.Time, tz *time.Location) (time.Time, timeutil.TimeGrain) {
	tg := timeutil.TimeGrainUnspecified
	if p.Relative != nil {
		tm, tg = p.Relative.eval(evalOpts, tm, tz)
	} else if p.Labelled != nil {
		tm = p.Labelled.eval(evalOpts, tm)
	}
	return tm, tg
}

func (s *StartingEndingInterval) eval(evalOpts EvalOptions, tm time.Time, tz *time.Location) (time.Time, time.Time, timeutil.TimeGrain) {
	start, _ := s.PointInTime.eval(evalOpts, tm, tz)
	end := start

	num := 0
	if s.Duration.Num != nil {
		num = *s.Duration.Num
	}

	tg := grainMap[s.Duration.Grain]

	if s.Starting {
		end = timeutil.OffsetTime(start, tg, num)
	} else if s.Ending {
		start = timeutil.OffsetTime(end, tg, -num)
	}

	return start, end, tg
}

func (i *DurationToInterval) eval(evalOpts EvalOptions, start time.Time, tz *time.Location) (time.Time, time.Time, timeutil.TimeGrain) {
	sign := -1
	if i.Prefix != nil && *i.Prefix == "+" {
		sign = 1
	}

	tg := grainMap[i.Duration.Grain]

	start = i.Duration.Offset(start, sign)
	start = truncateWithCorrection(start, tg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)

	end := timeutil.OffsetTime(start, tg, 1)

	return start, end, lowerOrderMap[tg]
}

func (i *IntervalPart) eval(evalOpts EvalOptions, start, end time.Time, tz *time.Location) (time.Time, time.Time) {
	if i.Ordinal != nil {
		return i.Ordinal.eval(evalOpts, start, tz)
	} else if i.Snapped != nil {
		return i.Snapped.eval(evalOpts, start, end, tz)
	}
	return start, end
}

func (i *IntervalPart) grain() timeutil.TimeGrain {
	if i.Ordinal != nil {
		return grainMap[i.Ordinal.Grain]
	} else if i.Snapped != nil {
		return grainMap[i.Snapped.Grain]
	}
	return timeutil.TimeGrainUnspecified
}

func (o *OrdinalIntervalPart) eval(evalOpts EvalOptions, start time.Time, tz *time.Location) (time.Time, time.Time) {
	tg := grainMap[o.Grain]
	offset := o.Num - 1

	start = timeutil.OffsetTime(start, tg, offset)
	start = truncateWithCorrection(start, tg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)

	end := timeutil.OffsetTime(start, tg, 1)

	return start, end
}

func (s *SnappedIntervalPart) eval(evalOpts EvalOptions, start, end time.Time, tz *time.Location) (time.Time, time.Time) {
	tg := grainMap[s.Grain]

	if s.Prefix == "<" {
		// Anchor the range to the beginning of the higher order start
		// EG: <4d of M : gives 1st 4 days of the current month regardless of current date.

		// Anchoring to start should follow week rules https://en.wikipedia.org/wiki/ISO_week_date#First_week
		start = truncateWithCorrection(start, tg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)
		end = timeutil.OffsetTime(start, tg, s.Num)
		end = timeutil.TruncateTime(end, tg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)
	} else {
		// Anchor the range to the end of the higher order end
		// EG: >4d of M : gives last 4 days of the current month regardless of current date.
		end = timeutil.CeilTime(end, tg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)
		start = timeutil.TruncateTime(end, tg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)
		start = timeutil.OffsetTime(start, tg, -s.Num)
	}

	return start, end
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

	if r.Suffix != nil {
		// `$` suffix means snap to end. So add 1 to the offset before truncating.
		if *r.Suffix == "$" {
			tm = timeutil.OffsetTime(tm, tg, 1)
		}
		tm = truncateWithCorrection(tm, tg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)
	}

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
