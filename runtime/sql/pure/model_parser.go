package pure

import (
	"regexp"
	"strings"
)

/**
 * Parses a model query but looks specifically for source table names
 */

type table struct {
	start     int
	end       int
	name      string
	substring string
}

var expressionTokens = []string{
	" where ",
	" group by ",
	" having ",
	" order by ",
	" left outer join ",
	" right outer join ",
	" inner join ",
	" outer join ",
	" natural join ",
	" right join ",
	" left join ",
	" limit ",
	" join ", // make join last,
}

var spacesRegex = regexp.MustCompile(`[\s\t\n\r]`)

func extractCTEs(query string) []*table {
	lowerQuery := strings.ToLower(query)

	var ctes []*table
	// return if the query is not a CTE
	if query == "" || strings.HasPrefix(strings.TrimSpace(lowerQuery), "select") {
		return ctes
	}

	// Find the very start of the expression.
	si := strings.Index(lowerQuery, "with ")
	// exit if there is no `WITH` statement.
	if si == -1 {
		return ctes
	}
	// this is the start of the tape, where the first CTE alias should be.
	si += len("with ")
	// set the tape at the start.
	ri := si
	// the expression index.
	var ei int
	// this will track the nested parens level.
	nest := 0
	inside := false
	curCte := &table{}

	for ri < len(lowerQuery) {
		char := lowerQuery[ri]

		// let's get the name of this thing.
		// we should only trigger this if nest === 1; otherwise we're selecting the CTEs of CTEs :)
		if nest == 1 && strings.HasSuffix(lowerQuery[si:ri], " as (") {
			curCte.name = strings.TrimSpace(strings.Replace(query[si:ri-4], ",", "", -1))
		}

		// Let's up the nest by one if we encounter an open paren.
		// we will set the inside flag to true, then set the expression index
		// to match the right index.
		if char == '(' {
			nest++
			if !inside {
				inside = true
				ei = ri
			}
		} else if char == ')' {
			// If we encounter a close paren, let's unnest.
			nest--
			// if we encounter a close parent AND the nest is at 0, then we've found the end of the CTE.
			if nest == 0 {
				// we reset.
				curCte.start = ei + 1
				// move up the start to the first non-whitespace char.
				firstRealCharIdx := indexOfFirstChar(query[curCte.start:])
				if firstRealCharIdx != -1 {
					curCte.start += firstRealCharIdx
				}
				curCte.end = ri
				curCte.substring = query[ei : ri+1]
				curCte.substring = strings.TrimSpace(curCte.substring[1 : len(curCte.substring)-1])
				ctes = append(ctes, curCte)

				si = ri + 1
				// reset the expression
				curCte = &table{}
				inside = false
			}
		}

		ri++
		// do we kill things if SELECT is at the end?
		if !inside && strings.HasPrefix(strings.TrimSpace(lowerQuery[si:ri]), "select ") {
			break
		}
	}

	return ctes
}

func extractFromStatements(query string) []*table {
	rest := spacesRegex.ReplaceAllString(query, " ")
	finds := getAllIndexes(strings.ToLower(rest), " from ")

	var tables []*table

	for _, find := range finds {
		ei := find + len(" from ")
		ri := ei
		for ri < len(rest) {
			ri++
			char := rest[ri-1]
			seqSoFar := rest[ei:ri]

			// skip if the FROM statement contains a nested statement inside it.
			if char == '(' {
				break
			}

			trimmedSeq, ok := trimmedExpression(seqSoFar)

			if ok || char == ';' || char == ')' || ri == len(rest) {
				if ok {
					// reset seqSoFar to not include the expression token.
					ri -= len(seqSoFar) - len(trimmedSeq)
					seqSoFar = trimmedSeq
				}

				// we hit the end of the table def.
				leftCorrection := 0
				rightCorrection := 0

				// can we flip this?
				// get the right side if there's extra characters;
				leftSide := indexOfFirstChar(seqSoFar)
				rightSide := indexOfLastChar(seqSoFar)

				if leftSide != -1 {
					leftCorrection = leftSide
				}
				if rightSide != -1 {
					rightCorrection = rightSide
				}

				if char == ')' {
					rightCorrection++
					// look for spaces b/t ) and statement.
					additionalRightBuffer := indexOfLastChar(seqSoFar[:len(seqSoFar)-1])
					if additionalRightBuffer != -1 {
						rightCorrection += additionalRightBuffer
					}
				}

				finalSeq := rest[ei+leftCorrection : ri-rightCorrection]
				words := strings.Split(finalSeq, " ")
				remainingChars := strings.TrimPrefix(finalSeq, words[0])
				tables = append(tables, &table{
					start: ei + leftCorrection,
					end:   ri - rightCorrection - len(remainingChars),
					name:  strings.TrimSpace(words[0]),
				})

				break
			}
		}
	}

	return tables
}

func extractJoins(query string) []*table {
	rest := spacesRegex.ReplaceAllString(query, " ")
	finds := getAllIndexes(strings.ToLower(rest), " join ")

	var matches []*table

	for _, find := range finds {
		ei := find + len(" join ")
		ri := ei

		for ri < len(rest) {
			ri++
			soFar := rest[ei:ri]
			// break if we're on a subquery, for now.
			if ri < len(rest) && rest[ri] == '(' {
				break
			}
			if !strings.HasSuffix(strings.ToLower(soFar), " on ") {
				continue
			}
			// we have a match.
			name := soFar[0 : len(soFar)-len(" on ")]
			matches = append(matches, &table{
				start:     ei + indexOfFirstChar(name),
				end:       ei + len(name) - indexOfLastChar(name),
				name:      strings.TrimSpace(name),
				substring: "",
			})
		}
	}

	return matches
}

func ExtractTableNames(query string) []string {
	ctes := extractCTEs(query)
	ignoreMap := make(map[string]bool)
	for _, cte := range ctes {
		ignoreMap[cte.name] = true
	}

	froms := extractFromStatements(query)
	joins := extractJoins(query)

	var all []string

	for _, from := range froms {
		if _, ok := ignoreMap[from.name]; !ok {
			ignoreMap[from.name] = true
			all = append(all, from.name)
		}
	}

	for _, join := range joins {
		if _, ok := ignoreMap[join.name]; !ok {
			all = append(all, join.name)
		}
	}

	return all
}

// indexOfFirstChar returns index of 1st non space or ';' char
func indexOfFirstChar(s string) int {
	return strings.IndexFunc(s, func(r rune) bool {
		switch r {
		case ' ', '\t', '\n', '\r', ';':
			return false
		default:
			return true
		}
	})
}

// indexOfLastChar returns index of last non space or ';' char
func indexOfLastChar(s string) int {
	i := strings.LastIndexFunc(s, func(r rune) bool {
		switch r {
		case ' ', '\t', '\n', '\r', ';':
			return false
		default:
			return true
		}
	})
	if i == -1 {
		return -1
	}
	return len(s) - 1 - i
}

func getAllIndexes(s string, substr string) []int {
	var indexes []int
	c := 0
	for i := strings.Index(s, substr); i != -1; {
		s = s[i+1:]
		indexes = append(indexes, i+c)
		c += i + 1
		i = strings.Index(s, substr)
	}
	return indexes
}

func trimmedExpression(s string) (string, bool) {
	l := strings.ToLower(s)
	for _, exp := range expressionTokens {
		if strings.HasSuffix(l, exp) {
			return s[:len(s)-len(exp)], true
		}
	}
	return s, false
}
