package pgwire

import (
	"fmt"
	"strconv"
	"strings"

	base "github.com/rilldata/rill/runtime/pkg/pgwire"
)

// parsedSQL is shared by routing, parameter inference and binding.
// Token offsets refer to text, which contains at most one statement.
type parsedSQL struct {
	text   string
	tokens []sqlToken
}

type sqlToken struct {
	start, end int
	kind       byte // w: word, i: quoted identifier, s: string, p: parameter, or punctuation
}

// parseSQL keeps source offsets so inference, interpolation and routing agree on
// which parts of the query are SQL, excluding comments and quoted values.
func parseSQL(query string) (*parsedSQL, error) {
	var tokens []sqlToken
	for i := 0; i < len(query); {
		start := i
		c := query[i]
		switch {
		case c == ' ' || c == '\t' || c == '\r' || c == '\n':
			i++
			continue
		case strings.HasPrefix(query[i:], "--"):
			for i < len(query) && query[i] != '\n' {
				i++
			}
			continue
		case strings.HasPrefix(query[i:], "/*"):
			i += 2
			depth := 1
			for i < len(query) && depth > 0 {
				switch {
				case strings.HasPrefix(query[i:], "/*"):
					depth++
					i += 2
				case strings.HasPrefix(query[i:], "*/"):
					depth--
					i += 2
				default:
					i++
				}
			}
			if depth != 0 {
				return nil, fmt.Errorf("unterminated SQL comment")
			}
			continue
		case c == '\'' || c == '"' || c == '`':
			escaped := c == '\'' && len(tokens) > 0 && tokens[len(tokens)-1].end == start && strings.EqualFold(query[tokens[len(tokens)-1].start:start], "E")
			i++
			closed := false
			for i < len(query) {
				if query[i] == c {
					i++
					if i < len(query) && query[i] == c {
						i++
						continue
					}
					closed = true
					break
				}
				if escaped && query[i] == '\\' && i+1 < len(query) {
					i++
				}
				i++
			}
			if !closed {
				return nil, fmt.Errorf("unterminated SQL quote")
			}
			kind := byte('i')
			if c == '\'' {
				kind = 's'
			}
			tokens = append(tokens, sqlToken{start, i, kind})
		case c == '$':
			i++
			if i < len(query) && query[i] >= '0' && query[i] <= '9' {
				for i < len(query) && query[i] >= '0' && query[i] <= '9' {
					i++
				}
				index, err := strconv.Atoi(query[start+1 : i])
				if err != nil || index < 1 || index > 65535 {
					return nil, &base.Error{Code: "42P02", Message: "parameter index must be between 1 and 65535"}
				}
				tokens = append(tokens, sqlToken{start, i, 'p'})
				continue
			}
			for i < len(query) && (query[i] == '_' || query[i] >= 'a' && query[i] <= 'z' || query[i] >= 'A' && query[i] <= 'Z' || i > start+1 && query[i] >= '0' && query[i] <= '9') {
				i++
			}
			if i < len(query) && query[i] == '$' {
				i++
				delimiter := query[start:i]
				end := strings.Index(query[i:], delimiter)
				if end < 0 {
					return nil, fmt.Errorf("unterminated dollar-quoted SQL string")
				}
				i += end + len(delimiter)
				tokens = append(tokens, sqlToken{start, i, 's'})
			} else {
				tokens = append(tokens, sqlToken{start, i, 'w'})
			}
		case c == '_' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z':
			i++
			for i < len(query) {
				c = query[i]
				if !(c == '_' || c == '$' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9') {
					break
				}
				i++
			}
			tokens = append(tokens, sqlToken{start, i, 'w'})
		default:
			i++
			tokens = append(tokens, sqlToken{start, i, c})
		}
	}
	for i, token := range tokens {
		if token.kind == ';' {
			if i != len(tokens)-1 {
				return nil, &base.Error{Code: "42601", Message: "expected one SQL statement"}
			}
			return &parsedSQL{text: query[:token.start], tokens: tokens[:i]}, nil
		}
	}
	return &parsedSQL{text: query, tokens: tokens}, nil
}

// paginationParameter recognizes both LIMIT count OFFSET offset and LIMIT offset, count.
func (q *parsedSQL) paginationParameter(pos int) bool {
	for i := pos - 1; i >= 0; i-- {
		token := q.tokens[i]
		if token.kind == 'w' {
			word := q.text[token.start:token.end]
			return strings.EqualFold(word, "LIMIT") || strings.EqualFold(word, "OFFSET")
		}
		if token.kind != 'p' && token.kind != ',' && (token.kind < '0' || token.kind > '9') {
			return false
		}
	}
	return false
}

func (q *parsedSQL) replaceParameters(render func(index, pos int) (string, error)) (string, error) {
	var out strings.Builder
	start := 0
	for pos, token := range q.tokens {
		if token.kind != 'p' {
			continue
		}
		index, _ := strconv.Atoi(q.text[token.start+1 : token.end])
		replacement, err := render(index-1, pos)
		if err != nil {
			return "", err
		}
		out.WriteString(q.text[start:token.start])
		out.WriteString(replacement)
		start = token.end
	}
	out.WriteString(q.text[start:])
	return out.String(), nil
}
