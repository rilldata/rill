package bigquery

import (
	"fmt"
	"strings"

	"github.com/rilldata/rill/runtime/drivers"
)

type dialect struct {
	drivers.BaseDialect
}

var DialectBigQuery drivers.Dialect = func() drivers.Dialect {
	d := &dialect{}
	d.BaseDialect = drivers.NewBaseDialect(drivers.DialectNameBigQuery, BigQueryEscapeIdentifier, BigQueryEscapeIdentifier)
	return d
}()

func BigQueryEscapeIdentifier(ident string) string {
	if ident == "" {
		return ident
	}
	// Bigquery uses backticks for quoting identifiers
	// Replace any backticks inside the identifier with double backticks
	return fmt.Sprintf("`%s`", strings.ReplaceAll(ident, "`", "``"))
}
