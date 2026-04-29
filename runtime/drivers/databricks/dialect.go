package databricks

import (
	"fmt"
	"strings"

	"github.com/rilldata/rill/runtime/drivers"
)

type dialect struct {
	drivers.BaseDialect
}

var DialectDatabricks drivers.Dialect = func() drivers.Dialect {
	d := &dialect{}
	d.BaseDialect = drivers.NewBaseDialect(drivers.DialectNameDatabricks, DatabricksEscapeIdentifier, DatabricksEscapeIdentifier)
	return d
}()

func DatabricksEscapeIdentifier(ident string) string {
	if ident == "" {
		return ident
	}
	// Databricks uses backticks for quoting identifiers
	// Replace any backticks inside the identifier with double backticks
	return fmt.Sprintf("`%s`", strings.ReplaceAll(ident, "`", "``"))
}
