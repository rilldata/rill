package canvas

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/rilldata/rill/runtime/metricsview"
)

// parseFilter parses a component's own filter (its dimension_filters), which uses the syntax of dashboard URL filters,
// such as "country IN ('US','CA') AND city HAVING (revenue GT 100)".
// It mirrors the frontend's parser (web-common/src/features/dashboards/url-state/filters/expression.ne),
// and returns the expression that the frontend adds to the component's queries, or nil for an empty filter.
// Unlike the frontend, it accepts any whitespace between tokens.
func parseFilter(filter string) (*metricsview.Expression, error) {
	tokens, err := lexFilter(filter)
	if err != nil {
		return nil, err
	}
	if len(tokens) == 0 {
		return nil, nil
	}
	p := &filterParser{tokens: tokens}
	expr, err := p.expr()
	if err != nil {
		return nil, err
	}
	if p.pos < len(p.tokens) {
		return nil, p.unexpected()
	}
	return expr, nil
}

// filterParser is a recursive descent parser for the tokens of a component filter.
type filterParser struct {
	tokens []filterToken
	pos    int
}

// expr parses boolean expressions joined by AND or OR.
// Like the frontend, it allows one switch between AND and OR, which nests the rest of the expressions:
// "a AND b OR c" means "a AND (b OR c)", and "a AND b OR c AND d" needs parentheses.
func (p *filterParser) expr() (*metricsview.Expression, error) {
	var operands []*metricsview.Expression
	var joiners []metricsview.Operator
	for {
		e, err := p.boolean()
		if err != nil {
			return nil, err
		}
		operands = append(operands, e)
		switch {
		case p.keyword("AND"):
			joiners = append(joiners, metricsview.OperatorAnd)
		case p.keyword("OR"):
			joiners = append(joiners, metricsview.OperatorOr)
		default:
			if len(joiners) == 0 {
				return operands[0], nil
			}
			split := len(joiners)
			for i, op := range joiners {
				if op != joiners[0] {
					split = i
					break
				}
			}
			if split == len(joiners) {
				return filterCondition(joiners[0], operands...), nil
			}
			for _, op := range joiners[split:] {
				if op != joiners[split] {
					return nil, errors.New("mixing AND and OR more than once requires parentheses")
				}
			}
			nested := filterCondition(joiners[split], operands[split:]...)
			return filterCondition(joiners[0], append(operands[:split:split], nested)...), nil
		}
	}
}

// boolean parses a parenthesized expression, an IN filter, a HAVING (measure) filter, or a comparison.
func (p *filterParser) boolean() (*metricsview.Expression, error) {
	if p.punct("(") {
		e, err := p.expr()
		if err != nil {
			return nil, err
		}
		return e, p.expect(")")
	}

	t, ok := p.peek()
	if !ok {
		return nil, p.unexpected()
	}
	switch t.kind {
	case filterTokenWord, filterTokenField:
		// A field, even if it's named like a keyword: "in IN ('x')" filters a field named "in".
		p.pos++
		if exclude, ok := p.inOperator(); ok {
			values, err := p.list("(", ")")
			if err != nil {
				return nil, err
			}
			exprs := []*metricsview.Expression{{Name: t.text}}
			for _, v := range values {
				exprs = append(exprs, &metricsview.Expression{Value: v})
			}
			if exclude {
				return filterCondition(metricsview.OperatorNin, exprs...), nil
			}
			return filterCondition(metricsview.OperatorIn, exprs...), nil
		}
		if p.keyword("HAVING") {
			if err := p.expect("("); err != nil {
				return nil, err
			}
			having, err := p.expr()
			if err != nil {
				return nil, err
			}
			if err := p.expect(")"); err != nil {
				return nil, err
			}
			// Like the frontend, the subquery selects every field that the measure filter references.
			var measures []metricsview.Measure
			for _, name := range filterIdentifiers(having) {
				measures = append(measures, metricsview.Measure{Name: name})
			}
			subquery := &metricsview.Subquery{Dimension: metricsview.Dimension{Name: t.text}, Measures: measures, Having: having}
			return filterCondition(metricsview.OperatorIn, &metricsview.Expression{Name: t.text}, &metricsview.Expression{Subquery: subquery}), nil
		}
		return p.comparison(&metricsview.Expression{Name: t.text})
	case filterTokenString:
		// The frontend also reads a quoted string on the left of a comparison as a field name.
		p.pos++
		return p.comparison(&metricsview.Expression{Name: t.text})
	default:
		v, err := p.value()
		if err != nil {
			return nil, err
		}
		return p.comparison(&metricsview.Expression{Value: v})
	}
}

// comparison parses the operator and value of a comparison of the left operand.
func (p *filterParser) comparison(left *metricsview.Expression) (*metricsview.Expression, error) {
	op, ok := p.comparisonOperator()
	if !ok {
		return nil, p.unexpected()
	}
	v, err := p.value()
	if err != nil {
		return nil, err
	}
	return filterCondition(op, left, &metricsview.Expression{Value: v}), nil
}

