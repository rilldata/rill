package clickhouse

import (
	"errors"
	"fmt"
	"strings"

	"github.com/rilldata/rill/runtime/drivers"
)

// aggregateFunctionCombinators lists the ClickHouse aggregate function combinators (see system.aggregate_function_combinators).
// Longer names come first so that suffix stripping is unambiguous, e.g. OrNull before Null and SimpleState before State.
var aggregateFunctionCombinators = []string{"SimpleState", "OrDefault", "Distinct", "Resample", "ForEach", "ArgMax", "ArgMin", "OrNull", "Merge", "State", "Array", "Null", "Map", "If"}

// splitKeywords are the identifiers that may appear outside aggregate calls in a measure expression without referencing a column:
// literals, operators and the CASE and INTERVAL syntax. Function names and types after AS or "::" are recognized separately.
var splitKeywords = map[string]bool{
	"null": true, "true": true, "false": true,
	"and": true, "or": true, "not": true, "in": true, "is": true, "like": true, "ilike": true, "between": true,
	"case": true, "when": true, "then": true, "else": true, "end": true, "as": true,
	"interval": true, "nanosecond": true, "microsecond": true, "millisecond": true, "second": true, "minute": true, "hour": true, "day": true, "week": true, "month": true, "quarter": true, "year": true,
}

