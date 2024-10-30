package clickhouse

import (
	"github.com/rilldata/rill/runtime/drivers"
)

func safeSQLName(name string) string {
	return drivers.DialectClickHouse.EscapeIdentifier(name)
}
