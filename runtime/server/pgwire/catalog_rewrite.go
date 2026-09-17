package pgwire

import (
	"regexp"
	"strings"
)

// These rules cover introspection SQL emitted by Superset and Metabase.
// Matches and capture groups must span complete SQL tokens so the rules cannot
// rewrite fragments of literals, quoted identifiers, comments, or longer names.
var catalogRewrites = []struct {
	pattern     *regexp.Regexp
	replacement string
}{
	{regexp.MustCompile(regexp.QuoteMeta("ix.indrelid = c.conrelid and\n                                ix.indexrelid = c.conindid and\n                                c.contype in ('p', 'u', 'x')")), "ix.indrelid = c.conrelid"},
	{regexp.MustCompile(regexp.QuoteMeta("t.oid = a.attrelid and a.attnum = ANY(ix.indkey)")), "t.oid = a.attrelid"},
	{regexp.MustCompile(regexp.QuoteMeta("pg_get_constraintdef(cons.oid)")), "pg_get_constraintdef(cons.oid, false)"},
	{regexp.MustCompile(regexp.QuoteMeta("pg_catalog.format_type(a.atttypid, a.atttypmod)")), `CASE pg_catalog.format_type(a.atttypid, a.atttypmod)
		WHEN 'bool' THEN 'boolean'
		WHEN 'float4' THEN 'real'
		WHEN 'float8' THEN 'double precision'
		WHEN 'hugeint' THEN 'bigint'
		WHEN 'int2' THEN 'smallint'
		WHEN 'int4' THEN 'integer'
		WHEN 'int8' THEN 'bigint'
		WHEN 'timestamptz' THEN 'timestamp with time zone'
		WHEN 'timetz' THEN 'time with time zone'
		WHEN 'varchar' THEN 'character varying'
		ELSE pg_catalog.format_type(a.atttypid, a.atttypmod)
	END`},
	{regexp.MustCompile(`(?i)^SELECT nspname FROM pg_namespace WHERE nspname NOT LIKE 'pg_%' ORDER BY nspname$`), "SELECT nspname FROM pg_namespace WHERE nspname NOT IN ('pg_catalog', 'information_schema', 'main') ORDER BY nspname"},
	{regexp.MustCompile(regexp.QuoteMeta("t.schemaname <> 'information_schema'")), "t.schemaname <> 'information_schema' AND t.schemaname <> 'pg_catalog' AND t.schemaname <> 'main'"},
	{regexp.MustCompile(regexp.QuoteMeta("(information_schema._pg_expandarray(i.indkey)).n")), "generate_subscripts(i.indkey, 1)"},
	{regexp.MustCompile(`pg_catalog\.(has_any_column_privilege|has_column_privilege|has_database_privilege|has_foreign_data_wrapper_privilege|has_function_privilege|has_language_privilege|has_parameter_privilege|has_schema_privilege|has_sequence_privilege|has_server_privilege|has_table_privilege|has_tablespace_privilege|has_type_privilege|pg_has_role)\(([^,]+), ([^,]+), ([^)]+)\)`), `(select pg_catalog.$1($3, $4))`},
	{regexp.MustCompile(regexp.QuoteMeta("pg_catalog.pg_matviews")), "pg_matviews"},
	{regexp.MustCompile(`(?i)pg_catalog\.pg_get_serial_sequence\([^)]*\)`), "NULL"},
	{regexp.MustCompile(`(?i)(?:pg_catalog\.)?pg_get_indexdef\([^)]*\)`), "NULL"},
	{regexp.MustCompile(regexp.QuoteMeta("'pg_class'::regclass")), `(SELECT oid FROM pg_class WHERE relname = 'pg_class')`},
	{regexp.MustCompile(`(?is)\(SELECT\s+json_build_object\([^)]*\)\s*FROM[^)]*\)\s+as\s+identity_options`), "NULL AS identity_options"},
}