// inOperator parses IN, IN LIST, NIN, NOT IN or NOT IN LIST, and reports whether it excludes the values.
func (p *filterParser) inOperator() (exclude, ok bool) {
	start := p.pos
	switch {
	case p.keyword("IN"):
		p.keyword("LIST")
		return false, true
	case p.keyword("NIN"):
		return true, true
	case p.keyword("NOT") && p.keyword("IN"):
		p.keyword("LIST")
		return true, true
	}
	p.pos = start
	return false, false
}

// filterComparisonOperators maps the comparison operators of component filters to the query engine's operators.
var filterComparisonOperators = map[string]metricsview.Operator{
	"EQ":    metricsview.OperatorEq,
	"NEQ":   metricsview.OperatorNeq,
	"GT":    metricsview.OperatorGt,
	"GTE":   metricsview.OperatorGte,
	"LT":    metricsview.OperatorLt,
	"LTE":   metricsview.OperatorLte,
	"LIKE":  metricsview.OperatorIlike,
	"NLIKE": metricsview.OperatorNilike,
}

// comparisonOperator parses a comparison operator, such as EQ or NOT LIKE.
func (p *filterParser) comparisonOperator() (metricsview.Operator, bool) {
	start := p.pos
	if p.keyword("NOT") {
		if p.keyword("LIKE") {
			return metricsview.OperatorNilike, true
		}
		p.pos = start
		return "", false
	}
	t, ok := p.peek()
	if !ok || t.kind != filterTokenWord {
		return "", false
	}
	op, ok := filterComparisonOperators[strings.ToUpper(t.text)]
	if ok {
		p.pos++
	}
	return op, ok
}

// value parses a string, number, true, false, null, list or object.
func (p *filterParser) value() (any, error) {
	t, ok := p.peek()
	if !ok {
		return nil, p.unexpected()
	}
	if t.kind == filterTokenPunct {
		switch t.text {
		case "[":
			return p.list("[", "]")
		case "{":
			return p.object()
		}
		return nil, p.unexpected()
	}
	switch t.kind {
	case filterTokenString:
		p.pos++
		return t.text, nil
	case filterTokenNumber:
		p.pos++
		return strconv.ParseFloat(t.text, 64)
	case filterTokenWord:
		switch strings.ToLower(t.text) {
		case "true":
			p.pos++
			return true, nil
		case "false":
			p.pos++
			return false, nil
		case "null":
			p.pos++
			return nil, nil
		}
	}
	return nil, p.unexpected()
}

// list parses a non-empty list of values between the open and close punctuation.
func (p *filterParser) list(open, close string) ([]any, error) {
	if err := p.expect(open); err != nil {
		return nil, err
	}
	var values []any
	for {
		v, err := p.value()
		if err != nil {
			return nil, err
		}
		values = append(values, v)
		if !p.punct(",") {
			return values, p.expect(close)
		}
	}
}

// object parses an object of quoted keys and values, such as {'US':['SF','LA']}.
func (p *filterParser) object() (map[string]any, error) {
	if err := p.expect("{"); err != nil {
		return nil, err
	}
	obj := make(map[string]any)
	for {
		t, ok := p.peek()
		if !ok || t.kind != filterTokenString {
			return nil, p.unexpected()
		}
		p.pos++
		if err := p.expect(":"); err != nil {
			return nil, err
		}
		v, err := p.value()
		if err != nil {
			return nil, err
		}
		obj[t.text] = v
		if !p.punct(",") {
			return obj, p.expect("}")
		}
	}
}

// peek returns the next token without consuming it.
func (p *filterParser) peek() (filterToken, bool) {
	if p.pos >= len(p.tokens) {
		return filterToken{}, false
	}
	return p.tokens[p.pos], true
}

// keyword consumes the next token if it's the keyword, in any case.
func (p *filterParser) keyword(word string) bool {
	t, ok := p.peek()
	if ok && t.kind == filterTokenWord && strings.EqualFold(t.text, word) {
		p.pos++
		return true
	}
	return false
}

// punct consumes the next token if it's the punctuation.
func (p *filterParser) punct(s string) bool {
	t, ok := p.peek()
	if ok && t.kind == filterTokenPunct && t.text == s {
		p.pos++
		return true
	}
	return false
}

// expect consumes the punctuation, or returns an error if it's not next.
func (p *filterParser) expect(s string) error {
	if p.punct(s) {
		return nil
	}
	return p.unexpected()
}

// unexpected returns an error for the next token.
func (p *filterParser) unexpected() error {
	t, ok := p.peek()
	if !ok {
		return errors.New("unexpected end of filter")
	}
	return fmt.Errorf("unexpected %q", t.text)
}

