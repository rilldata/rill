package clickhouse

import (
	"strings"

	"github.com/rilldata/rill/runtime/drivers"
)

func safeSQLName(name string) string {
	return drivers.DialectClickHouse.EscapeIdentifier(name)
}

func localTableName(name string) string {
	return name + "_local"
}

func localToActualTableName(name string) string {
	return strings.TrimSuffix(name, "_local")
}
