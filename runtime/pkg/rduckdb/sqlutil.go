package rduckdb

import (
	"fmt"
	"strings"
)

func safeSQLString(s string) string {
	return fmt.Sprintf("'%s'", strings.ReplaceAll(s, "'", "''"))
}

func safeSQLName(ident string) string {
	if ident == "" {
		return ident
	}
	return fmt.Sprintf("\"%s\"", strings.ReplaceAll(ident, "\"", "\"\"")) // nolint:gocritic // Because SQL escaping is different
}
