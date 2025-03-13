package rilltime

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/alecthomas/participle/v2/lexer"
	"github.com/rilldata/rill/runtime/pkg/timeutil"
)

type ExpressionV2 struct {
	From           *Link           `parser:"@@"`
	To             *Link           `parser:"(To @@)?"`
	Grain          *string         `parser:"(By @Grain)?"`
	AnchorOverride *AnchorOverride `parser:"('@' @@)?"`

	timeZone *time.Location
}

// Link represents a link of grains specifying the customisable anchors.
// EG: 7d of -1M : The 7day period of last month. 7day is relative to watermark unless something else is specified.
type Link struct {
	Parts []*LinkPart `parser:"@@ (Of @@)*"`
}

type LinkPart struct {
	Pos lexer.Position

	Ordinal      *Ordinal      `parser:"( @@"`
	Anchor       *TimeAnchor   `parser:"| @@"`
	AbsoluteTime *AbsoluteTime `parser:"| @@)"`
}

type AnchorOverride struct {
	Earliest     bool          `parser:"( @Earliest"`
	Now          bool          `parser:"| @Now"`
	Latest       bool          `parser:"| @Latest"`
	Watermark    bool          `parser:"| @Watermark"`
	AbsoluteTime *AbsoluteTime `parser:"| @@)"`
}

type TimeAnchor struct {
	Pos lexer.Position

	Prefix    *string `parser:"@AnchorPrefix?"`
	Num       *int    `parser:"@Number?"`
	Grain     string  `parser:"@Grain"`
	IsCurrent bool    `parser:"@Current?"`
}

// Ordinal represent a particular sequence of a grain in the next order grain.
// EG: W2 - week 2 of the month.
//     M5 - month 5 of the year.
type Ordinal struct {
	Grain string `parser:"@Grain"`
	Num   int    `parser:"@Number"`
}

type AbsoluteTime struct {
	ISO    string `parser:"@ISOTime"`
	year   int
	month  int
	week   int
	day    int
	hour   int
	minute int
	second int
	tg     timeutil.TimeGrain
}

func ParseV2(from string, parseOpts ParseOptions) (*ExpressionV2, error) {
	//fmt.Println(rillTimeV2Parser.String())
	//tokens, err := rillTimeV2Parser.Lex("", strings.NewReader(from))
	//if err != nil {
	//	return nil, err
	//}
	//for _, token := range tokens {
	//	name := ""
	//	for n, t := range rillTimeV2Parser.Lexer().Symbols() {
	//		if t == token.Type {
	//			name = n
	//			break
	//		}
	//	}
	//	fmt.Println(name, token.Value)
	//}

	rt, err := rillTimeV2Parser.ParseString("", from)
	if err != nil {
		return nil, err
	}

	// TODO: validation per link and link-part
	if rt.From != nil {
		for _, part := range rt.From.Parts {
			if part.AbsoluteTime != nil {
				err = part.AbsoluteTime.parse()
				if err != nil {
					return nil, err
				}
			}
		}
	} else {
		return nil, errors.New("invalid range: missing from")
	}

	if rt.To != nil {
		for _, part := range rt.To.Parts {
			if part.AbsoluteTime != nil {
				err = part.AbsoluteTime.parse()
				if err != nil {
					return nil, err
				}
			}
		}
	}

	if rt.AnchorOverride != nil && rt.AnchorOverride.AbsoluteTime != nil {
		err = rt.AnchorOverride.AbsoluteTime.parse()
		if err != nil {
			return nil, err
		}
	}

	rt.timeZone = time.UTC
	if parseOpts.DefaultTimeZone != nil {
		rt.timeZone = parseOpts.DefaultTimeZone
	}

	return rt, nil
}

func (e *ExpressionV2) Eval(evalOpts EvalOptions) (time.Time, time.Time, timeutil.TimeGrain) {
	anchor := evalOpts.Watermark
	if e.AnchorOverride != nil {
		anchor = e.AnchorOverride.anchor(evalOpts, e.timeZone)
	}

	start, end, tg := e.From.time(evalOpts, anchor, anchor, e.timeZone)
	if e.To != nil {
		_, end, _ = e.To.time(evalOpts, anchor, anchor, e.timeZone)
	}

	if e.Grain != nil {
		tg = grainMap[*e.Grain]
	} else {
		tg = lowerOrderMap[tg]
	}

	return start, end, tg
}

func (l *Link) time(evalOpts EvalOptions, start, end time.Time, tz *time.Location) (time.Time, time.Time, timeutil.TimeGrain) {
	tg := timeutil.TimeGrainUnspecified
	i := len(l.Parts) - 1
	for i >= 0 {
		start, end, tg = l.Parts[i].time(evalOpts, start, end, tz, tg, i == 0)
		i--
	}

	return start, end, tg
}

