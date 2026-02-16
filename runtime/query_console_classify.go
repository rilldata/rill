package runtime

import (
	"context"
	"fmt"
	"strings"
	"unicode"
)

// ModelType represents the classification of a published model.
type ModelType int

const (
	// ModelTypeSource indicates the model reads from external connector tables.
	ModelTypeSource ModelType = iota
	// ModelTypeDerived indicates the model reads only from existing Rill models.
	ModelTypeDerived
)

// String returns a human-readable representation of the ModelType.
func (mt ModelType) String() string {
	switch mt {
	case ModelTypeSource:
		return "source_model"
	case ModelTypeDerived:
		return "derived_model"
	default:
		return "unknown"
	}
}

// CatalogLookup provides an interface for checking whether an object exists
// as an internal Rill resource (model, source, metrics view) in the catalog.
type CatalogLookup interface {
	// IsInternalResource returns true if the given name corresponds to an existing
	// Rill-managed resource (source, model, or metrics view).
	IsInternalResource(ctx context.Context, name string) bool
}

// ClassifyModelType determines whether a SQL query should be classified as a
// source_model or derived_model based on the table references it contains.
//
// The classification rules are:
//   - If all referenced tables are existing internal Rill resources → derived_model
//   - If any referenced table is an external connector table → source_model
//   - Mixed references (internal + external) default to source_model
//   - If no table references are found, defaults to derived_model
func ClassifyModelType(ctx context.Context, catalog CatalogLookup, sql string) (ModelType, error) {
	if strings.TrimSpace(sql) == "" {
		return ModelTypeDerived, fmt.Errorf("empty SQL statement")
	}

	refs, err := extractTableReferences(sql)
	if err != nil {
		return ModelTypeSource, fmt.Errorf("failed to extract table references: %w", err)
	}

	// No table references found (e.g., SELECT 1+1); treat as derived.
	if len(refs) == 0 {
		return ModelTypeDerived, nil
	}

	// Check each reference against the catalog.
	allInternal := true
	for _, ref := range refs {
		// Use the unqualified table name for catalog lookup.
		// For qualified names like "schema.table" or "db.schema.table",
		// we check the full qualified name first, then fall back to the last part.
		if !isInternalRef(ctx, catalog, ref) {
			allInternal = false
			break
		}
	}

	if allInternal {
		return ModelTypeDerived, nil
	}
	return ModelTypeSource, nil
}

// isInternalRef checks whether a table reference corresponds to an internal
// Rill resource. It tries the full reference first, then the unqualified name.
func isInternalRef(ctx context.Context, catalog CatalogLookup, ref string) bool {
	// Try full qualified reference
	if catalog.IsInternalResource(ctx, ref) {
		return true
	}

	// Try just the table name (last segment after dot)
	parts := strings.Split(ref, ".")
	if len(parts) > 1 {
		tableName := parts[len(parts)-1]
		tableName = unquoteIdentifier(tableName)
		if catalog.IsInternalResource(ctx, tableName) {
			return true
		}
	}

	// Try unquoted version of the full reference
	unquoted := unquoteIdentifier(ref)
	if unquoted != ref && catalog.IsInternalResource(ctx, unquoted) {
		return true
	}

	return false
}

