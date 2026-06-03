package clickhouse

import (
	"strings"
)

func safeSQLName(name string) string {
	return DialectClickhouse.EscapeIdentifier(name)
}

func localTableName(name string) string {
	return name + "_local"
}

func localToActualTableName(name string) string {
	return strings.TrimSuffix(name, "_local")
}
