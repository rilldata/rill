package duckdbsql

import (
	"regexp"
)

var (
	PivotRegex          = regexp.MustCompile(`(?i)(?:^|\s|UN)PIVOT`)
	PivotDirectRefRegex = regexp.MustCompile(`(?i)(?:^|\s|UN)PIVOT\s+([a-zA-Z0-9_\-.]+?)\s+ON`)
	OtherReferences     = regexp.MustCompile(`(?im)(?:from|join)\s+([a-zA-z0-9_.]+|"[^"]+")`)
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

	// get other refs from select statements
	otherRefs := OtherReferences.FindAllStringSubmatch(sql, -1)
	for _, or := range otherRefs {
		if len(or) < 1 {
			continue
		}
		// add an empty ref
		ast.newFromNode(nil, nil, "", &TableRef{
			Name: or[1],
		})
	}

	return ast, nil
}