// filterTokenKind is the kind of a token of a component filter.
type filterTokenKind int

const (
	filterTokenWord   filterTokenKind = iota // an identifier: a field name or a keyword
	filterTokenField                         // a field name in double quotes
	filterTokenString                        // a string in single quotes
	filterTokenNumber                        // an integer or decimal
	filterTokenPunct                         // one of ()[]{},:
)

type filterToken struct {
	kind filterTokenKind
	// text is the token's text, without the quotes and escapes of strings.
	text string
}

// lexFilter splits a component filter into tokens.
func lexFilter(filter string) ([]filterToken, error) {
	isLetter := func(c byte) bool { return 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' }
	isDigit := func(c byte) bool { return '0' <= c && c <= '9' }
	var tokens []filterToken
	for i := 0; i < len(filter); {
		c := filter[i]
		switch {
		case strings.IndexByte(" \t\n\v\f", c) >= 0:
			i++
		case strings.IndexByte("()[]{},:", c) >= 0:
			tokens = append(tokens, filterToken{kind: filterTokenPunct, text: filter[i : i+1]})
			i++
		case c == '\'' || c == '"':
			text, n, err := lexFilterString(filter[i:])
			if err != nil {
				return nil, err
			}
			kind := filterTokenString
			if c == '"' {
				kind = filterTokenField
			}
			tokens = append(tokens, filterToken{kind: kind, text: text})
			i += n
		case isLetter(c):
			end := i + 1
			for end < len(filter) && (isLetter(filter[end]) || isDigit(filter[end]) || filter[end] == '_') {
				end++
			}
			tokens = append(tokens, filterToken{kind: filterTokenWord, text: filter[i:end]})
			i = end
		case isDigit(c) || c == '-' && i+1 < len(filter) && isDigit(filter[i+1]):
			end := i + 1
			for end < len(filter) && isDigit(filter[end]) {
				end++
			}
			if end+1 < len(filter) && filter[end] == '.' && isDigit(filter[end+1]) {
				end += 2
				for end < len(filter) && isDigit(filter[end]) {
					end++
				}
			}
			tokens = append(tokens, filterToken{kind: filterTokenNumber, text: filter[i:end]})
			i = end
		default:
			return nil, fmt.Errorf("unexpected character %q", c)
		}
	}
	return tokens, nil
}

// lexFilterString decodes the quoted string at the start of s, and returns it with its length in s.
// Like the frontend's strings, it can't contain a newline, and it supports JSON escapes, plus \' in single quotes.
func lexFilterString(s string) (string, int, error) {
	quote := s[0]
	var b strings.Builder
	for i := 1; i < len(s); i++ {
		switch c := s[i]; c {
		case quote:
			return b.String(), i + 1, nil
		case '\n':
			return "", 0, errors.New("unterminated string")
		case '\\':
			if i+1 >= len(s) {
				return "", 0, errors.New("unterminated string")
			}
			i++
			switch e := s[i]; e {
			case '"', '\\', '/':
				b.WriteByte(e)
			case 'b':
				b.WriteByte('\b')
			case 'f':
				b.WriteByte('\f')
			case 'n':
				b.WriteByte('\n')
			case 'r':
				b.WriteByte('\r')
			case 't':
				b.WriteByte('\t')
			case 'u':
				if i+4 >= len(s) {
					return "", 0, errors.New("unterminated string")
				}
				r, err := strconv.ParseUint(s[i+1:i+5], 16, 32)
				if err != nil {
					return "", 0, fmt.Errorf("invalid escape \\u%s", s[i+1:i+5])
				}
				b.WriteRune(rune(r))
				i += 4
			case '\'':
				if quote != '\'' {
					return "", 0, errors.New(`invalid escape \'`)
				}
				b.WriteByte('\'')
			default:
				return "", 0, fmt.Errorf("invalid escape \\%c", e)
			}
		default:
			b.WriteByte(c)
		}
	}
	return "", 0, errors.New("unterminated string")
}

// filterCondition returns a condition expression with the operator and operands.
func filterCondition(op metricsview.Operator, exprs ...*metricsview.Expression) *metricsview.Expression {
	return &metricsview.Expression{Condition: &metricsview.Condition{Operator: op, Expressions: exprs}}
}

// filterIdentifiers returns the distinct field names in the expression, without descending into subqueries, like the frontend's getAllIdentifiers.
func filterIdentifiers(e *metricsview.Expression) []string {
	var names []string
	seen := make(map[string]bool)
	var walk func(e *metricsview.Expression)
	walk = func(e *metricsview.Expression) {
		if e.Name != "" && !seen[e.Name] {
			seen[e.Name] = true
			names = append(names, e.Name)
		}
		if e.Condition != nil {
			for _, sub := range e.Condition.Expressions {
				walk(sub)
			}
		}
	}
	walk(e)
	return names
}
