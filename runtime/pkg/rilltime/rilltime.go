package rilltime

import (
	"fmt"

	"github.com/alecthomas/participle/v2"
)

// RillTime => Time (, Time)? (: Modifiers)?
type RillTime struct {
	Start     *Time      `parser:"  @@"`
	End       *Time      `parser:"(',' @@)?"`
	Modifiers *Modifiers `parser:"(':' @@)?"`
}

// Time => -?[1-9][0-9]* (/\W+)?
//       | now (/\W+)?
type Time struct {
	Neg   bool    `parser:"( @'-'?"`
	Num   *int    `parser:"  @Int"`
	Grain *string `parser:"  @Ident"`
	Now   bool    `parser:"| @'now')"`
	Trunc *string `parser:"  ('/' @Ident)?"`
}

// Modifiers => \W+ | |\W+|
type Modifiers struct {
	Grain  *string `parser:"(@Ident | '|' @Ident '|')?"`
	Offset *Time   `parser:"('@' @@)?"`
}

func (t *RillTime) String() string {
	time := ""
	if t.Start != nil {
		time = t.Start.String()
	}
	if t.End != nil {
		time += ", " + t.End.String()
	}
	if t.Modifiers != nil {
		time += " :" + t.Modifiers.String()
	}
	return time
}

func (t *Time) String() string {
	time := ""
	if t.Now {
		time = "now"
	} else {
		if t.Neg {
			time += "-"
		}
		if t.Num != nil {
			time += fmt.Sprintf("%d", *t.Num)
		}
		if t.Grain != nil {
			time += *t.Grain
		}
	}
	if t.Trunc != nil {
		time += "/" + *t.Trunc
	}
	return time
}

func (m *Modifiers) String() string {
	time := ""
	if m.Grain != nil {
		time += " |" + *m.Grain + "|"
	}
	if m.Offset != nil {
		time += " @" + m.Offset.String()
	}
	return time
}

func Parse(timeRange string) (*RillTime, error) {
	parser, err := participle.Build[RillTime]()
	if err != nil {
		return nil, err
	}
	ast, err := parser.ParseString("", timeRange)
	if err != nil {
		return nil, err
	}
	return ast, nil
}
