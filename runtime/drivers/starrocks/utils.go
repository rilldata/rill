package starrocks

import (
	"strings"
)

// StarRocks reserved keywords and their aliases.
// Maps reserved keywords to safe alternatives for use in SQL queries.
// Reference: https://docs.starrocks.io/docs/sql-reference/sql-statements/keywords/
var reservedKeywordAliases = map[string]string{
	"range":  "valRange",  // Used in histogram queries - conflicts with RANGE keyword
	"values": "vals",   // Used in histogram queries - conflicts with VALUES keyword
}

// safeSQLName escapes an identifier (catalog, database, table name) for StarRocks.
// StarRocks uses backticks (`) to escape identifiers, similar to MySQL.
func safeSQLName(name string) string {
	if name == "" {
		return name
	}
	// Escape backticks inside the name by doubling them
	escaped := strings.ReplaceAll(name, "`", "``")
	return "`" + escaped + "`"
}

// EscapeReservedKeyword returns a safe alias for SQL reserved keywords.
// For reserved keywords, returns the predefined alias; otherwise returns the original keyword.
//
// Example:
//
//	Input:  "range"  → Output: "valRange"
//	Input:  "values" → Output: "vals"
//	Input:  "other"  → Output: "other"
func EscapeReservedKeyword(keyword string) string {
	if alias, ok := reservedKeywordAliases[strings.ToLower(keyword)]; ok {
		return alias
	}
	return keyword
}
