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

// GetTypeCast returns the type casting syntax for StarRocks.
//
// StarRocks uses MySQL-style CAST() function instead of PostgreSQL-style ::TYPE syntax.
// This function returns an empty string to rely on StarRocks' implicit type conversion
// for numeric operations, which works correctly for histogram queries.
//
// The calling code uses suffix concatenation pattern (column + castSuffix), so returning
// an empty string means no explicit cast is applied, allowing implicit conversion.
//
// Example usage in column_numeric_histogram.go:
//
//	PostgreSQL: column::DOUBLE  (explicit cast using suffix)
//	StarRocks:  column          (implicit conversion - works for numeric histogram)
//
// Note: StarRocks' implicit type conversion handles numeric operations correctly.
// If explicit CAST(column AS TYPE) is needed in the future, the calling pattern
// would need to change from suffix concatenation to function wrapping.
func GetTypeCast(typeName string) string {
	// Return empty string to rely on implicit type conversion
	// StarRocks handles numeric operations without explicit CAST() for histogram queries
	return ""
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

// Note: switchCatalogContext is intentionally NOT included.
// This driver uses fully qualified table names (catalog.database.table)
// instead of SET CATALOG/USE commands for better compatibility with
// external catalogs and connection pooling.
