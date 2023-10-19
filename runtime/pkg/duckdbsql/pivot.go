package duckdbsql

import (
	"regexp"
)

var (
	PivotDirectRefRegex    = regexp.MustCompile(`(?i)(?:^|\s|UN)PIVOT\s+([a-zA-Z0-9_\-.]+?)\s+ON`)
	PivotSubSelectRefRegex = regexp.MustCompile(`(?is)(?:^|\s|UN)PIVOT\s+\((.*?)\)\s+ON`)
)

func parsePivotLikeStatements(sql string) (*AST, error) {
	ast := &AST{
		sql:       sql,
		rootNodes: make([]*selectNode, 0),
		aliases:   map[string]bool{},
		added:     map[string]bool{},
		fromNodes: make([]*fromNode, 0),
		columns:   make([]*columnNode, 0),
	}

	directMatches := PivotDirectRefRegex.FindAllStringSubmatch(sql, -1)
	for _, match := range directMatches {
		if len(match) < 1 {
			continue
		}
		// add an empty ref
		ast.newFromNode(nil, nil, "", &TableRef{
			Name: match[1],
		})
	}

	subSelectMatches := PivotSubSelectRefRegex.FindAllStringSubmatch(sql, -1)
	for _, match := range subSelectMatches {
		if len(match) < 1 {
			continue
		}
		// parse the sub query
		sa, err := Parse(match[1])
		if err != nil {
			// ignore errors
			continue
		}
		// add all sub refs
		for _, ref := range sa.GetTableRefs() {
			ast.newFromNode(nil, nil, "", ref)
		}
	}

	return ast, nil
}