func rewriteCatalogSQL(query string) (string, error) {
	// Parameter binding changes source offsets, so tokenize the bound SQL once
	// and apply all compatibility edits against that immutable text.
	parsed, err := parseSQL(strings.TrimSpace(query))
	if err != nil {
		return "", err
	}
	query = parsed.text
	starts, ends := make(map[int]int), make(map[int]int)
	for i, token := range parsed.tokens {
		starts[token.start], ends[token.end] = i, i+1
	}
	type replacement struct {
		end  int
		text string
	}
	replacements := make(map[int]replacement)
	for _, rule := range catalogRewrites {
		for _, match := range rule.pattern.FindAllStringSubmatchIndex(query, -1) {
			valid := true
			for i := 0; i < len(match); i += 2 {
				if match[i] < 0 {
					continue
				}
				raw := query[match[i]:match[i+1]]
				trimmed := strings.TrimSpace(raw)
				start := match[i] + strings.Index(raw, trimmed)
				first, startOK := starts[start]
				last, endOK := ends[start+len(trimmed)]
				if !startOK || !endOK || !balancedTokens(parsed.tokens[first:last]) {
					valid = false
					break
				}
			}
			if !valid {
				continue
			}
			first := starts[match[0]]
			if first > 0 && parsed.tokens[first-1].kind == '.' {
				continue
			}
			if _, exists := replacements[match[0]]; !exists {
				replacements[match[0]] = replacement{match[1], string(rule.pattern.ExpandString(nil, rule.replacement, query, match))}
			}
		}
	}

	var out strings.Builder
	start := 0
	// Track SELECT projection lists separately from parentheses in expressions.
	// A default column alias is valid only for a complete, unaliased projection.
	selections := []bool{false}
	for i := 0; i < len(parsed.tokens); i++ {
		token := parsed.tokens[i]
		if token.start < start {
			continue
		}
		if edit, ok := replacements[token.start]; ok {
			out.WriteString(query[start:token.start])
			out.WriteString(edit.text)
			start = edit.end
			continue
		}
		nameIndex := i
		word := strings.ToLower(query[token.start:token.end])
		if word == "pg_catalog" && i+2 < len(parsed.tokens) && parsed.tokens[i+1].kind == '.' {
			nameIndex += 2
			name := parsed.tokens[nameIndex]
			word = strings.ToLower(query[name.start:name.end])
		}
		if (word == "version" || word == "pg_backend_pid") && (i == 0 || parsed.tokens[i-1].kind != '.') && nameIndex+2 < len(parsed.tokens) && parsed.tokens[nameIndex+1].kind == '(' && parsed.tokens[nameIndex+2].kind == ')' {
			end := nameIndex + 2
			value := "1234"
			if word == "version" {
				value = "'PostgreSQL 16.3 (Rill pgwire)'"
			}
			out.WriteString(query[start:token.start])
			out.WriteString("(SELECT " + value + ")")
			previous := ""
			if i > 0 {
				previous = strings.ToUpper(query[parsed.tokens[i-1].start:parsed.tokens[i-1].end])
			}
			next := ""
			if end+1 < len(parsed.tokens) {
				next = strings.ToUpper(query[parsed.tokens[end+1].start:parsed.tokens[end+1].end])
			}
			if selections[len(selections)-1] && (previous == "SELECT" || previous == "DISTINCT" || previous == "ALL" || previous == ",") {
				switch next {
				case "", ",", ")", "FROM", "WHERE", "GROUP", "HAVING", "ORDER", "LIMIT", "OFFSET", "UNION", "EXCEPT", "INTERSECT":
					out.WriteString(" AS " + word)
				}
			}
			start = parsed.tokens[end].end
			i = end
			continue
		}
		switch token.kind {
		case '(':
			selections = append(selections, false)
		case ')':
			if len(selections) > 1 {
				selections = selections[:len(selections)-1]
			}
		case 'w':
			switch word {
			case "select":
				selections[len(selections)-1] = true
			case "from", "where", "group", "having", "order", "limit", "offset", "union", "except", "intersect":
				selections[len(selections)-1] = false
			}
		}
	}
	out.WriteString(query[start:])
	return out.String(), nil
}

func balancedTokens(tokens []sqlToken) bool {
	depth := 0
	for _, token := range tokens {
		switch token.kind {
		case '(':
			depth++
		case ')':
			depth--
			if depth < 0 {
				return false
			}
		}
	}
	return depth == 0
}