// extractTableReferences performs a lightweight parse of SQL to extract table
// names referenced in FROM and JOIN clauses. This is not a full SQL parser;
// it handles common patterns including:
//   - Simple: SELECT ... FROM table_name
//   - Qualified: SELECT ... FROM schema.table_name
//   - Quoted: SELECT ... FROM "table_name"
//   - Joins: ... JOIN table_name ON ...
//   - Subqueries: FROM (SELECT ...) AS alias — skips subqueries
//   - CTEs: WITH cte AS (...) SELECT ... FROM cte — recognizes CTE names
//
// It returns a deduplicated list of table references.
func extractTableReferences(sql string) ([]string, error) {
	tokens := tokenizeSQL(sql)
	if len(tokens) == 0 {
		return nil, nil
	}

	// Collect CTE names so we can exclude them from external references.
	cteNames := make(map[string]bool)
	collectCTENames(tokens, cteNames)

	refs := make(map[string]bool)
	for i := 0; i < len(tokens); i++ {
		upper := strings.ToUpper(tokens[i])

		// Look for FROM or JOIN keywords
		isFrom := upper == "FROM"
		isJoin := upper == "JOIN"
		// Also handle LEFT JOIN, RIGHT JOIN, INNER JOIN, OUTER JOIN, CROSS JOIN, FULL JOIN
		if !isFrom && !isJoin {
			continue
		}

		// Get the next non-whitespace token
		next := i + 1
		if next >= len(tokens) {
			continue
		}

		nextToken := tokens[next]

		// Skip subqueries: FROM (
		if nextToken == "(" {
			continue
		}

		// Skip LATERAL keyword
		if strings.ToUpper(nextToken) == "LATERAL" {
			continue
		}

		// Build possibly qualified table name (handles schema.table, db.schema.table)
		tableName := nextToken
		j := next + 1
		for j+1 < len(tokens) && tokens[j] == "." {
			tableName += "." + tokens[j+1]
			j += 2
		}

		// Skip if it's a CTE name
		normalized := strings.ToLower(unquoteIdentifier(tableName))
		if cteNames[normalized] {
			continue
		}

		// Skip known SQL keywords that might follow FROM/JOIN erroneously
		if isSQLKeyword(nextToken) {
			continue
		}

		refs[tableName] = true
	}

	result := make([]string, 0, len(refs))
	for ref := range refs {
		result = append(result, ref)
	}
	return result, nil
}

// collectCTENames extracts CTE names from a WITH clause.
// Pattern: WITH name AS (...), name AS (...)
func collectCTENames(tokens []string, cteNames map[string]bool) {
	if len(tokens) == 0 {
		return
	}

	// Check if the query starts with WITH (possibly after whitespace tokens)
	startIdx := -1
	for i, t := range tokens {
		if strings.ToUpper(t) == "WITH" {
			startIdx = i
			break
		}
		// Only skip if we haven't hit a real keyword yet
		if isSQLKeyword(t) {
			break
		}
	}

	if startIdx < 0 {
		return
	}

	// Check for WITH RECURSIVE
	i := startIdx + 1
	if i < len(tokens) && strings.ToUpper(tokens[i]) == "RECURSIVE" {
		i++
	}

	// Parse CTE names: expect "name AS (" pattern
	for i < len(tokens) {
		// CTE name
		if i >= len(tokens) {
			break
		}
		cteName := strings.ToLower(unquoteIdentifier(tokens[i]))
		i++

		// Expect AS
		if i >= len(tokens) || strings.ToUpper(tokens[i]) != "AS" {
			break
		}
		i++

		// Optional NOT MATERIALIZED or MATERIALIZED
		if i < len(tokens) && strings.ToUpper(tokens[i]) == "NOT" {
			i++
			if i < len(tokens) && strings.ToUpper(tokens[i]) == "MATERIALIZED" {
				i++
			}
		} else if i < len(tokens) && strings.ToUpper(tokens[i]) == "MATERIALIZED" {
			i++
		}

		// Expect opening parenthesis
		if i >= len(tokens) || tokens[i] != "(" {
			break
		}

		// Register the CTE name
		cteNames[cteName] = true

		// Skip to matching closing parenthesis
		depth := 1
		i++
		for i < len(tokens) && depth > 0 {
			if tokens[i] == "(" {
				depth++
			} else if tokens[i] == ")" {
				depth--
			}
			i++
		}

		// Check for comma (more CTEs) or end of WITH clause
		if i < len(tokens) && tokens[i] == "," {
			i++
			continue
		}

		// No comma means end of CTE list
		break
	}
}