// SplitAggregateExpression splits a measure expression into partial aggregates for an inner SELECT and a final expression that merges them,
// using the -State and -Merge aggregate function combinators: "sum(a)/sum(b)" becomes the partials "sumState(a)" and "sumState(b)",
// and the final expression "sumMerge(t.m)/sumMerge(t.m__1)". Parameters and arguments are copied as written, so every aggregate function works,
// including combinators ("sumIf") and parametric functions ("quantile(0.5)(x)").
// A measure that already merges states ("uniqMerge(s)") is pre-merged with -MergeState and merged with the function as written.
//
// The DISTINCT keyword needs care: ClickHouse applies it as a -Distinct combinator outside -State, so "countState(DISTINCT x)" holds a plain count
// and merging it adds up the distinct counts of the groups. Instead, "count(DISTINCT x)" becomes "uniqExactState(x)", which is what ClickHouse
// evaluates count(DISTINCT) as by default (the count_distinct_implementation setting), and "f(DISTINCT x)" becomes "fDistinctState(x)",
// whose state keeps the distinct values and merges correctly.
//
// It returns ok=false when the expression cannot be split safely: it has no aggregate call, uses a window, nests aggregates, contains comments,
// produces a state (-State, -SimpleState), or references an identifier outside an aggregate call that is not a keyword, a function or a type.
// The last rule is the safety net for aggregate functions that are missing from aggregateFunctions: their column arguments would be left outside any -State call.
func (d *dialect) SplitAggregateExpression(expr, measureName, innerAlias string) ([]drivers.PartialAggregate, string, bool) {
	toks, err := lexSQLExpression(expr)
	if err != nil {
		return nil, "", false
	}

	var partials []drivers.PartialAggregate
	var final strings.Builder
	pos := 0 // Offset in expr up to which the expression has been copied to final.
	for i := 0; i < len(toks); {
		t := toks[i]
		if t.kind == sqlTokenQuotedIdent {
			// A quoted identifier is always a reference to a column, which cannot be re-aggregated outside an aggregate call.
			return nil, "", false
		}
		if t.kind != sqlTokenIdent {
			i++
			continue
		}

		// A type after AS or "::" is skipped as a whole, including an argument list like Decimal(10, 2) or Nullable(Float64).
		afterAs := i >= 1 && toks[i-1].kind == sqlTokenIdent && strings.EqualFold(toks[i-1].text, "AS")
		afterCast := i >= 2 && toks[i-1].kind == sqlTokenOther && toks[i-2].kind == sqlTokenOther && expr[toks[i-2].start:toks[i-1].end] == "::"
		if afterAs || afterCast {
			i++
			if i < len(toks) && toks[i].kind == sqlTokenLParen {
				i = matchingParen(toks, i) + 1
			}
			continue
		}

		name := t.text
		if i+1 >= len(toks) || toks[i+1].kind != sqlTokenLParen {
			// A bare identifier is either a keyword or a reference to a column outside an aggregate call, which cannot be re-aggregated.
			if !splitKeywords[strings.ToLower(name)] {
				return nil, "", false
			}
			i++
			continue
		}

		combinators, ok := aggregateCombinators(name)
		if !ok {
			// A scalar function call; its arguments are scanned like the rest of the expression.
			i++
			continue
		}
		for j, c := range combinators {
			if c == "State" || c == "SimpleState" || (c == "Merge" && j > 0) {
				return nil, "", false
			}
		}

		// Parse the argument list, and the parameter list before it for parametric functions like quantile(0.5)(x).
		params := ""
		argsStart := i + 1
		argsEnd := matchingParen(toks, argsStart)
		if argsEnd+1 < len(toks) && toks[argsEnd+1].kind == sqlTokenLParen {
			params = expr[toks[argsStart].start:toks[argsEnd].end]
			argsStart = argsEnd + 1
			argsEnd = matchingParen(toks, argsStart)
		}

		// Nested aggregates are invalid SQL; leave the expression alone so the usual error surfaces.
		for j := argsStart + 1; j < argsEnd; j++ {
			if toks[j].kind == sqlTokenIdent && toks[j+1].kind == sqlTokenLParen {
				if _, nested := aggregateCombinators(toks[j].text); nested {
					return nil, "", false
				}
			}
		}

		// Window functions aggregate over frames, not groups.
		if argsEnd+1 < len(toks) && toks[argsEnd+1].kind == sqlTokenIdent && strings.EqualFold(toks[argsEnd+1].text, "OVER") {
			return nil, "", false
		}

		args := expr[toks[argsStart].end:toks[argsEnd].start]

		stateName := name + "State"
		mergeName := name + "Merge"
		if len(combinators) > 0 && combinators[0] == "Merge" {
			mergeName = name
		}
		if argsEnd > argsStart+1 && toks[argsStart+1].kind == sqlTokenIdent && strings.EqualFold(toks[argsStart+1].text, "DISTINCT") {
			if len(combinators) > 0 {
				return nil, "", false
			}
			args = expr[toks[argsStart+2].start:toks[argsEnd].start]
			if strings.EqualFold(name, "count") {
				stateName, mergeName = "uniqExactState", "uniqExactMerge"
			} else {
				stateName, mergeName = name+"DistinctState", name+"DistinctMerge"
			}
		}

		alias := measureName
		if len(partials) > 0 {
			alias = fmt.Sprintf("%s__%d", measureName, len(partials))
		}
		partials = append(partials, drivers.PartialAggregate{
			Alias: alias,
			Expr:  fmt.Sprintf("%s%s(%s)", stateName, params, args),
		})
		final.WriteString(expr[pos:t.start])
		fmt.Fprintf(&final, "%s%s(%s)", mergeName, params, d.EscapeMemberAlias(innerAlias, alias))
		pos = toks[argsEnd].end
		i = argsEnd + 1
	}
	if len(partials) == 0 {
		return nil, "", false
	}
	final.WriteString(expr[pos:])

	return partials, final.String(), true
}

// aggregateCombinators reports whether name is an aggregate function, possibly with combinator suffixes,
// and returns the combinators applied to it from the outermost (last suffix) inwards.
func aggregateCombinators(name string) ([]string, bool) {
	var combinators []string
	for {
		if _, ok := aggregateFunctions[name]; ok || aggregateFunctionsCaseInsensitive[strings.ToLower(name)] {
			return combinators, true
		}
		stripped := false
		for _, c := range aggregateFunctionCombinators {
			if len(name) > len(c) && strings.HasSuffix(name, c) {
				name = strings.TrimSuffix(name, c)
				combinators = append(combinators, c)
				stripped = true
				break
			}
		}
		if !stripped {
			return nil, false
		}
	}
}

