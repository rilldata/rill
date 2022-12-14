package queries

import (
	"fmt"
	"strings"
)

func quoteName(name string) string {
	return fmt.Sprintf("\"%s\"", name)
}

func EscapeSingleQuotes(value string) string {
	return strings.ReplaceAll(value, "'", "''")
}

func EscapeDoubleQuotes(column string) string {
	return strings.ReplaceAll(column, "\"", "\"\"")
}