// tokenizeSQL performs a simple tokenization of SQL text. It handles:
// - Quoted identifiers (double quotes)
// - String literals (single quotes)
// - Parentheses as separate tokens
// - Dot as a separate token
// - Commas as separate tokens
// - Line comments (--) and block comments (/* */)
// - Identifiers and keywords as tokens
func tokenizeSQL(sql string) []string {
	var tokens []string
	runes := []rune(sql)
	n := len(runes)
	i := 0

	for i < n {
		ch := runes[i]

		// Skip whitespace
		if unicode.IsSpace(ch) {
			i++
			continue
		}

		// Line comment: --
		if ch == '-' && i+1 < n && runes[i+1] == '-' {
			for i < n && runes[i] != '\n' {
				i++
			}
			continue
		}

		// Block comment: /* ... */
		if ch == '/' && i+1 < n && runes[i+1] == '*' {
			i += 2
			for i+1 < n {
				if runes[i] == '*' && runes[i+1] == '/' {
					i += 2
					break
				}
				i++
			}
			continue
		}

		// String literal: '...'
		if ch == '\'' {
			i++ // skip opening quote
			for i < n {
				if runes[i] == '\'' {
					if i+1 < n && runes[i+1] == '\'' {
						// Escaped quote
						i += 2
						continue
					}
					i++ // skip closing quote
					break
				}
				i++
			}
			// Don't add string literals as tokens — they're not table references
			continue
		}

		// Quoted identifier: "..."
		if ch == '"' {
			start := i
			i++ // skip opening quote
			for i < n {
				if runes[i] == '"' {
					if i+1 < n && runes[i+1] == '"' {
						// Escaped quote within identifier
						i += 2
						continue
					}
					i++ // skip closing quote
					break
				}
				i++
			}
			tokens = append(tokens, string(runes[start:i]))
			continue
		}

		// Backtick-quoted identifier: `...`
		if ch == '`' {
			start := i
			i++ // skip opening backtick
			for i < n && runes[i] != '`' {
				i++
			}
			if i < n {
				i++ // skip closing backtick
			}
			tokens = append(tokens, string(runes[start:i]))
			continue
		}

		// Single-character tokens
		if ch == '(' || ch == ')' || ch == '.' || ch == ',' || ch == ';' {
			tokens = append(tokens, string(ch))
			i++
			continue
		}

		// Operators and other symbols — skip
		if !isIdentRune(ch) {
			i++
			continue
		}

		// Identifier or keyword
		start := i
		for i < n && isIdentRune(runes[i]) {
			i++
		}
		tokens = append(tokens, string(runes[start:i]))
	}

	return tokens
}

// isIdentRune returns true if the rune can be part of an identifier.
func isIdentRune(ch rune) bool {
	return unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_'
}

// unquoteIdentifier removes surrounding double quotes or backticks from an identifier.
func unquoteIdentifier(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '`' && s[len(s)-1] == '`') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

// isSQLKeyword returns true if the token is a reserved SQL keyword that cannot
// be a table name in the context immediately after FROM/JOIN.
func isSQLKeyword(token string) bool {
	switch strings.ToUpper(token) {
	case "SELECT", "WHERE", "GROUP", "ORDER", "HAVING", "LIMIT", "OFFSET",
		"UNION", "INTERSECT", "EXCEPT", "INSERT", "UPDATE", "DELETE",
		"CREATE", "ALTER", "DROP", "SET", "VALUES", "INTO",
		"ON", "USING", "CASE", "WHEN", "THEN", "ELSE", "END",
		"AND", "OR", "NOT", "IN", "EXISTS", "BETWEEN", "LIKE",
		"IS", "NULL", "TRUE", "FALSE", "AS", "WITH", "RECURSIVE",
		"FROM", "JOIN", "LEFT", "RIGHT", "INNER", "OUTER", "CROSS", "FULL",
		"NATURAL", "LATERAL", "FETCH", "FOR", "WINDOW", "PARTITION",
		"ROWS", "RANGE", "GROUPS", "OVER", "FILTER", "WITHIN":
		return true
	default:
		return false
	}
}