// matchingParen returns the index of the parenthesis closing the one at index i.
// lexSQLExpression rejects unbalanced parentheses, so the closing parenthesis always exists.
func matchingParen(toks []sqlToken, i int) int {
	depth := 0
	for j := i; ; j++ {
		switch toks[j].kind {
		case sqlTokenLParen:
			depth++
		case sqlTokenRParen:
			depth--
			if depth == 0 {
				return j
			}
		}
	}
}

type sqlTokenKind int

const (
	sqlTokenIdent sqlTokenKind = iota
	sqlTokenQuotedIdent
	sqlTokenString
	sqlTokenNumber
	sqlTokenLParen
	sqlTokenRParen
	sqlTokenOther // A single character that is not a parenthesis, such as an operator or a comma.
)

// sqlToken is a token of a SQL expression, with its text and byte offsets in the expression.
type sqlToken struct {
	kind       sqlTokenKind
	text       string
	start, end int
}

// lexSQLExpression splits a ClickHouse SQL expression into tokens.
// Strings and quoted identifiers are single tokens, so parentheses and keywords inside them are never misread.
// Comments and unbalanced parentheses are rejected because they could hide part of the expression from the split.
func lexSQLExpression(expr string) ([]sqlToken, error) {
	var toks []sqlToken
	depth := 0
	for i := 0; i < len(expr); {
		c := expr[i]
		switch {
		case c == ' ' || c == '\t' || c == '\n' || c == '\r':
			i++
		case c == '\'' || c == '"' || c == '`':
			end, ok := scanQuoted(expr, i)
			if !ok {
				return nil, errors.New("unterminated quote")
			}
			kind := sqlTokenQuotedIdent
			if c == '\'' {
				kind = sqlTokenString
			}
			toks = append(toks, sqlToken{kind: kind, text: expr[i:end], start: i, end: end})
			i = end
		case c == '#' || (c == '-' && i+1 < len(expr) && expr[i+1] == '-') || (c == '/' && i+1 < len(expr) && expr[i+1] == '*'):
			return nil, errors.New("comments are not supported")
		case isIdentStart(c):
			end := i + 1
			for end < len(expr) && isIdentChar(expr[end]) {
				end++
			}
			toks = append(toks, sqlToken{kind: sqlTokenIdent, text: expr[i:end], start: i, end: end})
			i = end
		case isDigit(c) || (c == '.' && i+1 < len(expr) && isDigit(expr[i+1])):
			end := i + 1
			for end < len(expr) && (isIdentChar(expr[end]) || expr[end] == '.') {
				end++
			}
			toks = append(toks, sqlToken{kind: sqlTokenNumber, text: expr[i:end], start: i, end: end})
			i = end
		case c == '(':
			toks = append(toks, sqlToken{kind: sqlTokenLParen, text: expr[i : i+1], start: i, end: i + 1})
			depth++
			i++
		case c == ')':
			toks = append(toks, sqlToken{kind: sqlTokenRParen, text: expr[i : i+1], start: i, end: i + 1})
			depth--
			if depth < 0 {
				return nil, errors.New("unbalanced parentheses")
			}
			i++
		default:
			toks = append(toks, sqlToken{kind: sqlTokenOther, text: expr[i : i+1], start: i, end: i + 1})
			i++
		}
	}
	if depth != 0 {
		return nil, errors.New("unbalanced parentheses")
	}
	return toks, nil
}

// scanQuoted returns the end offset of the string or quoted identifier starting at offset start, honoring backslash escapes and doubled quotes.
func scanQuoted(expr string, start int) (int, bool) {
	quote := expr[start]
	for i := start + 1; i < len(expr); i++ {
		switch expr[i] {
		case '\\':
			i++
		case quote:
			if i+1 < len(expr) && expr[i+1] == quote {
				i++
				continue
			}
			return i + 1, true
		}
	}
	return 0, false
}

func isIdentStart(c byte) bool {
	return c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isIdentChar(c byte) bool {
	return isIdentStart(c) || isDigit(c)
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}
