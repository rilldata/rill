package starrocks

import (
	"strings"
)

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
