package bigquery

import (
	"fmt"
	"strings"

	"github.com/rilldata/rill/runtime/drivers"
)

type dialect struct {
	drivers.BaseDialect
}

func newDialect() *dialect {
	d := &dialect{}
	d.InitBase(d)
	return d
}

func (d *dialect) String() string { return "bigquery" }

func (d *dialect) EscapeIdentifier(ident string) string {
	if ident == "" {
		return ident
	}
	// Bigquery uses backticks for quoting identifiers
	// Replace any backticks inside the identifier with double backticks
	return fmt.Sprintf("`%s`", strings.ReplaceAll(ident, "`", "``"))
}
