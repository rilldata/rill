package mysql

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

func (d *dialect) String() string { return "mysql" }

func (d *dialect) EscapeIdentifier(ident string) string {
	if ident == "" {
		return ident
	}
	// MySQL uses backticks for quoting identifiers.
	return fmt.Sprintf("`%s`", strings.ReplaceAll(ident, "`", "``"))
}

func (d *dialect) SelectInlineResults(_ *drivers.Result) (string, []any, []any, error) {
	return "", nil, nil, fmt.Errorf("SelectInlineResults not implemented for MySQL")
}
