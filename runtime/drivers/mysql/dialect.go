package mysql

import (
	"fmt"
	"strings"

	"github.com/rilldata/rill/runtime/drivers"
)

type dialect struct {
	drivers.BaseDialect
}

var DialectMySQL drivers.Dialect = func() drivers.Dialect {
	d := &dialect{}
	d.BaseDialect = drivers.NewBaseDialect(drivers.DialectNameMySQL, EscapeIdentifier, EscapeIdentifier)
	return d
}()

func (d *dialect) SupportsILike() bool {
	return false
}

func EscapeIdentifier(ident string) string {
	if ident == "" {
		return ident
	}
	// MySQL uses backticks for quoting identifiers
	// Replace any backticks inside the identifier with double backticks.
	return fmt.Sprintf("`%s`", strings.ReplaceAll(ident, "`", "``"))
}
