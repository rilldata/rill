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
	d.InitBase(d)
	return d
}()

func (d *dialect) String() string { return "mysql" }

func (d *dialect) EscapeIdentifier(ident string) string {
	if ident == "" {
		return ident
	}
	// MySQL uses backticks for quoting identifiers
	// Replace any backticks inside the identifier with double backticks.
	return fmt.Sprintf("`%s`", strings.ReplaceAll(ident, "`", "``"))
}

func (d *dialect) SupportsILike() bool {
	return false
}