func (l *LinkPart) time(evalOpts EvalOptions, start, end time.Time, tz *time.Location, tg timeutil.TimeGrain, isFinal bool) (time.Time, time.Time, timeutil.TimeGrain) {
	if l.Anchor != nil {
		return l.Anchor.time(evalOpts, start, end, tz, tg, isFinal)
	} else if l.Ordinal != nil {
		return l.Ordinal.time(evalOpts, start, end, tz, tg)
	} else if l.AbsoluteTime != nil {
		return l.AbsoluteTime.time(tz, isFinal)
	}
	return time.Time{}, time.Time{}, tg
}

func (a *AnchorOverride) anchor(evalOpts EvalOptions, tz *time.Location) time.Time {
	if a.Earliest {
		return evalOpts.MinTime
	} else if a.Now {
		return evalOpts.Now
	} else if a.Latest {
		return evalOpts.MaxTime
	} else if a.Watermark {
		return evalOpts.Watermark
	} else {
		tm, _, _ := a.AbsoluteTime.time(tz, false)
		return tm
	}
}

func (t *TimeAnchor) time(evalOpts EvalOptions, start, end time.Time, tz *time.Location, higherTg timeutil.TimeGrain, isFinal bool) (time.Time, time.Time, timeutil.TimeGrain) {
	num := 1
	if t.Num != nil {
		num = *t.Num
	}

	if t.IsCurrent {
		num -= 1
	}

	curTg := grainMap[t.Grain]
	if higherTg == timeutil.TimeGrainUnspecified {
		higherTg = higherOrderMap[curTg]
	}

	if t.Prefix == nil {
		if num > 0 {
			start = timeutil.OffsetTime(start, curTg, -num)
		}

		start = timeutil.TruncateTime(start, curTg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)

		if !t.IsCurrent {
			end = timeutil.TruncateTime(end, curTg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)
		} else if isFinal {
			end = timeutil.OffsetTime(end, timeutil.TimeGrainSecond, 1)
		}
	} else {
		switch *t.Prefix {
		// -<grain> is used as an offset rather than a range.
		// So we subtract <num> from start and <num-1> from end.
		case "-":
			start = timeutil.OffsetTime(start, curTg, -num)
			if isFinal {
				start = timeutil.TruncateTime(start, curTg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)

				end = timeutil.OffsetTime(end, curTg, -num+1)
				end = timeutil.TruncateTime(end, curTg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)
			} else {
				end = timeutil.OffsetTime(end, curTg, -num)
			}

		// Same with +<grain> is used as an offset.
		// So we add <num-1> to start and <num> to end.
		case "+":
			end = timeutil.OffsetTime(end, curTg, num)
			if isFinal {
				start = timeutil.OffsetTime(start, curTg, num-1)
				start = timeutil.CeilTime(start, curTg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)
				end = timeutil.CeilTime(end, curTg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)
			} else {
				start = timeutil.OffsetTime(start, curTg, num)
			}

		// Anchor the range to the beginning of the higher order grain
		// EG: <4d of M : gives 1st 4 days of the current month regardless of current date.
		case "<":
			start = timeutil.TruncateTime(start, higherTg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)

			end = timeutil.OffsetTime(start, curTg, num)
			end = timeutil.TruncateTime(end, curTg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)

		// Anchor the range to the end of the higher order grain
		// EG: >4d of M : gives last 4 days of the current month regardless of current date.
		case ">":
			end = timeutil.CeilTime(end, higherTg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)

			start = timeutil.OffsetTime(end, curTg, -num)
			start = timeutil.TruncateTime(start, curTg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)
		}
	}

	return start, end, curTg
}

func (o *Ordinal) time(evalOpts EvalOptions, start, end time.Time, tz *time.Location, higherTg timeutil.TimeGrain) (time.Time, time.Time, timeutil.TimeGrain) {
	curTg := grainMap[o.Grain]
	if higherTg == timeutil.TimeGrainUnspecified {
		higherTg = higherOrderMap[curTg]
	}

	start = timeutil.TruncateTime(start, higherTg, tz, evalOpts.FirstDay, evalOpts.FirstMonth)

	offset := o.Num - 1
	if curTg == timeutil.TimeGrainWeek {
		// https://en.wikipedia.org/wiki/ISO_week_date#First_week
		if start.Weekday() >= 5 {
			offset++
		}

		start = timeutil.TruncateTime(start, timeutil.TimeGrainWeek, tz, evalOpts.FirstDay, evalOpts.FirstMonth)
	}

	start = timeutil.OffsetTime(start, curTg, offset)
	end = timeutil.OffsetTime(start, curTg, 1)

	return start, end, curTg
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

	// Since we use this to build a time, month and day cannot be zero
	if a.month == 0 {
		a.month = 1
	}
	if a.day == 0 {
		a.day = 1
	}

	return nil
}

func (a *AbsoluteTime) time(tz *time.Location, isFinal bool) (time.Time, time.Time, timeutil.TimeGrain) {
	start := time.Date(a.year, time.Month(a.month), a.day, a.hour, a.minute, a.second, 0, tz)
	end := start

	if isFinal {
		end = timeutil.OffsetTime(start, a.tg, 1)
	}

	return start, end, a.tg
}
