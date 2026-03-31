package postgres

import (
	"fmt"
	"strings"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/sqldialect"
)

type dialect struct {
	sqldialect.Base
}

func newDialect() *dialect {
	d := &dialect{}
	d.InitBase(d)
	return d
}

func (d *dialect) String() string { return "postgres" }

func (d *dialect) EscapeIdentifier(ident string) string {
	if ident == "" {
		return ident
	}
	return fmt.Sprintf(`"%s"`, strings.ReplaceAll(ident, `"`, `""`)) // nolint:gocritic
}

func (d *dialect) SelectInlineResults(_ *drivers.Result) (string, []any, []any, error) {
	return "", nil, nil, fmt.Errorf("SelectInlineResults not implemented for Postgres")
}
