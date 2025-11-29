package starrocks

import (
	"strings"
)

// StarRocks reserved keywords that have been observed to cause conflicts.
// Only keywords that have actually caused issues in production are included here.
// Reference: https://docs.starrocks.io/docs/sql-reference/sql-statements/keywords/
//
// Add new keywords here only when:
// 1. A query fails with syntax error due to the keyword
// 2. You've verified escaping fixes the issue
// 3. You've added a comment explaining the context
var reservedKeywords = map[string]bool{
	"range":  true, // Used in histogram queries - conflicts with RANGE keyword
	"values": true, // Used in histogram queries - conflicts with VALUES keyword
}

// GetTypeCast returns the type casting syntax for StarRocks.
//
// StarRocks uses MySQL-style CAST() function instead of PostgreSQL-style ::TYPE syntax.
// Since the queries package handles type casting differently per dialect,
// this function returns an empty string to indicate no suffix-style casting is needed.
//
// Example:
//
//	PostgreSQL: column::DOUBLE
//	StarRocks:  CAST(column AS DOUBLE) -- handled elsewhere
func GetTypeCast(typeName string) string {
	return "" // StarRocks uses CAST() function, not suffix notation
}

// EscapeReservedKeyword escapes SQL reserved keywords for StarRocks.
//
// StarRocks uses backticks (`) to escape identifiers, similar to MySQL.
// Only keywords that have been observed to cause conflicts are escaped.
//
// Example:
//
//	Input:  "range"
//	Output: "`range`"
func EscapeReservedKeyword(keyword string) string {
	if reservedKeywords[strings.ToLower(keyword)] {
		return "`" + keyword + "`"
	}
	return keyword
}
